package context

import (
	"slices"
	"sort"

	"github.com/Permify/permify/internal/storage/context/utils"
	"github.com/Permify/permify/pkg/database"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// ContextualTuples - A collection of tuples with context.
type ContextualTuples struct {
	Tuples []*base.Tuple
}

// NewContextualTuples - Creates a new collection of tuples with context.
func NewContextualTuples(tuples ...*base.Tuple) *ContextualTuples {
	return &ContextualTuples{
		Tuples: tuples,
	}
}

// QueryRelationships filters the ContextualTuples based on the provided TupleFilter
// and returns a TupleIterator for the filtered tuples.
// QueryRelationships filters the ContextualTuples based on the provided TupleFilter, applies cursor-based pagination, and returns a TupleIterator for the filtered tuples.
func (c *ContextualTuples) QueryRelationships(filter *base.TupleFilter, pagination database.CursorPagination) (*database.TupleIterator, error) {
	// Sort tuples based on the provided order field
	sort.Slice(c.Tuples, func(i, j int) bool {
		switch pagination.Sort() {
		case "entity_id":
			return c.Tuples[i].GetEntity().GetId() < c.Tuples[j].GetEntity().GetId()
		case "subject_id":
			return c.Tuples[i].GetSubject().GetId() < c.Tuples[j].GetSubject().GetId()
		default:
			return false // If no valid order is provided, no sorting is applied
		}
	})

	cursor := ""
	if pagination.Cursor() != "" {
		t, err := utils.EncodedContinuousToken{Value: pagination.Cursor()}.Decode()
		if err != nil {
			return nil, err
		}
		cursor = t.(utils.ContinuousToken).Value
	}

	// Filter the tuples based on the provided filter and cursor
	filtered := c.filterTuples(filter, cursor, pagination.Sort())

	// Return a new TupleIterator for the filtered tuples
	return database.NewTupleIterator(filtered...), nil
}

// filterTuples applies the provided filter to c's Tuples and returns a slice of Tuples that match the filter.
func (c *ContextualTuples) filterTuples(filter *base.TupleFilter, cursor, order string) []*base.Tuple {
	var filtered []*base.Tuple // Initialize a slice to hold the filtered tuples

	// Iterate over the tuples
	for _, tup := range c.Tuples {
		// Skip tuples that come before the cursor based on the specified order field
		if cursor != "" && !isTupleAfterCursor(tup, cursor, order) {
			continue
		}

		// If a tuple matches the Entity, Relation, and Subject filters, add it to the filtered slice
		if matchesEntityFilterForTuples(tup, filter.GetEntity()) &&
			matchesRelationFilter(tup, filter.GetRelation()) &&
			matchesSubjectFilter(tup, filter.GetSubject()) {
			filtered = append(filtered, tup)
		}
	}

	return filtered // Return the filtered tuples
}

// isAfterCursor checks if the tuple's ID (based on the order field) comes after the cursor.
func isTupleAfterCursor(tup *base.Tuple, cursor, order string) bool {
	switch order {
	case "entity_id":
		return tup.GetEntity().GetId() >= cursor
	case "subject_id":
		return tup.GetSubject().GetId() >= cursor
	default:
		// If the order field is not recognized, default to not skipping any tuples
		return true
	}
}

// matchesEntityFilterForTuples checks if a Tuple matches the conditions in an EntityFilter.
func matchesEntityFilterForTuples(tup *base.Tuple, filter *base.EntityFilter) bool {
	// Return true if the filter is empty or the tuple's entity matches the filter
	return (filter.GetType() == "" || tup.GetEntity().GetType() == filter.GetType()) &&
		(len(filter.GetIds()) == 0 || slices.Contains(filter.GetIds(), tup.GetEntity().GetId()))
}

// matchesRelationFilter checks if a Tuple matches the condition in a RelationFilter.
func matchesRelationFilter(tup *base.Tuple, filter string) bool {
	// Return true if the filter is empty or the tuple's relation matches the filter
	return filter == "" || tup.GetRelation() == filter
}

// matchesSubjectFilter checks if a Tuple matches the conditions in a SubjectFilter.
func matchesSubjectFilter(tup *base.Tuple, filter *base.SubjectFilter) bool {
	// Return true if the filter is empty or the tuple's subject matches the filter
	return (filter.GetType() == "" || tup.GetSubject().GetType() == filter.GetType()) &&
		(len(filter.GetIds()) == 0 || slices.Contains(filter.GetIds(), tup.GetSubject().GetId())) &&
		(filter.GetRelation() == "" || tup.GetSubject().GetRelation() == filter.GetRelation())
}
