package keys

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/cespare/xxhash/v2"

	"github.com/Permify/permify/internal/invoke"
	"github.com/Permify/permify/pkg/cache"
	"github.com/Permify/permify/pkg/logger"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/tuple"
)

// CheckEngineWithKeys is a struct that holds an instance of a cache.Cache for managing engine keys.
type CheckEngineWithKeys struct {
	checker invoke.Check
	cache   cache.Cache
	l       *logger.Logger
}

// NewCheckEngineWithKeys creates a new instance of EngineKeyManager by initializing an EngineKeys
// struct with the provided cache.Cache instance.
func NewCheckEngineWithKeys(checker invoke.Check, cache cache.Cache, l *logger.Logger) invoke.Check {
	return &CheckEngineWithKeys{
		checker: checker,
		cache:   cache,
		l:       l,
	}
}

// Check invokes the permission check on the engine with the given
func (c *CheckEngineWithKeys) Check(ctx context.Context, request *base.PermissionCheckRequest) (response *base.PermissionCheckResponse, err error) {
	res, found := c.getCheckKey(request)
	if found {
		if request.GetMetadata().GetExclusion() {
			if res.GetCan() == base.PermissionCheckResponse_RESULT_ALLOWED {
				return &base.PermissionCheckResponse{
					Can: base.PermissionCheckResponse_RESULT_DENIED,
					Metadata: &base.PermissionCheckResponseMetadata{
						CheckCount: 0,
					},
				}, nil
			}
			return &base.PermissionCheckResponse{
				Can: base.PermissionCheckResponse_RESULT_ALLOWED,
				Metadata: &base.PermissionCheckResponseMetadata{
					CheckCount: 0,
				},
			}, nil
		}
		return &base.PermissionCheckResponse{
			Can:      res.GetCan(),
			Metadata: &base.PermissionCheckResponseMetadata{},
		}, nil
	}

	res, err = c.checker.Check(ctx, request)
	if err != nil {
		return &base.PermissionCheckResponse{
			Can: base.PermissionCheckResponse_RESULT_ALLOWED,
			Metadata: &base.PermissionCheckResponseMetadata{
				CheckCount: 0,
			},
		}, err
	}

	c.setCheckKey(request, &base.PermissionCheckResponse{
		Can:      res.GetCan(),
		Metadata: &base.PermissionCheckResponseMetadata{},
	})

	return res, err
}

// GetCheckKey retrieves the value for the given key from the EngineKeys cache.
// It returns the PermissionCheckResponse if the key is found, and a boolean value
// indicating whether the key was found or not.
func (c *CheckEngineWithKeys) getCheckKey(key *base.PermissionCheckRequest) (*base.PermissionCheckResponse, bool) {
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

func (c *CheckEngineWithKeys) setCheckKey(key *base.PermissionCheckRequest, value *base.PermissionCheckResponse) bool {
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
