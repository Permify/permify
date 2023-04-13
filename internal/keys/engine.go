package keys

import (
	"encoding/hex"
	"fmt"
	"github.com/buraksezer/consistent"

	"github.com/cespare/xxhash/v2"

	"github.com/Permify/permify/pkg/cache"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/tuple"
)

// EngineKeys is a struct that holds an instance of a cache.Cache for managing engine keys.
type EngineKeys struct {
	keys             []uint64
	cache            cache.Cache
	consistent       *consistent.Consistent
	localNodeAddress string
}

type hashed struct{}

// NewCheckEngineKeys creates a new instance of EngineKeyManager by initializing an EngineKeys
// struct with the provided cache.Cache instance.
func NewCheckEngineKeys(cache cache.Cache) EngineKeyManager {
	// Create a new consistent instance
	cfg := consistent.Config{
		PartitionCount:    7,
		ReplicationFactor: 20,
		Load:              1.25,
		Hasher:            hashed{},
	}

	// Return a new instance of EngineKeys with the provided cache
	return &EngineKeys{
		consistent: consistent.New(nil, cfg),
		cache:      cache,
	}
}

// SetCheckKey sets the value for the given key in the EngineKeys cache.
// It returns true if the operation is successful, false otherwise.
func (c *EngineKeys) SetCheckKey(key *base.PermissionCheckRequest, value *base.PermissionCheckResponse) bool {
	if key == nil || value == nil {
		// If either the key or value is nil, return false
		return false
	}

	// Generate a unique checkKey string based on the provided PermissionCheckRequest
	checkKey := fmt.Sprintf("check_%s_%s:%s:%s@%s", key.GetTenantId(), key.GetMetadata().GetSchemaVersion(), key.GetMetadata().GetSnapToken(), tuple.EntityAndRelationToString(&base.EntityAndRelation{
		Entity:   key.GetEntity(),
		Relation: key.GetPermission(),
	}), tuple.SubjectToString(key.GetSubject()))

	// Initialize a new xxhash object
	h := xxhash.New()

	// Write the checkKey string to the hash object
	size, err := h.Write([]byte(checkKey))
	if err != nil {
		// If there's an error, return false
		return false
	}

	// Generate the final cache key by encoding the hash object's sum as a hexadecimal string
	k := hex.EncodeToString(h.Sum(nil))

	// Set the cache key with the given value and size, then return the result
	return c.cache.Set(k, value.Can, int64(size))
}

// GetCheckKey retrieves the value for the given key from the EngineKeys cache.
// It returns the PermissionCheckResponse if the key is found, and a boolean value
// indicating whether the key was found or not.
func (c *EngineKeys) GetCheckKey(key *base.PermissionCheckRequest) (*base.PermissionCheckResponse, bool) {

	if key == nil {
		// If either the key or value is nil, return false
		return nil, false
	}

	// Generate a unique checkKey string based on the provided PermissionCheckRequest
	checkKey := fmt.Sprintf("check_%s_%s:%s:%s@%s", key.GetTenantId(), key.GetMetadata().GetSchemaVersion(), key.GetMetadata().GetSnapToken(), tuple.EntityAndRelationToString(&base.EntityAndRelation{
		Entity:   key.GetEntity(),
		Relation: key.GetPermission(),
	}), tuple.SubjectToString(key.GetSubject()))

	// Initialize a new xxhash object
	h := xxhash.New()

	// Write the checkKey string to the hash object
	_, err := h.Write([]byte(checkKey))
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
			Can: resp.(base.PermissionCheckResponse_Result),
			Metadata: &base.PermissionCheckResponseMetadata{
				CheckCount: 0,
			},
		}, true
	}

	// If the key is not found, return nil and false
	return nil, false
}

// NoopEngineKeys is an empty struct that implements the EngineKeyManager interface
// with no-op (no operation) methods, meaning they do not perform any real work or caching.
type NoopEngineKeys struct{}

// NewNoopCheckEngineKeys creates a new noop instance of EngineKeyManager by
// initializing a NoopEngineKeys struct. This instance will not perform any caching.
func NewNoopCheckEngineKeys() EngineKeyManager {
	// Return a new instance of NoopEngineKeys
	return &NoopEngineKeys{}
}

// SetCheckKey is a no-op method that implements the SetCheckKey method for the
// EngineKeyManager interface. It always returns true, indicating success, but
// performs no actual caching or operations.
func (c *NoopEngineKeys) SetCheckKey(*base.PermissionCheckRequest, *base.PermissionCheckResponse) bool {
	return true
}

// GetCheckKey is a no-op method that implements the GetCheckKey method for the
// EngineKeyManager interface. It always returns nil and false, indicating that
// the key is not found, as it performs no actual caching or operations.
func (c *NoopEngineKeys) GetCheckKey(*base.PermissionCheckRequest) (*base.PermissionCheckResponse, bool) {
	return nil, false
}

func (h hashed) Sum64(data []byte) uint64 {
	// you should use a proper hash function for uniformity.
	return xxhash.Sum64(data)
}

// AddNode adds a new node to the EngineKeys cache for consistent hashing.
func (c *EngineKeys) AddNode(node string) {
	c.AddNode(node)
}

// RemoveNode removes a node from the EngineKeys cache for consistent hashing.
func (c *EngineKeys) RemoveNode(node string) {
	c.RemoveNode(node)
}
