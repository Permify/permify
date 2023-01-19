package postgres

import (
	"context"
	"database/sql"
	"errors"

	"go.opentelemetry.io/otel/codes"

	"github.com/Permify/permify/internal/repositories"
	"github.com/Permify/permify/internal/repositories/postgres/snapshot"
	"github.com/Permify/permify/internal/repositories/postgres/types"
	"github.com/Permify/permify/internal/repositories/postgres/utils"
	"github.com/Permify/permify/pkg/database"
	db "github.com/Permify/permify/pkg/database/postgres"
	"github.com/Permify/permify/pkg/logger"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/token"
)

type RelationshipReader struct {
	database *db.Postgres
	// options
	txOptions sql.TxOptions
	// logger
	logger logger.Interface
}

// NewRelationshipReader - Creates a new RelationshipReader
func NewRelationshipReader(database *db.Postgres, logger logger.Interface) *RelationshipReader {
	return &RelationshipReader{
		database:  database,
		txOptions: sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true},
		logger:    logger,
	}
}

// QueryRelationships - Gets all relationships for a given filter
func (r *RelationshipReader) QueryRelationships(ctx context.Context, filter *base.TupleFilter, t string) (tuples database.ITupleCollection, err error) {
	ctx, span := tracer.Start(ctx, "relationship-reader.query-relationships")
	defer span.End()

	var st token.SnapToken
	st, err = snapshot.EncodedToken{Value: t}.Decode()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	var tx *sql.Tx
	tx, err = r.database.DB.BeginTx(ctx, &r.txOptions)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	defer utils.Rollback(tx, r.logger)

	var args []interface{}

	builder := r.database.Builder.Select("entity_type, entity_id, relation, subject_type, subject_id, subject_relation").From(RelationTuplesTable)
	builder = utils.FilterQueryForSelectBuilder(builder, filter)

	builder = utils.SnapshotQuery(builder, st.(snapshot.Token).Value.Uint)

	var query string
	query, args, err = builder.ToSql()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, errors.New(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String())
	}

	var rows *sql.Rows
	rows, err = tx.QueryContext(ctx, query, args...)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
	}
	defer rows.Close()

	collection := database.NewTupleCollection()
	for rows.Next() {
		rt := repositories.RelationTuple{}
		err = rows.Scan(&rt.EntityType, &rt.EntityID, &rt.Relation, &rt.SubjectType, &rt.SubjectID, &rt.SubjectRelation)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return nil, err
		}
		collection.Add(rt.ToTuple())
	}
	if err = rows.Err(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	return collection, nil
}

// GetUniqueEntityIDsByEntityType - Gets all unique entity ids for a given entity type
func (r *RelationshipReader) GetUniqueEntityIDsByEntityType(ctx context.Context, typ, t string) (ids []string, err error) {
	ctx, span := tracer.Start(ctx, "relationship-reader.get-unique-entity-ids-by-entity-type")
	defer span.End()

	var st token.SnapToken
	st, err = snapshot.EncodedToken{Value: t}.Decode()
	if err != nil {
		return nil, err
	}

	var tx *sql.Tx
	tx, err = r.database.DB.BeginTx(ctx, &r.txOptions)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	defer utils.Rollback(tx, r.logger)

	var args []interface{}

	builder := r.database.Builder.Select("entity_id").Distinct().From(RelationTuplesTable).Where("entity_type = ?", typ)
	builder = utils.SnapshotQuery(builder, st.(snapshot.Token).Value.Uint)

	var query string
	query, args, err = builder.ToSql()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, errors.New(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String())
	}

	var rows *sql.Rows
	rows, err = tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
	}
	defer rows.Close()

	var result []string
	for rows.Next() {
		var id string
		err = rows.Scan(&id)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return nil, err
		}
		result = append(result, id)
	}
	if err = rows.Err(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	return result, nil
}

// HeadSnapshot - Gets the latest token
func (r *RelationshipReader) HeadSnapshot(ctx context.Context) (token.SnapToken, error) {
	ctx, span := tracer.Start(ctx, "relationship-reader.head-snapshot")
	defer span.End()

	var xid types.XID8
	builder := r.database.Builder.Select("id").From(TransactionsTable).OrderBy("id DESC").Limit(1)
	query, args, err := builder.ToSql()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, errors.New(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String())
	}

	row := r.database.DB.QueryRowContext(ctx, query, args...)
	err = row.Scan(&xid)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	return snapshot.Token{Value: xid}, nil
}
