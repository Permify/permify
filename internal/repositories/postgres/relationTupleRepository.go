package postgres

import (
	"context"
	e "errors"

	"github.com/jackc/pgconn"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v4"

	"github.com/Permify/permify/internal/repositories/entities"
	"github.com/Permify/permify/internal/repositories/filters"
	"github.com/Permify/permify/internal/repositories/postgres/migrations"
	"github.com/Permify/permify/pkg/database"
	db "github.com/Permify/permify/pkg/database/postgres"
	"github.com/Permify/permify/pkg/errors"
)

// RelationTupleRepository -.
type RelationTupleRepository struct {
	Database *db.Postgres
}

// NewRelationTupleRepository -.
func NewRelationTupleRepository(pg *db.Postgres) *RelationTupleRepository {
	return &RelationTupleRepository{pg}
}

// Migrate -
func (r *RelationTupleRepository) Migrate() errors.Error {
	var err error
	ctx := context.Background()

	var tx pgx.Tx
	tx, err = r.Database.Pool.Begin(ctx)
	if err != nil {
		return errors.NewError(errors.Database).SetSubKind(database.ErrMigration)
	}

	_, err = tx.Exec(context.Background(), migrations.CreateRelationTupleMigration())
	if err != nil {
		return errors.NewError(errors.Database).SetSubKind(database.ErrMigration)
	}

	_, err = tx.Exec(context.Background(), migrations.DropRelationTupleTypeColumnIfExistMigration())
	if err != nil {
		return errors.NewError(errors.Database).SetSubKind(database.ErrMigration)
	}

	_, err = tx.Exec(context.Background(), migrations.CreateRelationTupleUserSetIndexIfNotExistMigration())
	if err != nil {
		return errors.NewError(errors.Database).SetSubKind(database.ErrMigration)
	}

	_, err = tx.Exec(context.Background(), migrations.CreateRelationTupleUserSetRelationIndexIfNotExistMigration())
	if err != nil {
		return errors.NewError(errors.Database).SetSubKind(database.ErrMigration)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return errors.NewError(errors.Database).SetSubKind(database.ErrMigration)
	}

	return nil
}

// QueryTuples -
func (r *RelationTupleRepository) QueryTuples(ctx context.Context, entity string, objectID string, relation string) (entities.RelationTuples, errors.Error) {
	var err error
	var sql string
	var args []interface{}
	sql, args, err = r.Database.Builder.
		Select("entity, object_id, relation, userset_entity, userset_object_id, userset_relation").From(entities.RelationTuple{}.Table()).Where(squirrel.Eq{"entity": entity, "object_id": objectID, "relation": relation}).OrderBy("userset_entity, userset_relation ASC").
		ToSql()
	if err != nil {
		return nil, errors.NewError(errors.Database).SetSubKind(database.ErrBuilder)
	}

	var rows pgx.Rows
	rows, err = r.Database.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, errors.NewError(errors.Database).SetSubKind(database.ErrExecution)
	}
	defer rows.Close()

	ent := make([]entities.RelationTuple, 0, _defaultEntityCap)

	for rows.Next() {
		rt := entities.RelationTuple{}
		err = rows.Scan(&rt.Entity, &rt.ObjectID, &rt.Relation, &rt.UsersetEntity, &rt.UsersetObjectID, &rt.UsersetRelation)
		if err != nil {
			return nil, errors.NewError(errors.Database).SetMessage(err.Error())
		}
		ent = append(ent, rt)
	}

	return ent, nil
}

