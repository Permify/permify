package memory

import (
	"context"
	"errors"
	`github.com/hashicorp/go-memdb`
	"time"

	"github.com/Permify/permify/internal/repositories"
	"github.com/Permify/permify/internal/repositories/memory/snapshot"
	"github.com/Permify/permify/internal/repositories/memory/utils"
	"github.com/Permify/permify/pkg/database"
	db "github.com/Permify/permify/pkg/database/memory"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/token"
)

type RelationshipWriter struct {
	database *db.Memory
}

// NewRelationshipWriter - Creates a new RelationshipReader
func NewRelationshipWriter(database *db.Memory) *RelationshipWriter {
	return &RelationshipWriter{
		database: database,
	}
}

// WriteRelationships - Write a Relation to repository
func (r *RelationshipWriter) WriteRelationships(ctx context.Context, collection database.ITupleCollection) (token.EncodedSnapToken, error) {
	var err error

	iterator := collection.CreateTupleIterator()
	if !iterator.HasNext() {
		return nil, nil
	}

	txn := r.database.DB.Txn(true)
	defer txn.Abort()

	for iterator.HasNext() {
		bt := iterator.GetNext()
		t := repositories.RelationTuple{
			EntityType:      bt.GetEntity().GetType(),
			EntityID:        bt.GetEntity().GetId(),
			Relation:        bt.GetRelation(),
			SubjectType:     bt.GetSubject().GetType(),
			SubjectID:       bt.GetSubject().GetId(),
			SubjectRelation: bt.GetSubject().GetRelation(),
		}
		if err = txn.Insert(RelationTuplesTable, t); err != nil {
			return nil, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
		}
	}

	txn.Commit()
	return snapshot.NewToken(time.Now()).Encode(), nil
}

// DeleteRelationships - Delete relationship from repository
func (r *RelationshipWriter) DeleteRelationships(ctx context.Context, filter *base.TupleFilter) (token.EncodedSnapToken, error) {
	var err error
	txn := r.database.DB.Txn(true)
	defer txn.Abort()

	index, args := utils.GetIndexNameAndArgsByFilters(filter)
	var it memdb.ResultIterator
	it, err = txn.Get(RelationTuplesTable, index, args...)
	if err != nil {
		return nil, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
	}

	fit := memdb.NewFilterIterator(it, utils.FilterQuery(filter))
	for obj := fit.Next(); obj != nil; obj = it.Next() {
		t := obj.(repositories.RelationTuple)
		err = txn.Delete(RelationTuplesTable, t)
		if err != nil {
			return nil, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
		}
	}

	txn.Commit()
	return snapshot.NewToken(time.Now()).Encode(), nil
}
