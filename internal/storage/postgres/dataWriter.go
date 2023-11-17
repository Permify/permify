package postgres

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"strings"

	"github.com/golang/protobuf/jsonpb"

	"github.com/Masterminds/squirrel"
	otelCodes "go.opentelemetry.io/otel/codes"

	"github.com/Permify/permify/internal/storage/postgres/snapshot"
	"github.com/Permify/permify/internal/storage/postgres/types"
	"github.com/Permify/permify/internal/storage/postgres/utils"
	"github.com/Permify/permify/internal/validation"
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
	txOptions       sql.TxOptions
	maxDataPerWrite int
	maxRetries      int
}

func NewDataWriter(database *db.Postgres) *DataWriter {
	return &DataWriter{
		database:        database,
		txOptions:       sql.TxOptions{Isolation: sql.LevelSerializable, ReadOnly: false},
		maxDataPerWrite: _defaultMaxDataPerWrite,
		maxRetries:      _defaultMaxRetries,
	}
}

func (w *DataWriter) Write(ctx context.Context, tenantID string, tupleCollection *database.TupleCollection, attributeCollection *database.AttributeCollection) (token token.EncodedSnapToken, err error) {
	ctx, span := tracer.Start(ctx, "data-writer.write")
	defer span.End()

	slog.Info("Writing data to the database. TenantID: ", slog.String("tenant_id", tenantID), "Max Retries: ", slog.Any("max_retries", w.maxRetries))

	if len(tupleCollection.GetTuples())+len(attributeCollection.GetAttributes()) > w.maxDataPerWrite {
		return nil, errors.New("max data per write exceeded")
	}

	for i := 0; i <= w.maxRetries; i++ {
		var tx *sql.Tx
		tx, err = w.database.DB.BeginTx(ctx, &w.txOptions)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(otelCodes.Error, err.Error())

			slog.Error("Failed to begin transaction: ", slog.Any("error", err))

			return nil, err
		}

		slog.Debug("Inserting transaction record for tenant: ", slog.String("tenant_id", tenantID))

		transaction := w.database.Builder.Insert("transactions").
			Columns("tenant_id").
			Values(tenantID).
			Suffix("RETURNING id").RunWith(tx)
		if err != nil {
			utils.Rollback(tx)
			span.RecordError(err)
			span.SetStatus(otelCodes.Error, err.Error())

			slog.Error("Failed to insert transaction record for tenant: ", slog.Any("error", err))

			return nil, errors.New(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String())
		}

		var xid types.XID8
		err = transaction.QueryRowContext(ctx).Scan(&xid)
		if err != nil {
			utils.Rollback(tx)
			span.RecordError(err)
			span.SetStatus(otelCodes.Error, err.Error())

			slog.Error("Failed to query row context: ", slog.Any("error", err))

			return nil, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
		}

		slog.Debug("Retrieved transaction: ", slog.Any("transaction", transaction), "for tenant: ", slog.Any("tenant_id", tenantID))

		slog.Debug("Processing tuples and executing insert query. ")
		if len(tupleCollection.GetTuples()) > 0 {

			tuplesInsertBuilder := w.database.Builder.Insert(RelationTuplesTable).Columns("entity_type, entity_id, relation, subject_type, subject_id, subject_relation, created_tx_id, tenant_id")

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

				tuplesInsertBuilder = tuplesInsertBuilder.Values(t.GetEntity().GetType(), t.GetEntity().GetId(), t.GetRelation(), t.GetSubject().GetType(), t.GetSubject().GetId(), srelation, xid, tenantID)
			}

			tDeleteBuilder := w.database.Builder.Update(RelationTuplesTable).Set("expired_tx_id", xid).Where(squirrel.Eq{
				"expired_tx_id": "0",
				"tenant_id":     tenantID,
			}).Where(deleteClauses)

			var tdquery string
			var tdargs []interface{}

			tdquery, tdargs, err = tDeleteBuilder.ToSql()
			if err != nil {
				utils.Rollback(tx)
				span.RecordError(err)
				span.SetStatus(otelCodes.Error, err.Error())

				slog.Error("Failed to build SQL query for tuple deletion: ", slog.Any("error", err))

				return nil, errors.New(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String())
			}

			_, err = tx.ExecContext(ctx, tdquery, tdargs...)
			if err != nil {
				utils.Rollback(tx)
				span.RecordError(err)
				span.SetStatus(otelCodes.Error, err.Error())
				if strings.Contains(err.Error(), "could not serialize") {
					continue
				}

				slog.Error("Failed to execute context query: ", slog.Any("error", err))

				return nil, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
			}

			var tiquery string
			var tiargs []interface{}

			tiquery, tiargs, err = tuplesInsertBuilder.ToSql()
			if err != nil {
				utils.Rollback(tx)
				span.RecordError(err)
				span.SetStatus(otelCodes.Error, err.Error())

				slog.Error("Failed to build SQL query for tuples insert: ", slog.Any("error", err))

				return nil, errors.New(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String())
			}

			_, err = tx.ExecContext(ctx, tiquery, tiargs...)
			if err != nil {
				utils.Rollback(tx)
				span.RecordError(err)
				span.SetStatus(otelCodes.Error, err.Error())
				if strings.Contains(err.Error(), "could not serialize") {
					continue
				}

				slog.Error("Failed to execute context query: ", slog.Any("error", err))

				return nil, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
			}
		}

		if len(attributeCollection.GetAttributes()) > 0 {

			attributesInsertBuilder := w.database.Builder.Insert(AttributesTable).Columns("entity_type, entity_id, attribute, value, created_tx_id, tenant_id")

			deleteClauses := squirrel.Or{}

			aiter := attributeCollection.CreateAttributeIterator()
			for aiter.HasNext() {
				a := aiter.GetNext()

				m := jsonpb.Marshaler{}
				jsonStr, err := m.MarshalToString(a.GetValue())
				if err != nil {
					utils.Rollback(tx)
					span.RecordError(err)
					span.SetStatus(otelCodes.Error, err.Error())

					slog.Error("Failed to convert the value to string: ", slog.Any("error", err))

					return nil, errors.New(base.ErrorCode_ERROR_CODE_INVALID_ARGUMENT.String())
				}

				// Build the condition for this tuple.
				condition := squirrel.Eq{
					"entity_type": a.GetEntity().GetType(),
					"entity_id":   a.GetEntity().GetId(),
					"attribute":   a.GetAttribute(),
				}

				// Add the condition to the OR slice.
				deleteClauses = append(deleteClauses, condition)

				attributesInsertBuilder = attributesInsertBuilder.Values(a.GetEntity().GetType(), a.GetEntity().GetId(), a.GetAttribute(), jsonStr, xid, tenantID)
			}

			tDeleteBuilder := w.database.Builder.Update(AttributesTable).Set("expired_tx_id", xid).Where(squirrel.Eq{
				"expired_tx_id": "0",
				"tenant_id":     tenantID,
			}).Where(deleteClauses)

			var adquery string
			var adargs []interface{}

			adquery, adargs, err = tDeleteBuilder.ToSql()
			if err != nil {
				utils.Rollback(tx)
				span.RecordError(err)
				span.SetStatus(otelCodes.Error, err.Error())

				slog.Error("Failed to build SQL query for attribute delete: ", slog.Any("error", err))

				return nil, errors.New(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String())
			}

			_, err = tx.ExecContext(ctx, adquery, adargs...)
			if err != nil {
				utils.Rollback(tx)
				span.RecordError(err)
				span.SetStatus(otelCodes.Error, err.Error())
				if strings.Contains(err.Error(), "could not serialize") {
					continue
				}

				slog.Error("Failed to execute context query: ", slog.Any("error", err))

				return nil, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
			}

			var aquery string
			var aargs []interface{}

			aquery, aargs, err = attributesInsertBuilder.ToSql()
			if err != nil {
				utils.Rollback(tx)
				span.RecordError(err)
				span.SetStatus(otelCodes.Error, err.Error())

				slog.Error("Failed to build query for attribute insertion: ", slog.Any("error", err))

				return nil, errors.New(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String())
			}

			_, err = tx.ExecContext(ctx, aquery, aargs...)
			if err != nil {
				utils.Rollback(tx)
				span.RecordError(err)
				span.SetStatus(otelCodes.Error, err.Error())
				if strings.Contains(err.Error(), "could not serialize") {
					continue
				}
				slog.Error("Failed to execute context query: ", slog.Any("error", err))

				return nil, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
			}
		}

		if err = tx.Commit(); err != nil {
			utils.Rollback(tx)
			span.RecordError(err)
			span.SetStatus(otelCodes.Error, err.Error())
			if strings.Contains(err.Error(), "could not serialize") {
				continue
			}
			slog.Error("Failed to commiting database transaction: ", slog.Any("error", err))

			return nil, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
		}

		slog.Info("Data successfully written to the database.")

		return snapshot.NewToken(xid).Encode(), nil
	}

	slog.Error("Failed to write data to the database. Max retries reached. Aborting operation. ", slog.Any("error", errors.New(base.ErrorCode_ERROR_CODE_ERROR_MAX_RETRIES.String())))

	return nil, errors.New(base.ErrorCode_ERROR_CODE_ERROR_MAX_RETRIES.String())
}

