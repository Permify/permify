package postgres

import (
	"context"
	e "errors"

	"github.com/Masterminds/squirrel"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"

	"github.com/Permify/permify/internal/repositories/entities"
	"github.com/Permify/permify/internal/repositories/postgres/migrations"
	"github.com/Permify/permify/pkg/database"
	db "github.com/Permify/permify/pkg/database/postgres"
	"github.com/Permify/permify/pkg/errors"
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
func (r *EntityConfigRepository) Migrate() errors.Error {
	var err error
	ctx := context.Background()

	var tx pgx.Tx
	tx, err = r.Database.Pool.Begin(ctx)
	if err != nil {
		return errors.DatabaseError.SetSubKind(database.ErrMigration)
	}

	_, err = tx.Exec(context.Background(), migrations.CreateEntityConfigMigration())
	if err != nil {
		return errors.DatabaseError.SetSubKind(database.ErrMigration)
	}

	_, err = tx.Exec(context.Background(), migrations.CreateEntityConfigVersionField())
	if err != nil {
		return errors.DatabaseError.SetSubKind(database.ErrMigration)
	}

	_, err = tx.Exec(context.Background(), migrations.CreateEntityConfigChangePrimaryKey())
	if err != nil {
		return errors.DatabaseError.SetSubKind(database.ErrMigration)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return errors.DatabaseError.SetSubKind(database.ErrMigration)
	}

	return nil
}

// All -
func (r *EntityConfigRepository) All(ctx context.Context, version string) (entities.EntityConfigs, errors.Error) {
	var configs entities.EntityConfigs
	var err error
	if version == "" {
		version, err = r.findLastVersion(ctx)
		if err != nil {
			return configs, errors.DatabaseError.SetMessage(err.Error())
		}
	}

	var sql string
	var args []interface{}
	sql, args, err = r.Database.Builder.
		Select("entity, serialized_config, version").From(entities.EntityConfig{}.Table()).Where(squirrel.Eq{"version": version}).
		ToSql()
	if err != nil {
		return nil, errors.DatabaseError.SetSubKind(database.ErrBuilder)
	}

	var rows pgx.Rows
	rows, err = r.Database.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, errors.DatabaseError.SetSubKind(database.ErrExecution)
	}
	defer rows.Close()

	ent := make([]entities.EntityConfig, 0, _defaultEntityCap)

	for rows.Next() {
		c := entities.EntityConfig{}
		err = rows.Scan(&c.Entity, &c.SerializedConfig, &c.Version)
		if err != nil {
			return configs, errors.DatabaseError.SetSubKind(database.ErrScan)
		}
		ent = append(ent, c)
	}

	return ent, nil
}

// Read -
func (r *EntityConfigRepository) Read(ctx context.Context, name string, version string) (entities.EntityConfig, errors.Error) {
	var config entities.EntityConfig
	var err error
	if version == "" {
		version, err = r.findLastVersion(ctx)
		if err != nil {
			return config, errors.DatabaseError.SetMessage(err.Error())
		}
	}
	var sql string
	var args []interface{}
	sql, args, err = r.Database.Builder.
		Select("entity, serialized_config, version").From(entities.EntityConfig{}.Table()).Where(squirrel.Eq{"entity": name, "version": version}).Limit(1).
		ToSql()
	if err != nil {
		return config, errors.DatabaseError.SetSubKind(database.ErrBuilder)
	}
	var row pgx.Row
	row = r.Database.Pool.QueryRow(ctx, sql, args...)
	err = row.Scan(&config.Entity, &config.SerializedConfig, &config.Version)
	if err != nil {
		if e.Is(err, pgx.ErrNoRows) {
			return config, errors.DatabaseError.SetSubKind(database.ErrRecordNotFound)
		}
		return config, errors.DatabaseError.SetMessage(err.Error())
	}

	return config, nil
}

// Write -
func (r *EntityConfigRepository) Write(ctx context.Context, configs entities.EntityConfigs, version string) errors.Error {
	var err error
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
		return errors.NewError(errors.Database).SetSubKind(database.ErrBuilder)
	}

	_, err = r.Database.Pool.Exec(ctx, query, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if e.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				return errors.DatabaseError.SetSubKind(database.ErrUniqueConstraint)
			default:
				return errors.DatabaseError.SetMessage(err.Error())
			}
		}
		return errors.DatabaseError.SetMessage(err.Error())
	}

	return nil
}

// findLastVersion -
func (r *EntityConfigRepository) findLastVersion(ctx context.Context) (string, errors.Error) {
	var version string
	var err error
	var sql string
	var args []interface{}
	sql, args, err = r.Database.Builder.
		Select("version").From(entities.EntityConfig{}.Table()).OrderBy("version DESC").Limit(1).
		ToSql()
	if err != nil {
		return "", errors.DatabaseError.SetSubKind(database.ErrBuilder)
	}
	var row pgx.Row
	row = r.Database.Pool.QueryRow(ctx, sql, args...)
	err = row.Scan(&version)
	if err != nil {
		if e.Is(err, pgx.ErrNoRows) {
			return version, errors.DatabaseError.SetSubKind(database.ErrRecordNotFound)
		}
		return version, errors.DatabaseError.SetMessage(err.Error())
	}
	return version, nil
}
