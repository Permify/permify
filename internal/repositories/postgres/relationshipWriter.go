package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/Masterminds/squirrel"
	otelCodes "go.opentelemetry.io/otel/codes"

	"github.com/Permify/permify/internal/repositories/postgres/snapshot"
	"github.com/Permify/permify/internal/repositories/postgres/types"
	"github.com/Permify/permify/internal/repositories/postgres/utils"
	"github.com/Permify/permify/pkg/database"
	db "github.com/Permify/permify/pkg/database/postgres"
	"github.com/Permify/permify/pkg/helper"
	"github.com/Permify/permify/pkg/logger"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/token"
)

// RelationshipWriter - Structure for Relationship Writer
type RelationshipWriter struct {
	database *db.Postgres
	// options
	txOptions         sql.TxOptions
	maxTuplesPerWrite int
	maxRetries        int
	// logger
	logger logger.Interface
}

// NewRelationshipWriter - Creates a new RelationTupleReader
func NewRelationshipWriter(database *db.Postgres, logger logger.Interface) *RelationshipWriter {
	return &RelationshipWriter{
		database:          database,
		txOptions:         sql.TxOptions{Isolation: sql.LevelSerializable, ReadOnly: false},
		maxTuplesPerWrite: _defaultMaxTuplesPerWrite,
		maxRetries:        _defaultMaxRetries,
		logger:            logger,
	}
}

// WriteRelationships - Writes a collection of relationships to the database
func (w *RelationshipWriter) WriteRelationships(ctx context.Context, tenantID uint64, collection *database.TupleCollection) (token token.EncodedSnapToken, err error) {
	ctx, span := tracer.Start(ctx, "relationship-writer.write-relationships")
	defer span.End()

	if len(collection.GetTuples()) > w.maxTuplesPerWrite {
		return nil, errors.New("")
	}

	for i := 0; i <= w.maxRetries; i++ {
		var tx *sql.Tx
		tx, err = w.database.DB.BeginTx(ctx, &w.txOptions)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(otelCodes.Error, err.Error())
			return nil, err
		}

		insertBuilder := w.database.Builder.Insert(RelationTuplesTable).Columns("entity_type, entity_id, relation, subject_type, subject_id, subject_relation, tenant_id")

		iter := collection.CreateTupleIterator()
		for iter.HasNext() {
			t := iter.GetNext()
			insertBuilder = insertBuilder.Values(t.GetEntity().GetType(), t.GetEntity().GetId(), t.GetRelation(), t.GetSubject().GetType(), t.GetSubject().GetId(), t.GetSubject().GetRelation(), tenantID)
		}

		var query string
		var args []interface{}

		query, args, err = insertBuilder.ToSql()
		if err != nil {
			utils.Rollback(tx, w.logger)
			span.RecordError(err)
			span.SetStatus(otelCodes.Error, err.Error())
			return nil, errors.New(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String())
		}

		_, err = tx.ExecContext(ctx, query, args...)
		if err != nil {
			utils.Rollback(tx, w.logger)
			span.RecordError(err)
			span.SetStatus(otelCodes.Error, err.Error())
			if strings.Contains(err.Error(), "could not serialize") {
				continue
			} else if strings.Contains(err.Error(), "duplicate key value") {
				return nil, errors.New(base.ErrorCode_ERROR_CODE_UNIQUE_CONSTRAINT.String())
			} else {
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
			helper.Pre(err)
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

// DeleteRelationships - Deletes a collection of relationships to the database
func (w *RelationshipWriter) DeleteRelationships(ctx context.Context, tenantID uint64, filter *base.TupleFilter) (token token.EncodedSnapToken, err error) {
	ctx, span := tracer.Start(ctx, "relationship-writer.delete-relationships")
	defer span.End()

	for i := 0; i <= w.maxRetries; i++ {
		var tx *sql.Tx
		tx, err = w.database.DB.BeginTx(ctx, &w.txOptions)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(otelCodes.Error, err.Error())
			return nil, err
		}

		builder := w.database.Builder.Update(RelationTuplesTable).Set("expired_tx_id", squirrel.Expr("pg_current_xact_id()")).Where(squirrel.Eq{"expired_tx_id": "0"})
		builder = utils.FilterQueryForUpdateBuilder(builder, filter)

		var query string
		var args []interface{}

		query, args, err = builder.ToSql()
		if err != nil {
			utils.Rollback(tx, w.logger)
			span.RecordError(err)
			span.SetStatus(otelCodes.Error, err.Error())
			return nil, errors.New(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String())
		}

		_, err = tx.ExecContext(ctx, query, args...)
		if err != nil {
			utils.Rollback(tx, w.logger)
			span.RecordError(err)
			span.SetStatus(otelCodes.Error, err.Error())
			if strings.Contains(err.Error(), "could not serialize") {
				continue
			} else {
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
			helper.Pre(err)
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
