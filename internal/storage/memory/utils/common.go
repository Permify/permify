package utils

import (
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

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
