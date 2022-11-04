package database

import (
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// TUPLE

// TupleIterator tuple iterator -
type TupleIterator struct {
	index  int
	tuples []*base.Tuple
}

// HasNext -
func (i *TupleIterator) HasNext() bool {
	if i.index < len(i.tuples) {
		return true
	}
	return false
}

// GetNext -
func (i *TupleIterator) GetNext() *base.Tuple {
	if i.HasNext() {
		tuple := i.tuples[i.index]
		i.index++
		return tuple
	}
	return nil
}

// SUBJECT

// SubjectIterator -
type SubjectIterator struct {
	index    int
	subjects []*base.Subject
}

// HasNext -
func (u *SubjectIterator) HasNext() bool {
	if u.index < len(u.subjects) {
		return true
	}
	return false
}

// GetNext -
func (u *SubjectIterator) GetNext() *base.Subject {
	if u.HasNext() {
		subject := u.subjects[u.index]
		u.index++
		return subject
	}
	return nil
}

// ENTITY

// EntityIterator -
type EntityIterator struct {
	index    int
	entities []*base.Entity
}

// HasNext -
func (u *EntityIterator) HasNext() bool {
	if u.index < len(u.entities) {
		return true
	}
	return false
}

// GetNext -
func (u *EntityIterator) GetNext() *base.Entity {
	if u.HasNext() {
		entity := u.entities[u.index]
		u.index++
		return entity
	}
	return nil
}
