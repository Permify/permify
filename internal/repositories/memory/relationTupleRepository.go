package memory

import (
	"context"
	"github.com/hashicorp/go-memdb"

	"github.com/Permify/permify/internal/repositories/entities"
	"github.com/Permify/permify/internal/repositories/filters"
	"github.com/Permify/permify/pkg/database"
	db "github.com/Permify/permify/pkg/database/memory"
	"github.com/Permify/permify/pkg/errors"
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
func (r *RelationTupleRepository) Migrate() (err errors.Error) {
	return nil
}

// QueryTuples -
func (r *RelationTupleRepository) QueryTuples(ctx context.Context, namespace string, objectID string, relation string) (entities.RelationTuples, errors.Error) {
	var tuples entities.RelationTuples
	var err error
	txn := r.Database.DB.Txn(false)
	defer txn.Abort()

	var it memdb.ResultIterator
	it, err = txn.Get(entities.RelationTuple{}.Table(), "entity-index", namespace, objectID, relation)
	if err != nil {
		return tuples, errors.DatabaseError.SetSubKind(database.ErrExecution)
	}

	for obj := it.Next(); obj != nil; obj = it.Next() {
		tuples = append(tuples, obj.(entities.RelationTuple))
	}

	return tuples, nil
}

// Read -
func (r *RelationTupleRepository) Read(ctx context.Context, filter filters.RelationTupleFilter) (entities.RelationTuples, errors.Error) {
	var tuples entities.RelationTuples
	var err error
	txn := r.Database.DB.Txn(false)
	defer txn.Abort()

	filterFactory := func(filter filters.RelationTupleFilter) func(interface{}) bool {
		return func(raw interface{}) bool {
			obj, ok := raw.(entities.RelationTuple)
			if !ok {
				return true
			}

			if filter.Entity.ID != "" && filter.Entity.ID != obj.ObjectID {
				return true
			}

			if filter.Relation != "" && filter.Relation != obj.Relation {
				return true
			}

			if filter.Subject != (filters.SubjectFilter{}) {
				if filter.Subject.Type != "" && filter.Subject.Type != obj.UsersetEntity {
					return true
				}

				if filter.Subject.ID != "" && filter.Subject.ID != obj.UsersetObjectID {
					return true
				}

				if filter.Subject.Relation != "" && filter.Subject.Relation != obj.UsersetRelation {
					return true
				}
			}

			return false
		}
	}

	var it memdb.ResultIterator
	it, err = txn.Get(entities.RelationTuple{}.Table(), "entity", filter.Entity.Type)
	if err != nil {
		return tuples, errors.DatabaseError.SetSubKind(database.ErrExecution)
	}

	filtered := memdb.NewFilterIterator(it, filterFactory(filter))
	for obj := filtered.Next(); obj != nil; obj = it.Next() {
		tuples = append(tuples, obj.(entities.RelationTuple))
	}

	return tuples, nil
}

// Write -
func (r *RelationTupleRepository) Write(ctx context.Context, tuples entities.RelationTuples) errors.Error {
	var err error
	txn := r.Database.DB.Txn(true)
	defer txn.Abort()
	for _, tuple := range tuples {
		if err = txn.Insert(entities.RelationTuple{}.Table(), tuple); err != nil {
			return errors.DatabaseError.SetSubKind(database.ErrExecution)
		}
	}
	txn.Commit()
	return nil
}

// Delete -
func (r *RelationTupleRepository) Delete(ctx context.Context, tuples entities.RelationTuples) errors.Error {
	var err error
	txn := r.Database.DB.Txn(true)
	defer txn.Abort()
	for _, tuple := range tuples {
		if err = txn.Delete(entities.RelationTuple{}.Table(), tuple); err != nil {
			if err.Error() == "not found" {
				//errors.DatabaseError.SetSubKind(database.ErrRecordNotFound)
				return nil
			}
			return errors.DatabaseError.SetSubKind(database.ErrUniqueConstraint)
		}
	}
	txn.Commit()
	return nil
}
