package postgres

import (
	"context"
	"errors"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	otelCodes "go.opentelemetry.io/otel/codes"

	"github.com/Permify/permify/internal/repositories/postgres/snapshot"
	"github.com/Permify/permify/internal/repositories/postgres/types"
	"github.com/Permify/permify/internal/repositories/postgres/utils"
	"github.com/Permify/permify/pkg/database"
	db "github.com/Permify/permify/pkg/database/postgres"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/token"
)

// RelationshipWriter - Structure for Relationship Writer
type RelationshipWriter struct {
	database  *db.Postgres
	txOptions pgx.TxOptions
}

// NewRelationshipWriter - Creates a new RelationTupleReader
func NewRelationshipWriter(database *db.Postgres) *RelationshipWriter {
	return &RelationshipWriter{
		database:  database,
		txOptions: pgx.TxOptions{IsoLevel: pgx.Serializable, AccessMode: pgx.ReadWrite},
	}
}

// WriteRelationships - Writes a collection of relationships to the database
func (w *RelationshipWriter) WriteRelationships(ctx context.Context, collection database.ITupleCollection) (token.EncodedSnapToken, error) {
	ctx, span := tracer.Start(ctx, "relationship-writer.write-relationships")
	defer span.End()

	for i := 0; i <= 10; i++ {
		tx, bErr := w.database.Pool.BeginTx(ctx, w.txOptions)
		if bErr != nil {
			return nil, bErr
		}

		batch := &pgx.Batch{}
		iter := collection.CreateTupleIterator()
		for iter.HasNext() {
			t := iter.GetNext()
			query, args, err := w.database.Builder.
				Insert(RelationTuplesTable).
				Columns("entity_type, entity_id, relation, subject_type, subject_id, subject_relation").Values(t.GetEntity().GetType(), t.GetEntity().GetId(), t.GetRelation(), t.GetSubject().GetType(), t.GetSubject().GetId(), t.GetSubject().GetRelation()).ToSql()
			if err != nil {
				return nil, errors.New(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String())
			}
			batch.Queue(query, args...)
		}

		var xid types.XID8
		err := tx.QueryRow(ctx, utils.NewTransactionQuery()).Scan(&xid)
		if err != nil {
			_ = tx.Rollback(ctx)
			return nil, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
		}

		results := tx.SendBatch(ctx, batch)
		if err = results.Close(); err != nil {
			_ = tx.Rollback(ctx)
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				switch pgErr.Code {
				case "40001":
					continue
				case "23505":
					return nil, errors.New(base.ErrorCode_ERROR_CODE_UNIQUE_CONSTRAINT.String())
				default:
					return nil, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
				}
			}
		}

		if err = tx.Commit(ctx); err != nil {
			return nil, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
		}

		return snapshot.NewToken(xid).Encode(), nil
	}

	return nil, errors.New(base.ErrorCode_ERROR_CODE_ERROR_MAX_RETRIES.String())
}

// DeleteRelationships - Deletes a collection of relationships to the database
func (w *RelationshipWriter) DeleteRelationships(ctx context.Context, filter *base.TupleFilter) (token.EncodedSnapToken, error) {
	ctx, span := tracer.Start(ctx, "relationship-writer.delete-relationships")
	defer span.End()

	for i := 0; i <= 10; i++ {
		tx, err := w.database.Pool.BeginTx(ctx, w.txOptions)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(otelCodes.Error, err.Error())
			return nil, err
		}

		batch := &pgx.Batch{}
		builder := w.database.Builder.Update(RelationTuplesTable).Set("expired_tx_id", squirrel.Expr("pg_current_xact_id()")).Where(squirrel.Eq{"expired_tx_id": "0"})
		builder = utils.FilterQueryForUpdateBuilder(builder, filter)

		query, args, err := builder.ToSql()
		if err != nil {
			span.RecordError(err)
			span.SetStatus(otelCodes.Error, err.Error())
			return nil, errors.New(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String())
		}
		batch.Queue(query, args...)

		var xid types.XID8
		err = tx.QueryRow(ctx, utils.NewTransactionQuery()).Scan(&xid)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(otelCodes.Error, err.Error())
			_ = tx.Rollback(ctx)
			return nil, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
		}

		results := tx.SendBatch(ctx, batch)
		if err = results.Close(); err != nil {
			span.RecordError(err)
			span.SetStatus(otelCodes.Error, err.Error())
			_ = tx.Rollback(ctx)
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				switch pgErr.Code {
				case "40001":
					continue
				default:
					return nil, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
				}
			}
		}

		if err = tx.Commit(ctx); err != nil {
			return nil, err
		}

		return snapshot.NewToken(xid).Encode(), nil
	}

	return nil, errors.New(base.ErrorCode_ERROR_CODE_ERROR_MAX_RETRIES.String())
}
