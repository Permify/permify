package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/anypb"

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
	i, _ := uniqueIterator.GetNext()
	if i != tuple1 {
		t.Error("Expected tuple1 for GetNext(), but got something else")
	}
	if !uniqueIterator.HasNext() {
		t.Error("Expected true for HasNext(), but got false")
	}
	i, _ = uniqueIterator.GetNext()
	if i != tuple2 {
		t.Error("Expected tuple2 for GetNext(), but got something else")
	}
	if !uniqueIterator.HasNext() {
		t.Error("Expected true for HasNext(), but got false")
	}
	i, _ = uniqueIterator.GetNext()
	if i != tuple3 {
		t.Error("Expected tuple3 for GetNext(), but got something else")
	}
	if !uniqueIterator.HasNext() {
		t.Error("Expected false for HasNext(), but got true")
	}
	i, _ = uniqueIterator.GetNext()
	if i != tuple6 {
		t.Error("Expected tuple6 for GetNext(), but got something else")
	}
	if !uniqueIterator.HasNext() {
		t.Error("Expected false for HasNext(), but got true")
	}
	i, _ = uniqueIterator.GetNext()
	if i != tuple4 {
		t.Error("Expected tuple4 for GetNext(), but got something else")
	}
	if !uniqueIterator.HasNext() {
		t.Error("Expected false for HasNext(), but got true")
	}
	i, _ = uniqueIterator.GetNext()
	if i != tuple5 {
		t.Error("Expected tuple5 for GetNext(), but got something else")
	}
	if uniqueIterator.HasNext() {
		t.Error("Expected false for HasNext(), but got true")
	}
}

func TestAttributeIterator(t *testing.T) {
	isPublic, err := anypb.New(&base.BooleanValue{Data: true})
	assert.NoError(t, err)

	isPublic2, err := anypb.New(&base.BooleanValue{Data: false})
	assert.NoError(t, err)

	// Create some attributes
	attribute1 := &base.Attribute{
		Entity: &base.Entity{
			Type: "entity",
			Id:   "e1",
		},
		Attribute: "public",
		Value:     isPublic,
	}

	attribute2 := &base.Attribute{
		Entity: &base.Entity{
			Type: "entity",
			Id:   "e2",
		},
		Attribute: "public",
		Value:     isPublic2,
	}

	attribute3 := &base.Attribute{
		Entity: &base.Entity{
			Type: "entity",
			Id:   "e3",
		},
		Attribute: "public",
		Value:     isPublic,
	}

	// Create an attribute collection and add the tuples
	attributeCollection := NewAttributeCollection(attribute1, attribute2, attribute3)

	// Create a attribute iterator
	attributeIterator := attributeCollection.CreateAttributeIterator()

	// Test HasNext() and GetNext() methods
	if !attributeIterator.HasNext() {
		t.Error("Expected true for HasNext(), but got false")
	}
	if attributeIterator.GetNext() != attribute1 {
		t.Error("Expected tuple1 for GetNext(), but got something else")
	}
	if !attributeIterator.HasNext() {
		t.Error("Expected true for HasNext(), but got false")
	}
	if attributeIterator.GetNext() != attribute2 {
		t.Error("Expected tuple2 for GetNext(), but got something else")
	}
	if !attributeIterator.HasNext() {
		t.Error("Expected true for HasNext(), but got false")
	}
	if attributeIterator.GetNext() != attribute3 {
		t.Error("Expected tuple3 for GetNext(), but got something else")
	}
	if attributeIterator.HasNext() {
		t.Error("Expected false for HasNext(), but got true")
	}
}

