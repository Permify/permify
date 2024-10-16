package engines

import (
	"context"
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

type BulkCheckerType string

const (
	BULK_SUBJECT BulkCheckerType = "subject"
	BULK_ENTITY  BulkCheckerType = "entity"
)

// BulkCheckerRequest is a struct for a permission check request and the channel to send the result.
type BulkCheckerRequest struct {
	Request *base.PermissionCheckRequest
	Result  base.CheckResult
}

// BulkChecker is a struct for checking permissions in bulk.
// It processes permission check requests concurrently and maintains a sorted list of these requests.
type BulkChecker struct {
	// typ defines the type of bulk checking being performed.
	// It distinguishes between different modes of operation within the BulkChecker,
	// such as whether the check is focused on entities, subjects, or another criterion.
	typ BulkCheckerType

	checker invoke.Check
	// RequestChan is the input queue for permission check requests.
	// Incoming requests are received on this channel and processed by the BulkChecker.
	RequestChan chan BulkCheckerRequest

	// ctx is the context used to manage goroutines and handle cancellation.
	// It allows for graceful shutdown of all goroutines when the context is canceled.
	ctx context.Context

	// g is an errgroup used for managing multiple goroutines.
	// It allows BulkChecker to wait for all goroutines to finish and to capture any errors they may return.
	g *errgroup.Group

	// concurrencyLimit is the maximum number of concurrent permission checks that can be processed at one time.
	// It controls the level of parallelism within the BulkChecker.
	concurrencyLimit int

	// callback is a function that handles the result of each permission check.
	// It is called with the entity ID and the result of the permission check (e.g., allowed or denied).
	callback func(entityID, ct string)

	// sortedList is a slice that stores BulkCheckerRequest objects.
	// This list is maintained in a sorted order based on some criteria, such as the entity ID.
	list []BulkCheckerRequest

	// mu is a mutex used for thread-safe access to the sortedList.
	// It ensures that only one goroutine can modify the sortedList at a time, preventing data races.
	mu sync.Mutex

	// wg is a WaitGroup used to coordinate the completion of request collection.
	// It ensures that all requests are collected and processed before ExecuteRequests begins execution.
	// The WaitGroup helps to synchronize the collection of requests with the execution of those requests,
	// preventing race conditions where the execution starts before all requests are collected.
	wg *sync.WaitGroup
}

// NewBulkChecker creates a new BulkChecker instance.
// ctx: context for managing goroutines and cancellation
// engine: the CheckEngine to use for permission checks
// callback: a callback function that handles the result of each permission check
// concurrencyLimit: the maximum number of concurrent permission checks
func NewBulkChecker(ctx context.Context, checker invoke.Check, typ BulkCheckerType, callback func(entityID, ct string), concurrencyLimit int) *BulkChecker {
	bc := &BulkChecker{
		RequestChan:      make(chan BulkCheckerRequest),
		checker:          checker,
		g:                &errgroup.Group{},
		concurrencyLimit: concurrencyLimit,
		ctx:              ctx,
		callback:         callback,
		typ:              typ,
		wg:               &sync.WaitGroup{},
	}

	// Start processing requests in a separate goroutine
	// Use a WaitGroup to ensure all requests are collected before proceeding
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done() // Signal that the request collection is finished
		bc.CollectAndSortRequests()
	}()

	bc.wg = &wg // Store the WaitGroup for future use

	return bc
}

// CollectAndSortRequests processes incoming requests and maintains a sorted list.
func (bc *BulkChecker) CollectAndSortRequests() {
	for {
		select {
		case req, ok := <-bc.RequestChan:
			if !ok {
				// Channel closed, process remaining requests
				return
			}

			bc.mu.Lock()
			bc.list = append(bc.list, req)
			// Optionally, you could sort here or later in ExecuteRequests
			bc.mu.Unlock()

		case <-bc.ctx.Done():
			return
		}
	}
}

// prepareRequestList sorts and prepares a copy of the request list.
func (bc *BulkChecker) prepareRequestList() []BulkCheckerRequest {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	bc.sortRequests()                                 // Sort requests based on ID
	return append([]BulkCheckerRequest{}, bc.list...) // Return a copy of the list to avoid modifying the original
}

