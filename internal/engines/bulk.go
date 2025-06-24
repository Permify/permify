package engines

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"

	"github.com/pkg/errors"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"

	"github.com/Permify/permify/internal/invoke"
	"github.com/Permify/permify/internal/storage/memory/utils"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// BulkCheckerType defines the type of bulk checking operation.
// This enum determines how requests are sorted and processed.
type BulkCheckerType string

const (
	// BulkCheckerTypeSubject indicates that requests should be sorted and processed by subject ID
	BulkCheckerTypeSubject BulkCheckerType = "subject"
	// BulkCheckerTypeEntity indicates that requests should be sorted and processed by entity ID
	BulkCheckerTypeEntity BulkCheckerType = "entity"
)

// BulkCheckerRequest represents a permission check request with optional pre-computed result.
// This struct encapsulates both the permission check request and an optional pre-determined result,
// allowing for optimization when results are already known (e.g., from caching).
type BulkCheckerRequest struct {
	// Request contains the actual permission check request
	Request *base.PermissionCheckRequest
	// Result holds a pre-computed result if available, otherwise CHECK_RESULT_UNSPECIFIED
	Result base.CheckResult
}

// BulkCheckerConfig holds configuration parameters for the BulkChecker.
// This struct allows for fine-tuning the behavior and performance characteristics
// of the bulk permission checking system.
type BulkCheckerConfig struct {
	// ConcurrencyLimit defines the maximum number of concurrent permission checks
	// that can be processed simultaneously. Higher values increase throughput
	// but may consume more system resources.
	ConcurrencyLimit int
	// BufferSize defines the size of the internal request buffer.
	// This should be set based on expected request volume to avoid blocking.
	BufferSize int
}

// DefaultBulkCheckerConfig returns a sensible default configuration
// that balances performance and resource usage for most use cases.
func DefaultBulkCheckerConfig() BulkCheckerConfig {
	return BulkCheckerConfig{
		ConcurrencyLimit: 10,   // Moderate concurrency for most workloads
		BufferSize:       1000, // Buffer for 1000 requests
	}
}

// BulkChecker handles concurrent permission checks with ordered result processing.
// This struct implements a high-performance bulk permission checking system that:
// - Collects permission check requests asynchronously
// - Processes them concurrently with controlled parallelism
// - Maintains strict ordering of results based on request sorting
// - Provides efficient resource management and error handling
type BulkChecker struct {
	// typ determines the sorting strategy and processing behavior
	typ BulkCheckerType
	// checker is the underlying permission checking engine
	checker invoke.Check
	// config holds the operational configuration
	config BulkCheckerConfig
	// ctx provides context for cancellation and timeout management
	ctx context.Context
	// cancel allows for graceful shutdown of all operations
	cancel context.CancelFunc

	// Request handling components
	// requestChan is the input channel for receiving permission check requests
	requestChan chan BulkCheckerRequest
	// requests stores all collected requests before processing
	requests []BulkCheckerRequest
	// requestsMu provides thread-safe access to the requests slice
	requestsMu sync.RWMutex

	// Execution state management
	// executionState tracks the progress and results of request processing
	executionState *executionState
	// collectionDone signals when request collection has completed
	collectionDone chan struct{}

	// Callback for processing results
	// callback is invoked for each successful permission check with the entity/subject ID and continuous token
	callback func(entityID, continuousToken string)
}

// executionState manages the execution of requests and maintains processing order.
// This struct ensures that results are processed in the correct sequence
// even when requests complete out of order due to concurrent processing.
type executionState struct {
	// mu protects access to the execution state
	mu sync.Mutex
	// results stores the results of all requests in their original order
	results []base.CheckResult
	// processedIndex tracks the next result to be processed in order
	processedIndex int
	// successCount tracks the number of successful permission checks
	successCount int64
	// limit defines the maximum number of successful results to process
	limit int64
}

