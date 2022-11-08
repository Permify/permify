package postgres

import (
	"context"
	"errors"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v4"

	"github.com/Permify/permify/internal/repositories"
	db "github.com/Permify/permify/pkg/database/postgres"
	"github.com/Permify/permify/pkg/dsl/compiler"
	"github.com/Permify/permify/pkg/dsl/schema"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

type SchemaReader struct {
	database  *db.Postgres
	txOptions pgx.TxOptions
}

// NewSchemaReader creates a new SchemaReader
func NewSchemaReader(database *db.Postgres) *SchemaReader {
	return &SchemaReader{
		database:  database,
		txOptions: pgx.TxOptions{IsoLevel: pgx.ReadCommitted, AccessMode: pgx.ReadOnly},
	}
}

// ReadSchema reads entity config from the repository.
func (r *SchemaReader) ReadSchema(ctx context.Context, version string) (schema *base.IndexedSchema, err error) {
	if version == "" {
		version, err = r.headVersion(ctx)
		if err != nil {
			return nil, err
		}
	}

	tx, err := r.database.Pool.BeginTx(ctx, r.txOptions)
	if err != nil {
		return nil, err
	}

	var sql string
	var args []interface{}

	query := r.database.Builder.Select("entity_type, serialized_definition").From(schemaDefinitionTable).Where(squirrel.Eq{"version": version})
	sql, args, err = query.ToSql()

	var rows pgx.Rows
	rows, err = tx.Query(ctx, sql, args...)
	if err != nil {
		return nil, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
	}
	defer rows.Close()

	var definitions []string
	for rows.Next() {
		sd := repositories.SchemaDefinition{}
		err = rows.Scan(&sd.EntityType, &sd.SerializedDefinition)
		if err != nil {
			return nil, err
		}
		definitions = append(definitions, sd.Serialized())
	}

	schema, err = compiler.NewSchema(definitions...)
	if err != nil {
		return nil, err
	}

	return nil, err
}

// ReadSchemaDefinition reads entity config from the repository.
func (r *SchemaReader) ReadSchemaDefinition(ctx context.Context, entityType string, version string) (*base.EntityDefinition, error) {
	var err error
	if version == "" {
		version, err = r.headVersion(ctx)
		if err != nil {
			return nil, err
		}
	}

	var tx pgx.Tx
	tx, err = r.database.Pool.BeginTx(ctx, r.txOptions)
	if err != nil {
		return nil, err
	}

	var sql string
	var args []interface{}

	query := r.database.Builder.Select("entity_type, serialized_definition").Where(squirrel.Eq{"entity_type": entityType, "version": version}).From(schemaDefinitionTable).Limit(1)
	sql, args, err = query.ToSql()

	var def repositories.SchemaDefinition
	var row pgx.Row
	row = tx.QueryRow(ctx, sql, args...)
	if err = row.Scan(&def.EntityType, &def.SerializedDefinition); err != nil {
		return nil, err
	}

	var sch *base.IndexedSchema
	sch, err = compiler.NewSchemaWithoutReferenceValidation(def.Serialized())
	if err != nil {
		return nil, err
	}

	return schema.GetEntityByName(sch, entityType)
}

// headVersion finds the latest version of the schema.
func (r *SchemaReader) headVersion(ctx context.Context) (version string, err error) {
	var sql string
	var args []interface{}
	sql, args, err = r.database.Builder.
		Select("version").From(schemaDefinitionTable).OrderBy("version DESC").Limit(1).
		ToSql()
	if err != nil {
		return "", errors.New(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String())
	}
	var row pgx.Row
	row = r.database.Pool.QueryRow(ctx, sql, args...)
	err = row.Scan(&version)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", errors.New(base.ErrorCode_ERROR_CODE_SCHEMA_NOT_FOUND.String())
		}
		return "", err
	}
	return version, nil
}
