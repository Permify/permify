package postgres

import (
	"context"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5"

	"github.com/Masterminds/squirrel"
	"github.com/golang/protobuf/jsonpb"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/Permify/permify/internal/storage/postgres/snapshot"
	"github.com/Permify/permify/internal/storage/postgres/types"
	"github.com/Permify/permify/internal/storage/postgres/utils"
	"github.com/Permify/permify/internal/validation"
	"github.com/Permify/permify/pkg/bundle"
	"github.com/Permify/permify/pkg/database"
	db "github.com/Permify/permify/pkg/database/postgres"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/token"
	"github.com/Permify/permify/pkg/tuple"
)

// DataWriter - Structure for Data Writer
type DataWriter struct {
	database *db.Postgres
	// options
	txOptions pgx.TxOptions
}

func NewDataWriter(database *db.Postgres) *DataWriter {
	return &DataWriter{
		database:  database,
		txOptions: pgx.TxOptions{IsoLevel: pgx.Serializable, AccessMode: pgx.ReadWrite},
	}
}

// Write method writes a collection of tuples and attributes to the database for a specific tenant.
// It returns an EncodedSnapToken upon successful write or an error if the write fails.
func (w *DataWriter) Write(
	ctx context.Context,
	tenantID string,
	tupleCollection *database.TupleCollection,
	attributeCollection *database.AttributeCollection,
) (token token.EncodedSnapToken, err error) {
	// Start a new tracing span for this operation.
	ctx, span := tracer.Start(ctx, "data-writer.write")
	defer span.End() // Ensure that the span is ended when the function returns.

	// Log the start of a data write operation.
	slog.Debug("writing data for tenant_id", slog.String("tenant_id", tenantID), "max retries", slog.Any("max_retries", w.database.GetMaxRetries()))

	// Check if the total number of tuples and attributes exceeds the maximum allowed per write.
	if len(tupleCollection.GetTuples())+len(attributeCollection.GetAttributes()) > w.database.GetMaxDataPerWrite() {
		return nil, errors.New(base.ErrorCode_ERROR_CODE_MAX_DATA_PER_WRITE_EXCEEDED.String())
	}

	// Retry loop for handling transient errors like serialization issues.
	for i := 0; i <= w.database.GetMaxRetries(); i++ {
		// Attempt to write the data to the database.
		tkn, err := w.write(ctx, tenantID, tupleCollection, attributeCollection)
		if err != nil {
			// Check if the error is due to serialization, and if so, retry.
			if utils.IsSerializationRelatedError(err) || pgconn.SafeToRetry(err) {
				slog.Warn("serialization error occurred", slog.String("tenant_id", tenantID), slog.Int("retry", i))
				utils.WaitWithBackoff(ctx, tenantID, i)
				continue // Retry the operation.
			}
			// If the error is not serialization-related, handle it and return.
			return nil, utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_DATASTORE)
		}
		// If to write is successful, return the token.
		return tkn, nil
	}

	// Log an error if the operation failed after reaching the maximum number of retries.
	slog.Error("max retries reached", slog.Any("error", errors.New(base.ErrorCode_ERROR_CODE_ERROR_MAX_RETRIES.String())))

	// Return an error indicating that the maximum number of retries has been reached.
	return nil, errors.New(base.ErrorCode_ERROR_CODE_ERROR_MAX_RETRIES.String())
}

// write handles the database writing of tuple and attribute collections for a given tenant.
// It returns an EncodedSnapToken upon successful write or an error if the write fails.
func (w *DataWriter) write(
	ctx context.Context,
	tenantID string,
	tupleCollection *database.TupleCollection,
	attributeCollection *database.AttributeCollection,
) (token token.EncodedSnapToken, err error) {
	tx, err := w.database.WritePool.BeginTx(ctx, w.txOptions)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var xid types.XID8
	err = tx.QueryRow(ctx, utils.TransactionTemplate, tenantID).Scan(&xid)
	if err != nil {
		return nil, err
	}

	slog.Debug("retrieved transaction", slog.Any("xid", xid), "for tenant", slog.Any("tenant_id", tenantID))

	// Batch insert tuples
	if len(tupleCollection.GetTuples()) > 0 {
		batch := &pgx.Batch{}
		deleteClauses := squirrel.Or{}

		titer := tupleCollection.CreateTupleIterator()
		for titer.HasNext() {
			t := titer.GetNext()
			srelation := t.GetSubject().GetRelation()
			if srelation == tuple.ELLIPSIS {
				srelation = ""
			}

			// Build the condition for this tuple.
			condition := squirrel.Eq{
				"entity_type":      t.GetEntity().GetType(),
				"entity_id":        t.GetEntity().GetId(),
				"relation":         t.GetRelation(),
				"subject_type":     t.GetSubject().GetType(),
				"subject_id":       t.GetSubject().GetId(),
				"subject_relation": srelation,
			}

			// Add the condition to the OR slice.
			deleteClauses = append(deleteClauses, condition)

			// Queue the insert statement in the batch.
			batch.Queue(
				"INSERT INTO relation_tuples (entity_type, entity_id, relation, subject_type, subject_id, subject_relation, created_tx_id, tenant_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
				t.GetEntity().GetType(), t.GetEntity().GetId(), t.GetRelation(), t.GetSubject().GetType(), t.GetSubject().GetId(), srelation, xid, tenantID,
			)
		}

		// Execute the batch insert.
		br := tx.SendBatch(ctx, batch)
		defer br.Close()
		_, err = br.Exec()
		if err != nil {
			return nil, err
		}

		// Handle tuple deletions.
		tDeleteBuilder := w.database.Builder.Update(RelationTuplesTable).Set("expired_tx_id", xid).Where(squirrel.Eq{
			"expired_tx_id": "0",
			"tenant_id":     tenantID,
		}).Where(deleteClauses)

		tdquery, tdargs, err := tDeleteBuilder.ToSql()
		if err != nil {
			return nil, err
		}
		_, err = tx.Exec(ctx, tdquery, tdargs...)
		if err != nil {
			return nil, err
		}
	}

	// Batch insert attributes
	if len(attributeCollection.GetAttributes()) > 0 {
		batch := &pgx.Batch{}
		deleteClauses := squirrel.Or{}