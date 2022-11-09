package utils

import (
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// GetIndexNameAndArgsByFilters -
func GetIndexNameAndArgsByFilters(filter *base.TupleFilter) (string, []any) {
	if filter.GetEntity().GetType() != "" && filter.GetRelation() != "" {
		return "entity-type-and-relation-index", []any{filter.GetEntity().GetType(), filter.GetRelation()}
	}
	if filter.GetEntity().GetType() != "" {
		return "entity-type-index", []any{filter.GetEntity().GetType()}
	}
	return "id", nil
}
