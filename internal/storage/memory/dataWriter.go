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
	"github.com/Permify/permify/internal/validation"
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
func (w *DataWriter) Write(_ context.Context, tenantID string, tupleCollection *database.TupleCollection, attributesCollection *database.AttributeCollection) (token.EncodedSnapToken, error) {
	var err error

	tupleIterator := tupleCollection.CreateTupleIterator()
	attributeIterator := attributesCollection.CreateAttributeIterator()
	if !tupleIterator.HasNext() && !attributeIterator.HasNext() {
		return token.NewNoopToken().Encode(), nil
	}

	txn := w.database.DB.Txn(true)
	defer txn.Abort()

	for tupleIterator.HasNext() {
		bt := tupleIterator.GetNext()
		srelation := bt.GetSubject().GetRelation()
		if srelation == tuple.ELLIPSIS {
			srelation = ""
		}
		if err = txn.Insert(constants.RelationTuplesTable, storage.RelationTuple{
			ID:              w.database.RelationTupleID(),
			TenantID:        tenantID,
			EntityType:      bt.GetEntity().GetType(),
			EntityID:        bt.GetEntity().GetId(),
			Relation:        bt.GetRelation(),
			SubjectType:     bt.GetSubject().GetType(),
			SubjectID:       bt.GetSubject().GetId(),
			SubjectRelation: srelation,
		}); err != nil {
			return nil, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
		}
	}

	for attributeIterator.HasNext() {
		at := attributeIterator.GetNext()
		if err = txn.Insert(constants.AttributesTable, storage.Attribute{
			ID:         w.database.AttributeID(),
			TenantID:   tenantID,
			EntityType: at.GetEntity().GetType(),
			EntityID:   at.GetEntity().GetId(),
			Attribute:  at.GetAttribute(),
			Value:      at.GetValue(),
		}); err != nil {
			return nil, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
		}
	}

	txn.Commit()
	return snapshot.NewToken(time.Now()).Encode(), nil
}

// Delete - Delete relationship from repository
func (w *DataWriter) Delete(_ context.Context, tenantID string, tupleFilter *base.TupleFilter, attributeFilter *base.AttributeFilter) (token.EncodedSnapToken, error) {
	var err error
	txn := w.database.DB.Txn(true)
	defer txn.Abort()

	if !validation.IsTupleFilterEmpty(tupleFilter) {
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
	}

	if !validation.IsAttributeFilterEmpty(attributeFilter) {
		aIndex, args := utils.GetAttributesIndexNameAndArgsByFilters(tenantID, attributeFilter)
		var aIt memdb.ResultIterator
		aIt, err = txn.Get(constants.AttributesTable, aIndex, args...)
		if err != nil {
			return nil, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
		}

		fit := memdb.NewFilterIterator(aIt, utils.FilterAttributesQuery(tenantID, attributeFilter))
		for obj := fit.Next(); obj != nil; obj = fit.Next() {
			t, ok := obj.(storage.Attribute)
			if !ok {
				return nil, errors.New(base.ErrorCode_ERROR_CODE_TYPE_CONVERSATION.String())
			}
			err = txn.Delete(constants.AttributesTable, t)
			if err != nil {
				return nil, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
			}
		}
	}

	txn.Commit()
	return snapshot.NewToken(time.Now()).Encode(), nil
}

// RunBundle executes a bundle of operations in the context of a given tenant.
// It returns an EncodedSnapToken upon successful completion or an error if the operation fails.
func (w *DataWriter) RunBundle(
	ctx context.Context,
	tenantID string,
	arguments map[string]string,
	b *base.DataBundle,
) (token.EncodedSnapToken, error) {
	txn := w.database.DB.Txn(true)
	defer txn.Abort()

	for _, op := range b.GetOperations() {
		tb, ab, err := bundle.Operation(arguments, op)
		if err != nil {
			return nil, err
		}

		err = w.runOperation(ctx, txn, tenantID, tb, ab)
		if err != nil {
			return nil, err
		}
	}

	// Commit the final transaction
	txn.Commit()

	return snapshot.NewToken(time.Now()).Encode(), nil
}