// NewBulkChecker creates a new BulkChecker instance with comprehensive validation and error handling.
// This constructor ensures that all dependencies are properly initialized and validates
// configuration parameters to prevent runtime errors.
//
// Parameters:
//   - ctx: Context for managing the lifecycle of the BulkChecker
//   - checker: The permission checking engine to use for actual permission checks
//   - typ: The type of bulk checking operation (entity or subject)
//   - callback: Function called for each successful permission check
//   - config: Configuration parameters for tuning performance and behavior
//
// Returns:
//   - *BulkChecker: The initialized BulkChecker instance
//   - error: Any error that occurred during initialization
func NewBulkChecker(ctx context.Context, checker invoke.Check, typ BulkCheckerType, callback func(entityID, ct string), config BulkCheckerConfig) (*BulkChecker, error) {
	// Validate all required parameters
	if ctx == nil {
		return nil, fmt.Errorf("context cannot be nil")
	}
	if checker == nil {
		return nil, fmt.Errorf("checker cannot be nil")
	}
	if callback == nil {
		return nil, fmt.Errorf("callback cannot be nil")
	}

	// Apply default values for invalid configuration
	if config.ConcurrencyLimit <= 0 {
		config.ConcurrencyLimit = DefaultBulkCheckerConfig().ConcurrencyLimit
	}
	if config.BufferSize <= 0 {
		config.BufferSize = DefaultBulkCheckerConfig().BufferSize
	}

	// Create a cancellable context for the BulkChecker
	ctx, cancel := context.WithCancel(ctx)

	// Initialize the BulkChecker with all components
	bc := &BulkChecker{
		typ:            typ,
		checker:        checker,
		config:         config,
		ctx:            ctx,
		cancel:         cancel,
		requestChan:    make(chan BulkCheckerRequest, config.BufferSize),
		requests:       make([]BulkCheckerRequest, 0, config.BufferSize),
		callback:       callback,
		collectionDone: make(chan struct{}),
	}

	// Start the background request collection goroutine
	go bc.collectRequests()

	return bc, nil
}

// collectRequests safely collects requests until the channel is closed.
// This method runs in a separate goroutine and continuously processes
// incoming requests until either the channel is closed or the context is cancelled.
// It ensures thread-safe addition of requests to the internal collection.
func (bc *BulkChecker) collectRequests() {
	// Signal completion when this goroutine exits
	defer close(bc.collectionDone)

	for {
		select {
		case req, ok := <-bc.requestChan:
			if !ok {
				// Channel closed, stop collecting
				return
			}
			bc.addRequest(req)
		case <-bc.ctx.Done():
			// Context cancelled, stop collecting
			return
		}
	}
}

// addRequest safely adds a request to the internal list.
// This method uses a mutex to ensure thread-safe access to the requests slice,
// preventing race conditions when multiple goroutines are adding requests.
func (bc *BulkChecker) addRequest(req BulkCheckerRequest) {
	bc.requestsMu.Lock()
	defer bc.requestsMu.Unlock()
	bc.requests = append(bc.requests, req)
}

// StopCollectingRequests safely stops request collection and waits for completion.
// This method closes the input channel and waits for the collection goroutine
// to finish processing any remaining requests. This ensures that no requests
// are lost during shutdown.
func (bc *BulkChecker) StopCollectingRequests() {
	close(bc.requestChan)
	<-bc.collectionDone // Wait for collection to complete
}

// getSortedRequests returns a sorted copy of requests based on the checker type.
// This method creates a copy of the requests to avoid modifying the original
// collection and sorts them according to the BulkCheckerType (entity ID or subject ID).
// The sorting ensures consistent and predictable result ordering.
func (bc *BulkChecker) getSortedRequests() []BulkCheckerRequest {
	bc.requestsMu.RLock()
	defer bc.requestsMu.RUnlock()

	// Create a copy to avoid modifying the original
	requests := make([]BulkCheckerRequest, len(bc.requests))
	copy(requests, bc.requests)

	// Sort the copy based on the checker type
	bc.sortRequests(requests)
	return requests
}