// StopCollectingRequests Signal to stop collecting requests and close the channel
func (bc *BulkChecker) StopCollectingRequests() {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	// Close the channel to signal no more requests will be sent
	close(bc.RequestChan)
}

// sortRequests sorts the sortedList based on the type (entity or subject).
func (bc *BulkChecker) sortRequests() {
	if bc.typ == BULK_ENTITY {
		sort.Slice(bc.list, func(i, j int) bool {
			return bc.list[i].Request.GetEntity().GetId() < bc.list[j].Request.GetEntity().GetId()
		})
	} else if bc.typ == BULK_SUBJECT {
		sort.Slice(bc.list, func(i, j int) bool {
			return bc.list[i].Request.GetSubject().GetId() < bc.list[j].Request.GetSubject().GetId()
		})
	}
}

// ExecuteRequests begins processing permission check requests from the sorted list.
func (bc *BulkChecker) ExecuteRequests(size uint32) error {
	// Stop collecting new requests and close the RequestChan to ensure no more requests are added
	bc.StopCollectingRequests()

	// Wait for request collection to complete before proceeding
	bc.wg.Wait()

	// Create a context with cancellation that will be used to stop further processing
	ctx, cancel := context.WithCancel(bc.ctx)
	defer cancel() // Ensure the context is canceled when done

	// Track the number of successful permission checks
	successCount := int64(0)
	// Semaphore to control the maximum number of concurrent permission checks
	sem := semaphore.NewWeighted(int64(bc.concurrencyLimit))
	var mu sync.Mutex

	// Sort and copy the list of requests
	listCopy := bc.prepareRequestList()

	// Pre-allocate a slice to store the results of the permission checks
	results := make([]base.CheckResult, len(listCopy))
	// Track the index of the last processed request to ensure results are processed in order
	processedIndex := 0

	// Loop through each request in the copied list
	for i, currentRequest := range listCopy {
		// If we've reached the success limit, stop processing further requests
		if atomic.LoadInt64(&successCount) >= int64(size) {
			cancel() // Cancel the context to stop further goroutines
			break
		}

		index := i
		req := currentRequest

		// Use errgroup to manage the goroutines, which allows for error handling and synchronization
		bc.g.Go(func() error {
			// Check if the context has been canceled, and return without error if so
			if err := ctx.Err(); err == context.Canceled {
				return nil // Gracefully exit the goroutine if context is canceled
			}

			// Acquire a slot in the semaphore to control concurrency
			if err := sem.Acquire(ctx, 1); err != nil {
				// Return nil instead of error if the context was canceled during semaphore acquisition
				if IsContextRelatedError(ctx, err) {
					return nil
				}
				return err // Return an error if semaphore acquisition fails for other reasons
			}
			defer sem.Release(1) // Ensure the semaphore slot is released after processing

			var result base.CheckResult
			if req.Result == base.CheckResult_CHECK_RESULT_UNSPECIFIED {
				// Perform the permission check if the result is not already specified
				cr, err := bc.checker.Check(ctx, req.Request)
				if err != nil {
					// Handle context cancellation error here
					if IsContextRelatedError(ctx, err) {
						return nil // Ignore the cancellation error
					}
					return err // Return the actual error if it's not due to cancellation
				}
				result = cr.GetCan() // Get the result from the check
			} else {
				// Use the already specified result
				result = req.Result
			}

			// Lock the mutex to safely update shared resources
			mu.Lock()
			results[index] = result // Store the result in the pre-allocated slice

			// Process the results in order, starting from the current processed index
			for processedIndex < len(listCopy) && results[processedIndex] != base.CheckResult_CHECK_RESULT_UNSPECIFIED {
				// If the result at the processed index is allowed, call the callback function
				if results[processedIndex] == base.CheckResult_CHECK_RESULT_ALLOWED {
					if atomic.AddInt64(&successCount, 1) <= int64(size) {
						ct := ""
						if processedIndex+1 < len(listCopy) {
							// If there is a next item, create a continuous token with the next ID
							if bc.typ == BULK_ENTITY {
								ct = utils.NewContinuousToken(listCopy[processedIndex+1].Request.GetEntity().GetId()).Encode().String()
							} else if bc.typ == BULK_SUBJECT {
								ct = utils.NewContinuousToken(listCopy[processedIndex+1].Request.GetSubject().GetId()).Encode().String()
							}
						}
						// Depending on the type of check (entity or subject), call the appropriate callback
						if bc.typ == BULK_ENTITY {
							bc.callback(listCopy[processedIndex].Request.GetEntity().GetId(), ct)
						} else if bc.typ == BULK_SUBJECT {
							bc.callback(listCopy[processedIndex].Request.GetSubject().GetId(), ct)
						}
					}
				}
				processedIndex++ // Move to the next index for processing
			}
			mu.Unlock() // Unlock the mutex after updating the results and processed index

			// If the success limit has been reached, cancel the context to stop further processing
			if atomic.LoadInt64(&successCount) >= int64(size) {
				cancel()
			}

			return nil // Return nil to indicate successful processing
		})
	}

	// Wait for all goroutines to complete and check for any errors
	if err := bc.g.Wait(); err != nil {
		// Ignore context cancellation as an error
		if IsContextRelatedError(ctx, err) {
			return nil
		}
		return err // Return other errors if any goroutine returned an error
	}

	return nil // Return nil if all processing completed successfully
}

