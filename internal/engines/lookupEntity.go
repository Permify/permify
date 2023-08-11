package engines

import (
	"context"
	"errors"
	"sync"

	"github.com/Permify/permify/internal/schema"
	"github.com/Permify/permify/internal/storage"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// LookupEntityEngine is a struct that performs permission checks on a set of entities
// and returns the entities that have the requested permission.
type LookupEntityEngine struct {
	// schemaReader is responsible for reading schema information
	schemaReader storage.SchemaReader
	// checkEngine is responsible for performing permission checks
	checkEngine *CheckEngine
	// schemaBasedEntityFilter is responsible for performing entity filter operations
	schemaBasedEntityFilter *SchemaBasedEntityFilter
	// massEntityFilter is responsible for performing entity filter operations
	massEntityFilter *MassEntityFilter
	// concurrencyLimit is the maximum number of concurrent permission checks allowed
	concurrencyLimit int
}

// NewLookupEntityEngine creates a new LookupEntityEngine instance.
// engine: the CheckEngine to use for permission checks
// reader: the RelationshipReader to retrieve entity relationships
func NewLookupEntityEngine(check *CheckEngine, schemaReader storage.SchemaReader, schemaBasedEntityFilter *SchemaBasedEntityFilter, massEntityFilter *MassEntityFilter, opts ...LookupEntityOption) *LookupEntityEngine {
	engine := &LookupEntityEngine{
		schemaReader:            schemaReader,
		checkEngine:             check,
		schemaBasedEntityFilter: schemaBasedEntityFilter,
		massEntityFilter:        massEntityFilter,
		concurrencyLimit:        _defaultConcurrencyLimit,
	}

	// options
	for _, opt := range opts {
		opt(engine)
	}

	return engine
}

// LookupEntity performs a permission check on a set of entities and returns a response
// containing the IDs of the entities that have the requested permission.
func (engine *LookupEntityEngine) LookupEntity(ctx context.Context, request *base.PermissionLookupEntityRequest) (response *base.PermissionLookupEntityResponse, err error) {
	// A mutex and slice are declared to safely store entity IDs from concurrent callbacks
	var mu sync.Mutex
	var entityIDs []string

	// Callback function which is called for each entity. If the entity passes the permission check,
	// the entity ID is appended to the entityIDs slice.
	callback := func(entityID string, result base.CheckResult) {
		if result == base.CheckResult_CHECK_RESULT_ALLOWED {
			mu.Lock()         // Safeguard access to the shared slice with a mutex
			defer mu.Unlock() // Ensure the lock is released after appending the ID
			entityIDs = append(entityIDs, entityID)
		}
	}

	// Create and start BulkChecker. It performs permission checks in parallel.
	checker := NewBulkChecker(ctx, engine.checkEngine, callback, engine.concurrencyLimit)
	checker.Start(BULK_ENTITY)

	// Create and start BulkPublisher. It receives entities and passes them to BulkChecker.
	publisher := NewBulkEntityPublisher(ctx, request, checker)

	// Retrieve the schema of the entity based on the tenantId and schema version
	var sc *base.SchemaDefinition
	sc, err = engine.schemaReader.ReadSchema(ctx, request.GetTenantId(), request.GetMetadata().GetSchemaVersion())
	if err != nil {
		return nil, err
	}

	// Perform a walk of the entity schema for the permission check
	err = schema.NewWalker(sc).Walk(request.GetEntityType(), request.GetPermission())
	if err != nil {
		// If the error is unimplemented, handle it with a MassEntityFilter
		if errors.Is(err, schema.ErrUnimplemented) {
			err = engine.massEntityFilter.EntityFilter(ctx, request, publisher)
			if err != nil {
				return nil, err
			}
		} else { // For other errors, simply return the error
			return nil, err
		}
	} else {

		// Map to keep track of visited entities
		visits := &ERMap{}

		// Perform an entity filter operation based on the permission request
		err = engine.schemaBasedEntityFilter.EntityFilter(ctx, &base.PermissionEntityFilterRequest{
			TenantId: request.GetTenantId(),
			Metadata: &base.PermissionEntityFilterRequestMetadata{
				SnapToken:     request.GetMetadata().GetSnapToken(),
				SchemaVersion: request.GetMetadata().GetSchemaVersion(),
				Depth:         request.GetMetadata().GetDepth(),
			},
			EntityReference: &base.RelationReference{
				Type:     request.GetEntityType(),
				Relation: request.GetPermission(),
			},
			Subject: request.GetSubject(),
			Context: request.GetContext(),
		}, visits, publisher)

		if err != nil {
			return nil, err
		}
	}

	// Stop the BulkChecker and wait for it to finish processing entities
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

// LookupEntityStream performs a permission check on a set of entities and streams the results
// containing the IDs of the entities that have the requested permission.
func (engine *LookupEntityEngine) LookupEntityStream(ctx context.Context, request *base.PermissionLookupEntityRequest, server base.Permission_LookupEntityStreamServer) (err error) {
	// Define a callback function that will be called for each entity that passes the permission check.
	// If the check result is allowed, it sends the entity ID to the server stream.
	callback := func(entityID string, result base.CheckResult) {
		if result == base.CheckResult_CHECK_RESULT_ALLOWED {
			err := server.Send(&base.PermissionLookupEntityStreamResponse{
				EntityId: entityID,
			})
			// If there is an error in sending the response, the function will return
			if err != nil {
				return
			}
		}
	}

	// Create and start BulkChecker. It performs permission checks concurrently.
	checker := NewBulkChecker(ctx, engine.checkEngine, callback, engine.concurrencyLimit)
	checker.Start(BULK_ENTITY)

	// Create and start BulkPublisher. It receives entities and passes them to BulkChecker.
	publisher := NewBulkEntityPublisher(ctx, request, checker)

	// Retrieve the entity definition schema based on the tenantId and schema version
	var sc *base.SchemaDefinition
	sc, err = engine.schemaReader.ReadSchema(ctx, request.GetTenantId(), request.GetMetadata().GetSchemaVersion())
	if err != nil {
		return err
	}

	// Perform a permission check walk through the entity schema
	err = schema.NewWalker(sc).Walk(request.GetEntityType(), request.GetPermission())

	// If error exists in permission check walk
	if err != nil {
		// If the error is unimplemented, handle it with a MassEntityFilter
		if errors.Is(err, schema.ErrUnimplemented) {
			err = engine.massEntityFilter.EntityFilter(ctx, request, publisher)
			if err != nil {
				return err
			}
		} else { // For other types of errors, simply return the error
			return err
		}
	} else { // If there was no error in permission check walk
		visits := &ERMap{}

		// Perform an entity filter operation based on the permission request
		err = engine.schemaBasedEntityFilter.EntityFilter(ctx, &base.PermissionEntityFilterRequest{
			TenantId: request.GetTenantId(),
			Metadata: &base.PermissionEntityFilterRequestMetadata{
				SnapToken:     request.GetMetadata().GetSnapToken(),
				SchemaVersion: request.GetMetadata().GetSchemaVersion(),
				Depth:         request.GetMetadata().GetDepth(),
			},
			EntityReference: &base.RelationReference{
				Type:     request.GetEntityType(),
				Relation: request.GetPermission(),
			},
			Subject: request.GetSubject(),
			Context: request.GetContext(),
		}, visits, publisher)

		if err != nil {
			return err
		}
	}

	// Stop the BulkChecker and wait for it to finish processing entities
	checker.Stop()
	err = checker.Wait()
	if err != nil {
		return
	}

	return err
}
