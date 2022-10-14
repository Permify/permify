package memory

import (
	"context"
	"errors"

	"github.com/hashicorp/go-memdb"

	"github.com/Permify/permify/internal/repositories"
	db "github.com/Permify/permify/pkg/database/memory"
	"github.com/Permify/permify/pkg/helper"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/tuple"
)

// RelationTupleRepository -.
type RelationTupleRepository struct {
	Database *db.Memory
}

// NewRelationTupleRepository -.
func NewRelationTupleRepository(mm *db.Memory) *RelationTupleRepository {
	return &RelationTupleRepository{mm}
}

// Migrate -
func (r *RelationTupleRepository) Migrate() (err error) {
	return nil
}

// ReverseQueryTuples -
func (r *RelationTupleRepository) ReverseQueryTuples(ctx context.Context, entity string, relation string, subjectEntity string, subjectIDs []string, subjectRelation string) (tuple.ITupleIterator, error) {
	var err error
	txn := r.Database.DB.Txn(false)
	defer txn.Abort()

	filterFactory := func(subjectRelation string, subjectIDs []string) func(interface{}) bool {
		return func(raw interface{}) bool {
			obj, ok := raw.(repositories.RelationTuple)
			if !ok {
				return true
			}

			if subjectRelation != "" && subjectRelation != obj.SubjectRelation {
				return true
			}

			if len(subjectIDs) > 0 && !helper.InArray(obj.SubjectID, subjectIDs) {
				return true
			}

			return false
		}
	}

	var it memdb.ResultIterator
	it, err = txn.Get("relation_tuple", "subject-index", entity, relation, subjectEntity)
	if err != nil {
		return nil, errors.New(base.ErrorCode_execution.String())
	}

	collection := tuple.NewTupleCollection()

	filtered := memdb.NewFilterIterator(it, filterFactory(subjectRelation, subjectIDs))
	for obj := filtered.Next(); obj != nil; obj = it.Next() {
		t := obj.(repositories.RelationTuple)
		collection.Add(t.ToTuple())
	}

	return collection.CreateTupleIterator(), nil
}

// QueryTuples -
func (r *RelationTupleRepository) QueryTuples(ctx context.Context, entity string, objectID string, relation string) (tuple.ITupleIterator, error) {
	var err error
	txn := r.Database.DB.Txn(false)
	defer txn.Abort()

	var it memdb.ResultIterator
	it, err = txn.Get("relation_tuple", "entity-index", entity, objectID, relation)
	if err != nil {
		return nil, errors.New(base.ErrorCode_execution.String())
	}

	collection := tuple.NewTupleCollection()

	for obj := it.Next(); obj != nil; obj = it.Next() {
		t := obj.(repositories.RelationTuple)
		collection.Add(t.ToTuple())
	}

	return collection.CreateTupleIterator(), nil
}

// Read -
func (r *RelationTupleRepository) Read(ctx context.Context, filter *base.TupleFilter) (tuple.ITupleCollection, error) {
	var err error
	txn := r.Database.DB.Txn(false)
	defer txn.Abort()

	filterFactory := func(filter *base.TupleFilter) func(interface{}) bool {
		return func(raw interface{}) bool {
			obj, ok := raw.(repositories.RelationTuple)
			if !ok {
				return true
			}

			if filter.GetEntity().GetId() != "" && filter.GetEntity().GetId() != obj.EntityID {
				return true
			}

			if filter.Relation != "" && filter.Relation != obj.Relation {
				return true
			}

			if filter.GetSubject() != nil {
				if filter.GetSubject().GetType() != "" && filter.GetSubject().GetType() != obj.SubjectEntity {
					return true
				}

				if filter.GetSubject().GetId() != "" && filter.GetSubject().GetId() != obj.SubjectID {
					return true
				}

				if filter.GetSubject().GetRelation() != "" && filter.GetSubject().GetRelation() != obj.SubjectRelation {
					return true
				}
			}

			return false
		}
	}

	var it memdb.ResultIterator
	it, err = txn.Get("relation_tuple", "entity", filter.Entity.Type)
	if err != nil {
		return nil, errors.New(base.ErrorCode_execution.String())
	}

	collection := tuple.NewTupleCollection()

	filtered := memdb.NewFilterIterator(it, filterFactory(filter))
	for obj := filtered.Next(); obj != nil; obj = it.Next() {
		t := obj.(repositories.RelationTuple)
		collection.Add(t.ToTuple())
	}

	return collection, nil
}

// Write -
func (r *RelationTupleRepository) Write(ctx context.Context, iterator tuple.ITupleIterator) error {
	var err error

	if !iterator.HasNext() {
		return nil
	}

	txn := r.Database.DB.Txn(true)
	defer txn.Abort()

	for iterator.HasNext() {
		bt := iterator.GetNext()
		t := repositories.RelationTuple{
			Entity:          bt.GetEntity().GetType(),
			EntityID:        bt.GetEntity().GetId(),
			Relation:        bt.GetRelation(),
			SubjectEntity:   bt.GetSubject().GetType(),
			SubjectID:       bt.GetSubject().GetId(),
			SubjectRelation: bt.GetSubject().GetRelation(),
		}
		if err = txn.Insert("relation_tuple", t); err != nil {
			return errors.New(base.ErrorCode_execution.String())
		}
	}

	txn.Commit()
	return nil
}

// Delete -
func (r *RelationTupleRepository) Delete(ctx context.Context, iterator tuple.ITupleIterator) error {
	if !iterator.HasNext() {
		return nil
	}

	var err error
	txn := r.Database.DB.Txn(true)
	defer txn.Abort()

	for iterator.HasNext() {
		bt := iterator.GetNext()
		t := repositories.RelationTuple{
			Entity:          bt.GetEntity().GetType(),
			EntityID:        bt.GetEntity().GetId(),
			Relation:        bt.GetRelation(),
			SubjectEntity:   bt.GetSubject().GetType(),
			SubjectID:       bt.GetSubject().GetId(),
			SubjectRelation: bt.GetSubject().GetRelation(),
		}
		if err = txn.Delete("relation_tuple", t); err != nil {
			if err.Error() == "not found" {
				// errors.DatabaseError.SetSubKind(database.ErrRecordNotFound)
				return nil
			}
			return errors.New(base.ErrorCode_unique_constraint.String())
		}
	}

	txn.Commit()
	return nil
}
