package utils

import (
	"sync"

	base "github.com/Permify/permify/pkg/pb/base/v1"
)

var (
	RelationTuplesID AutoIncForRelationTuples
	AttributesID     AutoIncForRelationTuples
)

type AutoIncForRelationTuples struct {
	sync.Mutex
	id uint64
}

// ID -
func (a *AutoIncForRelationTuples) ID() (id uint64) {
	a.Lock()
	defer a.Unlock()
	if a.id == 0 {
		a.id++
	}
	id = a.id
	a.id++
	return
}

type AutoIncForAttributes struct {
	sync.Mutex
	id uint64
}

// ID -
func (a *AutoIncForAttributes) ID() (id uint64) {
	a.Lock()
	defer a.Unlock()
	if a.id == 0 {
		a.id++
	}
	id = a.id
	a.id++
	return
}

// GetRelationTuplesIndexNameAndArgsByFilters - Get index name and arguments by filters
func GetRelationTuplesIndexNameAndArgsByFilters(tenantID string, filter *base.TupleFilter) (string, []any) {
	if filter.GetEntity().GetType() != "" && filter.GetRelation() != "" {
		return "entity-type-and-relation-index", []any{tenantID, filter.GetEntity().GetType(), filter.GetRelation()}
	}
	if filter.GetEntity().GetType() != "" {
		return "entity-type-index", []any{tenantID, filter.GetEntity().GetType()}
	}
	return "id", nil
}

func GetAttributesIndexNameAndArgsByFilters(tenantID string, filter *base.AttributeFilter) (string, []any) {
	if filter.GetEntity().GetType() != "" {
		return "entity-type-index", []any{tenantID, filter.GetEntity().GetType()}
	}
	return "id", nil
}