// Read -.
func (r *RelationTupleRepository) Read(ctx context.Context, filter filters.RelationTupleFilter) (entities.RelationTuples, errors.Error) {
	var err error
	var sql string

	eq := squirrel.Eq{}
	eq["entity"] = filter.Entity.Type

	if filter.Entity.ID != "" {
		eq["object_id"] = filter.Entity.ID
	}

	if filter.Relation != "" {
		eq["relation"] = filter.Relation
	}

	if filter.Subject.Type != "" {
		eq["userset_entity"] = filter.Subject.Type
	}

	if filter.Subject.ID != "" {
		eq["userset_object_id"] = filter.Subject.ID
	}

	if filter.Subject.Relation != "" {
		eq["userset_relation"] = filter.Subject.Relation
	}

	var args []interface{}
	sql, args, err = r.Database.Builder.
		Select("entity, object_id, relation, userset_entity, userset_object_id, userset_relation, commit_time").From(entities.RelationTuple{}.Table()).Where(eq).OrderBy("userset_entity, userset_relation ASC").
		ToSql()
	if err != nil {
		return nil, errors.NewError(errors.Database).SetSubKind(database.ErrBuilder)
	}

	var rows pgx.Rows
	rows, err = r.Database.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, errors.NewError(errors.Database).SetSubKind(database.ErrExecution)
	}
	defer rows.Close()

	ent := make([]entities.RelationTuple, 0, _defaultEntityCap)

	for rows.Next() {
		e := entities.RelationTuple{}

		err = rows.Scan(&e.Entity, &e.ObjectID, &e.Relation, &e.UsersetEntity, &e.UsersetObjectID, &e.UsersetRelation, &e.CommitTime)
		if err != nil {
			return ent, errors.NewError(errors.Database).SetSubKind(database.ErrScan)
		}

		ent = append(ent, e)
	}

	return ent, nil
}

// Write -.
func (r *RelationTupleRepository) Write(ctx context.Context, tuples entities.RelationTuples) errors.Error {
	var err error

	if len(tuples) < 1 {
		return nil
	}

	sql := r.Database.Builder.
		Insert(entities.RelationTuple{}.Table()).
		Columns("entity, object_id, relation, userset_entity, userset_object_id, userset_relation")

	for _, entity := range tuples {
		sql = sql.Values(entity.Entity, entity.ObjectID, entity.Relation, entity.UsersetEntity, entity.UsersetObjectID, entity.UsersetRelation)
	}

	var query string
	var args []interface{}
	query, args, err = sql.ToSql()
	if err != nil {
		return errors.NewError(errors.Database).SetSubKind(database.ErrBuilder)
	}

	_, err = r.Database.Pool.Exec(ctx, query, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if e.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				return errors.NewError(errors.Database).SetSubKind(database.ErrUniqueConstraint)
			default:
				return errors.NewError(errors.Database).SetSubKind(database.ErrExecution)
			}
		}
		return errors.NewError(errors.Database).SetMessage(err.Error())
	}

	return nil
}

// Delete -.
func (r *RelationTupleRepository) Delete(ctx context.Context, tuples entities.RelationTuples) errors.Error {
	tx, err := r.Database.Pool.Begin(ctx)
	if err != nil {
		return errors.NewError(errors.Database).SetMessage(err.Error())
	}
	batch := &pgx.Batch{}
	for _, tuple := range tuples {
		sql, args, err := r.Database.Builder.
			Delete(entities.RelationTuple{}.Table()).Where(squirrel.Eq{"entity": tuple.Entity, "object_id": tuple.ObjectID, "relation": tuple.Relation, "userset_entity": tuple.UsersetEntity, "userset_object_id": tuple.UsersetObjectID, "userset_relation": tuple.UsersetRelation}).
			ToSql()
		if err != nil {
			return errors.NewError(errors.Database).SetSubKind(database.ErrBuilder)
		}
		batch.Queue(sql, args...)
	}
	results := tx.SendBatch(ctx, batch)
	if err = results.Close(); err != nil {
		return errors.NewError(errors.Database).SetMessage(err.Error())
	}
	if err = tx.Commit(ctx); err != nil {
		return errors.NewError(errors.Database).SetSubKind(database.ErrExecution)
	}
	return nil
}
