package database

import (
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/tuple"
)

// TUPLE

// TupleIterator - Tuple iterator -
type TupleIterator struct {
	index  int
	tuples []*base.Tuple
}

// NewTupleIterator -
func NewTupleIterator(tuples ...*base.Tuple) *TupleIterator {
	return &TupleIterator{
		index:  0,
		tuples: tuples,
	}
}

// HasNext - Checks whether next tuple exists
func (i *TupleIterator) HasNext() bool {
	return i.index < len(i.tuples)
}

// GetNext - Get next tuple
func (i *TupleIterator) GetNext() *base.Tuple {
	if i.HasNext() {
		tup := i.tuples[i.index]
		i.index++
		return tup
	}
	return nil
}

// UniqueTupleIterator combines two TupleIterators and ensures that only unique Tuples are returned.
// Uniqueness is based on the Tuple's pointer address.
type UniqueTupleIterator struct {
	iterator1, iterator2 *TupleIterator  // The two TupleIterators to combine
	visited              map[string]bool // A map to track Tuples that have been returned
}

// NewUniqueTupleIterator creates a new UniqueTupleIterator from two TupleIterators.
func NewUniqueTupleIterator(iterator1, iterator2 *TupleIterator) *UniqueTupleIterator {
	return &UniqueTupleIterator{
		iterator1: iterator1,
		iterator2: iterator2,
		visited:   make(map[string]bool), // Initialize an empty map for visited Tuples
	}
}

// HasNext checks if there is a next Tuple in either of the two TupleIterators.
func (i *UniqueTupleIterator) HasNext() bool {
	// If either iterator has a next Tuple, return true
	return i.iterator1.HasNext() || i.iterator2.HasNext()
}

// GetNext returns the next unique Tuple from the two TupleIterators.
func (i *UniqueTupleIterator) GetNext() (*base.Tuple, bool) {
	// Check the first iterator for any next Tuples
	for i.iterator1.HasNext() {
		tup := i.iterator1.GetNext() // Get the next Tuple
		key := tuple.ToString(tup)
		// If the Tuple hasn't been visited yet, mark it as visited and return it
		if _, found := i.visited[key]; !found {
			i.visited[key] = true
			return tup, true
		}
	}

	// If no more unique Tuples are in the first iterator, check the second one
	for i.iterator2.HasNext() {
		tup := i.iterator2.GetNext() // Get the next Tuple
		key := tuple.ToString(tup)
		// If the Tuple hasn't been visited yet, mark it as visited and return it
		if _, found := i.visited[key]; !found {
			i.visited[key] = true
			return tup, true
		}
	}

	// If no more unique Tuples are in either of the iterators, return nil
	return &base.Tuple{}, false
}

// UniqueAttributeIterator combines two AttributeIterators and ensures that only unique Attributes are returned.
// Uniqueness is based on the Attribute's pointer address.
type UniqueAttributeIterator struct {
	iterator1, iterator2 *AttributeIterator // The two AttributeIterators to combine
	visited              map[string]bool    // A map to track Attributes that have been returned
}

// NewUniqueAttributeIterator creates a new UniqueAttributeIterator from two AttributeIterators.
func NewUniqueAttributeIterator(iterator1, iterator2 *AttributeIterator) *UniqueAttributeIterator {
	return &UniqueAttributeIterator{
		iterator1: iterator1,
		iterator2: iterator2,
		visited:   make(map[string]bool), // Initialize an empty map for visited Tuples
	}
}

// HasNext checks if there is a next Tuple in either of the two TupleIterators.
func (i *UniqueAttributeIterator) HasNext() bool {
	// If either iterator has a next Tuple, return true
	return i.iterator1.HasNext() || i.iterator2.HasNext()
}

// GetNext returns the next unique Tuple from the two TupleIterators.
func (i *UniqueAttributeIterator) GetNext() (*base.Attribute, bool) {
	// Check the first iterator for any next Tuples
	for i.iterator1.HasNext() {
		attr := i.iterator1.GetNext() // Get the next Tuple
		key := attr.GetEntity().GetType() + ":" + attr.GetEntity().GetId() + "$" + attr.GetAttribute()
		// If the Tuple hasn't been visited yet, mark it as visited and return it
		if _, found := i.visited[key]; !found {
			i.visited[key] = true
			return attr, true
		}
	}

	// If no more unique Tuples are in the first iterator, check the second one
	for i.iterator2.HasNext() {
		attr := i.iterator2.GetNext() // Get the next Tuple
		key := attr.GetEntity().GetType() + ":" + attr.GetEntity().GetId() + "$" + attr.GetAttribute()
		// If the Tuple hasn't been visited yet, mark it as visited and return it
		if _, found := i.visited[key]; !found {
			i.visited[key] = true
			return attr, true
		}
	}

	// If no more unique Tuples are in either of the iterators, return nil
	return &base.Attribute{}, false
}

// SUBJECT

// SubjectIterator - Structure for subject iterator
type SubjectIterator struct {
	index    int
	subjects []*base.Subject
}

// NewSubjectIterator -
func NewSubjectIterator(subjects []*base.Subject) *SubjectIterator {
	return &SubjectIterator{
		index:    0,
		subjects: subjects,
	}
}

// HasNext - Checks whether next subject exists
func (u *SubjectIterator) HasNext() bool {
	return u.index < len(u.subjects)
}

// GetNext - Get next tuple
func (u *SubjectIterator) GetNext() *base.Subject {
	if u.HasNext() {
		subject := u.subjects[u.index]
		u.index++
		return subject
	}
	return nil
}

// ENTITY

// EntityIterator - Structure for entity iterator
type EntityIterator struct {
	index    int
	entities []*base.Entity
}

// NewEntityIterator -
func NewEntityIterator(entities []*base.Entity) *EntityIterator {
	return &EntityIterator{
		index:    0,
		entities: entities,
	}
}

// HasNext - Checks whether next entity exists
func (u *EntityIterator) HasNext() bool {
	return u.index < len(u.entities)
}

// GetNext - Get next entity
func (u *EntityIterator) GetNext() *base.Entity {
	if u.HasNext() {
		entity := u.entities[u.index]
		u.index++
		return entity
	}
	return nil
}

// AttributeIterator - Attribute iterator -
type AttributeIterator struct {
	index      int
	attributes []*base.Attribute
}

// NewAttributeIterator -
func NewAttributeIterator(attributes ...*base.Attribute) *AttributeIterator {
	return &AttributeIterator{
		index:      0,
		attributes: attributes,
	}
}

// HasNext - Checks whether next entity exists
func (u *AttributeIterator) HasNext() bool {
	return u.index < len(u.attributes)
}

// GetNext - Get next entity
func (u *AttributeIterator) GetNext() *base.Attribute {
	if u.HasNext() {
		attribute := u.attributes[u.index]
		u.index++
		return attribute
	}
	return nil
}
