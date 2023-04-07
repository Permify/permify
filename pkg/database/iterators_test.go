package database

import (
	`testing`

	base `github.com/Permify/permify/pkg/pb/base/v1`
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
