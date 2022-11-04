package postgres

import (
	"context"
	"errors"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"

	"github.com/Permify/permify/internal/repositories/postgres/builders"
	"github.com/Permify/permify/internal/repositories/postgres/types"
	"github.com/Permify/permify/pkg/database"
	db "github.com/Permify/permify/pkg/database/postgres"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/token"
)

type RelationshipWriter struct {
	database  *db.Postgres
	txOptions pgx.TxOptions
}

// NewRelationshipWriter creates a new RelationTupleReader
func NewRelationshipWriter(database *db.Postgres) *RelationshipWriter {
	return &RelationshipWriter{
		database:  database,
		txOptions: pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadWrite},
	}
}

// WriteRelationships writes a collection of relationships to the database
func (w *RelationshipWriter) WriteRelationships(ctx context.Context, collection database.ITupleCollection) (token.SnapToken, error) {
	tx, err := w.database.Pool.BeginTx(ctx, w.txOptions)
	if err != nil {
		return token.SnapToken{}, err
	}

	batch := &pgx.Batch{}
	iter := collection.CreateTupleIterator()
	for iter.HasNext() {
		t := iter.GetNext()
		query, args, err := w.database.Builder.
			Insert(relationTuplesTable).
			Columns("entity_type, entity_id, relation, subject_type, subject_id, subject_relation").Values(t.GetEntity().GetType(), t.GetEntity().GetId(), t.GetRelation(), t.GetSubject().GetType(), t.GetSubject().GetId(), t.GetSubject().GetRelation()).ToSql()
		if err != nil {
			return token.SnapToken{}, errors.New(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String())
		}
		batch.Queue(query, args...)
	}

	var xid types.XID8
	err = tx.QueryRow(ctx, builders.NewTransactionQuery()).Scan(&xid)

	results := tx.SendBatch(ctx, batch)
	if err = results.Close(); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				return token.SnapToken{}, errors.New(base.ErrorCode_ERROR_CODE_UNIQUE_CONSTRAINT.String())
			default:
				return token.SnapToken{}, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
			}
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return token.SnapToken{}, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
	}

	return token.New(xid.Uint), nil
}

// DeleteRelationships deletes a collection of relationships to the database
func (w *RelationshipWriter) DeleteRelationships(ctx context.Context, filter *base.TupleFilter) (token.SnapToken, error) {
	tx, err := w.database.Pool.BeginTx(ctx, w.txOptions)
	if err != nil {
		return token.SnapToken{}, err
	}

	batch := &pgx.Batch{}
	builder := w.database.Builder.Update(relationTuplesTable).Set("expired_tx_id", squirrel.Expr("pg_current_xact_id()")).Where(squirrel.Eq{"expired_tx_id": "0"})
	builder = builders.FilterQueryForUpdateBuilder(builder, filter)

	query, args, err := builder.ToSql()
	batch.Queue(query, args...)

	var xid types.XID8
	err = tx.QueryRow(ctx, builders.NewTransactionQuery()).Scan(&xid)

	results := tx.SendBatch(ctx, batch)
	if err = results.Close(); err != nil {
		return token.SnapToken{}, err
	}

	if err = tx.Commit(ctx); err != nil {
		return token.SnapToken{}, err
	}

	return token.New(xid.Uint), nil
}
