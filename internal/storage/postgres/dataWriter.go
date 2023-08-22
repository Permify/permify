package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/golang/protobuf/jsonpb"

	"github.com/Masterminds/squirrel"
	otelCodes "go.opentelemetry.io/otel/codes"

	"github.com/Permify/permify/internal/storage/postgres/snapshot"
	"github.com/Permify/permify/internal/storage/postgres/types"
	"github.com/Permify/permify/internal/storage/postgres/utils"
	"github.com/Permify/permify/pkg/database"
	db "github.com/Permify/permify/pkg/database/postgres"
	"github.com/Permify/permify/pkg/logger"
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
	// logger
	logger logger.Interface
}

func NewDataWriter(database *db.Postgres, logger logger.Interface) *DataWriter {
	return &DataWriter{
		database:        database,
		txOptions:       sql.TxOptions{Isolation: sql.LevelSerializable, ReadOnly: false},
		maxDataPerWrite: _defaultMaxDataPerWrite,
		maxRetries:      _defaultMaxRetries,
		logger:          logger,
	}
}

func (w *DataWriter) Write(ctx context.Context, tenantID string, tupleCollection *database.TupleCollection, attributeCollection *database.AttributeCollection) (token token.EncodedSnapToken, err error) {
	ctx, span := tracer.Start(ctx, "data-writer.write")
	defer span.End()

	if len(tupleCollection.GetTuples())+len(attributeCollection.GetAttributes()) > w.maxDataPerWrite {
		return nil, errors.New("max data per write exceeded")
	}

	for i := 0; i <= w.maxRetries; i++ {
		var tx *sql.Tx
		tx, err = w.database.DB.BeginTx(ctx, &w.txOptions)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(otelCodes.Error, err.Error())
			return nil, err
		}

		if len(tupleCollection.GetTuples()) > 0 {

			tuplesInsertBuilder := w.database.Builder.Insert(RelationTuplesTable).Columns("entity_type, entity_id, relation, subject_type, subject_id, subject_relation, tenant_id").Suffix("ON CONFLICT DO NOTHING")

			titer := tupleCollection.CreateTupleIterator()
			for titer.HasNext() {
				t := titer.GetNext()
				srelation := t.GetSubject().GetRelation()
				if srelation == tuple.ELLIPSIS {
					srelation = ""
				}
				tuplesInsertBuilder = tuplesInsertBuilder.Values(t.GetEntity().GetType(), t.GetEntity().GetId(), t.GetRelation(), t.GetSubject().GetType(), t.GetSubject().GetId(), srelation, tenantID)
			}

			var tquery string
			var targs []interface{}

			tquery, targs, err = tuplesInsertBuilder.ToSql()
			if err != nil {
				utils.Rollback(tx, w.logger)
				span.RecordError(err)
				span.SetStatus(otelCodes.Error, err.Error())
				return nil, errors.New(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String())
			}

			_, err = tx.ExecContext(ctx, tquery, targs...)
			if err != nil {
				utils.Rollback(tx, w.logger)
				span.RecordError(err)
				span.SetStatus(otelCodes.Error, err.Error())
				if strings.Contains(err.Error(), "could not serialize") {
					continue
				}
				return nil, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
			}

		}

		if len(attributeCollection.GetAttributes()) > 0 {

			attributesInsertBuilder := w.database.Builder.Insert(AttributesTable).Columns("entity_type, entity_id, attribute, type, value, tenant_id").Suffix("ON CONFLICT DO NOTHING")

			aiter := attributeCollection.CreateAttributeIterator()
			for aiter.HasNext() {
				a := aiter.GetNext()

				m := jsonpb.Marshaler{}
				jsonStr, err := m.MarshalToString(a.GetValue())
				if err != nil {
					utils.Rollback(tx, w.logger)
					span.RecordError(err)
					span.SetStatus(otelCodes.Error, err.Error())
					return nil, errors.New(base.ErrorCode_ERROR_CODE_INVALID_ARGUMENT.String())
				}

				attributesInsertBuilder = attributesInsertBuilder.Values(a.GetEntity().GetType(), a.GetEntity().GetId(), a.GetAttribute(), a.GetType(), jsonStr, tenantID)
			}

			var aquery string
			var aargs []interface{}

			aquery, aargs, err = attributesInsertBuilder.ToSql()
			if err != nil {
				utils.Rollback(tx, w.logger)
				span.RecordError(err)
				span.SetStatus(otelCodes.Error, err.Error())
				return nil, errors.New(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String())
			}

			_, err = tx.ExecContext(ctx, aquery, aargs...)
			if err != nil {
				utils.Rollback(tx, w.logger)
				span.RecordError(err)
				span.SetStatus(otelCodes.Error, err.Error())
				if strings.Contains(err.Error(), "could not serialize") {
					continue
				}
				return nil, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
			}

		}

		transaction := w.database.Builder.Insert("transactions").
			Columns("tenant_id").
			Values(tenantID).
			Suffix("RETURNING id").RunWith(tx)
		if err != nil {
			utils.Rollback(tx, w.logger)
			span.RecordError(err)
			span.SetStatus(otelCodes.Error, err.Error())
			return nil, errors.New(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String())
		}

		var xid types.XID8
		err = transaction.QueryRowContext(ctx).Scan(&xid)
		if err != nil {
			utils.Rollback(tx, w.logger)
			span.RecordError(err)
			span.SetStatus(otelCodes.Error, err.Error())
			return nil, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
		}

		if err = tx.Commit(); err != nil {
			utils.Rollback(tx, w.logger)
			span.RecordError(err)
			span.SetStatus(otelCodes.Error, err.Error())
			return nil, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
		}

		return snapshot.NewToken(xid).Encode(), nil
	}

	return nil, errors.New(base.ErrorCode_ERROR_CODE_ERROR_MAX_RETRIES.String())
}

