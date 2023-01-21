package utils

import (
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// GetIndexNameAndArgsByFilters - Get index name and arguments by filters
func GetIndexNameAndArgsByFilters(tenantID string, filter *base.TupleFilter) (string, []any) {
	if filter.GetEntity().GetType() != "" && filter.GetRelation() != "" {
		return "entity-type-and-relation-index", []any{tenantID, filter.GetEntity().GetType(), filter.GetRelation()}
	}
	if filter.GetEntity().GetType() != "" {
		return "entity-type-index", []any{tenantID, filter.GetEntity().GetType()}
	}
	return "id", nil
}
