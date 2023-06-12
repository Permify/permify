package utils

import (
	"github.com/Masterminds/squirrel"

	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// FilterQueryForSelectBuilder -
func FilterQueryForSelectBuilder(sl squirrel.SelectBuilder, filter *base.TupleFilter) squirrel.SelectBuilder {
	eq := squirrel.Eq{}

	if filter.GetEntity().GetType() != "" {
		eq["entity_type"] = filter.GetEntity().GetType()
	}

	if len(filter.GetEntity().GetIds()) > 0 {
		eq["entity_id"] = filter.GetEntity().GetIds()
	}

	if filter.GetRelation() != "" {
		eq["relation"] = filter.GetRelation()
	}

	if filter.GetSubject().GetType() != "" {
		eq["subject_type"] = filter.GetSubject().GetType()
	}

	if len(filter.GetSubject().GetIds()) > 0 {
		eq["subject_id"] = filter.GetSubject().GetIds()
	}

	if filter.GetSubject().GetRelation() != "" {
		eq["subject_relation"] = filter.GetSubject().GetRelation()
	}

	// If eq is empty, return the original squirrel.UpdateBuilder without attaching a WHERE clause.
	// This ensures we don't accidentally update every row in the table.
	if len(eq) == 0 {
		return sl
	}

	return sl.Where(eq)
}

// FilterQueryForUpdateBuilder -
func FilterQueryForUpdateBuilder(sl squirrel.UpdateBuilder, filter *base.TupleFilter) squirrel.UpdateBuilder {
	eq := squirrel.Eq{}

	if filter.GetEntity().GetType() != "" {
		eq["entity_type"] = filter.GetEntity().GetType()
	}

	if len(filter.GetEntity().GetIds()) > 0 {
		eq["entity_id"] = filter.GetEntity().GetIds()
	}

	if filter.GetRelation() != "" {
		eq["relation"] = filter.GetRelation()
	}

	if filter.GetSubject().GetType() != "" {
		eq["subject_type"] = filter.GetSubject().GetType()
	}

	if len(filter.GetSubject().GetIds()) > 0 {
		eq["subject_id"] = filter.GetSubject().GetIds()
	}

	if filter.GetSubject().GetRelation() != "" {
		eq["subject_relation"] = filter.GetSubject().GetRelation()
	}

	// If eq is empty, return the original squirrel.UpdateBuilder without attaching a WHERE clause.
	// This ensures we don't accidentally update every row in the table.
	if len(eq) == 0 {
		return sl
	}

	return sl.Where(eq)
}
