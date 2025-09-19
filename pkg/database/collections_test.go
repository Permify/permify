package database

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Permify/permify/pkg/attribute"
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

func TestAttributeCollection(t *testing.T) {
	// Create a new AttributeCollection with two attributes
	attr1, err := attribute.Attribute("book:b1$status|string:published")
	assert.NoError(t, err)

	attr2, err := attribute.Attribute("book:b2$status|string:draft")
	assert.NoError(t, err)

	attrColl := NewAttributeCollection(attr1, attr2)

	// Test the CreateAttributeIterator method
	iter := attrColl.CreateAttributeIterator()
	assert.NotNil(t, iter)

	// Test the GetAttributes method
	attributes := attrColl.GetAttributes()
	assert.Equal(t, 2, len(attributes))
	assert.Equal(t, attr1, attributes[0])
	assert.Equal(t, attr2, attributes[1])

	// Test the Add method
	attr3, err := attribute.Attribute("book:b3$status|string:archived")
	assert.NoError(t, err)

	attrColl.Add(attr3)
	attributes = attrColl.GetAttributes()
	assert.Equal(t, 3, len(attributes))
	assert.Equal(t, attr3, attributes[2])
}

func TestEmptyCollections(t *testing.T) {
	// Test empty TupleCollection
	emptyTupleColl := NewTupleCollection()
	assert.NotNil(t, emptyTupleColl)
	assert.Equal(t, 0, len(emptyTupleColl.GetTuples()))

	// Test empty SubjectCollection
	emptySubjectColl := NewSubjectCollection()
	assert.NotNil(t, emptySubjectColl)
	assert.Equal(t, 0, len(emptySubjectColl.GetSubjects()))

	// Test empty EntityCollection
	emptyEntityColl := NewEntityCollection()
	assert.NotNil(t, emptyEntityColl)
	assert.Equal(t, 0, len(emptyEntityColl.GetEntities()))

	// Test empty AttributeCollection
	emptyAttrColl := NewAttributeCollection()
	assert.NotNil(t, emptyAttrColl)
	assert.Equal(t, 0, len(emptyAttrColl.GetAttributes()))
}

func TestToSubjectCollection(t *testing.T) {
	// Create a TupleCollection with tuples
	tuple1 := &base.Tuple{
		Subject: &base.Subject{Type: "user", Id: "u1"},
		Entity:  &base.Entity{Type: "book", Id: "b1"},
	}
	tuple2 := &base.Tuple{
		Subject: &base.Subject{Type: "user", Id: "u2"},
		Entity:  &base.Entity{Type: "book", Id: "b2"},
	}
	tupleColl := NewTupleCollection(tuple1, tuple2)

	// Test ToSubjectCollection method
	subjectColl := tupleColl.ToSubjectCollection()
	assert.NotNil(t, subjectColl)
	assert.Equal(t, 2, len(subjectColl.GetSubjects()))
	assert.Equal(t, tuple1.Subject, subjectColl.GetSubjects()[0])
	assert.Equal(t, tuple2.Subject, subjectColl.GetSubjects()[1])

	// Test with empty tuple collection
	emptyTupleColl := NewTupleCollection()
	emptySubjectColl := emptyTupleColl.ToSubjectCollection()
	assert.NotNil(t, emptySubjectColl)
	assert.Equal(t, 0, len(emptySubjectColl.GetSubjects()))
}

func TestBundles(t *testing.T) {
	// Test TupleBundle
	tuple1 := &base.Tuple{
		Subject: &base.Subject{Type: "user", Id: "u1"},
		Entity:  &base.Entity{Type: "book", Id: "b1"},
	}
	tuple2 := &base.Tuple{
		Subject: &base.Subject{Type: "user", Id: "u2"},
		Entity:  &base.Entity{Type: "book", Id: "b2"},
	}

	tupleBundle := TupleBundle{
		Write:  *NewTupleCollection(tuple1),
		Delete: *NewTupleCollection(tuple2),
	}

	assert.NotNil(t, tupleBundle.Write)
	assert.NotNil(t, tupleBundle.Delete)
	assert.Equal(t, 1, len(tupleBundle.Write.GetTuples()))
	assert.Equal(t, 1, len(tupleBundle.Delete.GetTuples()))
	assert.Equal(t, tuple1, tupleBundle.Write.GetTuples()[0])
	assert.Equal(t, tuple2, tupleBundle.Delete.GetTuples()[0])

	// Test AttributeBundle
	attr1, err := attribute.Attribute("book:b1$status|string:published")
	assert.NoError(t, err)

	attr2, err := attribute.Attribute("book:b2$status|string:draft")
	assert.NoError(t, err)

	attrBundle := AttributeBundle{
		Write:  *NewAttributeCollection(attr1),
		Delete: *NewAttributeCollection(attr2),
	}

	assert.NotNil(t, attrBundle.Write)
	assert.NotNil(t, attrBundle.Delete)
	assert.Equal(t, 1, len(attrBundle.Write.GetAttributes()))
	assert.Equal(t, 1, len(attrBundle.Delete.GetAttributes()))
	assert.Equal(t, attr1, attrBundle.Write.GetAttributes()[0])
	assert.Equal(t, attr2, attrBundle.Delete.GetAttributes()[0])
}
