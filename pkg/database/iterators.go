package database

import (
	base "github.com/Permify/permify/pkg/pb/base/v1"
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
		tuple := i.tuples[i.index]
		i.index++
		return tuple
	}
	return nil
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
