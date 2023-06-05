package database

import (
	"testing"

	base "github.com/Permify/permify/pkg/pb/base/v1"
)

func TestTupleIterator(t *testing.T) {
	// Create some tuples
	tuple1 := &base.Tuple{
		Subject:  &base.Subject{Type: "user", Id: "u1"},
		Relation: "rel1",
		Entity: &base.Entity{
			Type: "entity",
			Id:   "e1",
		},
	}

	tuple2 := &base.Tuple{
		Subject:  &base.Subject{Type: "user", Id: "u2"},
		Relation: "rel2",
		Entity: &base.Entity{
			Type: "entity",
			Id:   "e2",
		},
	}

	tuple3 := &base.Tuple{
		Subject:  &base.Subject{Type: "user", Id: "u3"},
		Relation: "rel3",
		Entity: &base.Entity{
			Type: "entity",
			Id:   "e3",
		},
	}

	// Create a tuple collection and add the tuples
	tupleCollection := NewTupleCollection(tuple1, tuple2, tuple3)

	// Create a tuple iterator
	tupleIterator := tupleCollection.CreateTupleIterator()

	// Test HasNext() and GetNext() methods
	if !tupleIterator.HasNext() {
		t.Error("Expected true for HasNext(), but got false")
	}
	if tupleIterator.GetNext() != tuple1 {
		t.Error("Expected tuple1 for GetNext(), but got something else")
	}
	if !tupleIterator.HasNext() {
		t.Error("Expected true for HasNext(), but got false")
	}
	if tupleIterator.GetNext() != tuple2 {
		t.Error("Expected tuple2 for GetNext(), but got something else")
	}
	if !tupleIterator.HasNext() {
		t.Error("Expected true for HasNext(), but got false")
	}
	if tupleIterator.GetNext() != tuple3 {
		t.Error("Expected tuple3 for GetNext(), but got something else")
	}
	if tupleIterator.HasNext() {
		t.Error("Expected false for HasNext(), but got true")
	}
}

func TestSubjectIterator(t *testing.T) {
	// Create some subjects
	subject1 := &base.Subject{Type: "user", Id: "u1"}
	subject2 := &base.Subject{Type: "user", Id: "u2"}
	subject3 := &base.Subject{Type: "user", Id: "u3"}

	// Create a subject collection and add the subjects
	subjectCollection := NewSubjectCollection(subject1, subject2, subject3)

	// Create a subject iterator
	subjectIterator := subjectCollection.CreateSubjectIterator()

	// Test HasNext() and GetNext() methods
	if !subjectIterator.HasNext() {
		t.Error("Expected true for HasNext(), but got false")
	}
	if subjectIterator.GetNext() != subject1 {
		t.Error("Expected subject1 for GetNext(), but got something else")
	}
	if !subjectIterator.HasNext() {
		t.Error("Expected true for HasNext(), but got false")
	}
	if subjectIterator.GetNext() != subject2 {
		t.Error("Expected subject2 for GetNext(), but got something else")
	}
	if !subjectIterator.HasNext() {
		t.Error("Expected true for HasNext(), but got false")
	}
	if subjectIterator.GetNext() != subject3 {
		t.Error("Expected subject3 for GetNext(), but got something else")
	}
	if subjectIterator.HasNext() {
		t.Error("Expected false for HasNext(), but got true")
	}
}

func TestUniqueTupleIterator(t *testing.T) {
	// Create some tuples
	tuple1 := &base.Tuple{
		Entity: &base.Entity{
			Type: "entity",
			Id:   "e1",
		},
		Relation: "rel1",
		Subject:  &base.Subject{Type: "user", Id: "u1"},
	}

	tuple2 := &base.Tuple{
		Entity: &base.Entity{
			Type: "entity",
			Id:   "e2",
		},
		Relation: "rel2",
		Subject:  &base.Subject{Type: "user", Id: "u2"},
	}

	tuple3 := &base.Tuple{
		Entity: &base.Entity{
			Type: "entity",
			Id:   "e3",
		},
		Relation: "rel3",
		Subject:  &base.Subject{Type: "user", Id: "u3"},
	}

	tuple4 := &base.Tuple{
		Entity: &base.Entity{
			Type: "entity",
			Id:   "e4",
		},
		Relation: "rel4",
		Subject:  &base.Subject{Type: "organization", Id: "o4", Relation: "member"},
	}

	tuple5 := &base.Tuple{
		Entity: &base.Entity{
			Type: "entity",
			Id:   "e5",
		},
		Relation: "rel5",
		Subject:  &base.Subject{Type: "user", Id: "u5"},
	}

	tuple6 := &base.Tuple{
		Entity: &base.Entity{
			Type: "entity",
			Id:   "e6",
		},
		Relation: "rel6",
		Subject:  &base.Subject{Type: "organization", Id: "o6", Relation: "admin"},
	}

	// Create a tuple iterators
	tupleIterator1 := NewTupleIterator(tuple1, tuple2, tuple3, tuple6)
	tupleIterator2 := NewTupleIterator(tuple6, tuple1, tuple2, tuple4, tuple5)

	// Create a unique iterator
	uniqueIterator := NewUniqueTupleIterator(tupleIterator1, tupleIterator2)

	// Test HasNext() and GetNext() methods
	if !uniqueIterator.HasNext() {
		t.Error("Expected true for HasNext(), but got false")
	}
	if uniqueIterator.GetNext() != tuple1 {
		t.Error("Expected tuple1 for GetNext(), but got something else")
	}
	if !uniqueIterator.HasNext() {
		t.Error("Expected true for HasNext(), but got false")
	}
	if uniqueIterator.GetNext() != tuple2 {
		t.Error("Expected tuple2 for GetNext(), but got something else")
	}
	if !uniqueIterator.HasNext() {
		t.Error("Expected true for HasNext(), but got false")
	}
	if uniqueIterator.GetNext() != tuple3 {
		t.Error("Expected tuple3 for GetNext(), but got something else")
	}
	if !uniqueIterator.HasNext() {
		t.Error("Expected false for HasNext(), but got true")
	}
	if uniqueIterator.GetNext() != tuple6 {
		t.Error("Expected tuple6 for GetNext(), but got something else")
	}
	if !uniqueIterator.HasNext() {
		t.Error("Expected false for HasNext(), but got true")
	}
	if uniqueIterator.GetNext() != tuple4 {
		t.Error("Expected tuple4 for GetNext(), but got something else")
	}
	if !uniqueIterator.HasNext() {
		t.Error("Expected false for HasNext(), but got true")
	}
	if uniqueIterator.GetNext() != tuple5 {
		t.Error("Expected tuple5 for GetNext(), but got something else")
	}
	if uniqueIterator.HasNext() {
		t.Error("Expected false for HasNext(), but got true")
	}
}
