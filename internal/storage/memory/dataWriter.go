package memory

import (
	"context"
	"errors"
	"log/slog"
	"strings"
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
	database   *db.Memory
	maxRetries int
}

// NewDataWriter - Create a new DataWriter
func NewDataWriter(database *db.Memory) *DataWriter {
	return &DataWriter{
		database:   database,
		maxRetries: constants.DefaultMaxRetries,
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

// func (r *DataWriter) RunBundle(ctx context.Context, tenantID string, arguments map[string]string, b *base.DataBundle) (token token.EncodedSnapToken, err error) {
// 	for _, op := range b.GetOperations() {
// 		tupleCollection, attributeCollection, err := bundle.Operation(arguments, op)
// 		if err != nil {
// 			return nil, err
// 		}

// 		// Write operation
// 		if _, err = r.Write(ctx, tenantID, &tupleCollection.Write, &attributeCollection.Write); err != nil {
// 			return nil, err
// 		}

// 		// Delete operation
// 		// if _, err = r.Delete(ctx, tenantID, &tupleCollection.Delete, &attributeCollection.Delete); err != nil {
// 		// 	return nil, err
// 		// }
// 	}

// 	return snapshot.NewToken(time.Now()).Encode(), nil
// }

// RunBundle executes a bundle of operations in the context of a given tenant.
// It returns an EncodedSnapToken upon successful completion or an error if the operation fails.
func (w *DataWriter) RunBundle(
	ctx context.Context,
	tenantID string,
	arguments map[string]string,
	b *base.DataBundle,
) (token.EncodedSnapToken, error) {
	// Start a new tracing span for this operation.
	ctx, span := tracer.Start(ctx, "data-writer.run-bundle")
	defer span.End() // Ensure that the span is ended when the function returns.

	// Log the start of running a bundle operation.
	slog.Info("Running bundle from the database. TenantID: ", slog.String("tenant_id", tenantID), "Max Retries: ", slog.Any("max_retries", w.maxRetries))

	// Retry loop for handling transient errors like serialization issues.
	for i := 0; i <= w.maxRetries; i++ {
		// Attempt to run the bundle operation.
		tkn, err := w.runBundle(ctx, tenantID, arguments, b)
		if err != nil {
			// Check if the error is due to serialization, and if so, retry.
			if strings.Contains(err.Error(), "could not serialize") {
				slog.Warn("Serialization error occurred. Retrying...", slog.String("tenant_id", tenantID), slog.Int("retry", i))
				continue // Retry the operation.
			}
			// If the error is not serialization-related, handle it and return.
			return nil, utils.HandleError(span, err, base.ErrorCode_ERROR_CODE_DATASTORE)
		}
		// If the operation is successful, return the token.
		return tkn, err
	}

	// Log an error if the operation failed after reaching the maximum number of retries.
	slog.Error("Failed to run bundle from the database. Max retries reached. Aborting operation. ", slog.Any("error", errors.New(base.ErrorCode_ERROR_CODE_ERROR_MAX_RETRIES.String())))

	// Return an error indicating that the maximum number of retries has been reached.
	return nil, errors.New(base.ErrorCode_ERROR_CODE_ERROR_MAX_RETRIES.String())
}

// runBundle executes a series of operations defined in a DataBundle within a single database transaction.
// It returns an EncodedSnapToken upon successful execution of all operations or an error if any operation fails.
func (w *DataWriter) runBundle(
	ctx context.Context,
	tenantID string,
	arguments map[string]string,
	b *base.DataBundle,
) (token.EncodedSnapToken, error) {
	// Create a new write transaction
	txn := w.database.DB.Txn(true)
	defer txn.Abort()

	// Create a transaction object in memdb
	transaction := &storage.Transaction{
		ID:       int64(time.Now().UnixNano()),
		TenantID: tenantID,
	}
	if err := txn.Insert(constants.TransactionsTable, transaction); err != nil {
		return nil, err
	}

	xid := transaction.ID // Assuming xid is the ID of the transaction

	txid := time.Unix(0, xid)
	// Commit the transaction
	txn.Commit()

	for _, op := range b.GetOperations() {
		tb, ab, err := bundle.Operation(arguments, op)
		if err != nil {
			return nil, err
		}

		err = w.runOperation(ctx, xid, tenantID, tb, ab)
		if err != nil {
			return nil, err
		}
	}

	// Create a new write transaction for committing the changes
	txn = w.database.DB.Txn(true)
	defer txn.Abort()

	if err := txn.Insert(constants.TransactionsTable, transaction); err != nil {
		return nil, err
	}

	// Commit the final transaction
	txn.Commit()

	return snapshot.NewToken(txid).Encode(), nil
}

// runOperation processes and executes database operations defined in TupleBundle and AttributeBundle within a given transaction.
func (w *DataWriter) runOperation(
	ctx context.Context,
	xid int64,
	tenantID string,
	tb database.TupleBundle,
	ab database.AttributeBundle,
) error {
	slog.Debug("Processing bundles queries. ")
	transaction := w.database.DB.Txn(true)
	defer transaction.Abort() // Defer the
	if len(tb.Write.GetTuples()) > 0 {
		for titer := tb.Write.CreateTupleIterator(); titer.HasNext(); {
			t := titer.GetNext()
			srelation := t.GetSubject().GetRelation()
			if srelation == tuple.ELLIPSIS {
				srelation = ""
			}

			tupleData := &storage.RelationTuple{
				EntityID:        t.GetEntity().GetId(),
				EntityType:      t.GetEntity().GetType(),
				Relation:        t.GetRelation(),
				SubjectID:       t.GetSubject().GetId(),
				SubjectType:     t.GetSubject().GetType(),
				SubjectRelation: srelation,
				ID:              uint64(xid),
				TenantID:        tenantID,
			}

			// Insert the tuple into the RelationTuplesTable

			// Insert the tuple into the RelationTuplesTable
			if err := transaction.Insert(constants.RelationTuplesTable, tupleData); err != nil {
				return err
			}

			// Commit the transaction to persist the changes

		}
	}

	if len(ab.Write.GetAttributes()) > 0 {
		for aiter := ab.Write.CreateAttributeIterator(); aiter.HasNext(); {
			a := aiter.GetNext()

			attributeData := &storage.Attribute{
				EntityID:   a.GetEntity().GetId(),
				EntityType: a.GetEntity().GetType(),
				Attribute:  a.GetAttribute(),
				Value:      a.GetValue(),
				ID:         uint64(xid),
				TenantID:   tenantID,
			}

			if err := transaction.Insert(constants.AttributesTable, attributeData); err != nil {
				return err
			}

		}
	}

	if len(tb.Delete.GetTuples()) > 0 {
		for titer := tb.Delete.CreateTupleIterator(); titer.HasNext(); {
			t := titer.GetNext()
			srelation := t.GetSubject().GetRelation()
			if srelation == tuple.ELLIPSIS {
				srelation = ""
			}

			condition := func(tupleData interface{}) bool {
				tuple := tupleData.(*storage.RelationTuple)
				return (tuple.EntityID == t.GetEntity().GetId() &&
					tuple.EntityType == t.GetEntity().GetType() &&
					tuple.Relation == t.GetRelation() &&
					tuple.SubjectID == t.GetSubject().GetId() &&
					tuple.SubjectType == t.GetSubject().GetType() &&
					tuple.SubjectRelation == srelation) ||
					((srelation == "" && tuple.SubjectRelation == "") &&
						tuple.TenantID == tenantID &&
						tuple.ExpiredTxID == "0")
			}

			// Delete the matching tuples from the RelationTuplesTable

			if err := transaction.Delete(constants.RelationTuplesTable, condition); err != nil {
				return err
			}

		}
	}

	if len(ab.Delete.GetAttributes()) > 0 {
		for aiter := ab.Delete.CreateAttributeIterator(); aiter.HasNext(); {
			a := aiter.GetNext()

			// Define a condition to identify the attribute to be deleted
			condition := func(attributeData interface{}) bool {
				attribute := attributeData.(*storage.Attribute)
				return attribute.EntityID == a.GetEntity().GetId() &&
					attribute.EntityType == a.GetEntity().GetType() &&
					attribute.Attribute == a.GetAttribute() &&
					attribute.TenantID == tenantID &&
					attribute.ExpiredTxID == "0"
			}

			// Delete the matching attributes from the AttributesTable
			if err := transaction.Delete(constants.AttributesTable, condition); err != nil {
				return err
			}

		}
	}
	transaction.Commit()
	return nil
}
