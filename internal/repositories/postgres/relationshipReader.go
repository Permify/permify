package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v4"

	otelCodes "go.opentelemetry.io/otel/codes"

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

// NewRelationshipReader - Creates a new RelationshipReader
func NewRelationshipReader(database *db.Postgres) *RelationshipReader {
	return &RelationshipReader{
		database:  database,
		txOptions: pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly},
	}
}

// QueryRelationships - Gets all relationships for a given filter
func (r *RelationshipReader) QueryRelationships(ctx context.Context, filter *base.TupleFilter, t string) (database.ITupleCollection, error) {
	ctx, span := tracer.Start(ctx, "relationship-reader.query-relationships")
	defer span.End()

	var err error
	var st token.SnapToken
	st, err = snapshot.EncodedToken{Value: t}.Decode()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		return nil, err
	}

	var tx pgx.Tx
	tx, err = r.database.Pool.BeginTx(ctx, r.txOptions)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
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
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
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
			span.RecordError(err)
			span.SetStatus(otelCodes.Error, err.Error())
			return nil, err
		}
		collection.Add(rt.ToTuple())
	}

	return collection, nil
}

// GetUniqueEntityIDsByEntityType - Gets all unique entity ids for a given entity type
func (r *RelationshipReader) GetUniqueEntityIDsByEntityType(ctx context.Context, typ, t string) (ids []string, err error) {
	ctx, span := tracer.Start(ctx, "relationship-reader.get-unique-entity-ids-by-entity_type")
	defer span.End()

	var st token.SnapToken
	st, err = snapshot.EncodedToken{Value: t}.Decode()
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

	query := r.database.Builder.Select("DISTINCT entity_id").From(RelationTuplesTable).Where("entity_type = ?", typ)
	query = utils.SnapshotQuery(query, st.(snapshot.Token).Value.Uint)

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

	var result []string
	for rows.Next() {
		var id string
		err = rows.Scan(&id)
		if err != nil {
			return nil, err
		}
		result = append(result, id)
	}

	return result, nil
}

// HeadSnapshot - Gets the latest token
func (r *RelationshipReader) HeadSnapshot(ctx context.Context) (token.SnapToken, error) {
	ctx, span := tracer.Start(ctx, "relationship-reader.head-snapshot")
	defer span.End()

	var xid types.XID8
	query := r.database.Builder.Select("id").From(TransactionsTable).OrderBy("id DESC").Limit(1)
	sql, args, err := query.ToSql()
	if err != nil {
		return nil, errors.New(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String())
	}
	row := r.database.Pool.QueryRow(ctx, sql, args...)
	err = row.Scan(&xid)
	if err != nil {
		return nil, err
	}
	return snapshot.Token{Value: xid}, nil
}
