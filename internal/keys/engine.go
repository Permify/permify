package keys

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Permify/permify/internal/config"
	"github.com/Permify/permify/pkg/cache"
	"github.com/Permify/permify/pkg/consistent"
	"github.com/Permify/permify/pkg/gossip"
	"github.com/Permify/permify/pkg/logger"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/tuple"
	"github.com/cespare/xxhash/v2"
	"github.com/hashicorp/memberlist"
	"net"
	"net/http"
	"time"
)

// EngineKeys is a struct that holds an instance of a cache.Cache for managing engine keys.
type EngineKeys struct {
	distribution     bool
	keys             []uint64
	cache            cache.Cache
	gossip           gossip.IGossip
	consistent       hash.ConsistentEngine
	localNodeAddress string
	l                *logger.Logger
}

// NewCheckEngineKeys creates a new instance of EngineKeyManager by initializing an EngineKeys
// struct with the provided cache.Cache instance.
func NewCheckEngineKeys(cache cache.Cache, consistent *hash.ConsistentHash, gossip *gossip.Engine, cfg config.Server, l *logger.Logger, distributed bool) *EngineKeys {
	// Return a new instance of EngineKeys with the provided cache
	return &EngineKeys{
		localNodeAddress: ExternalIP() + ":" + cfg.HTTP.Port,
		gossip:           gossip,
		consistent:       consistent,
		cache:            cache,
		distribution:     distributed,
		l:                l,
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
		c.l.Error("error writing to hash object: %v", err)
		// If there's an error, return false
		return false
	}

	// Generate the final cache key by encoding the hash object's sum as a hexadecimal string
	k := hex.EncodeToString(h.Sum(nil))

	if c.distribution {
		ok := c.consistent.AddKey(k)
		if !ok {
			c.l.Error("error adding key %s to consistent hash", k)
			// If there's an error, return false
			return false
		}
		c.l.Info("added key %s to consistent hash", k)

		// Use Consistent Hashing to find the responsible node for the given key
		node, found := c.consistent.Get(k)
		if !found {
			c.l.Error("node not found for key %s", k)
			// If the responsible node is not found, return false
			return false
		}
		c.l.Info("node %s found for key %s", node, k)

		// Check if the node is the local node
		if node == c.localNodeAddress {
			// Set the cache key with the given value and size, then return the result
			return c.cache.Set(k, value.Can, int64(size))
		} else {
			_, err := c.forwardRequestSetToNode(node, value, key)
			if err != nil {
				c.l.Error("error forwarding request to node %s: %s", node, err.Error())
				return false
			}

			c.l.Info("forwarded request to node %s , %s", node, k)

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

	if c.distribution {
		// Find the responsible node for the given key
		node, ok := c.consistent.Get(k)
		if !ok {
			c.l.Error("node not found for key %s", k)
			// If the node is not found, return nil and false
			return nil, false
		}

		c.l.Info("node %s found for key %s", node, k)

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
			toNode, err := c.forwardRequestGetToNode(node, key)
			if err != nil {
				c.l.Error("error forwarding request to node %s: %s", node, err)
				return nil, false
			}

			c.l.Info("forwarded Get request to node %s , %s", node, k)

			return toNode.PermissionCheckResponse, true
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

func (c *EngineKeys) SetKey(key *base.PermissionCheckRequest, value *base.PermissionCheckResponse) bool {
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

func (c *EngineKeys) GetKey(key *base.PermissionCheckRequest) (*base.PermissionCheckResponse, bool) {
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

func (c *EngineKeys) forwardRequestSetToNode(node string, value *base.PermissionCheckResponse, key *base.PermissionCheckRequest) (bool, error) {
	// Create a new HTTP client with a timeout
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Prepare the payload to be sent in the request body
	payload := struct {
		Key   *base.PermissionCheckRequest  `json:"permission_check_request"`
		Value *base.PermissionCheckResponse `json:"permission_check_response"`
	}{
		Value: value,
		Key:   key,
	}

	// Encode the payload as JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {

		return false, err
	}

	// Create a new HTTP request to the responsible node
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s/v1/consistent/set", node), bytes.NewBuffer(payloadBytes))
	if err != nil {
		c.l.Error("failed to set check key on the responsible node: %s", err.Error())
		return false, err
	}

	// Set the request content type to JSON
	req.Header.Set("Content-Type", "application/json")

	// Send the request to the responsible node
	resp, err := client.Do(req)
	if err != nil {
		c.l.Error("failed to set check key on the responsible node: %s", err.Error())
		return false, err
	}

	// Close the response body when the function returns
	defer resp.Body.Close()

	// Check if the response status code is 200 OK
	if resp.StatusCode == http.StatusOK {
		return true, nil
	} else {
		return false, errors.New("failed to set check key on the responsible node")
	}
}

func (c *EngineKeys) forwardRequestGetToNode(node string, checkKey *base.PermissionCheckRequest) (*base.ConsistentGetResponse, error) {
	// Create a new HTTP client with a timeout
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Prepare the payload to be sent in the request body
	payload := struct {
		CheckKey *base.PermissionCheckRequest `json:"permission_check_request"`
	}{
		CheckKey: checkKey,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	// Create a new HTTP request to the responsible node
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s/v1/consistent/get", node), bytes.NewBuffer(payloadBytes))
	if err != nil {
		c.l.Error("failed to set check key on the responsible node: %s", err.Error())
		return nil, err
	}

	// Set the request content type to JSON
	req.Header.Set("Content-Type", "application/json")

	// Send the request to the responsible node
	resp, err := client.Do(req)
	if err != nil {
		c.l.Error("failed to set check key on the responsible node: %s", err.Error())
		return nil, err
	}

	// unmarshall body
	var body base.ConsistentGetResponse
	err = json.NewDecoder(resp.Body).Decode(&body)
	if err != nil {
		c.l.Error("failed to set check key on the responsible node: %s", err.Error())
		return nil, err
	}

	// Close the response body when the function returns
	defer resp.Body.Close()

	// Check if the response status code is 200 OK
	if resp.StatusCode == http.StatusOK {
		return &body, nil
	} else {
		return nil, errors.New("failed to set check key on the responsible node")
	}
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

func (c *NoopEngineKeys) SetKey(*base.PermissionCheckRequest, *base.PermissionCheckResponse) bool {
	return true
}

func (c *NoopEngineKeys) GetKey(*base.PermissionCheckRequest) (*base.PermissionCheckResponse, bool) {
	return nil, false
}

// SyncPeers is a no-op method that implements the SyncPeers method for the
// EngineKeyManager interface. It does nothing, as it performs no actual caching
// or operations.
func (c *NoopEngineKeys) SyncPeers(*memberlist.Memberlist) {}

func ExternalIP() string {
	faces, err := net.Interfaces()
	if err != nil {
		return ""
	}
	for _, iface := range faces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		address, err := iface.Addrs()
		if err != nil {
			return ""
		}
		for _, addr := range address {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip.String()
		}
	}
	return ""
}
