package postgres

import (
	"context"
	`errors`
	"fmt"
	`github.com/jackc/pgconn`

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v4"

	"github.com/Permify/permify/internal/entities"
	`github.com/Permify/permify/internal/repositories`
	"github.com/Permify/permify/pkg/database/postgres"
)

// RelationTupleRepository -.
type RelationTupleRepository struct {
	Database *postgres.Postgres
}

// NewRelationTupleRepository -.
func NewRelationTupleRepository(pg *postgres.Postgres) *RelationTupleRepository {
	return &RelationTupleRepository{pg}
}

// QueryTuples -
func (r *RelationTupleRepository) QueryTuples(ctx context.Context, entity string, objectID string, relation string) (tuples []entities.RelationTuple, err error) {
	var sql string
	var args []interface{}
	sql, args, err = r.Database.Builder.
		Select("entity, object_id, relation, userset_entity, userset_object_id, userset_relation").From(entities.RelationTuple{}.Table()).Where(squirrel.Eq{"entity": entity, "object_id": objectID, "relation": relation}).OrderBy("userset_entity, userset_relation ASC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("RelationTupleRepo - QueryTuples - r.Builder: %w", err)
	}

	var rows pgx.Rows
	rows, err = r.Database.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("RelationTupleRepo - QueryTuples - r.Pool.Query: %w", err)
	}
	defer rows.Close()

	ent := make([]entities.RelationTuple, 0, _defaultEntityCap)

	for rows.Next() {
		e := entities.RelationTuple{}

		err = rows.Scan(&e.Entity, &e.ObjectID, &e.Relation, &e.UsersetEntity, &e.UsersetObjectID, &e.UsersetRelation)
		if err != nil {
			return []entities.RelationTuple{}, fmt.Errorf("RelationTupleRepo - QueryTuples - rows.Scan: %w", err)
		}

		ent = append(ent, e)
	}

	return ent, nil
}

// Write -.
func (r *RelationTupleRepository) Write(ctx context.Context, tuples []entities.RelationTuple) (err error) {

	if len(tuples) < 1 {
		return nil
	}

	sql := r.Database.Builder.
		Insert(entities.RelationTuple{}.Table()).
		Columns("entity, object_id, relation, userset_entity, userset_object_id, userset_relation, type")

	for _, entity := range tuples {
		sql = sql.Values(entity.Entity, entity.ObjectID, entity.Relation, entity.UsersetEntity, entity.UsersetObjectID, entity.UsersetRelation, entity.Type)
	}

	var query string
	var args []interface{}
	query, args, err = sql.ToSql()
	if err != nil {
		return fmt.Errorf("RelationTupleRepo - Write - r.ToSql: %w", err)
	}

	_, err = r.Database.Pool.Exec(ctx, query, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				return repositories.ErrUniqueConstraint
			default:
				return fmt.Errorf("RelationTupleRepo - Write - r.Pool.Exec: %w", err)
			}
		}

		return fmt.Errorf("RelationTupleRepo - Write - r.Pool.Exec: %w", err)
	}

	return nil
}

// Delete -.
func (r *RelationTupleRepository) Delete(ctx context.Context, tuples []entities.RelationTuple) error {

	for _, tuple := range tuples {
		sql, args, err := r.Database.Builder.
			Delete(entities.RelationTuple{}.Table()).Where(squirrel.Eq{"entity": tuple.Entity, "object_id": tuple.ObjectID, "relation": tuple.Relation, "userset_entity": tuple.UsersetEntity, "userset_object_id": tuple.UsersetObjectID, "userset_relation": tuple.UsersetRelation}).
			ToSql()
		if err != nil {
			return fmt.Errorf("RelationTupleRepo - Delete - r.Builder: %w", err)
		}

		_, err = r.Database.Pool.Exec(ctx, sql, args...)
		if err != nil {
			return fmt.Errorf("RelationTupleRepo - Delete - r.Pool.Exec: %w", err)
		}
	}

	return nil
}
