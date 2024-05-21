package utils

import (
	"github.com/Masterminds/squirrel"

	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// TuplesFilterQueryForSelectBuilder -
func TuplesFilterQueryForSelectBuilder(sl squirrel.SelectBuilder, filter *base.TupleFilter) squirrel.SelectBuilder {
	eq := squirrel.Eq{}

	if filter.GetEntity().GetType() != "" {
		eq["entity_type"] = filter.GetEntity().GetType()
	}

	entityIds := filter.GetEntity().GetIds()
	if len(entityIds) > 0 {
		if len(entityIds) == 1 {
			// If there is only one ID, use the = operator
			eq["entity_id"] = entityIds[0]
		} else {
			// If there are multiple IDs, use the IN operator
			eq["entity_id"] = entityIds
		}
	}

	if filter.GetRelation() != "" {
		eq["relation"] = filter.GetRelation()
	}

	if filter.GetSubject().GetType() != "" {
		eq["subject_type"] = filter.GetSubject().GetType()
	}

	subjectIds := filter.GetSubject().GetIds()
	if len(subjectIds) > 0 {
		if len(subjectIds) == 1 {
			// If there is only one ID, use the = operator
			eq["subject_id"] = subjectIds[0]
		} else {
			// If there are multiple IDs, use the IN operator
			eq["subject_id"] = subjectIds
		}
	}

	if filter.GetSubject().GetRelation() != "" {
		eq["subject_relation"] = filter.GetSubject().GetRelation()
	}

	// If eq is empty, return the original squirrel.SelectBuilder without attaching a WHERE clause.
	// This ensures we don't accidentally select every row in the table.
	if len(eq) == 0 {
		return sl
	}

	return sl.Where(eq)
}

// AttributesFilterQueryForSelectBuilder -
func AttributesFilterQueryForSelectBuilder(sl squirrel.SelectBuilder, filter *base.AttributeFilter) squirrel.SelectBuilder {
	eq := squirrel.Eq{}

	if filter.GetEntity().GetType() != "" {
		eq["entity_type"] = filter.GetEntity().GetType()
	}

	entityIds := filter.GetEntity().GetIds()
	if len(entityIds) > 0 {
		if len(entityIds) == 1 {
			// If there is only one ID, use the = operator
			eq["entity_id"] = entityIds[0]
		} else {
			// If there are multiple IDs, use the IN operator
			eq["entity_id"] = entityIds
		}
	}

	attributes := filter.GetAttributes()
	if len(attributes) > 0 {
		if len(attributes) == 1 {
			// If there is only one attribute, use the = operator
			eq["attribute"] = attributes[0]
		} else {
			// If there are multiple attributes, use the IN operator
			eq["attribute"] = attributes
		}
	}

	// If eq is empty, return the original squirrel.SelectBuilder without attaching a WHERE clause.
	// This ensures we don't accidentally select every row in the table.
	if len(eq) == 0 {
		return sl
	}

	return sl.Where(eq)
}

// TuplesFilterQueryForUpdateBuilder -
func TuplesFilterQueryForUpdateBuilder(sl squirrel.UpdateBuilder, filter *base.TupleFilter) squirrel.UpdateBuilder {
	eq := squirrel.Eq{}

	if filter.GetEntity().GetType() != "" {
		eq["entity_type"] = filter.GetEntity().GetType()
	}

	entityIds := filter.GetEntity().GetIds()
	if len(entityIds) > 0 {
		if len(entityIds) == 1 {
			// If there is only one ID, use the = operator
			eq["entity_id"] = entityIds[0]
		} else {
			// If there are multiple IDs, use the IN operator
			eq["entity_id"] = entityIds
		}
	}

	if filter.GetRelation() != "" {
		eq["relation"] = filter.GetRelation()
	}

	if filter.GetSubject().GetType() != "" {
		eq["subject_type"] = filter.GetSubject().GetType()
	}

	subjectIds := filter.GetSubject().GetIds()
	if len(subjectIds) > 0 {
		if len(subjectIds) == 1 {
			// If there is only one ID, use the = operator
			eq["subject_id"] = subjectIds[0]
		} else {
			// If there are multiple IDs, use the IN operator
			eq["subject_id"] = subjectIds
		}
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

// AttributesFilterQueryForUpdateBuilder -
func AttributesFilterQueryForUpdateBuilder(sl squirrel.UpdateBuilder, filter *base.AttributeFilter) squirrel.UpdateBuilder {
	eq := squirrel.Eq{}

	if filter.GetEntity().GetType() != "" {
		eq["entity_type"] = filter.GetEntity().GetType()
	}

	entityIds := filter.GetEntity().GetIds()
	if len(entityIds) > 0 {
		if len(entityIds) == 1 {
			// If there is only one ID, use the = operator
			eq["entity_id"] = entityIds[0]
		} else {
			// If there are multiple IDs, use the IN operator
			eq["entity_id"] = entityIds
		}
	}

	attributes := filter.GetAttributes()
	if len(attributes) > 0 {
		if len(attributes) == 1 {
			// If there is only one attribute, use the = operator
			eq["attribute"] = attributes[0]
		} else {
			// If there are multiple attributes, use the IN operator
			eq["attribute"] = attributes
		}
	}

	// If eq is empty, return the original squirrel.UpdateBuilder without attaching a WHERE clause.
	// This ensures we don't accidentally update every row in the table.
	if len(eq) == 0 {
		return sl
	}

	return sl.Where(eq)
}
