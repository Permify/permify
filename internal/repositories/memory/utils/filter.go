package utils

import (
	"github.com/hashicorp/go-memdb"
	"golang.org/x/exp/slices"

	"github.com/Permify/permify/internal/repositories"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// FilterQuery - Filter relation tuples according to given filter
func FilterQuery(filter *base.TupleFilter) memdb.FilterFunc {
	return func(tupleRaw interface{}) bool {
		tuple, ok := tupleRaw.(repositories.RelationTuple)
		if !ok {
			return true
		}
		switch {
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
