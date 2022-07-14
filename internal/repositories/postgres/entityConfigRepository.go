package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"

	"github.com/Permify/permify/internal/entities"
	"github.com/Permify/permify/internal/repositories"
	"github.com/Permify/permify/pkg/database/postgres"
)

// EntityConfigRepository -.
type EntityConfigRepository struct {
	Database *postgres.Postgres
}

// NewEntityConfigRepository -.
func NewEntityConfigRepository(pg *postgres.Postgres) *EntityConfigRepository {
	return &EntityConfigRepository{pg}
}

// All -
func (r *EntityConfigRepository) All(ctx context.Context) (configs []entities.EntityConfig, err error) {
	var sql string
	var args []interface{}
	sql, args, err = r.Database.Builder.
		Select("entity, serialized_config").From(entities.EntityConfig{}.Table()).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("EntityConfigRepo - AllEntityConfig - r.Builder: %w", err)
	}

	var rows pgx.Rows
	rows, err = r.Database.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("EntityConfigRepo - AllEntityConfig - r.Pool.Query: %w", err)
	}
	defer rows.Close()

	ent := make([]entities.EntityConfig, 0, _defaultEntityCap)

	for rows.Next() {
		e := entities.EntityConfig{}
		err = rows.Scan(&e.Entity, &e.SerializedConfig)
		if err != nil {
			return []entities.EntityConfig{}, fmt.Errorf("RelationTupleRepo - AllEntityConfig - rows.Scan: %w", err)
		}
		ent = append(ent, e)
	}

	return ent, nil
}

// First -
func (r *EntityConfigRepository) First(ctx context.Context, entity string) (config entities.EntityConfig, err error) {
	var sql string
	var args []interface{}
	sql, args, err = r.Database.Builder.
		Select("entity, serialized_config").From(entities.EntityConfig{}.Table()).Where(squirrel.Eq{"entity": entity}).
		ToSql()
	if err != nil {
		return entities.EntityConfig{}, fmt.Errorf("EntityConfigRepo - EntityConfig - r.Builder: %w", err)
	}

	row := r.Database.Pool.QueryRow(ctx, sql, args...)

	e := entities.EntityConfig{}

	switch err := row.Scan(&e.Entity, &e.SerializedConfig); err {
	case pgx.ErrNoRows:
		fmt.Println("No rows were returned!")
	default:
		return entities.EntityConfig{}, fmt.Errorf("EntityConfigRepo - Scan: %w", err)
	}

	return e, nil
}

// Replace -
func (r *EntityConfigRepository) Replace(ctx context.Context, configs []entities.EntityConfig) (err error) {
	if len(configs) < 1 {
		return nil
	}

	sql := r.Database.Builder.
		Insert(entities.EntityConfig{}.Table()).
		Columns("entity, serialized_config")

	for _, config := range configs {
		sql = sql.Values(config.Entity, config.SerializedConfig)
	}

	var query string
	var args []interface{}
	query, args, err = sql.ToSql()
	if err != nil {
		return fmt.Errorf("EntityConfigRepo - Replace - r.ToSql: %w", err)
	}

	_, err = r.Database.Pool.Exec(ctx, query, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				return repositories.ErrUniqueConstraint
			default:
				return fmt.Errorf("EntityConfigRepo - Replace - r.Pool.Exec: %w", err)
			}
		}
		return fmt.Errorf("EntityConfigRepo - Replace - r.Pool.Exec: %w", err)
	}

	return nil
}
