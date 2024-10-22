package engines

import (
	"context"
	"slices"
	"sort"
	"strings"
	"sync"

	"github.com/Permify/permify/internal/invoke"
	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/internal/storage/context/utils"
	"github.com/Permify/permify/pkg/database"
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
	callback := func(entityID, token string) {
		mu.Lock()         // Safeguard access to the shared slice with a mutex
		defer mu.Unlock() // Ensure the lock is released after appending the ID
		entityIDs = append(entityIDs, entityID)
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

	// Perform an entity filter operation based on the permission request
	err = NewEntityFilter(engine.dataReader, sc).EntityFilter(ctx, &base.PermissionEntityFilterRequest{
		TenantId: request.GetTenantId(),
		Metadata: &base.PermissionEntityFilterRequestMetadata{
			SnapToken:     request.GetMetadata().GetSnapToken(),
			SchemaVersion: request.GetMetadata().GetSchemaVersion(),
			Depth:         request.GetMetadata().GetDepth(),
		},
		Entrance: &base.Entrance{
			Type:  request.GetEntityType(),
			Value: request.GetPermission(),
		},
		Subject: request.GetSubject(),
		Context: request.GetContext(),
		Scope:   request.GetScope(),
		Cursor:  request.GetContinuousToken(),
	}, visits, publisher)
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

// LookupEntityStream performs a permission check on a set of entities and streams the results
// containing the IDs of the entities that have the requested permission.
func (engine *LookupEngine) LookupEntityStream(ctx context.Context, request *base.PermissionLookupEntityRequest, server base.Permission_LookupEntityStreamServer) (err error) {
	size := request.GetPageSize()
	if size == 0 {
		size = 1000
	}

	// Define a callback function that will be called for each entity that passes the permission check.
	// If the check result is allowed, it sends the entity ID to the server stream.
	callback := func(entityID, token string) {
		err := server.Send(&base.PermissionLookupEntityStreamResponse{
			EntityId:        entityID,
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

	// Perform an entity filter operation based on the permission request
	err = NewEntityFilter(engine.dataReader, sc).EntityFilter(ctx, &base.PermissionEntityFilterRequest{
		TenantId: request.GetTenantId(),
		Metadata: &base.PermissionEntityFilterRequestMetadata{
			SnapToken:     request.GetMetadata().GetSnapToken(),
			SchemaVersion: request.GetMetadata().GetSchemaVersion(),
			Depth:         request.GetMetadata().GetDepth(),
		},
		Entrance: &base.Entrance{
			Type:  request.GetEntityType(),
			Value: request.GetPermission(),
		},
		Subject: request.GetSubject(),
		Context: request.GetContext(),
		Cursor:  request.GetContinuousToken(),
	}, visits, publisher)
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

	var ids []string
	var ct string

	// Use the schema-based subject filter to get the list of subjects with the requested permission.
	ids, err = NewSubjectFilter(engine.schemaReader, engine.dataReader, SubjectFilterConcurrencyLimit(engine.concurrencyLimit)).SubjectFilter(ctx, request)
	if err != nil {
		return nil, err
	}

	// Initialize excludedIds to be used in the query
	var excludedIds []string

	// Check if the wildcard '<>' is present in the ids.Ids or if it's formatted like "<>-1,2,3"
	for _, id := range ids {
		if id == ALL {
			// Handle '<>' case: no exclusions, include all resources
			excludedIds = nil
			break
		} else if strings.HasPrefix(id, ALL+"-") {
			// Handle '<>-1,2,3' case: parse exclusions after '-'
			excludedIds = strings.Split(strings.TrimPrefix(id, ALL+"-"), ",")
			break
		}
	}

	// If '<>' was found, query all subjects with exclusions if provided
	if excludedIds != nil || slices.Contains(ids, ALL) {
		resp, pct, err := engine.dataReader.QueryUniqueSubjectReferences(
			ctx,
			request.GetTenantId(),
			request.GetSubjectReference(),
			excludedIds, // Pass the exclusions if any
			request.GetMetadata().GetSnapToken(),
			database.NewPagination(database.Size(size), database.Token(request.GetContinuousToken())),
		)
		if err != nil {
			return nil, err
		}
		ct = pct.String()

		// Return the list of entity IDs that have the required permission.
		return &base.PermissionLookupSubjectResponse{
			SubjectIds:      resp,
			ContinuousToken: ct,
		}, nil
	}

	// Sort the IDs
	sort.Strings(ids)

	// Convert page size to int for compatibility with startIndex
	pageSize := int(size)

	// Determine the end index based on the page size and total number of IDs
	end := min(pageSize, len(ids))

	// Generate the next continuous token if there are more results
	if end < len(ids) {
		ct = utils.NewContinuousToken(ids[end]).Encode().String()
	} else {
		ct = ""
	}

	// Return the paginated list of IDs
	return &base.PermissionLookupSubjectResponse{
		SubjectIds:      ids[:end], // Slice the IDs based on pagination
		ContinuousToken: ct,        // Return the next continuous token
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
