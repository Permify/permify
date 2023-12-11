package memory

import (
	"context"
	"errors"
	"time"

	"github.com/hashicorp/go-memdb"

	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/internal/storage/memory/constants"
	"github.com/Permify/permify/internal/storage/memory/snapshot"
	"github.com/Permify/permify/internal/storage/memory/utils"
	"github.com/Permify/permify/pkg/bundle"
	"github.com/Permify/permify/pkg/database"
	db "github.com/Permify/permify/pkg/database/memory"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/token"
	"github.com/Permify/permify/pkg/tuple"
)

type DataWriter struct {
	database *db.Memory
}

// NewDataWriter - Create a new DataWriter
func NewDataWriter(database *db.Memory) *DataWriter {
	return &DataWriter{
		database: database,
	}
}

// WriteRelationships - Write a Relation to repository
func (r *DataWriter) Write(_ context.Context, tenantID string, tupleCollection *database.TupleCollection, attributesCollection *database.AttributeCollection) (token.EncodedSnapToken, error) {
	var err error

	tupleIterator := tupleCollection.CreateTupleIterator()
	attributeIterator := attributesCollection.CreateAttributeIterator()
	if !tupleIterator.HasNext() && !attributeIterator.HasNext() {
		return token.NewNoopToken().Encode(), nil
	}

	txn := r.database.DB.Txn(true)
	defer txn.Abort()

	for tupleIterator.HasNext() {
		bt := tupleIterator.GetNext()
		srelation := bt.GetSubject().GetRelation()
		if srelation == tuple.ELLIPSIS {
			srelation = ""
		}
		t := storage.RelationTuple{
			ID:              utils.RelationTuplesID.ID(),
			TenantID:        tenantID,
			EntityType:      bt.GetEntity().GetType(),
			EntityID:        bt.GetEntity().GetId(),
			Relation:        bt.GetRelation(),
			SubjectType:     bt.GetSubject().GetType(),
			SubjectID:       bt.GetSubject().GetId(),
			SubjectRelation: srelation,
		}
		if err = txn.Insert(constants.RelationTuplesTable, t); err != nil {
			return nil, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
		}
	}

	for attributeIterator.HasNext() {
		at := attributeIterator.GetNext()

		t := storage.Attribute{
			ID:         utils.AttributesID.ID(),
			TenantID:   tenantID,
			EntityType: at.GetEntity().GetType(),
			EntityID:   at.GetEntity().GetId(),
			Attribute:  at.GetAttribute(),
			Value:      at.GetValue(),
		}
		if err = txn.Insert(constants.AttributesTable, t); err != nil {
			return nil, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
		}
	}

	txn.Commit()
	return snapshot.NewToken(time.Now()).Encode(), nil
}

// Delete - Delete relationship from repository
func (r *DataWriter) Delete(_ context.Context, tenantID string, tupleFilter *base.TupleFilter, attributeFilter *base.AttributeFilter) (token.EncodedSnapToken, error) {
	var err error
	txn := r.database.DB.Txn(true)
	defer txn.Abort()

	tIndex, tArgs := utils.GetRelationTuplesIndexNameAndArgsByFilters(tenantID, tupleFilter)
	var tit memdb.ResultIterator
	tit, err = txn.Get(constants.RelationTuplesTable, tIndex, tArgs...)
	if err != nil {
		return nil, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
	}

	tFit := memdb.NewFilterIterator(tit, utils.FilterRelationTuplesQuery(tenantID, tupleFilter))
	for obj := tFit.Next(); obj != nil; obj = tFit.Next() {
		t, ok := obj.(storage.RelationTuple)
		if !ok {
			return nil, errors.New(base.ErrorCode_ERROR_CODE_TYPE_CONVERSATION.String())
		}
		err = txn.Delete(constants.RelationTuplesTable, t)
		if err != nil {
			return nil, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
		}
	}

	aIndex, args := utils.GetAttributesIndexNameAndArgsByFilters(tenantID, attributeFilter)
	var aIt memdb.ResultIterator
	aIt, err = txn.Get(constants.AttributesTable, aIndex, args...)
	if err != nil {
		return nil, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
	}

	fit := memdb.NewFilterIterator(aIt, utils.FilterAttributesQuery(tenantID, attributeFilter))
	for obj := fit.Next(); obj != nil; obj = fit.Next() {
		t, ok := obj.(storage.RelationTuple)
		if !ok {
			return nil, errors.New(base.ErrorCode_ERROR_CODE_TYPE_CONVERSATION.String())
		}
		err = txn.Delete(constants.RelationTuplesTable, t)
		if err != nil {
			return nil, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
		}
	}

	txn.Commit()
	return snapshot.NewToken(time.Now()).Encode(), nil
}

func (r *DataWriter) RunBundle(ctx context.Context, tenantID string, arguments map[string]string, b *base.DataBundle) (token token.EncodedSnapToken, err error) {
	for _, op := range b.GetOperations() {
		tupleCollection, attributeCollection, err := bundle.Operation(arguments, op)
		if err != nil {
			return nil, err
		}

		// Write operation
		if _, err = r.Write(ctx, tenantID, &tupleCollection.Write, &attributeCollection.Write); err != nil {
			return nil, err
		}

		// Delete operation
		// if _, err = r.Delete(ctx, tenantID, &tupleCollection.Delete, &attributeCollection.Delete); err != nil {
		// 	return nil, err
		// }
	}

	return snapshot.NewToken(time.Now()).Encode(), nil
}
