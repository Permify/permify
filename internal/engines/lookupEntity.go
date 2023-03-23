package engines

import (
	"context"
	"sync"

	"github.com/Permify/permify/internal/repositories"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/token"
)

// LookupEntityEngine is a struct that performs permission checks on a set of entities
// and returns the entities that have the requested permission.
type LookupEntityEngine struct {
	checkEngine *CheckEngine
	// relationshipReader is responsible for reading relationship information
	relationshipReader repositories.RelationshipReader
	// concurrencyLimit is the maximum number of concurrent permission checks allowed
	concurrencyLimit int
}

// NewLookupEntityEngine creates a new LookupEntityEngine instance.
// engine: the CheckEngine to use for permission checks
// reader: the RelationshipReader to retrieve entity relationships
func NewLookupEntityEngine(checker *CheckEngine, reader repositories.RelationshipReader, opts ...LookupEntityOption) *LookupEntityEngine {
	engine := &LookupEntityEngine{
		checkEngine:        checker,
		relationshipReader: reader,
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

	// Retrieve the snapshot token if not provided
	if request.GetMetadata().GetSnapToken() == "" {
		var st token.SnapToken
		st, err = engine.relationshipReader.HeadSnapshot(ctx, request.GetTenantId())
		if err != nil {
			return response, err
		}
		request.Metadata.SnapToken = st.Encode().String()
	}

	// Get unique entity IDs by entity type
	var ids []string
	ids, err = engine.relationshipReader.GetUniqueEntityIDsByEntityType(ctx, request.GetTenantId(), request.GetEntityType(), request.GetMetadata().GetSnapToken())
	if err != nil {
		return nil, err
	}

	// Mutex and slice for storing allowed entity IDs
	var mu sync.Mutex
	entityIDs := make([]string, 0, len(ids))

	callback := func(entityID string, result base.PermissionCheckResponse_Result) {
		if result == base.PermissionCheckResponse_RESULT_ALLOWED {
			mu.Lock()
			defer mu.Unlock()
			entityIDs = append(entityIDs, entityID)
		}
	}

	checker := NewBulkChecker(ctx, engine.checkEngine, callback, engine.concurrencyLimit)
	checker.Start()

	for _, id := range ids {
		checker.RequestChan <- &base.PermissionCheckRequest{
			TenantId: request.GetTenantId(),
			Metadata: &base.PermissionCheckRequestMetadata{
				SnapToken:     request.GetMetadata().GetSnapToken(),
				SchemaVersion: request.GetMetadata().GetSchemaVersion(),
				Depth:         request.GetMetadata().GetDepth(),
				Exclusion:     false,
			},
			Entity: &base.Entity{
				Type: request.GetEntityType(),
				Id:   id,
			},
			Permission: request.GetPermission(),
			Subject:    request.GetSubject(),
		}
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

	// Retrieve the snapshot token if not provided
	if request.GetMetadata().GetSnapToken() == "" {
		var st token.SnapToken
		st, err = engine.relationshipReader.HeadSnapshot(ctx, request.GetTenantId())
		if err != nil {
			return err
		}
		request.Metadata.SnapToken = st.Encode().String()
	}

	// Get unique entity IDs by entity type
	var ids []string
	ids, err = engine.relationshipReader.GetUniqueEntityIDsByEntityType(ctx, request.GetTenantId(), request.GetEntityType(), request.GetMetadata().GetSnapToken())
	if err != nil {
		return err
	}

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

	// Add permission check requests to the BulkChecker's RequestChan
	for _, id := range ids {
		checker.RequestChan <- &base.PermissionCheckRequest{
			TenantId: request.GetTenantId(),
			Metadata: &base.PermissionCheckRequestMetadata{
				SnapToken:     request.GetMetadata().GetSnapToken(),
				SchemaVersion: request.GetMetadata().GetSchemaVersion(),
				Depth:         request.GetMetadata().GetDepth(),
				Exclusion:     false,
			},
			Entity: &base.Entity{
				Type: request.GetEntityType(),
				Id:   id,
			},
			Permission: request.GetPermission(),
			Subject:    request.GetSubject(),
		}
	}

	// Stop input and wait for BulkChecker to finish
	checker.Stop()
	err = checker.Wait()
	if err != nil {
		return
	}

	return err
}
