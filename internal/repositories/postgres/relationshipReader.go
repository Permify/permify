package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v4"

	"github.com/Permify/permify/internal/repositories"
	"github.com/Permify/permify/internal/repositories/postgres/snapshot"
	"github.com/Permify/permify/internal/repositories/postgres/types"
	"github.com/Permify/permify/internal/repositories/postgres/utils"
	"github.com/Permify/permify/pkg/database"
	db "github.com/Permify/permify/pkg/database/postgres"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/token"
)

type RelationshipReader struct {
	database  *db.Postgres
	txOptions pgx.TxOptions
}

// NewRelationshipReader creates a new RelationshipReader
func NewRelationshipReader(database *db.Postgres) *RelationshipReader {
	return &RelationshipReader{
		database:  database,
		txOptions: pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly},
	}
}

// QueryRelationships gets all relationships for a given filter
func (r *RelationshipReader) QueryRelationships(ctx context.Context, filter *base.TupleFilter, t string) (database.ITupleCollection, error) {
	var err error

	var st token.SnapToken
	st, err = r.snapshotToken(ctx, t)
	if err != nil {
		return nil, err
	}

	var tx pgx.Tx
	tx, err = r.database.Pool.BeginTx(ctx, r.txOptions)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var sql string
	var args []interface{}

	query := r.database.Builder.Select("entity_type, entity_id, relation, subject_type, subject_id, subject_relation").From(RelationTuplesTable)
	query = utils.FilterQueryForSelectBuilder(query, filter)

	query = utils.SnapshotQuery(query, st.(snapshot.Token).Value.Uint)
	query = query.OrderBy("subject_type, subject_relation ASC")

	sql, args, err = query.ToSql()
	if err != nil {
		return nil, errors.New(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String())
	}

	var rows pgx.Rows
	rows, err = tx.Query(ctx, sql, args...)
	if err != nil {
		return nil, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
	}
	defer rows.Close()

	collection := database.NewTupleCollection()
	for rows.Next() {
		rt := repositories.RelationTuple{}
		err = rows.Scan(&rt.EntityType, &rt.EntityID, &rt.Relation, &rt.SubjectType, &rt.SubjectID, &rt.SubjectRelation)
		if err != nil {
			return nil, err
		}
		collection.Add(rt.ToTuple())
	}

	return collection, nil
}

// SnapshotToken gets the token for a given snapshot
func (r *RelationshipReader) snapshotToken(ctx context.Context, token string) (token.SnapToken, error) {
	encoded := snapshot.EncodedToken{Value: token}
	return encoded.Decode()
}

// HeadSnapshot gets the latest token
func (r *RelationshipReader) HeadSnapshot(ctx context.Context) (token.SnapToken, error) {
	var xid types.XID8
	query := r.database.Builder.Select("id").From(TransactionsTable).OrderBy("id DESC").Limit(1)
	sql, args, err := query.ToSql()
	if err != nil {
		return nil, errors.New(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String())
	}
	var row pgx.Row
	row = r.database.Pool.QueryRow(ctx, sql, args...)
	err = row.Scan(&xid)
	if err != nil {
		return nil, err
	}
	return snapshot.Token{Value: xid}, nil
}
