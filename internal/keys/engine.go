package keys

import (
	"encoding/hex"
	"fmt"

	"github.com/cespare/xxhash"

	"github.com/Permify/permify/pkg/cache"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/tuple"
)

type EngineKeys struct {
	cache cache.Cache
}

// NewCheckEngineKeys new instance of CheckEngineKeys
func NewCheckEngineKeys(cache cache.Cache) EngineKeyManager {
	return &EngineKeys{
		cache: cache,
	}
}

// SetCheckKey - Sets the value for the given key.
func (c *EngineKeys) SetCheckKey(key *base.PermissionCheckRequest, value *base.PermissionCheckResponse) bool {
	checkKey := fmt.Sprintf("check_%s_%s:%s:%s@%s", key.GetTenantId(), key.GetMetadata().GetSchemaVersion(), key.GetMetadata().GetSnapToken(), tuple.EntityAndRelationToString(&base.EntityAndRelation{
		Entity:   key.GetEntity(),
		Relation: key.GetPermission(),
	}), tuple.SubjectToString(key.GetSubject()))
	h := xxhash.New()
	size, err := h.Write([]byte(checkKey))
	if err != nil {
		return false
	}
	k := hex.EncodeToString(h.Sum(nil))
	return c.cache.Set(k, value, int64(size))
}

// GetCheckKey - Gets the value for the given key.
func (c *EngineKeys) GetCheckKey(key *base.PermissionCheckRequest) (*base.PermissionCheckResponse, bool) {
	checkKey := fmt.Sprintf("check_%s_%s:%s:%s@%s", key.GetTenantId(), key.GetMetadata().GetSchemaVersion(), key.GetMetadata().GetSnapToken(), tuple.EntityAndRelationToString(&base.EntityAndRelation{
		Entity:   key.GetEntity(),
		Relation: key.GetPermission(),
	}), tuple.SubjectToString(key.GetSubject()))
	h := xxhash.New()
	_, err := h.Write([]byte(checkKey))
	if err != nil {
		return nil, false
	}
	k := hex.EncodeToString(h.Sum(nil))
	resp, found := c.cache.Get(k)
	if found {
		return resp.(*base.PermissionCheckResponse), true
	}
	return nil, false
}

// NoopEngineKeys -
type NoopEngineKeys struct{}

// NewNoopCheckEngineKeys new noop instance of CheckEngineKeys
func NewNoopCheckEngineKeys() EngineKeyManager {
	return &NoopEngineKeys{}
}

// SetCheckKey sets the value for the given key.
func (c *NoopEngineKeys) SetCheckKey(*base.PermissionCheckRequest, *base.PermissionCheckResponse) bool {
	return true
}

// GetCheckKey gets the value for the given key.
func (c *NoopEngineKeys) GetCheckKey(*base.PermissionCheckRequest) (*base.PermissionCheckResponse, bool) {
	return nil, false
}
