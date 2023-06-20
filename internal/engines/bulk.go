package engines

import (
	"context"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"

	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// BulkCheckerRequest is a struct for a permission check request and the channel to send the result.
type BulkCheckerRequest struct {
	Request *base.PermissionCheckRequest
	Result  base.PermissionCheckResponse_Result
}

// BulkChecker is a struct for checking permissions in bulk.
type BulkChecker struct {
	checkEngine *CheckEngine
	// input queue for permission check requests
	RequestChan chan BulkCheckerRequest
	// context to manage goroutines and cancellation
	ctx context.Context
	// errgroup for managing multiple goroutines
	g *errgroup.Group
	// limit for concurrent permission checks
	concurrencyLimit int
	// callback function to handle the result of each permission check
	callback func(entityID string, result base.PermissionCheckResponse_Result)
}

// NewBulkChecker creates a new BulkChecker instance.
// ctx: context for managing goroutines and cancellation
// engine: the CheckEngine to use for permission checks
// callback: a callback function that handles the result of each permission check
// concurrencyLimit: the maximum number of concurrent permission checks
func NewBulkChecker(ctx context.Context, engine *CheckEngine, callback func(entityID string, result base.PermissionCheckResponse_Result), concurrencyLimit int) *BulkChecker {
	return &BulkChecker{
		RequestChan:      make(chan BulkCheckerRequest),
		checkEngine:      engine,
		g:                &errgroup.Group{},
		concurrencyLimit: concurrencyLimit,
		ctx:              ctx,
		callback:         callback,
	}
}

// Start begins processing permission check requests from the RequestChan.
// It starts an errgroup that manages multiple goroutines for performing permission checks.
func (c *BulkChecker) Start() {
	c.g.Go(func() error {
		sem := semaphore.NewWeighted(int64(c.concurrencyLimit))
		for {
			// acquire a semaphore before processing a request
			if err := sem.Acquire(c.ctx, 1); err != nil {
				return err
			}
			// read a request from the RequestChan
			req, ok := <-c.RequestChan
			if !ok {
				sem.Release(1)
				break
			}
			// run the permission check in a separate goroutine
			c.g.Go(func() error {
				defer sem.Release(1)
				if req.Result == base.PermissionCheckResponse_RESULT_UNKNOWN {
					result, err := c.checkEngine.Check(c.ctx, req.Request)
					if err != nil {
						return err
					}

					// call the callback with the result
					c.callback(req.Request.GetEntity().GetId(), result.Can)
				} else {
					c.callback(req.Request.GetEntity().GetId(), req.Result)
				}
				return nil
			})
		}
		// wait for all remaining semaphore resources to be released
		return sem.Acquire(c.ctx, int64(c.concurrencyLimit))
	})
}

// Stop stops input by closing the RequestChan.
func (c *BulkChecker) Stop() {
	close(c.RequestChan)
}

// Wait waits for all goroutines in the errgroup to finish.
// Returns an error if any of the goroutines encounter an error.
func (c *BulkChecker) Wait() error {
	return c.g.Wait()
}

// BulkPublisher is a struct for streaming permission check results.
type BulkPublisher struct {
	bulkChecker *BulkChecker

	request *base.PermissionLookupEntityRequest
	// context to manage goroutines and cancellation
	ctx context.Context
}

// NewBulkPublisher creates a new BulkStreamer instance.
func NewBulkPublisher(ctx context.Context, request *base.PermissionLookupEntityRequest, bulkChecker *BulkChecker) *BulkPublisher {
	return &BulkPublisher{
		bulkChecker: bulkChecker,
		request:     request,
		ctx:         ctx,
	}
}

// Publish publishes a permission check request to the BulkChecker.
func (s *BulkPublisher) Publish(entity *base.Entity, metadata *base.PermissionCheckRequestMetadata, contextual []*base.Tuple, result base.PermissionCheckResponse_Result) {
	s.bulkChecker.RequestChan <- BulkCheckerRequest{
		Request: &base.PermissionCheckRequest{
			TenantId:         s.request.GetTenantId(),
			Metadata:         metadata,
			Entity:           entity,
			Permission:       s.request.GetPermission(),
			Subject:          s.request.GetSubject(),
			ContextualTuples: contextual,
		},
		Result: result,
	}
}
