package postgres

import (
	"context"
	"errors"

	"github.com/Masterminds/squirrel"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"

	internalErrors "github.com/Permify/permify/internal/internal-errors"
	"github.com/Permify/permify/internal/repositories/entities"
	"github.com/Permify/permify/internal/repositories/postgres/migrations"
	"github.com/Permify/permify/pkg/database"
	db "github.com/Permify/permify/pkg/database/postgres"
)

// EntityConfigRepository -.
type EntityConfigRepository struct {
	Database *db.Postgres
}

// NewEntityConfigRepository -.
func NewEntityConfigRepository(pg *db.Postgres) *EntityConfigRepository {
	return &EntityConfigRepository{pg}
}

// Migrate -
func (r *EntityConfigRepository) Migrate() (err error) {
	ctx := context.Background()

	var tx pgx.Tx
	tx, err = r.Database.Pool.Begin(ctx)
	if err != nil {
		return err
	}

	_, err = tx.Exec(context.Background(), migrations.CreateEntityConfigMigration())
	if err != nil {
		return err
	}

	_, err = tx.Exec(context.Background(), migrations.CreateEntityConfigVersionField())
	if err != nil {
		return err
	}

	_, err = tx.Exec(context.Background(), migrations.CreateEntityConfigChangePrimaryKey())
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// All -
func (r *EntityConfigRepository) All(ctx context.Context, version string) (configs entities.EntityConfigs, err error) {
	if version == "" {
		version, err = r.findLastVersion(ctx)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return configs, internalErrors.EntityConfigCannotFoundError
			}
			return configs, err
		}
	}

	var sql string
	var args []interface{}
	sql, args, err = r.Database.Builder.
		Select("entity, serialized_config, version").From(entities.EntityConfig{}.Table()).Where(squirrel.Eq{"version": version}).
		ToSql()
	if err != nil {
		return nil, err
	}

	var rows pgx.Rows
	rows, err = r.Database.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ent := make([]entities.EntityConfig, 0, _defaultEntityCap)

	for rows.Next() {
		e := entities.EntityConfig{}
		err = rows.Scan(&e.Entity, &e.SerializedConfig, &e.Version)
		if err != nil {
			return []entities.EntityConfig{}, err
		}
		ent = append(ent, e)
	}

	return ent, nil
}

// Read -
func (r *EntityConfigRepository) Read(ctx context.Context, name string, version string) (config entities.EntityConfig, err error) {
	if version == "" {
		version, err = r.findLastVersion(ctx)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return config, internalErrors.EntityConfigCannotFoundError
			}
			return config, err
		}
	}
	var sql string
	var args []interface{}
	sql, args, err = r.Database.Builder.
		Select("entity, serialized_config, version").From(entities.EntityConfig{}.Table()).Where(squirrel.Eq{"entity": name, "version": version}).Limit(1).
		ToSql()
	if err != nil {
		return config, err
	}
	var row pgx.Row
	row = r.Database.Pool.QueryRow(ctx, sql, args...)
	err = row.Scan(&config.Entity, &config.SerializedConfig, &config.Version)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return config, internalErrors.EntityConfigCannotFoundError
		}
		return config, err
	}

	return config, nil
}

// Write -
func (r *EntityConfigRepository) Write(ctx context.Context, configs entities.EntityConfigs, version string) (err error) {
	if len(configs) < 1 {
		return nil
	}

	sql := r.Database.Builder.
		Insert(entities.EntityConfig{}.Table()).
		Columns("entity, serialized_config, version")

	for _, config := range configs {
		sql = sql.Values(config.Entity, config.SerializedConfig, version)
	}

	var query string
	var args []interface{}
	query, args, err = sql.ToSql()
	if err != nil {
		return err
	}

	_, err = r.Database.Pool.Exec(ctx, query, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				return database.ErrUniqueConstraint
			default:
				return err
			}
		}
		return err
	}

	return nil
}

// findLastVersion -
func (r *EntityConfigRepository) findLastVersion(ctx context.Context) (version string, err error) {
	var sql string
	var args []interface{}
	sql, args, err = r.Database.Builder.
		Select("version").From(entities.EntityConfig{}.Table()).OrderBy("version DESC").Limit(1).
		ToSql()
	if err != nil {
		return "", err
	}
	var row pgx.Row
	row = r.Database.Pool.QueryRow(ctx, sql, args...)
	err = row.Scan(&version)
	if err != nil {
		return version, err
	}
	return
}
