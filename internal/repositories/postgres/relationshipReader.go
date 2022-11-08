package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v4"

	"github.com/Permify/permify/internal/repositories"
	"github.com/Permify/permify/internal/repositories/postgres/builders"
	"github.com/Permify/permify/internal/repositories/postgres/snapshot"
	"github.com/Permify/permify/internal/repositories/postgres/types"
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
	tx, err := r.database.Pool.BeginTx(ctx, r.txOptions)
	if err != nil {
		return nil, err
	}

	var sql string
	var args []interface{}

	query := r.database.Builder.Select("entity_type, entity_id, relation, subject_type, subject_id, subject_relation").From(relationTuplesTable)
	query = builders.FilterQueryForSelectBuilder(query, filter)

	var st token.SnapToken
	st, err = r.snapToken(ctx, t)
	if err != nil {
		return nil, err
	}

	query = builders.SnapshotQuery(query, st.(snapshot.Token).Value.Uint)
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

// snapToken gets the token for a given snapshot
func (r *RelationshipReader) snapToken(ctx context.Context, token string) (token.SnapToken, error) {
	if token == "" {
		return r.headToken(ctx)
	}
	encoded := snapshot.EncodedToken{Value: token}
	return encoded.Decode()
}

// headToken gets the latest token
func (r *RelationshipReader) headToken(ctx context.Context) (token.SnapToken, error) {
	var xid types.XID8
	query := r.database.Builder.Select("id").From(transactionsTable).OrderBy("id DESC").Limit(1)
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
