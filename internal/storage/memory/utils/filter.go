package utils

import (
	"slices"

	"github.com/hashicorp/go-memdb"

	"github.com/Permify/permify/internal/storage"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// FilterRelationTuplesQuery - Filter relation tuples according to given filter
func FilterRelationTuplesQuery(tenantID string, filter *base.TupleFilter) memdb.FilterFunc {
	return func(tupleRaw interface{}) bool {
		tuple, ok := tupleRaw.(storage.RelationTuple)
		if !ok {
			return true
		}
		switch {
		case tuple.TenantID != tenantID:
			return true
		case filter.GetEntity().GetType() != "" && tuple.EntityType != filter.GetEntity().GetType():
			return true
		case len(filter.GetEntity().GetIds()) > 0 && !slices.Contains(filter.GetEntity().GetIds(), tuple.EntityID):
			return true
		case filter.GetRelation() != "" && tuple.Relation != filter.GetRelation():
			return true
		case filter.GetSubject().GetType() != "" && tuple.SubjectType != filter.GetSubject().GetType():
			return true
		case len(filter.GetSubject().GetIds()) > 0 && !slices.Contains(filter.GetSubject().GetIds(), tuple.SubjectID):
			return true
		case filter.GetSubject().GetRelation() != "" && tuple.SubjectRelation != filter.GetSubject().GetRelation():
			return true
		}
		return false
	}
}

// FilterAttributesQuery - Filter attributes according to given filter
func FilterAttributesQuery(tenantID string, filter *base.AttributeFilter) memdb.FilterFunc {
	return func(attributeRaw interface{}) bool {
		attribute, ok := attributeRaw.(storage.Attribute)
		if !ok {
			return true
		}
		switch {
		case attribute.TenantID != tenantID:
			return true
		case filter.GetEntity().GetType() != "" && attribute.EntityType != filter.GetEntity().GetType():
			return true
		case len(filter.GetEntity().GetIds()) > 0 && !slices.Contains(filter.GetEntity().GetIds(), attribute.EntityID):
			return true
		case len(filter.GetAttributes()) > 0 && !slices.Contains(filter.GetAttributes(), attribute.Attribute):
			return true
		}
		return false
	}
}