func TestUniqueAttributeIterator(t *testing.T) {
	isPublic, err := anypb.New(&base.BooleanValue{Data: true})
	assert.NoError(t, err)

	// Create some attributes
	attribute1 := &base.Attribute{
		Entity: &base.Entity{
			Type: "entity",
			Id:   "e1",
		},
		Attribute: "public",
		Value:     isPublic,
	}

	attribute2 := &base.Attribute{
		Entity: &base.Entity{
			Type: "entity",
			Id:   "e2",
		},
		Attribute: "public",
		Value:     isPublic,
	}

	attribute3 := &base.Attribute{
		Entity: &base.Entity{
			Type: "entity",
			Id:   "e3",
		},
		Attribute: "public",
		Value:     isPublic,
	}

	attribute4 := &base.Attribute{
		Entity: &base.Entity{
			Type: "entity",
			Id:   "e4",
		},
		Attribute: "public",
		Value:     isPublic,
	}

	attribute5 := &base.Attribute{
		Entity: &base.Entity{
			Type: "entity",
			Id:   "e5",
		},
		Attribute: "public",
		Value:     isPublic,
	}

	attribute6 := &base.Attribute{
		Entity: &base.Entity{
			Type: "entity",
			Id:   "e6",
		},
		Attribute: "public",
		Value:     isPublic,
	}

	// Create a tuple iterators
	attributeIterator1 := NewAttributeIterator(attribute1, attribute2, attribute3, attribute6)
	attributeIterator2 := NewAttributeIterator(attribute6, attribute1, attribute2, attribute4, attribute5)

	// Create a unique iterator
	uniqueIterator := NewUniqueAttributeIterator(attributeIterator1, attributeIterator2)

	// Test HasNext() and GetNext() methods
	if !uniqueIterator.HasNext() {
		t.Error("Expected true for HasNext(), but got false")
	}
	i, _ := uniqueIterator.GetNext()
	if i != attribute1 {
		t.Error("Expected tuple1 for GetNext(), but got something else")
	}
	if !uniqueIterator.HasNext() {
		t.Error("Expected true for HasNext(), but got false")
	}
	i, _ = uniqueIterator.GetNext()
	if i != attribute2 {
		t.Error("Expected tuple2 for GetNext(), but got something else")
	}
	if !uniqueIterator.HasNext() {
		t.Error("Expected true for HasNext(), but got false")
	}
	i, _ = uniqueIterator.GetNext()
	if i != attribute3 {
		t.Error("Expected tuple3 for GetNext(), but got something else")
	}
	if !uniqueIterator.HasNext() {
		t.Error("Expected false for HasNext(), but got true")
	}
	i, _ = uniqueIterator.GetNext()
	if i != attribute6 {
		t.Error("Expected tuple6 for GetNext(), but got something else")
	}
	if !uniqueIterator.HasNext() {
		t.Error("Expected false for HasNext(), but got true")
	}
	i, _ = uniqueIterator.GetNext()
	if i != attribute4 {
		t.Error("Expected tuple4 for GetNext(), but got something else")
	}
	if !uniqueIterator.HasNext() {
		t.Error("Expected false for HasNext(), but got true")
	}
	i, _ = uniqueIterator.GetNext()
	if i != attribute5 {
		t.Error("Expected tuple5 for GetNext(), but got something else")
	}
	if uniqueIterator.HasNext() {
		t.Error("Expected false for HasNext(), but got true")
	}
}

func TestEntityIterator(t *testing.T) {
	// Create some tuples
	en1 := &base.Entity{
		Type: "entity",
		Id:   "e1",
	}

	en2 := &base.Entity{
		Type: "entity",
		Id:   "e2",
	}

	en3 := &base.Entity{
		Type: "entity",
		Id:   "e3",
	}

	// Create a entity collection and add the tuples
	entityCollection := NewEntityCollection(en1, en2, en3)

	// Create an entity iterator
	entityIterator := entityCollection.CreateEntityIterator()

	// Test HasNext() and GetNext() methods
	if !entityIterator.HasNext() {
		t.Error("Expected true for HasNext(), but got false")
	}
	if entityIterator.GetNext() != en1 {
		t.Error("Expected tuple1 for GetNext(), but got something else")
	}
	if !entityIterator.HasNext() {
		t.Error("Expected true for HasNext(), but got false")
	}
	if entityIterator.GetNext() != en2 {
		t.Error("Expected tuple2 for GetNext(), but got something else")
	}
	if !entityIterator.HasNext() {
		t.Error("Expected true for HasNext(), but got false")
	}
	if entityIterator.GetNext() != en3 {
		t.Error("Expected tuple3 for GetNext(), but got something else")
	}
	if entityIterator.HasNext() {
		t.Error("Expected false for HasNext(), but got true")
	}
}
