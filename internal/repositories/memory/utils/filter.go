package utils

import (
	"github.com/hashicorp/go-memdb"

	"github.com/Permify/permify/internal/repositories"
	"github.com/Permify/permify/pkg/helper"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// FilterQuery -
func FilterQuery(filter *base.TupleFilter) memdb.FilterFunc {
	return func(tupleRaw interface{}) bool {
		tuple := tupleRaw.(repositories.RelationTuple)
		switch {
		case filter.GetEntity().GetType() != "" && tuple.EntityType != filter.GetEntity().GetType():
			return true
		case len(filter.GetEntity().GetIds()) > 0 && !helper.InArray(tuple.EntityID, filter.GetEntity().GetIds()):
			return true
		case filter.GetRelation() != "" && tuple.Relation != filter.GetRelation():
			return true
		case filter.GetSubject().GetType() != "" && tuple.SubjectType != filter.GetSubject().GetType():
			return true
		case len(filter.GetSubject().GetIds()) > 0 && !helper.InArray(tuple.SubjectID, filter.GetSubject().GetIds()):
			return true
		case filter.GetSubject().GetRelation() != "" && tuple.SubjectRelation != filter.GetSubject().GetRelation():
			return true
		}
		return false
	}
}
