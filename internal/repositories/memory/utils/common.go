package utils

import (
	"sync"

	base "github.com/Permify/permify/pkg/pb/base/v1"
)

var (
	TenantsID        AutoIncForTenants
	RelationTuplesID AutoIncForTenants
)

type AutoIncForTenants struct {
	sync.Mutex
	id uint64
}

// ID -
func (a *AutoIncForTenants) ID() (id uint64) {
	a.Lock()
	defer a.Unlock()

	id = a.id
	a.id++
	return
}

type AutoIncForRelationTuples struct {
	sync.Mutex
	id uint64
}

// ID -
func (a *AutoIncForRelationTuples) ID() (id uint64) {
	a.Lock()
	defer a.Unlock()

	id = a.id
	a.id++
	return
}

// GetIndexNameAndArgsByFilters - Get index name and arguments by filters
func GetIndexNameAndArgsByFilters(tenantID uint64, filter *base.TupleFilter) (string, []any) {
	if filter.GetEntity().GetType() != "" && filter.GetRelation() != "" {
		return "entity-type-and-relation-index", []any{tenantID, filter.GetEntity().GetType(), filter.GetRelation()}
	}
	if filter.GetEntity().GetType() != "" {
		return "entity-type-index", []any{tenantID, filter.GetEntity().GetType()}
	}
	return "id", nil
}