// sortRequests sorts requests based on the checker type.
// This method implements different sorting strategies:
// - For entity-based checks: sorts by entity ID
// - For subject-based checks: sorts by subject ID
// The sorting ensures that results are processed in a consistent order.
func (bc *BulkChecker) sortRequests(requests []BulkCheckerRequest) {
	switch bc.typ {
	case BulkCheckerTypeEntity:
		sort.Slice(requests, func(i, j int) bool {
			return requests[i].Request.GetEntity().GetId() < requests[j].Request.GetEntity().GetId()
		})
	case BulkCheckerTypeSubject:
		sort.Slice(requests, func(i, j int) bool {
			return requests[i].Request.GetSubject().GetId() < requests[j].Request.GetSubject().GetId()
		})
	}
}

// ExecuteRequests processes requests concurrently with comprehensive error handling and resource management.
// This method is the main entry point for bulk permission checking. It:
// 1. Stops collecting new requests
// 2. Sorts all collected requests
// 3. Processes them concurrently with controlled parallelism
// 4. Maintains strict ordering of results
// 5. Handles errors gracefully and manages resources properly
//
// Parameters:
//   - size: The maximum number of successful results to process
//
// Returns:
//   - error: Any error that occurred during processing (context cancellation is not considered an error)
func (bc *BulkChecker) ExecuteRequests(size uint32) error {
	if size == 0 {
		return fmt.Errorf("size must be greater than 0")
	}

	// Stop collecting new requests and wait for collection to complete
	bc.StopCollectingRequests()

	// Get sorted requests for processing
	requests := bc.getSortedRequests()
	if len(requests) == 0 {
		return nil // No requests to process
	}

	// Initialize execution state for tracking progress
	bc.executionState = &executionState{
		results: make([]base.CheckResult, len(requests)),
		limit:   int64(size),
	}

	// Create execution context with cancellation for graceful shutdown
	execCtx, execCancel := context.WithCancel(bc.ctx)
	defer execCancel()

	// Create semaphore to control concurrency
	sem := semaphore.NewWeighted(int64(bc.config.ConcurrencyLimit))

	// Create error group for managing goroutines and error propagation
	g, ctx := errgroup.WithContext(execCtx)

	// Process requests concurrently
	for i, req := range requests {
		// Check if we've reached the success limit
		if atomic.LoadInt64(&bc.executionState.successCount) >= int64(size) {
			break
		}

		index := i
		request := req

		// Launch goroutine for each request
		g.Go(func() error {
			return bc.processRequest(ctx, sem, index, request)
		})
	}

	// Wait for all goroutines to complete and handle any errors
	if err := g.Wait(); err != nil {
		if isContextError(err) {
			return nil // Context cancellation is not an error
		}
		return fmt.Errorf("bulk execution failed: %w", err)
	}

	return nil
}

// processRequest handles a single request with comprehensive error handling.
// This method is executed in a separate goroutine for each request and:
// 1. Acquires a semaphore slot to control concurrency
// 2. Performs the permission check or uses pre-computed result
// 3. Processes the result in the correct order
// 4. Handles context cancellation and other errors gracefully
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - sem: Semaphore for concurrency control
//   - index: The index of this request in the sorted list
//   - req: The permission check request to process
//
// Returns:
//   - error: Any error that occurred during processing
func (bc *BulkChecker) processRequest(ctx context.Context, sem *semaphore.Weighted, index int, req BulkCheckerRequest) error {
	// Check context before acquiring semaphore
	if err := ctx.Err(); err != nil {
		return nil
	}

	// Acquire semaphore slot to control concurrency
	if err := sem.Acquire(ctx, 1); err != nil {
		if isContextError(err) {
			return nil
		}
		return fmt.Errorf("failed to acquire semaphore: %w", err)
	}
	defer sem.Release(1)

	// Determine the result for this request
	result, err := bc.getRequestResult(ctx, req)
	if err != nil {
		if isContextError(err) {
			return nil
		}
		return fmt.Errorf("failed to get request result: %w", err)
	}

	// Process the result in the correct order
	return bc.processResult(index, result)
}

// getRequestResult determines the result for a request.
// This method either uses a pre-computed result if available,
// or performs the actual permission check using the underlying checker.
//
// Parameters:
//   - ctx: Context for the permission check
//   - req: The request to get the result for
//
// Returns:
//   - base.CheckResult: The result of the permission check
//   - error: Any error that occurred during the check
func (bc *BulkChecker) getRequestResult(ctx context.Context, req BulkCheckerRequest) (base.CheckResult, error) {
	// Use pre-computed result if available
	if req.Result != base.CheckResult_CHECK_RESULT_UNSPECIFIED {
		return req.Result, nil
	}

	// Perform the actual permission check
	response, err := bc.checker.Check(ctx, req.Request)
	if err != nil {
		return base.CheckResult_CHECK_RESULT_UNSPECIFIED, err
	}

	return response.GetCan(), nil
}