func (w *DataWriter) Delete(ctx context.Context, tenantID string, tupleFilter *base.TupleFilter, attributeFilter *base.AttributeFilter) (token token.EncodedSnapToken, err error) {
	ctx, span := tracer.Start(ctx, "data-writer.delete")
	defer span.End()

	slog.Info("Deleting data from the database. TenantID: ", slog.String("tenant_id", tenantID), "Max Retries: ", slog.Any("max_retries", w.maxRetries))

	for i := 0; i <= w.maxRetries; i++ {
		var tx *sql.Tx
		tx, err = w.database.DB.BeginTx(ctx, &w.txOptions)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(otelCodes.Error, err.Error())

			slog.Error("Failed to begin transaction: ", slog.Any("error", err))

			return nil, err
		}

		slog.Debug("Deleting transaction record for tenant: ", slog.String("tenant_id", tenantID))

		transaction := w.database.Builder.Insert("transactions").
			Columns("tenant_id").
			Values(tenantID).
			Suffix("RETURNING id").RunWith(tx)
		if err != nil {
			utils.Rollback(tx)
			span.RecordError(err)
			span.SetStatus(otelCodes.Error, err.Error())

			slog.Error("Failed to insert transaction record for tenant: ", slog.Any("error", err))

			return nil, errors.New(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String())
		}

		var xid types.XID8
		err = transaction.QueryRowContext(ctx).Scan(&xid)
		if err != nil {
			utils.Rollback(tx)
			span.RecordError(err)
			span.SetStatus(otelCodes.Error, err.Error())

			slog.Error("Failed to query row context: ", slog.Any("error", err))

			return nil, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
		}

		slog.Debug("Retrieved transaction: ", slog.Any("transaction", transaction), "for tenant: ", slog.Any("tenant_id", tenantID))

		slog.Debug("Processing tuple and executing update query. ")

		if !validation.IsTupleFilterEmpty(tupleFilter) {
			tbuilder := w.database.Builder.Update(RelationTuplesTable).Set("expired_tx_id", xid).Where(squirrel.Eq{"expired_tx_id": "0", "tenant_id": tenantID})
			tbuilder = utils.TuplesFilterQueryForUpdateBuilder(tbuilder, tupleFilter)

			var tquery string
			var targs []interface{}

			tquery, targs, err = tbuilder.ToSql()
			if err != nil {
				utils.Rollback(tx)
				span.RecordError(err)
				span.SetStatus(otelCodes.Error, err.Error())

				slog.Error("Failed to build SQL query for tuple updation: ", slog.Any("error", err))

				return nil, errors.New(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String())
			}

			_, err = tx.ExecContext(ctx, tquery, targs...)
			if err != nil {
				utils.Rollback(tx)
				span.RecordError(err)
				span.SetStatus(otelCodes.Error, err.Error())
				if strings.Contains(err.Error(), "could not serialize") {
					continue
				}

				slog.Error("Failed to execute context query: ", slog.Any("error", err))

				return nil, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
			}
		}

		slog.Debug("Processing attribute and executing update query.")

		if !validation.IsAttributeFilterEmpty(attributeFilter) {
			abuilder := w.database.Builder.Update(AttributesTable).Set("expired_tx_id", xid).Where(squirrel.Eq{"expired_tx_id": "0", "tenant_id": tenantID})
			abuilder = utils.AttributesFilterQueryForUpdateBuilder(abuilder, attributeFilter)

			var aquery string
			var aargs []interface{}

			aquery, aargs, err = abuilder.ToSql()
			if err != nil {
				utils.Rollback(tx)
				span.RecordError(err)
				span.SetStatus(otelCodes.Error, err.Error())

				slog.Error("Failed to build SQL query for attribute updation: ", slog.Any("error", err))

				return nil, errors.New(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String())
			}

			_, err = tx.ExecContext(ctx, aquery, aargs...)
			if err != nil {
				utils.Rollback(tx)
				span.RecordError(err)
				span.SetStatus(otelCodes.Error, err.Error())
				if strings.Contains(err.Error(), "could not serialize") {
					continue
				}

				slog.Error("Failed to execute context query: ", slog.Any("error", err))

				return nil, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
			}
		}

		if err = tx.Commit(); err != nil {
			utils.Rollback(tx)
			span.RecordError(err)
			span.SetStatus(otelCodes.Error, err.Error())
			if strings.Contains(err.Error(), "could not serialize") {
				continue
			}

			slog.Error("Failed to commiting database transaction: ", slog.Any("error", err))

			return nil, err
		}

		slog.Info("Data successfully deleted from the database.")

		return snapshot.NewToken(xid).Encode(), nil
	}

	slog.Error("Failed to delete data from the database. Max retries reached. Aborting operation. ", slog.Any("error", errors.New(base.ErrorCode_ERROR_CODE_ERROR_MAX_RETRIES.String())))

	return nil, errors.New(base.ErrorCode_ERROR_CODE_ERROR_MAX_RETRIES.String())
}
