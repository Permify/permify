package memory

import (
	"context"
	"errors"

	"github.com/Permify/permify/internal/repositories"
	"github.com/Permify/permify/pkg/database"
	db "github.com/Permify/permify/pkg/database/memory"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/token"
)

type RelationshipWriter struct {
	database *db.Memory
}

// NewRelationshipWriter creates a new RelationshipReader
func NewRelationshipWriter(database *db.Memory) *RelationshipWriter {
	return &RelationshipWriter{
		database: database,
	}
}

// WriteRelationships -
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
		if err = txn.Insert(relationTuplesTable, t); err != nil {
			return nil, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
		}
	}

	txn.Commit()
	return nil, nil
}

// DeleteRelationships -
func (r *RelationshipWriter) DeleteRelationships(ctx context.Context, filter *base.TupleFilter) (token.EncodedSnapToken, error) {
	//iterator := collection.CreateTupleIterator()
	//if !iterator.HasNext() {
	//	return nil
	//}
	//
	//var err error
	//txn := r.Database.DB.Txn(true)
	//defer txn.Abort()
	//
	//for iterator.HasNext() {
	//	bt := iterator.GetNext()
	//	t := repositories.RelationTuple{
	//		EntityType:      bt.GetEntity().GetType(),
	//		EntityID:        bt.GetEntity().GetId(),
	//		Relation:        bt.GetRelation(),
	//		SubjectType:     bt.GetSubject().GetType(),
	//		SubjectID:       bt.GetSubject().GetId(),
	//		SubjectRelation: bt.GetSubject().GetRelation(),
	//	}
	//	if err = txn.Delete(relationTuplesTable, t); err != nil {
	//		return snapshot.SnapToken{}, errors.New(base.ErrorCode_ERROR_CODE_UNIQUE_CONSTRAINT.String())
	//	}
	//}

	// txn.Commit()
	return nil, nil
}
