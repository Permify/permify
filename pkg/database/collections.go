package database

import (
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// TupleCollection tuple collection.
type TupleCollection struct {
	tuples []*base.Tuple
}

// NewTupleCollection create new tuple collection.
func NewTupleCollection(tuples ...*base.Tuple) *TupleCollection {
	if len(tuples) == 0 {
		return &TupleCollection{}
	}
	return &TupleCollection{
		tuples: tuples,
	}
}

// CreateTupleIterator create tuple iterator according to collection.
func (t *TupleCollection) CreateTupleIterator() ITupleIterator {
	return &TupleIterator{
		tuples: t.tuples,
	}
}

// GetTuples -
func (t *TupleCollection) GetTuples() []*base.Tuple {
	return t.tuples
}

// Add new subject to collection.
func (t *TupleCollection) Add(tuple *base.Tuple) {
	t.tuples = append(t.tuples, tuple)
	return
}

// SUBJECT

// SubjectCollection subject collection.
type SubjectCollection struct {
	subjects []*base.Subject
}

// NewSubjectCollection create new subject collection.
func NewSubjectCollection(subjects ...*base.Subject) ISubjectCollection {
	if len(subjects) == 0 {
		return &SubjectCollection{}
	}
	return &SubjectCollection{
		subjects: subjects,
	}
}

// CreateSubjectIterator create subject iterator according to collection.
func (s *SubjectCollection) CreateSubjectIterator() ISubjectIterator {
	return &SubjectIterator{
		subjects: s.subjects,
	}
}

// GetSubjects -
func (s *SubjectCollection) GetSubjects() []*base.Subject {
	return s.subjects
}

// Add new subject to collection.
func (s *SubjectCollection) Add(subject *base.Subject) {
	s.subjects = append(s.subjects, subject)
	return
}

// ENTITY

// EntityCollection entity collection.
type EntityCollection struct {
	entities []*base.Entity
}

// NewEntityCollection create new subject collection.
func NewEntityCollection(entities ...*base.Entity) IEntityCollection {
	if len(entities) == 0 {
		return &EntityCollection{}
	}
	return &EntityCollection{
		entities: entities,
	}
}

// CreateEntityIterator create entity iterator according to collection.
func (e *EntityCollection) CreateEntityIterator() IEntityIterator {
	return &EntityIterator{
		entities: e.entities,
	}
}

// GetEntities -
func (e *EntityCollection) GetEntities() []*base.Entity {
	return e.entities
}

// Add new subject to collection.
func (e *EntityCollection) Add(entity *base.Entity) {
	e.entities = append(e.entities, entity)
	return
}
