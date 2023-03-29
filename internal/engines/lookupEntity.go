package engines

import (
	"context"
	"sync"

	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// LookupEntityEngine is a struct that performs permission checks on a set of entities
// and returns the entities that have the requested permission.
type LookupEntityEngine struct {
	// checkEngine is responsible for performing permission checks
	checkEngine *CheckEngine
	// linkedEntityEngine is responsible for retrieving linked entities
	linkedEntityEngine *LinkedEntityEngine
	// concurrencyLimit is the maximum number of concurrent permission checks allowed
	concurrencyLimit int
}

// NewLookupEntityEngine creates a new LookupEntityEngine instance.
// engine: the CheckEngine to use for permission checks
// reader: the RelationshipReader to retrieve entity relationships
func NewLookupEntityEngine(check *CheckEngine, linked *LinkedEntityEngine, opts ...LookupEntityOption) *LookupEntityEngine {
	engine := &LookupEntityEngine{
		checkEngine:        check,
		linkedEntityEngine: linked,
		concurrencyLimit:   _defaultConcurrencyLimit,
	}

	// options
	for _, opt := range opts {
		opt(engine)
	}

	return engine
}

// Run performs a permission check on a set of entities and returns a response
// containing the IDs of the entities that have the requested permission.
func (engine *LookupEntityEngine) Run(ctx context.Context, request *base.PermissionLookupEntityRequest) (response *base.PermissionLookupEntityResponse, err error) {
	ctx, span := tracer.Start(ctx, "permissions.lookup-entity.execute")
	defer span.End()

	// Mutex and slice for storing allowed entity IDs
	var mu sync.Mutex
	var entityIDs []string

	callback := func(entityID string, result base.PermissionCheckResponse_Result) {
		if result == base.PermissionCheckResponse_RESULT_ALLOWED {
			mu.Lock()
			defer mu.Unlock()
			entityIDs = append(entityIDs, entityID)
		}
	}

	checker := NewBulkChecker(ctx, engine.checkEngine, callback, engine.concurrencyLimit)
	checker.Start()

	// Create and start BulkPublisher
	publisher := NewBulkPublisher(ctx, checker)

	// Create ERMap for storing visited entities
	visits := &ERMap{}

	// Get unique entity IDs by entity type
	err = engine.linkedEntityEngine.Run(ctx, request, visits, publisher)
	if err != nil {
		return nil, err
	}

	// Stop input and wait for BulkChecker to finish
	checker.Stop()
	err = checker.Wait()
	if err != nil {
		return
	}

	// Return response containing allowed entity IDs
	return &base.PermissionLookupEntityResponse{
		EntityIds: entityIDs,
	}, nil
}

// Stream performs a permission check on a set of entities and streams the results
// containing the IDs of the entities that have the requested permission.
func (engine *LookupEntityEngine) Stream(ctx context.Context, request *base.PermissionLookupEntityRequest, server base.Permission_LookupEntityStreamServer) (err error) {
	ctx, span := tracer.Start(ctx, "permissions.lookup-entity.stream")
	defer span.End()

	var mu sync.Mutex

	// Define callback function for handling permission check results
	callback := func(entityID string, result base.PermissionCheckResponse_Result) {
		if result == base.PermissionCheckResponse_RESULT_ALLOWED {
			// Send allowed entity ID to the stream
			mu.Lock()
			defer mu.Unlock()
			if err = server.Send(&base.PermissionLookupEntityStreamResponse{
				EntityId: entityID,
			}); err != nil {
				return
			}
		}
	}

	// Create and start BulkChecker with callback function
	checker := NewBulkChecker(ctx, engine.checkEngine, callback, engine.concurrencyLimit)
	checker.Start()

	// Create and start BulkPublisher
	publisher := NewBulkPublisher(ctx, checker)

	// Create ERMap for storing visited entities
	visits := &ERMap{}

	// Get unique entity IDs by entity type
	err = engine.linkedEntityEngine.Run(ctx, request, visits, publisher)
	if err != nil {
		return err
	}

	// Stop input and wait for BulkChecker to finish
	checker.Stop()
	err = checker.Wait()
	if err != nil {
		return
	}

	return err
}
