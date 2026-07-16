package cache

import (
	"context"
	"encoding/hex"

	"go.opentelemetry.io/otel/metric"

	"github.com/cespare/xxhash/v2"

	"github.com/Permify/permify/internal"
	"github.com/Permify/permify/internal/engines"
	"github.com/Permify/permify/internal/invoke"
	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/pkg/cache"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/telemetry"
)

// CheckEngineWithCache is a struct that holds an instance of a cache.Cache for managing engine cache.
type CheckEngineWithCache struct {
	// schemaReader is responsible for reading schema information
	schemaReader storage.SchemaReader
	checker      invoke.Check
	cache        cache.Cache

	// Metrics
	cacheHitHistogram metric.Int64Histogram
}

// NewCheckEngineWithCache creates a new instance of EngineKeyManager by initializing an EngineKeys
// struct with the provided cache.Cache instance.
func NewCheckEngineWithCache(
	checker invoke.Check,
	schemaReader storage.SchemaReader,
	cache cache.Cache,
) invoke.Check {
	return &CheckEngineWithCache{
		schemaReader:      schemaReader,
		checker:           checker,
		cache:             cache,
		cacheHitHistogram: telemetry.NewHistogram(internal.Meter, "cache_hit", "amount", "Number of cache hits"),
	}
}

// Check performs a permission check for a given request, using the cached results if available.
// Supports batch: request.EntityIDs may contain one or more entity IDs.
// For each entity, the cache is checked individually; uncached entities are delegated to the underlying checker.
func (c *CheckEngineWithCache) Check(ctx context.Context, request *invoke.BatchCheckRequest) (response *invoke.BatchCheckResponse, err error) {
	// Read entity definition once (all entities share the same type)
	en, _, err := c.schemaReader.ReadEntityDefinition(ctx, request.TenantID, request.EntityType, request.Metadata.GetSchemaVersion())
	if err != nil {
		return invoke.NewBatchCheckResponse(base.CheckResult_CHECK_RESULT_DENIED, request.EntityIDs...), err
	}

	isRelational := engines.IsRelational(en, request.Permission)

	// Per-entity cached results
	cachedResults := make(map[string]base.CheckResult)

	// Check cache per entity_id, collect uncached IDs
	var uncachedIDs []string
	for _, id := range request.EntityIDs {
		singleReq := request.ToPermissionCheckRequest(id)
		res, found := c.getCheckKey(singleReq, isRelational)
		if found {
			c.cacheHitHistogram.Record(ctx, 1)
			cachedResults[id] = res.GetCan()
			continue
		}
		uncachedIDs = append(uncachedIDs, id)
	}

	// All cached → return cached results
	if len(uncachedIDs) == 0 {
		return &invoke.BatchCheckResponse{
			Results:  cachedResults,
			Metadata: &base.PermissionCheckResponseMetadata{},
		}, nil
	}

	// Delegate uncached to underlying checker
	batchReq := request.Clone()
	batchReq.EntityIDs = uncachedIDs
	cres, err := c.checker.Check(ctx, batchReq)
	if err != nil {
		return invoke.NewBatchCheckResponse(base.CheckResult_CHECK_RESULT_DENIED, request.EntityIDs...), err
	}

	// Cache the per-entity results for each uncached entity
	for _, id := range uncachedIDs {
		singleReq := request.ToPermissionCheckRequest(id)
		result := base.CheckResult_CHECK_RESULT_DENIED
		if r, ok := cres.Results[id]; ok {
			result = r
		}
		c.setCheckKey(singleReq, &base.PermissionCheckResponse{
			Can:      result,
			Metadata: &base.PermissionCheckResponseMetadata{},
		}, isRelational)
	}

	// Merge cached results into the response
	for id, result := range cachedResults {
		cres.Results[id] = result
	}

	return cres, nil
}

// GetCheckKey retrieves the value for the given key from the EngineKeys cache.
// It returns the PermissionCheckResponse if the key is found, and a boolean value
// indicating whether the key was found or not.
func (c *CheckEngineWithCache) getCheckKey(key *base.PermissionCheckRequest, isRelational bool) (*base.PermissionCheckResponse, bool) {
	if key == nil {
		// If either the key or value is nil, return false
		return nil, false
	}

	// Initialize a new xxhash object
	h := xxhash.New()

	// Write the checkKey string to the hash object
	_, err := h.Write([]byte(engines.GenerateKey(key, isRelational)))
	if err != nil {
		// If there's an error, return nil and false
		return nil, false
	}

	// Generate the final cache key by encoding the hash object's sum as a hexadecimal string
	k := hex.EncodeToString(h.Sum(nil))

	// Get the value from the cache using the generated cache key
	resp, found := c.cache.Get(k)

	// If the key is found, return the value and true
	if found {
		// If permission is granted, return allowed response
		return &base.PermissionCheckResponse{
			Can: resp.(base.CheckResult),
			Metadata: &base.PermissionCheckResponseMetadata{
				CheckCount: 0,
			},
		}, true
	}

	// If the key is not found, return nil and false
	return nil, false
}

// setCheckKey is a function to set a check key in the cache of the CheckEngineWithKeys.
// It takes a permission check request as a key, a permission check response as a value,
// and returns a boolean value indicating if the operation was successful.
func (c *CheckEngineWithCache) setCheckKey(key *base.PermissionCheckRequest, value *base.PermissionCheckResponse, isRelational bool) bool {
	// If either the key or the value is nil, return false.
	if key == nil || value == nil {
		return false
	}

	// Create a new xxhash object for hashing.
	h := xxhash.New()

	// Generate a key string from the permission check request and write it to the hash.
	// If there's an error while writing to the hash, return false.
	size, err := h.Write([]byte(engines.GenerateKey(key, isRelational)))
	if err != nil {
		return false
	}

	// Compute the hash sum and encode it as a hexadecimal string.
	k := hex.EncodeToString(h.Sum(nil))

	// Set the hashed key and the check result in the cache, using the size of the hashed key as an expiry.
	// The Set method should return true if the operation was successful, so return the result.
	return c.cache.Set(k, value.GetCan(), int64(size))
}
