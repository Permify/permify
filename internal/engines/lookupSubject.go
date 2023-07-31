package engines

import (
	"context"
	"errors"
	"sync"

	"github.com/Permify/permify/internal/schema"
	"github.com/Permify/permify/internal/storage"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

type LookupSubjectEngine struct {
	// schemaReader is responsible for reading schema information
	schemaReader storage.SchemaReader
	// checkEngine is responsible for performing permission checks
	checkEngine *CheckEngine
	// schemaBasedEntityFilter is responsible for performing entity filter operations
	schemaBasedSubjectFilter *SchemaBasedSubjectFilter
	// massEntityFilter is responsible for performing entity filter operations
	massSubjectFilter *MassSubjectFilter
	// concurrencyLimit is the maximum number of concurrent permission checks allowed
	concurrencyLimit int
}

// NewLookupSubjectEngine is a constructor function for a LookupSubjectEngine.
// It initializes a new LookupSubjectEngine with a schema reader, a relationship reader,
// and an optional set of LookupSubjectOptions.
func NewLookupSubjectEngine(check *CheckEngine, sr storage.SchemaReader, schemaBasedSubjectFilter *SchemaBasedSubjectFilter, massSubjectFilter *MassSubjectFilter, opts ...LookupSubjectOption) *LookupSubjectEngine {
	// Creating a new LookupSubjectEngine and providing it with the schema reader and relationship reader.
	// By default, the concurrency limit is set to a predefined constant _defaultConcurrencyLimit.
	engine := &LookupSubjectEngine{
		checkEngine:              check,
		schemaReader:             sr,
		schemaBasedSubjectFilter: schemaBasedSubjectFilter,
		massSubjectFilter:        massSubjectFilter,
		concurrencyLimit:         _defaultConcurrencyLimit,
	}

	// opts is a variadic function argument, which means it can contain zero or more LookupSubjectOptions.
	// We loop through each option in opts and apply it to the engine.
	// This is a common way in Go to allow optional configuration of a new object.
	for _, opt := range opts {
		opt(engine)
	}

	// Finally, we return the newly configured LookupSubjectEngine.
	return engine
}

func (engine *LookupSubjectEngine) LookupSubject(ctx context.Context, request *base.PermissionLookupSubjectRequest) (response *base.PermissionLookupSubjectResponse, err error) {
	// Retrieve the schema of the entity based on the tenantId and schema version
	var sc *base.SchemaDefinition
	sc, err = engine.schemaReader.ReadSchema(ctx, request.GetTenantId(), request.GetMetadata().GetSchemaVersion())
	if err != nil {
		return nil, err
	}

	// Perform a walk of the entity schema for the permission check
	err = schema.NewWalker(sc).Walk(request.GetEntity().GetType(), request.GetPermission())
	if err != nil {
		// If the error is unimplemented, handle it with a MassEntityFilter
		if errors.Is(err, schema.ErrUnimplemented) {

			// A mutex and slice are declared to safely store entity IDs from concurrent callbacks
			var mu sync.Mutex
			var subjectIDs []string

			// Callback function which is called for each entity. If the entity passes the permission check,
			// the entity ID is appended to the entityIDs slice.
			callback := func(subjectID string, result base.CheckResult) {
				if result == base.CheckResult_CHECK_RESULT_ALLOWED {
					mu.Lock()         // Safeguard access to the shared slice with a mutex
					defer mu.Unlock() // Ensure the lock is released after appending the ID
					subjectIDs = append(subjectIDs, subjectID)
				}
			}

			// Create and start BulkChecker. It performs permission checks in parallel.
			checker := NewBulkChecker(ctx, engine.checkEngine, callback, engine.concurrencyLimit)
			checker.Start(BULK_SUBJECT)

			// Create and start BulkPublisher. It receives entities and passes them to BulkChecker.
			publisher := NewBulkSubjectPublisher(ctx, request, checker)

			err = engine.massSubjectFilter.SubjectFilter(ctx, request, publisher)
			if err != nil {
				return nil, err
			}

			// Stop the BulkChecker and wait for it to finish processing entities
			checker.Stop()
			err = checker.Wait()
			if err != nil {
				return
			}

			// Return response containing allowed entity IDs
			return &base.PermissionLookupSubjectResponse{
				SubjectIds: subjectIDs,
			}, nil
		}

		return nil, err
	}

	// Perform an entity filter operation based on the permission request
	ids, err := engine.schemaBasedSubjectFilter.SubjectFilter(ctx, request)
	if err != nil {
		return nil, err
	}

	// Return response containing allowed entity IDs
	return &base.PermissionLookupSubjectResponse{
		SubjectIds: ids,
	}, nil
}
