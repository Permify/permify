package keys

import (
	"encoding/hex"
	"fmt"

	"github.com/cespare/xxhash"

	"github.com/Permify/permify/pkg/cache"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/tuple"
)

type CommandKeys struct {
	cache cache.Cache
}

// NewCheckCommandKeys new instance of CheckCommandKeys
func NewCheckCommandKeys(cache cache.Cache) *CommandKeys {
	return &CommandKeys{
		cache: cache,
	}
}

// SetCheckKey sets the value for the given key.
func (c *CommandKeys) SetCheckKey(key *base.PermissionCheckRequest, value *base.PermissionCheckResponse) bool {
	checkKey := fmt.Sprintf("cc_%s:%s:%s@%s", key.GetSchemaVersion(), key.GetSnapToken(), tuple.EntityAndRelationToString(&base.EntityAndRelation{
		Entity:   key.GetEntity(),
		Relation: key.GetAction(),
	}), tuple.SubjectToString(key.GetSubject()))
	h := xxhash.New()
	size, err := h.Write([]byte(checkKey))
	if err != nil {
		return false
	}
	k := hex.EncodeToString(h.Sum(nil))
	return c.cache.Set(k, value, int64(size))
}

// GetCheckKey gets the value for the given key.
func (c *CommandKeys) GetCheckKey(key *base.PermissionCheckRequest) (*base.PermissionCheckResponse, bool) {
	checkKey := fmt.Sprintf("cc_%s:%s:%s@%s", key.GetSchemaVersion(), key.GetSnapToken(), tuple.EntityAndRelationToString(&base.EntityAndRelation{
		Entity:   key.GetEntity(),
		Relation: key.GetAction(),
	}), tuple.SubjectToString(key.GetSubject()))
	h := xxhash.New()
	_, err := h.Write([]byte(checkKey))
	if err != nil {
		return nil, false
	}
	k := hex.EncodeToString(h.Sum(nil))
	resp, ok := c.cache.Get(k)
	if ok {
		return resp.(*base.PermissionCheckResponse), true
	}
	return nil, false
}