// processResult processes a single result with thread-safe state updates.
// This method ensures that results are processed in the correct order
// even when requests complete out of order due to concurrent processing.
// It maintains the execution state and calls the callback for successful results.
//
// Parameters:
//   - index: The index of the result in the sorted list
//   - result: The result of the permission check
//
// Returns:
//   - error: Any error that occurred during processing
func (bc *BulkChecker) processResult(index int, result base.CheckResult) error {
	bc.executionState.mu.Lock()
	defer bc.executionState.mu.Unlock()

	// Store the result at the correct index
	bc.executionState.results[index] = result

	// Process results in order, starting from the current processed index
	for bc.executionState.processedIndex < len(bc.executionState.results) {
		currentResult := bc.executionState.results[bc.executionState.processedIndex]
		if currentResult == base.CheckResult_CHECK_RESULT_UNSPECIFIED {
			break // Wait for this result to be computed
		}

		// Process successful results
		if currentResult == base.CheckResult_CHECK_RESULT_ALLOWED {
			// Check if we've reached the success limit
			if atomic.LoadInt64(&bc.executionState.successCount) >= bc.executionState.limit {
				return nil
			}

			// Increment success count and call callback
			atomic.AddInt64(&bc.executionState.successCount, 1)
			bc.callbackWithToken(bc.executionState.processedIndex)
		}

		// Move to the next result
		bc.executionState.processedIndex++
	}

	return nil
}

// callbackWithToken calls the callback with the appropriate entity/subject ID and continuous token.
// This method retrieves the correct request from the sorted list and generates
// the appropriate continuous token for pagination support.
//
// Parameters:
//   - index: The index of the result in the sorted list
func (bc *BulkChecker) callbackWithToken(index int) {
	requests := bc.getSortedRequests()

	// Validate index bounds
	if index >= len(requests) {
		return
	}

	var id string
	var ct string

	// Extract ID and generate continuous token based on checker type
	switch bc.typ {
	case BulkCheckerTypeEntity:
		id = requests[index].Request.GetEntity().GetId()
		if index+1 < len(requests) {
			ct = utils.NewContinuousToken(requests[index+1].Request.GetEntity().GetId()).Encode().String()
		}
	case BulkCheckerTypeSubject:
		id = requests[index].Request.GetSubject().GetId()
		if index+1 < len(requests) {
			ct = utils.NewContinuousToken(requests[index+1].Request.GetSubject().GetId()).Encode().String()
		}
	}

	// Call the user-provided callback
	bc.callback(id, ct)
}

// Close properly cleans up resources and cancels all operations.
// This method should be called when the BulkChecker is no longer needed
// to ensure proper resource cleanup and prevent goroutine leaks.
//
// Returns:
//   - error: Any error that occurred during cleanup
func (bc *BulkChecker) Close() error {
	bc.cancel()
	return nil
}

// BulkEntityPublisher handles entity-based permission check publishing.
// This struct provides a convenient interface for publishing entity permission
// check requests to a BulkChecker instance.
type BulkEntityPublisher struct {
	// bulkChecker is the target BulkChecker for publishing requests
	bulkChecker *BulkChecker
	// request contains the base lookup request parameters
	request *base.PermissionLookupEntityRequest
}

// NewBulkEntityPublisher creates a new BulkEntityPublisher instance.
// This constructor initializes a publisher for entity-based permission checks.
//
// Parameters:
//   - ctx: Context for the publisher (currently unused but kept for API consistency)
//   - request: The base lookup request containing common parameters
//   - bulkChecker: The BulkChecker instance to publish to
//
// Returns:
//   - *BulkEntityPublisher: The initialized publisher instance
func NewBulkEntityPublisher(ctx context.Context, request *base.PermissionLookupEntityRequest, bulkChecker *BulkChecker) *BulkEntityPublisher {
	return &BulkEntityPublisher{
		bulkChecker: bulkChecker,
		request:     request,
	}
}

