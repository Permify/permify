package tuple

import (
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// TUPLE

// ITupleIterator abstract tuple iterator.
type ITupleIterator interface {
	HasNext() bool
	GetNext() *base.Tuple
}

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
		user := i.tuples[i.index]
		i.index++
		return user
	}
	return nil
}

// SUBJECT

// ISubjectIterator abstract subject iterator.
type ISubjectIterator interface {
	HasNext() bool
	GetNext() *base.Subject
}

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
		t := u.subjects[u.index]
		u.index++
		return t
	}
	return nil
}

// ENTITY

// IEntityIterator abstract subject iterator.
type IEntityIterator interface {
	HasNext() bool
	GetNext() *base.Entity
}

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
		t := u.entities[u.index]
		u.index++
		return t
	}
	return nil
}