func (w *DataWriter) Delete(ctx context.Context, tenantID string, tupleFilter *base.TupleFilter, attributeFilter *base.AttributeFilter) (token token.EncodedSnapToken, err error) {
	ctx, span := tracer.Start(ctx, "data-writer.delete")
	defer span.End()

	for i := 0; i <= w.maxRetries; i++ {
		var tx *sql.Tx
		tx, err = w.database.DB.BeginTx(ctx, &w.txOptions)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(otelCodes.Error, err.Error())
			return nil, err
		}

		tbuilder := w.database.Builder.Update(RelationTuplesTable).Set("expired_tx_id", squirrel.Expr("pg_current_xact_id()")).Where(squirrel.Eq{"expired_tx_id": "0"})
		tbuilder = utils.TuplesFilterQueryForUpdateBuilder(tbuilder, tupleFilter)

		var tquery string
		var targs []interface{}

		tquery, targs, err = tbuilder.ToSql()
		if err != nil {
			utils.Rollback(tx, w.logger)
			span.RecordError(err)
			span.SetStatus(otelCodes.Error, err.Error())
			return nil, errors.New(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String())
		}

		_, err = tx.ExecContext(ctx, tquery, targs...)
		if err != nil {
			utils.Rollback(tx, w.logger)
			span.RecordError(err)
			span.SetStatus(otelCodes.Error, err.Error())
			if strings.Contains(err.Error(), "could not serialize") {
				continue
			}
			return nil, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
		}

		abuilder := w.database.Builder.Update(AttributesTable).Set("expired_tx_id", squirrel.Expr("pg_current_xact_id()")).Where(squirrel.Eq{"expired_tx_id": "0"})
		abuilder = utils.AttributesFilterQueryForUpdateBuilder(tbuilder, attributeFilter)

		var aquery string
		var aargs []interface{}

		aquery, aargs, err = abuilder.ToSql()
		if err != nil {
			utils.Rollback(tx, w.logger)
			span.RecordError(err)
			span.SetStatus(otelCodes.Error, err.Error())
			return nil, errors.New(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String())
		}

		_, err = tx.ExecContext(ctx, aquery, aargs...)
		if err != nil {
			utils.Rollback(tx, w.logger)
			span.RecordError(err)
			span.SetStatus(otelCodes.Error, err.Error())
			if strings.Contains(err.Error(), "could not serialize") {
				continue
			}
			return nil, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
		}

		transaction := w.database.Builder.Insert("transactions").
			Columns("tenant_id").
			Values(tenantID).
			Suffix("RETURNING id").RunWith(tx)
		if err != nil {
			utils.Rollback(tx, w.logger)
			span.RecordError(err)
			span.SetStatus(otelCodes.Error, err.Error())
			return nil, errors.New(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String())
		}

		var xid types.XID8
		err = transaction.QueryRowContext(ctx).Scan(&xid)
		if err != nil {
			utils.Rollback(tx, w.logger)
			span.RecordError(err)
			span.SetStatus(otelCodes.Error, err.Error())
			return nil, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
		}

		if err = tx.Commit(); err != nil {
			utils.Rollback(tx, w.logger)
			span.RecordError(err)
			span.SetStatus(otelCodes.Error, err.Error())
			return nil, err
		}

		return snapshot.NewToken(xid).Encode(), nil
	}

	return nil, errors.New(base.ErrorCode_ERROR_CODE_ERROR_MAX_RETRIES.String())
}