// Publish sends an entity permission check request to the bulk checker.
// This method creates a permission check request from the provided parameters
// and sends it to the BulkChecker for processing. It handles context cancellation
// gracefully by dropping requests when the context is done.
//
// Parameters:
//   - entity: The entity to check permissions for
//   - metadata: Metadata for the permission check request
//   - context: Additional context for the permission check
//   - result: Optional pre-computed result
func (p *BulkEntityPublisher) Publish(entity *base.Entity, metadata *base.PermissionCheckRequestMetadata, context *base.Context, result base.CheckResult) {
	select {
	case p.bulkChecker.requestChan <- BulkCheckerRequest{
		Request: &base.PermissionCheckRequest{
			TenantId:   p.request.GetTenantId(),
			Metadata:   metadata,
			Entity:     entity,
			Permission: p.request.GetPermission(),
			Subject:    p.request.GetSubject(),
			Context:    context,
		},
		Result: result,
	}:
	case <-p.bulkChecker.ctx.Done():
		// Context cancelled, drop the request
	}
}

// BulkSubjectPublisher handles subject-based permission check publishing.
// This struct provides a convenient interface for publishing subject permission
// check requests to a BulkChecker instance.
type BulkSubjectPublisher struct {
	// bulkChecker is the target BulkChecker for publishing requests
	bulkChecker *BulkChecker
	// request contains the base lookup request parameters
	request *base.PermissionLookupSubjectRequest
}

// NewBulkSubjectPublisher creates a new BulkSubjectPublisher instance.
// This constructor initializes a publisher for subject-based permission checks.
//
// Parameters:
//   - ctx: Context for the publisher (currently unused but kept for API consistency)
//   - request: The base lookup request containing common parameters
//   - bulkChecker: The BulkChecker instance to publish to
//
// Returns:
//   - *BulkSubjectPublisher: The initialized publisher instance
func NewBulkSubjectPublisher(ctx context.Context, request *base.PermissionLookupSubjectRequest, bulkChecker *BulkChecker) *BulkSubjectPublisher {
	return &BulkSubjectPublisher{
		bulkChecker: bulkChecker,
		request:     request,
	}
}

// Publish sends a subject permission check request to the bulk checker.
// This method creates a permission check request from the provided parameters
// and sends it to the BulkChecker for processing. It handles context cancellation
// gracefully by dropping requests when the context is done.
//
// Parameters:
//   - subject: The subject to check permissions for
//   - metadata: Metadata for the permission check request
//   - context: Additional context for the permission check
//   - result: Optional pre-computed result
func (p *BulkSubjectPublisher) Publish(subject *base.Subject, metadata *base.PermissionCheckRequestMetadata, context *base.Context, result base.CheckResult) {
	select {
	case p.bulkChecker.requestChan <- BulkCheckerRequest{
		Request: &base.PermissionCheckRequest{
			TenantId:   p.request.GetTenantId(),
			Metadata:   metadata,
			Entity:     p.request.GetEntity(),
			Permission: p.request.GetPermission(),
			Subject:    subject,
			Context:    context,
		},
		Result: result,
	}:
	case <-p.bulkChecker.ctx.Done():
		// Context cancelled, drop the request
	}
}

// isContextError checks if an error is related to context cancellation or timeout.
// This helper function centralizes the logic for identifying context-related errors
// that should not be treated as actual errors in the bulk processing system.
//
// Parameters:
//   - err: The error to check
//
// Returns:
//   - bool: True if the error is context-related, false otherwise
func isContextError(err error) bool {
	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}

// IsContextRelatedError is a legacy function maintained for backward compatibility.
// This function provides the same functionality as isContextError but with
// the original signature to maintain compatibility with existing code.
//
// Parameters:
//   - ctx: Context (unused, kept for compatibility)
//   - err: The error to check
//
// Returns:
//   - bool: True if the error is context-related, false otherwise
func IsContextRelatedError(ctx context.Context, err error) bool {
	return isContextError(err)
}
