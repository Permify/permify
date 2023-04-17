package database

import (
	"testing"

	"github.com/stretchr/testify/assert"

	base "github.com/Permify/permify/pkg/pb/base/v1"
)

func TestTupleCollection(t *testing.T) {
	// Create a new TupleCollection with two tuples
	tuple1 := &base.Tuple{
		Subject: &base.Subject{Type: "user", Id: "u1"},
		Entity:  &base.Entity{Type: "book", Id: "b1"},
	}
	tuple2 := &base.Tuple{
		Subject: &base.Subject{Type: "user", Id: "u2"},
		Entity:  &base.Entity{Type: "book", Id: "b2"},
	}
	tupleColl := NewTupleCollection(tuple1, tuple2)

	// Test the CreateTupleIterator method
	iter := tupleColl.CreateTupleIterator()
	assert.NotNil(t, iter)

	// Test the GetTuples method
	tuples := tupleColl.GetTuples()
	assert.Equal(t, 2, len(tuples))
	assert.Equal(t, tuple1, tuples[0])
	assert.Equal(t, tuple2, tuples[1])

	// Test the Add method
	tuple3 := &base.Tuple{
		Subject: &base.Subject{Type: "user", Id: "u3"},
		Entity:  &base.Entity{Type: "book", Id: "b3"},
	}
	tupleColl.Add(tuple3)
	tuples = tupleColl.GetTuples()
	assert.Equal(t, 3, len(tuples))
	assert.Equal(t, tuple3, tuples[2])
}

func TestSubjectCollection(t *testing.T) {
	// Create a new SubjectCollection with two subjects
	subject1 := &base.Subject{Type: "user", Id: "u1"}
	subject2 := &base.Subject{Type: "user", Id: "u2"}
	subjectColl := NewSubjectCollection(subject1, subject2)

	// Test the CreateSubjectIterator method
	iter := subjectColl.CreateSubjectIterator()
	assert.NotNil(t, iter)

	// Test the GetSubjects method
	subjects := subjectColl.GetSubjects()
	assert.Equal(t, 2, len(subjects))
	assert.Equal(t, subject1, subjects[0])
	assert.Equal(t, subject2, subjects[1])

	// Test the Add method
	subject3 := &base.Subject{Type: "user", Id: "u3"}
	subjectColl.Add(subject3)
	subjects = subjectColl.GetSubjects()
	assert.Equal(t, 3, len(subjects))
	assert.Equal(t, subject3, subjects[2])
}

func TestEntityCollection(t *testing.T) {
	// Create a new EntityCollection with two entities
	entity1 := &base.Entity{Type: "book", Id: "b1"}
	entity2 := &base.Entity{Type: "book", Id: "b2"}
	entityColl := NewEntityCollection(entity1, entity2)

	// Test the CreateEntityIterator method
	iter := entityColl.CreateEntityIterator()
	assert.NotNil(t, iter)

	// Test the GetEntities method
	entities := entityColl.GetEntities()
	assert.Equal(t, 2, len(entities))
	assert.Equal(t, entity1, entities[0])
	assert.Equal(t, entity2, entities[1])

	// Test the Add method
	entity3 := &base.Entity{Type: "book", Id: "b3"}
	entityColl.Add(entity3)
	entities = entityColl.GetEntities()
	assert.Equal(t, 3, len(entities))
	assert.Equal(t, entity3, entities[2])
}
