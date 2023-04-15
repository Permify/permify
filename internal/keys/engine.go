package keys

import (
	"encoding/hex"
	"fmt"
	"github.com/Permify/permify/internal/config"
	"github.com/Permify/permify/pkg/consistent"
	"github.com/Permify/permify/pkg/gossip"
	"github.com/cespare/xxhash/v2"

	"github.com/Permify/permify/pkg/cache"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/tuple"
)

// EngineKeys is a struct that holds an instance of a cache.Cache for managing engine keys.
type EngineKeys struct {
	keys             []uint64
	cache            cache.Cache
	gossip           *gossip.Engine
	consistent       *hash.ConsistentHash
	localNodeAddress string
}

// NewCheckEngineKeys creates a new instance of EngineKeyManager by initializing an EngineKeys
// struct with the provided cache.Cache instance.
func NewCheckEngineKeys(cache cache.Cache, gossip *gossip.Engine, cfg config.Server) EngineKeyManager {

	// Initialize a new instance of ConsistentHash with 100 replicas
	consistent := hash.NewConsistentHash(100, nil)
	consistent.Add(cfg.Address + ":" + cfg.HTTP.Port)

	// Return a new instance of EngineKeys with the provided cache
	return &EngineKeys{
		localNodeAddress: cfg.Address + ":" + cfg.HTTP.Port,
		gossip:           gossip,
		consistent:       consistent,
		cache:            cache,
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

	if c.gossip.Enabled {

		// Use Consistent Hashing to find the responsible node for the given key
		node, found := c.consistent.Get(checkKey)
		if !found {
			// If the responsible node is not found, return false
			return false
		}

		// Check if the node is the local node
		if node == c.localNodeAddress {
			// Set the cache key with the given value and size, then return the result
			return c.cache.Set(k, value.Can, int64(size))
		} else {
			// if node is not local, forward the request to the responsible node
			return true
		}
	} else {
		// Set the cache key with the given value and size, then return the result
		return c.cache.Set(k, value.Can, int64(size))
	}
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

	if c.gossip.Enabled {
		// Find the responsible node for the given key
		node, ok := c.consistent.Get(k)
		if !ok {
			// If the node is not found, return nil and false
			return nil, false
		}

		// Check if the responsible node is the local node
		if node == c.localNodeAddress {
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
		} else {
			// if node is not local, forward the request to the responsible node
		}
	} else {
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
	}

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
