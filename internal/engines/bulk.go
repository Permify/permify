package engines

import (
	"context"
	"sort"
	"sync"
	"sync/atomic"

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
	// callback function to handle the result of each permission check
	callback func(entityID, permission string, result base.CheckResult)
}

// NewBulkChecker creates a new BulkChecker instance.
// ctx: context for managing goroutines and cancellation
// engine: the CheckEngine to use for permission checks
// callback: a callback function that handles the result of each permission check
// concurrencyLimit: the maximum number of concurrent permission checks
func NewBulkChecker(ctx context.Context, checker invoke.Check, callback func(entityID, permission string, result base.CheckResult), concurrencyLimit int) *BulkChecker {
	return &BulkChecker{
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

// StopCollectingRequests Signal to stop collecting requests and close the channel
func (bc *BulkChecker) StopCollectingRequests() {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	// Close the channel to signal no more requests will be sent
	close(bc.RequestChan)
}

					if typ == BULK_ENTITY {
						// call the callback with the result
						c.callback(req.Request.GetEntity().GetId(), req.Request.GetPermission(), result.Can)
					} else if typ == BULK_SUBJECT {
						c.callback(req.Request.GetSubject().GetId(), req.Request.GetPermission(), result.Can)
					}

				} else {
					if typ == BULK_ENTITY {
						// call the callback with the result
						c.callback(req.Request.GetEntity().GetId(), req.Request.GetPermission(), req.Result)
					} else if typ == BULK_SUBJECT {
						c.callback(req.Request.GetSubject().GetId(), req.Request.GetPermission(), req.Result)
					}
				}
				processedIndex++ // Move to the next index for processing
			}
			mu.Unlock() // Unlock the mutex after updating the results and processed index

			return nil // Return nil to indicate successful processing
		})
	}

	// Wait for all goroutines to complete and check for any errors
	if err := bc.g.Wait(); err != nil {
		return err // Return the error if any goroutine returned an error
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
func (s *BulkEntityPublisher) Publish(entity *base.Entity, metadata *base.PermissionCheckRequestMetadata, context *base.Context, result base.CheckResult, permissionChecks *ERMap) {
	for _, permission := range s.request.GetPermissions() {
		if !permissionChecks.Add(entity, permission) {
			continue
		}
		s.bulkChecker.RequestChan <- BulkCheckerRequest{
			Request: &base.PermissionCheckRequest{
				TenantId:   s.request.GetTenantId(),
				Metadata:   metadata,
				Entity:     entity,
				Permission: permission,
				Subject:    s.request.GetSubject(),
				Context:    context,
			},
			Result: result,
		}
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
