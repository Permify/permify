package engines

import (
	"context"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"

	"github.com/Permify/permify/internal/invoke"
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
type BulkChecker struct {
	checker invoke.Check
	// input queue for permission check requests
	RequestChan chan BulkCheckerRequest
	// context to manage goroutines and cancellation
	ctx context.Context
	// errgroup for managing multiple goroutines
	g *errgroup.Group
	// limit for concurrent permission checks
	concurrencyLimit int
	// callback function to handle the result of each permission check
	callback func(entityID string, permission string, result base.CheckResult)
}

// NewBulkChecker creates a new BulkChecker instance.
// ctx: context for managing goroutines and cancellation
// engine: the CheckEngine to use for permission checks
// callback: a callback function that handles the result of each permission check
// concurrencyLimit: the maximum number of concurrent permission checks
func NewBulkChecker(ctx context.Context, checker invoke.Check, callback func(entityID string, permission string, result base.CheckResult), concurrencyLimit int) *BulkChecker {
	return &BulkChecker{
		RequestChan:      make(chan BulkCheckerRequest),
		checker:          checker,
		g:                &errgroup.Group{},
		concurrencyLimit: concurrencyLimit,
		ctx:              ctx,
		callback:         callback,
	}
}

// Start begins processing permission check requests from the RequestChan.
// It starts an errgroup that manages multiple goroutines for performing permission checks.
func (c *BulkChecker) Start(typ BulkCheckerType) {
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
				if req.Result == base.CheckResult_CHECK_RESULT_UNSPECIFIED {
					result, err := c.checker.Check(c.ctx, req.Request)
					if err != nil {
						return err
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