// BulkEntityPublisher is a struct for streaming permission check results.
type BulkEntityPublisher struct {
	bulkChecker *BulkChecker

	request *base.PermissionLookupEntityRequest
	// context to manage goroutines and cancellation
	ctx context.Context
}

// NewBulkEntityPublisher creates a new BulkStreamer instance.
func NewBulkEntityPublisher(ctx context.Context, request *base.PermissionLookupEntityRequest, bulkChecker *BulkChecker) *BulkEntityPublisher {
	return &BulkEntityPublisher{
		bulkChecker: bulkChecker,
		request:     request,
		ctx:         ctx,
	}
}

// Publish publishes a permission check request to the BulkChecker.
func (s *BulkEntityPublisher) Publish(entity *base.Entity, metadata *base.PermissionCheckRequestMetadata, context *base.Context, result base.CheckResult) {
	s.bulkChecker.RequestChan <- BulkCheckerRequest{
		Request: &base.PermissionCheckRequest{
			TenantId:   s.request.GetTenantId(),
			Metadata:   metadata,
			Entity:     entity,
			Permission: s.request.GetPermission(),
			Subject:    s.request.GetSubject(),
			Context:    context,
		},
		Result: result,
	}
}

// BulkSubjectPublisher is a struct for streaming permission check results.
type BulkSubjectPublisher struct {
	bulkChecker *BulkChecker

	request *base.PermissionLookupSubjectRequest
	// context to manage goroutines and cancellation
	ctx context.Context
}

// NewBulkSubjectPublisher creates a new BulkStreamer instance.
func NewBulkSubjectPublisher(ctx context.Context, request *base.PermissionLookupSubjectRequest, bulkChecker *BulkChecker) *BulkSubjectPublisher {
	return &BulkSubjectPublisher{
		bulkChecker: bulkChecker,
		request:     request,
		ctx:         ctx,
	}
}

// Publish publishes a permission check request to the BulkChecker.
func (s *BulkSubjectPublisher) Publish(subject *base.Subject, metadata *base.PermissionCheckRequestMetadata, context *base.Context, result base.CheckResult) {
	s.bulkChecker.RequestChan <- BulkCheckerRequest{
		Request: &base.PermissionCheckRequest{
			TenantId:   s.request.GetTenantId(),
			Metadata:   metadata,
			Entity:     s.request.GetEntity(),
			Permission: s.request.GetPermission(),
			Subject:    subject,
			Context:    context,
		},
		Result: result,
	}
}

// IsContextRelatedError checks if the error is due to context cancellation, deadline exceedance, or closed connection
func IsContextRelatedError(ctx context.Context, err error) bool {
	return errors.Is(ctx.Err(), context.Canceled) || errors.Is(ctx.Err(), context.DeadlineExceeded)
}
