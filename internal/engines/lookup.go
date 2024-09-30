package engines

import (
	"context"
	"errors"
	"sync"

	"github.com/Permify/permify/internal/invoke"
	"github.com/Permify/permify/internal/schema"
	"github.com/Permify/permify/internal/storage"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

type LookupEngine struct {
	// schemaReader is responsible for reading schema information
	schemaReader storage.SchemaReader
	// schemaReader is responsible for reading data
	dataReader storage.DataReader
	// checkEngine is responsible for performing permission checks
	checkEngine invoke.Check
	// schemaMap is a map that keeps track of schema versions
	schemaMap sync.Map
	// concurrencyLimit is the maximum number of concurrent permission checks allowed
	concurrencyLimit int
}

func NewLookupEngine(
	check invoke.Check,
	schemaReader storage.SchemaReader,
	dataReader storage.DataReader,
	opts ...LookupOption,
) *LookupEngine {
	engine := &LookupEngine{
		schemaReader:     schemaReader,
		checkEngine:      check,
		dataReader:       dataReader,
		schemaMap:        sync.Map{},
		concurrencyLimit: _defaultConcurrencyLimit,
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
	var ct string

	size := request.GetPageSize()
	if size == 0 {
		size = 1000
	}

	// Callback function which is called for each entity. If the entity passes the permission check,
	// the entity ID is appended to the entityIDs slice.
	callback := func(entityID, permission, token string) {
		mu.Lock()         // Safeguard access to the shared slice with a mutex
		defer mu.Unlock() // Ensure the lock is released after appending the ID
		if _, exists := entityIDsByPermission[permission]; !exists {
			// If not, initialize it with an empty EntityIds struct
			entityIDsByPermission[permission] = &base.EntityIds{Ids: []string{}}
		}
		entityIDsByPermission[permission].Ids = append(entityIDsByPermission[permission].Ids, entityID)
		ct = token
	}

	// Create and start BulkChecker. It performs permission checks in parallel.
	checker := NewBulkChecker(ctx, engine.checkEngine, BULK_ENTITY, callback, engine.concurrencyLimit)

	// Create and start BulkPublisher. It receives entities and passes them to BulkChecker.
	publisher := NewBulkEntityPublisher(ctx, ConvertToPermissionsLookupEntityRequest(request), checker)

	// Retrieve the schema of the entity based on the tenantId and schema version
	var sc *base.SchemaDefinition
	sc, err = engine.readSchema(ctx, request.GetTenantId(), request.GetMetadata().GetSchemaVersion())
	if err != nil {
		return nil, err
	}

	// Create a map to keep track of visited entities
	visits := &VisitsMap{}

	permissionChecks := &VisitsMap{}

	// Perform an entity filter operation based on the permission request
	err = NewEntityFilter(engine.dataReader, sc).EntityFilter(ctx, &base.PermissionEntityFilterRequest{
		TenantId: request.GetTenantId(),
		Metadata: &base.PermissionEntityFilterRequestMetadata{
			SnapToken:     request.GetMetadata().GetSnapToken(),
			SchemaVersion: request.GetMetadata().GetSchemaVersion(),
			Depth:         request.GetMetadata().GetDepth(),
		},
		Entrances: []*base.Entrance{
			{
				Type:  request.GetEntityType(),
				Value: request.GetPermission(),
			},
		},
		Subject: request.GetSubject(),
		Context: request.GetContext(),
		Scope:   request.GetScope(),
		Cursor:  request.GetContinuousToken(),
	}, visits, publisher, permissionChecks)
	if err != nil {
		return nil, err
	}

	// At this point, the BulkChecker has collected and sorted requests
	err = checker.ExecuteRequests(size) // Execute the collected requests in parallel
	if err != nil {
		return nil, err
	}

	// Return response containing allowed entity IDs
	return &base.PermissionLookupEntityResponse{
		EntityIds:       entityIDs,
		ContinuousToken: ct,
	}, nil
}

// LookupEntity performs a permission check on a set of entities and returns a response
// containing the IDs of the entities that have the requested permission.
func (engine *LookupEngine) LookupEntities(ctx context.Context, request *base.PermissionsLookupEntityRequest) (response *base.PermissionsLookupEntityResponse, err error) {
	// A mutex and slice are declared to safely store entity IDs from concurrent callbacks
	var mu sync.Mutex
	entityIDsByPermission := make(map[string]*base.EntityIds)
	var ct string

	size := request.GetPageSize()
	if size == 0 {
		size = 1000
	}

	// Callback function which is called for each entity. If the entity passes the permission check,
	// the entity ID is appended to the entityIDs slice.
	callback := func(entityID, permission, token string) {
		mu.Lock()         // Safeguard access to the shared slice with a mutex
		defer mu.Unlock() // Ensure the lock is released after appending the ID
		if _, exists := entityIDsByPermission[permission]; !exists {
			// If not, initialize it with an empty EntityIds struct
			entityIDsByPermission[permission] = &base.EntityIds{Ids: []string{}}
		}
		entityIDsByPermission[permission].Ids = append(entityIDsByPermission[permission].Ids, entityID)
		ct = token
	}

	// Create and start BulkChecker. It performs permission checks in parallel.
	checker := NewBulkChecker(ctx, engine.checkEngine, BULK_ENTITY, callback, engine.concurrencyLimit)

	// Create and start BulkPublisher. It receives entities and passes them to BulkChecker.
	publisher := NewBulkEntityPublisher(ctx, request, checker)

	// Retrieve the schema of the entity based on the tenantId and schema version
	var sc *base.SchemaDefinition
	sc, err = engine.readSchema(ctx, request.GetTenantId(), request.GetMetadata().GetSchemaVersion())
	if err != nil {
		return nil, err
	}

	// Create a map to keep track of visited entities
	visits := &VisitsMap{}
	permissionChecks := &VisitsMap{}

	entrances := make([]*base.Entrance, 0)

	for _, permission := range request.GetPermissions() {
		entrances = append(entrances, &base.Entrance{
			Type:  request.GetEntityType(),
			Value: permission,
		})
	}

	// Perform an entity filter operation based on the permission request
	err = NewEntityFilter(engine.dataReader, sc).EntityFilter(ctx, &base.PermissionEntityFilterRequest{
		TenantId: request.GetTenantId(),
		Metadata: &base.PermissionEntityFilterRequestMetadata{
			SnapToken:     request.GetMetadata().GetSnapToken(),
			SchemaVersion: request.GetMetadata().GetSchemaVersion(),
			Depth:         request.GetMetadata().GetDepth(),
		},
		Entrances: entrances,
		Subject:   request.GetSubject(),
		Context:   request.GetContext(),
		Scope:     request.GetScope(),
		Cursor:    request.GetContinuousToken(),
	}, visits, publisher, permissionChecks)
	if err != nil {
		return nil, err
	}

	// At this point, the BulkChecker has collected and sorted requests
	err = checker.ExecuteRequests(size) // Execute the collected requests in parallel
	if err != nil {
		return nil, err
	}

	// Return response containing allowed entity IDs
	return &base.PermissionsLookupEntityResponse{
		EntityIds:       entityIDsByPermission,
		ContinuousToken: ct,
	}, nil
}

// LookupEntityStream performs a permission check on a set of entities and streams the results
// containing the IDs of the entities that have the requested permission.
func (engine *LookupEngine) LookupEntityStream(ctx context.Context, request *base.PermissionLookupEntityRequest, server base.Permission_LookupEntityStreamServer) (err error) {
	size := request.GetPageSize()
	if size == 0 {
		size = 1000
	}

	// Define a callback function that will be called for each entity that passes the permission check.
	// If the check result is allowed, it sends the entity ID to the server stream.
	callback := func(entityID, permission, token string) {
		err := server.Send(&base.PermissionLookupEntityStreamResponse{
			EntityId:        entityID,
			Permission:      permission,
			ContinuousToken: token,
		})
		// If there is an error in sending the response, the function will return
		if err != nil {
			return
		}
	}

	// Create and start BulkChecker. It performs permission checks concurrently.
	checker := NewBulkChecker(ctx, engine.checkEngine, BULK_ENTITY, callback, engine.concurrencyLimit)

	// Create and start BulkPublisher. It receives entities and passes them to BulkChecker.
	publisher := NewBulkEntityPublisher(ctx, ConvertToPermissionsLookupEntityRequest(request), checker)

	// Retrieve the entity definition schema based on the tenantId and schema version
	var sc *base.SchemaDefinition
	sc, err = engine.readSchema(ctx, request.GetTenantId(), request.GetMetadata().GetSchemaVersion())
	if err != nil {
		return err
	}

	visits := &VisitsMap{}
	permissionChecks := &VisitsMap{}

	// Perform an entity filter operation based on the permission request
	err = NewEntityFilter(engine.dataReader, sc).EntityFilter(ctx, &base.PermissionEntityFilterRequest{
		TenantId: request.GetTenantId(),
		Metadata: &base.PermissionEntityFilterRequestMetadata{
			SnapToken:     request.GetMetadata().GetSnapToken(),
			SchemaVersion: request.GetMetadata().GetSchemaVersion(),
			Depth:         request.GetMetadata().GetDepth(),
		},
		Entrances: []*base.Entrance{
			{
				Type:  request.GetEntityType(),
				Value: request.GetPermission(),
			},
		},
		Subject: request.GetSubject(),
		Context: request.GetContext(),
		Cursor:  request.GetContinuousToken(),
	}, visits, publisher, permissionChecks)
	if err != nil {
		return err
	}

	err = checker.ExecuteRequests(size)
	if err != nil {
		return err
	}

	return nil
}

// LookupEntityStream performs a permission check on a set of entities and streams the results
// containing the IDs of the entities that have the requested permission.
func (engine *LookupEngine) LookupEntitiesStream(ctx context.Context, request *base.PermissionsLookupEntityRequest, server base.Permission_LookupEntitiesStreamServer) (err error) {
	size := request.GetPageSize()
	if size == 0 {
		size = 1000
	}

	// Define a callback function that will be called for each entity that passes the permission check.
	// If the check result is allowed, it sends the entity ID to the server stream.
	callback := func(entityID, permission, token string) {
		err := server.Send(&base.PermissionsLookupEntityStreamResponse{
			EntityId:        entityID,
			Permission:      permission,
			ContinuousToken: token,
		})
		// If there is an error in sending the response, the function will return
		if err != nil {
			return
		}
	}

	// Create and start BulkChecker. It performs permission checks concurrently.
	checker := NewBulkChecker(ctx, engine.checkEngine, BULK_ENTITY, callback, engine.concurrencyLimit)

	// Create and start BulkPublisher. It receives entities and passes them to BulkChecker.
	publisher := NewBulkEntityPublisher(ctx, request, checker)

	// Retrieve the entity definition schema based on the tenantId and schema version
	var sc *base.SchemaDefinition
	sc, err = engine.readSchema(ctx, request.GetTenantId(), request.GetMetadata().GetSchemaVersion())
	if err != nil {
		return err
	}

	visits := &VisitsMap{}
	permissionChecks := &VisitsMap{}

	entrances := make([]*base.Entrance, 0)

	for _, permission := range request.GetPermissions() {
		entrances = append(entrances, &base.Entrance{
			Type:  request.GetEntityType(),
			Value: permission,
		})
	}

	// Perform an entity filter operation based on the permission request
	err = NewEntityFilter(engine.dataReader, sc).EntityFilter(ctx, &base.PermissionEntityFilterRequest{
		TenantId: request.GetTenantId(),
		Metadata: &base.PermissionEntityFilterRequestMetadata{
			SnapToken:     request.GetMetadata().GetSnapToken(),
			SchemaVersion: request.GetMetadata().GetSchemaVersion(),
			Depth:         request.GetMetadata().GetDepth(),
		},
		Entrances: entrances,
		Subject:   request.GetSubject(),
		Context:   request.GetContext(),
		Cursor:    request.GetContinuousToken(),
	}, visits, publisher, permissionChecks)
	if err != nil {
		return err
	}

	err = checker.ExecuteRequests(size)
	if err != nil {
		return err
	}

	return nil
}

// LookupSubject checks if a subject has a particular permission based on the schema and version.
// It returns a list of subjects that have the given permission.
func (engine *LookupEngine) LookupSubject(ctx context.Context, request *base.PermissionLookupSubjectRequest) (response *base.PermissionLookupSubjectResponse, err error) {
	size := request.GetPageSize()
	if size == 0 {
		size = 1000
	}

	// Use a mutex to protect concurrent writes to the subjectIDs slice.
	var mu sync.Mutex
	var subjectIDs []string
	var ct string

	// Callback function to handle the results of permission checks.
	// If an entity passes the permission check, its ID is stored in the subjectIDs slice.
	callback := func(subjectID, permission, token string) {
		mu.Lock()         // Lock to prevent concurrent modification of the slice.
		defer mu.Unlock() // Unlock after the ID is appended.
		subjectIDs = append(subjectIDs, subjectID)
		ct = token
	}

	// Create and initiate a BulkChecker to perform permission checks in parallel.
	checker := NewBulkChecker(ctx, engine.checkEngine, BULK_SUBJECT, callback, engine.concurrencyLimit)

	// Create and start a BulkPublisher to provide entities to the BulkChecker.
	publisher := NewBulkSubjectPublisher(ctx, request, checker)

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
			err = NewMassSubjectFilter(engine.dataReader).SubjectFilter(ctx, request, publisher)
			if err != nil {
				// Return an error if there was an issue with the subject filter.
				return nil, err
			}
		} else { // For other errors, simply return the error
			return nil, err
		}
	} else {
		// Use the schema-based subject filter to get the list of subjects with the requested permission.
		ids, err := NewSchemaBasedSubjectFilter(engine.schemaReader, engine.dataReader, SchemaBaseSubjectFilterConcurrencyLimit(engine.concurrencyLimit)).SubjectFilter(ctx, request)
		if err != nil {
			return nil, err
		}

		for _, id := range ids {
			publisher.Publish(&base.Subject{
				Type:     request.GetSubjectReference().GetType(),
				Id:       id,
				Relation: request.GetSubjectReference().GetRelation(),
			}, &base.PermissionCheckRequestMetadata{
				SnapToken:     request.GetMetadata().GetSnapToken(),
				SchemaVersion: request.GetMetadata().GetSchemaVersion(),
				Depth:         request.GetMetadata().GetDepth(),
			}, request.GetContext(), base.CheckResult_CHECK_RESULT_ALLOWED)
		}
	}

	err = checker.ExecuteRequests(size)
	if err != nil {
		// Return an error if there was an issue with the subject filter.
		return nil, err
	}

	// Return the list of entity IDs that have the required permission.
	return &base.PermissionLookupSubjectResponse{
		SubjectIds:      subjectIDs,
		ContinuousToken: ct,
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
