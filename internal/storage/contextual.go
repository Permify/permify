package storage

import (
	"golang.org/x/exp/slices"

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

// QueryRelationships filters the ContextualTuples based on the provided TupleFilter and returns a TupleIterator for the filtered tuples.
func (c *ContextualTuples) QueryRelationships(filter *base.TupleFilter) (*database.TupleIterator, error) {
	filtered := c.filterTuples(filter)                 // Filter the tuples based on the provided filter
	return database.NewTupleIterator(filtered...), nil // Return a new TupleIterator for the filtered tuples
}

// filterTuples applies the provided filter to c's Tuples and returns a slice of Tuples that match the filter.
func (c *ContextualTuples) filterTuples(filter *base.TupleFilter) []*base.Tuple {
	var filtered []*base.Tuple // Initialize a slice to hold the filtered tuples

	// Iterate over the tuples
	for _, tup := range c.Tuples {
		// If a tuple matches the Entity, Relation, and Subject filters, add it to the filtered slice
		if matchesEntityFilter(tup, filter.GetEntity()) &&
			matchesRelationFilter(tup, filter.GetRelation()) &&
			matchesSubjectFilter(tup, filter.GetSubject()) {
			filtered = append(filtered, tup)
		}
	}

	return filtered // Return the filtered tuples
}

// matchesEntityFilter checks if a Tuple matches the conditions in an EntityFilter.
func matchesEntityFilter(tup *base.Tuple, filter *base.EntityFilter) bool {
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