// runOperation processes and executes database operations defined in TupleBundle and AttributeBundle within a given transaction.
func (w *DataWriter) runOperation(
	ctx context.Context,
	txn *memdb.Txn,
	tenantID string,
	tb database.TupleBundle,
	ab database.AttributeBundle,
) (err error) {
	if len(tb.Write.GetTuples()) > 0 {
		for titer := tb.Write.CreateTupleIterator(); titer.HasNext(); {
			t := titer.GetNext()
			srelation := t.GetSubject().GetRelation()
			if srelation == tuple.ELLIPSIS {
				srelation = ""
			}

			// Insert the tuple into the RelationTuplesTable
			if err := txn.Insert(constants.RelationTuplesTable, storage.RelationTuple{
				ID:              w.database.RelationTupleID(),
				EntityID:        t.GetEntity().GetId(),
				EntityType:      t.GetEntity().GetType(),
				Relation:        t.GetRelation(),
				SubjectID:       t.GetSubject().GetId(),
				SubjectType:     t.GetSubject().GetType(),
				SubjectRelation: srelation,
				TenantID:        tenantID,
			}); err != nil {
				return err
			}
		}
	}

	if len(ab.Write.GetAttributes()) > 0 {
		for aiter := ab.Write.CreateAttributeIterator(); aiter.HasNext(); {
			a := aiter.GetNext()
			if err := txn.Insert(constants.AttributesTable, storage.Attribute{
				ID:         w.database.AttributeID(),
				EntityID:   a.GetEntity().GetId(),
				EntityType: a.GetEntity().GetType(),
				Attribute:  a.GetAttribute(),
				Value:      a.GetValue(),
				TenantID:   tenantID,
			}); err != nil {
				return err
			}
		}
	}

	if len(tb.Delete.GetTuples()) > 0 {
		for titer := tb.Delete.CreateTupleIterator(); titer.HasNext(); {
			next := titer.GetNext()
			tupleFilter := &base.TupleFilter{
				Entity: &base.EntityFilter{
					Type: next.GetEntity().GetType(),
					Ids:  []string{next.GetEntity().GetId()},
				},
				Relation: next.GetRelation(),
				Subject: &base.SubjectFilter{
					Type:     next.GetSubject().GetType(),
					Ids:      []string{next.GetSubject().GetId()},
					Relation: next.GetSubject().GetRelation(),
				},
			}

			tIndex, tArgs := utils.GetRelationTuplesIndexNameAndArgsByFilters(tenantID, tupleFilter)
			var tit memdb.ResultIterator
			tit, err = txn.Get(constants.RelationTuplesTable, tIndex, tArgs...)
			if err != nil {
				return errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
			}

			tFit := memdb.NewFilterIterator(tit, utils.FilterRelationTuplesQuery(tenantID, tupleFilter))
			for obj := tFit.Next(); obj != nil; obj = tFit.Next() {
				t, ok := obj.(storage.RelationTuple)
				if !ok {
					return errors.New(base.ErrorCode_ERROR_CODE_TYPE_CONVERSATION.String())
				}
				err = txn.Delete(constants.RelationTuplesTable, t)
				if err != nil {
					return errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
				}
			}
		}
	}

	if len(ab.Delete.GetAttributes()) > 0 {
		for aiter := ab.Delete.CreateAttributeIterator(); aiter.HasNext(); {
			next := aiter.GetNext()
			attributeFilter := &base.AttributeFilter{
				Entity: &base.EntityFilter{
					Type: next.GetEntity().GetType(),
					Ids:  []string{next.GetEntity().GetId()},
				},
				Attributes: []string{next.GetAttribute()},
			}

			aIndex, args := utils.GetAttributesIndexNameAndArgsByFilters(tenantID, attributeFilter)
			var aIt memdb.ResultIterator
			aIt, err = txn.Get(constants.AttributesTable, aIndex, args...)
			if err != nil {
				return errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
			}

			fit := memdb.NewFilterIterator(aIt, utils.FilterAttributesQuery(tenantID, attributeFilter))
			for obj := fit.Next(); obj != nil; obj = fit.Next() {
				t, ok := obj.(storage.Attribute)
				if !ok {
					return errors.New(base.ErrorCode_ERROR_CODE_TYPE_CONVERSATION.String())
				}
				err = txn.Delete(constants.AttributesTable, t)
				if err != nil {
					return errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
				}
			}
		}
	}

	return nil
}
