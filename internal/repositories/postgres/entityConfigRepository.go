package postgres

import (
	"context"
	e "errors"

	"github.com/pkg/errors"

	"github.com/Masterminds/squirrel"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"

	"github.com/Permify/permify/internal/repositories"
	"github.com/Permify/permify/internal/repositories/postgres/migrations"
	db "github.com/Permify/permify/pkg/database/postgres"
	base "github.com/Permify/permify/pkg/pb/base/v1"
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
func (r *EntityConfigRepository) Migrate() error {
	var err error
	ctx := context.Background()

	var tx pgx.Tx
	tx, err = r.Database.Pool.Begin(ctx)
	if err != nil {
		return errors.New(base.ErrorCode_migration.String())
	}

	_, err = tx.Exec(context.Background(), migrations.CreateEntityConfigMigration())
	if err != nil {
		return errors.New(base.ErrorCode_migration.String())
	}

	_, err = tx.Exec(context.Background(), migrations.CreateEntityConfigVersionField())
	if err != nil {
		return errors.New(base.ErrorCode_migration.String())
	}

	_, err = tx.Exec(context.Background(), migrations.CreateEntityConfigChangePrimaryKey())
	if err != nil {
		return errors.New(base.ErrorCode_migration.String())
	}

	err = tx.Commit(ctx)
	if err != nil {
		return errors.New(base.ErrorCode_migration.String())
	}

	return nil
}

// All -
func (r *EntityConfigRepository) All(ctx context.Context, version string) ([]repositories.EntityConfig, error) {
	var err error
	if version == "" {
		version, err = r.findLastVersion(ctx)
		if err != nil {
			return nil, err
		}
	}

	var sql string
	var args []interface{}
	sql, args, err = r.Database.Builder.
		Select("entity, serialized_config, version").From("entity_config").Where(squirrel.Eq{"version": version}).
		ToSql()
	if err != nil {
		return nil, errors.New(base.ErrorCode_sql_builder_error.String())
	}

	var rows pgx.Rows
	rows, err = r.Database.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, errors.New(base.ErrorCode_execution.String())
	}
	defer rows.Close()

	ent := make([]repositories.EntityConfig, 0, _defaultEntityCap)

	for rows.Next() {
		c := repositories.EntityConfig{}
		err = rows.Scan(&c.Entity, &c.SerializedConfig, &c.Version)
		if err != nil {
			return nil, errors.New(base.ErrorCode_scan.String())
		}
		ent = append(ent, c)
	}

	return ent, nil
}

// Read -
func (r *EntityConfigRepository) Read(ctx context.Context, name string, version string) (repositories.EntityConfig, error) {
	var config repositories.EntityConfig
	var err error
	if version == "" {
		version, err = r.findLastVersion(ctx)
		if err != nil {
			return config, err
		}
	}
	var sql string
	var args []interface{}
	sql, args, err = r.Database.Builder.
		Select("entity, serialized_config, version").From("entity_config").Where(squirrel.Eq{"entity": name, "version": version}).Limit(1).
		ToSql()
	if err != nil {
		return config, errors.New(base.ErrorCode_sql_builder_error.String())
	}
	var row pgx.Row
	row = r.Database.Pool.QueryRow(ctx, sql, args...)
	err = row.Scan(&config.Entity, &config.SerializedConfig, &config.Version)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return config, errors.New(base.ErrorCode_schema_not_found.String())
		}
		return config, err
	}

	return config, nil
}

// Write -
func (r *EntityConfigRepository) Write(ctx context.Context, configs []repositories.EntityConfig, version string) error {
	var err error
	if len(configs) < 1 {
		return nil
	}

	sql := r.Database.Builder.
		Insert("entity_config").
		Columns("entity, serialized_config, version")

	for _, config := range configs {
		sql = sql.Values(config.Entity, config.SerializedConfig, version)
	}

	var query string
	var args []interface{}
	query, args, err = sql.ToSql()
	if err != nil {
		return errors.New(base.ErrorCode_sql_builder_error.String())
	}

	_, err = r.Database.Pool.Exec(ctx, query, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if e.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				return errors.New(base.ErrorCode_unique_constraint.String())
			default:
				return err
			}
		}
		return err
	}

	return nil
}

// findLastVersion -
func (r *EntityConfigRepository) findLastVersion(ctx context.Context) (string, error) {
	var version string
	var err error
	var sql string
	var args []interface{}
	sql, args, err = r.Database.Builder.
		Select("version").From("entity_config").OrderBy("version DESC").Limit(1).
		ToSql()
	if err != nil {
		return "", errors.New(base.ErrorCode_sql_builder_error.String())
	}
	var row pgx.Row
	row = r.Database.Pool.QueryRow(ctx, sql, args...)
	err = row.Scan(&version)
	if err != nil {
		if e.Is(err, pgx.ErrNoRows) {
			return version, errors.New(base.ErrorCode_schema_not_found.String())
		}
		return version, err
	}
	return version, nil
}
