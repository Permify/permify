package database

import (
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// TupleCollection -Tuple collection.
type TupleCollection struct {
	tuples []*base.Tuple
}

// NewTupleCollection - Create new tuple collection.
func NewTupleCollection(tuples ...*base.Tuple) *TupleCollection {
	if len(tuples) == 0 {
		return &TupleCollection{}
	}
	return &TupleCollection{
		tuples: tuples,
	}
}

// CreateTupleIterator - Create tuple iterator according to collection.
func (t *TupleCollection) CreateTupleIterator() *TupleIterator {
	return &TupleIterator{
		tuples: t.tuples,
	}
}

// GetTuples - Get tuples
func (t *TupleCollection) GetTuples() []*base.Tuple {
	return t.tuples
}

// Add - New subject to collection.
func (t *TupleCollection) Add(tuple *base.Tuple) {
	t.tuples = append(t.tuples, tuple)
}

// ToSubjectCollection - Converts new subject collection from given tuple collection
func (t *TupleCollection) ToSubjectCollection() *SubjectCollection {
	subjects := make([]*base.Subject, len(t.tuples))
	for index, tuple := range t.tuples {
		subjects[index] = tuple.GetSubject()
	}
	return NewSubjectCollection(subjects...)
}

// SUBJECT

// SubjectCollection - Subject collection.
type SubjectCollection struct {
	subjects []*base.Subject
}

// NewSubjectCollection - Create new subject collection.
func NewSubjectCollection(subjects ...*base.Subject) *SubjectCollection {
	if len(subjects) == 0 {
		return &SubjectCollection{}
	}
	return &SubjectCollection{
		subjects: subjects,
	}
}

// CreateSubjectIterator - Create subject iterator according to collection.
func (s *SubjectCollection) CreateSubjectIterator() *SubjectIterator {
	return &SubjectIterator{
		subjects: s.subjects,
	}
}

// GetSubjects - Get subject collection
func (s *SubjectCollection) GetSubjects() []*base.Subject {
	return s.subjects
}

// Add - New subject to collection.
func (s *SubjectCollection) Add(subject *base.Subject) {
	s.subjects = append(s.subjects, subject)
}

// ENTITY

// EntityCollection - Entity collection.
type EntityCollection struct {
	entities []*base.Entity
}

// NewEntityCollection - Create new subject collection.
func NewEntityCollection(entities ...*base.Entity) *EntityCollection {
	if len(entities) == 0 {
		return &EntityCollection{}
	}
	return &EntityCollection{
		entities: entities,
	}
}

// CreateEntityIterator  - Create entity iterator according to collection.
func (e *EntityCollection) CreateEntityIterator() *EntityIterator {
	return &EntityIterator{
		entities: e.entities,
	}
}

// GetEntities - Get entities
func (e *EntityCollection) GetEntities() []*base.Entity {
	return e.entities
}

// Add - New subject to collection.
func (e *EntityCollection) Add(entity *base.Entity) {
	e.entities = append(e.entities, entity)
}

// AttributeCollection -Attribute collection.
type AttributeCollection struct {
	attributes []*base.Attribute
}

// NewAttributeCollection - Create new attribute collection.
func NewAttributeCollection(attributes ...*base.Attribute) *AttributeCollection {
	if len(attributes) == 0 {
		return &AttributeCollection{}
	}
	return &AttributeCollection{
		attributes: attributes,
	}
}

// GetAttributes - Get entities
func (t *AttributeCollection) GetAttributes() []*base.Attribute {
	return t.attributes
}

// CreateAttributeIterator - Create tuple iterator according to collection.
func (t *AttributeCollection) CreateAttributeIterator() *AttributeIterator {
	return &AttributeIterator{
		attributes: t.attributes,
	}
}

// Add - New subject to collection.
func (t *AttributeCollection) Add(attribute *base.Attribute) {
	t.attributes = append(t.attributes, attribute)
}

// TupleBundle defines a structure for managing collections of tuples,
// with separate collections for write (create/update) and delete operations.
type TupleBundle struct {
	// Write is a TupleCollection intended to hold tuples that are to be created or updated.
	Write TupleCollection

	// Delete is a TupleCollection intended to hold tuples that are to be deleted.
	Delete TupleCollection
}

// AttributeBundle defines a structure for managing collections of attributes,
// with separate collections for write (create/update) and delete operations.
type AttributeBundle struct {
	// Write is an AttributeCollection intended to hold attributes that are to be created or updated.
	Write AttributeCollection

	// Delete is an AttributeCollection intended to hold attributes that are to be deleted.
	Delete AttributeCollection
}
