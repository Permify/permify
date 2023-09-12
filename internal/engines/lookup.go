package engines

import (
	"context"
	"errors"
	"sync"

	"github.com/Permify/permify/internal/schema"
	"github.com/Permify/permify/internal/storage"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

type LookupEngine struct {
	// schemaReader is responsible for reading schema information
	schemaReader storage.SchemaReader
	// checkEngine is responsible for performing permission checks
	checkEngine *CheckEngine
	// schemaBasedEntityFilter is responsible for performing entity filter operations
	schemaBasedEntityFilter *SchemaBasedEntityFilter
	// massEntityFilter is responsible for performing entity filter operations
	massEntityFilter *MassEntityFilter
	// schemaBasedEntityFilter is responsible for performing entity filter operations
	schemaBasedSubjectFilter *SchemaBasedSubjectFilter
	// massEntityFilter is responsible for performing entity filter operations
	massSubjectFilter *MassSubjectFilter
	// schemaMap is a map that keeps track of schema versions
	schemaMap sync.Map
	// concurrencyLimit is the maximum number of concurrent permission checks allowed
	concurrencyLimit int
}

func NewLookupEngine(
	check *CheckEngine,
	schemaReader storage.SchemaReader,
	schemaBasedEntityFilter *SchemaBasedEntityFilter,
	massEntityFilter *MassEntityFilter,
	schemaBasedSubjectFilter *SchemaBasedSubjectFilter,
	massSubjectFilter *MassSubjectFilter,
	opts ...LookupOption,
) *LookupEngine {
	engine := &LookupEngine{
		schemaReader:             schemaReader,
		checkEngine:              check,
		schemaBasedEntityFilter:  schemaBasedEntityFilter,
		massEntityFilter:         massEntityFilter,
		schemaBasedSubjectFilter: schemaBasedSubjectFilter,
		massSubjectFilter:        massSubjectFilter,
		schemaMap:                sync.Map{},
		concurrencyLimit:         _defaultConcurrencyLimit,
	}

	// options
	for _, opt := range opts {
		opt(engine)
	}

	return engine
}

// LookupEntity performs a permission check on a set of entities and returns a response
// containing the IDs of the entities that have the requested permission.
func (engine *LookupEngine) LookupEntity(ctx context.Context, request *base.PermissionLookupEntityRequest) (response *base.PermissionLookupEntityResponse, err error) {
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
	sc, err = engine.readSchema(ctx, request.GetTenantId(), request.GetMetadata().GetSchemaVersion())
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

		// Create a map to keep track of visited entities
		visits := &ERMap{}

		// Set the schema for the schema-based entity filter in the engine
		engine.schemaBasedEntityFilter.SetSchema(sc)

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
func (engine *LookupEngine) LookupEntityStream(ctx context.Context, request *base.PermissionLookupEntityRequest, server base.Permission_LookupEntityStreamServer) (err error) {
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
	sc, err = engine.readSchema(ctx, request.GetTenantId(), request.GetMetadata().GetSchemaVersion())
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

// LookupSubject checks if a subject has a particular permission based on the schema and version.
// It returns a list of subjects that have the given permission.
func (engine *LookupEngine) LookupSubject(ctx context.Context, request *base.PermissionLookupSubjectRequest) (response *base.PermissionLookupSubjectResponse, err error) {
	// Retrieve the schema of the entity based on the provided tenantId and schema version.
	var sc *base.SchemaDefinition
	sc, err = engine.readSchema(ctx, request.GetTenantId(), request.GetMetadata().GetSchemaVersion())
	if err != nil {
		// Return an error if there was an issue retrieving the schema.
		return nil, err
	}

	// Walk the entity schema to perform a permission check.
	err = schema.NewWalker(sc).Walk(request.GetEntity().GetType(), request.GetPermission())
	if err != nil {
		// If the error indicates the schema walk is unimplemented, handle it with a MassEntityFilter.
		if errors.Is(err, schema.ErrUnimplemented) {

			// Use a mutex to protect concurrent writes to the subjectIDs slice.
			var mu sync.Mutex
			var subjectIDs []string

			// Callback function to handle the results of permission checks.
			// If an entity passes the permission check, its ID is stored in the subjectIDs slice.
			callback := func(subjectID string, result base.CheckResult) {
				if result == base.CheckResult_CHECK_RESULT_ALLOWED {
					mu.Lock()         // Lock to prevent concurrent modification of the slice.
					defer mu.Unlock() // Unlock after the ID is appended.
					subjectIDs = append(subjectIDs, subjectID)
				}
			}

			// Create and initiate a BulkChecker to perform permission checks in parallel.
			checker := NewBulkChecker(ctx, engine.checkEngine, callback, engine.concurrencyLimit)
			checker.Start(BULK_SUBJECT)

			// Create and start a BulkPublisher to provide entities to the BulkChecker.
			publisher := NewBulkSubjectPublisher(ctx, request, checker)

			err = engine.massSubjectFilter.SubjectFilter(ctx, request, publisher)
			if err != nil {
				// Return an error if there was an issue with the subject filter.
				return nil, err
			}

			// Stop the BulkChecker and ensure all entities have been processed.
			checker.Stop()
			err = checker.Wait()
			if err != nil {
				return nil, err
			}

			// Return the list of entity IDs that have the required permission.
			return &base.PermissionLookupSubjectResponse{
				SubjectIds: subjectIDs,
			}, nil
		}

		// If the error wasn't due to unimplemented schema walking, return the error.
		return nil, err
	}

	// Use the schema-based subject filter to get the list of subjects with the requested permission.
	ids, err := engine.schemaBasedSubjectFilter.SubjectFilter(ctx, request)
	if err != nil {
		return nil, err
	}

	// Return the list of entity IDs that have the required permission.
	return &base.PermissionLookupSubjectResponse{
		SubjectIds: ids,
	}, nil
}

// readSchema retrieves a SchemaDefinition for a given tenantID and schemaVersion.
// It first checks a cache (schemaMap) for the schema, and if not found, reads it using the schemaReader.
func (engine *LookupEngine) readSchema(ctx context.Context, tenantID, schemaVersion string) (*base.SchemaDefinition, error) {
	// Create a unique cache key by combining the tenantID and schemaVersion.
	// This ensures that different combinations of tenantID and schemaVersion get their own cache entries.
	cacheKey := tenantID + "|" + schemaVersion

	// Attempt to retrieve the schema from the cache (schemaMap) using the generated cacheKey.
	if sch, ok := engine.schemaMap.Load(cacheKey); ok {
		// If the schema is present in the cache, cast it to its correct type and return.
		return sch.(*base.SchemaDefinition), nil
	}

	// If the schema is not present in the cache, use the schemaReader to read it from the source (e.g., a database or file).
	sch, err := engine.schemaReader.ReadSchema(ctx, tenantID, schemaVersion)
	if err != nil {
		// If there's an error reading the schema (e.g., schema not found or database connection issue), return the error.
		return nil, err
	}

	// Cache the newly read schema in schemaMap so that subsequent reads can be faster.
	engine.schemaMap.Store(cacheKey, sch)

	// Return the freshly read schema.
	return sch, nil
}
