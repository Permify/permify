package postgres

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"github.com/Masterminds/squirrel"

	"github.com/Permify/permify/internal/schema"
	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/internal/storage/postgres/utils"
	db "github.com/Permify/permify/pkg/database/postgres"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// SchemaReader - Structure for SchemaReader
type SchemaReader struct {
	database *db.Postgres
	// options
	txOptions sql.TxOptions
}

// NewSchemaReader - Creates a new SchemaReader
func NewSchemaReader(database *db.Postgres) *SchemaReader {
	return &SchemaReader{
		database:  database,
		txOptions: sql.TxOptions{Isolation: sql.LevelReadCommitted, ReadOnly: true},
	}
}

// ReadSchema - Reads entity config from the repository.
func (r *SchemaReader) ReadSchema(ctx context.Context, tenantID, version string) (sch *base.SchemaDefinition, err error) {
	ctx, span := tracer.Start(ctx, "schema-reader.read-schema")
	defer span.End()

	slog.Info("Reading schema: ", slog.Any("tenant_id", tenantID), slog.Any("version", version))

	builder := r.database.Builder.Select("name, serialized_definition, version").From(SchemaDefinitionTable).Where(squirrel.Eq{"version": version, "tenant_id": tenantID})

	var query string
	var args []interface{}

	query, args, err = builder.ToSql()
	if err != nil {
		return nil, utils.HandleError(span, err, base.ErrorCode_ERROR_CODE_SQL_BUILDER)
	}

	slog.Debug("Executing SQL query: ", slog.Any("query", query), slog.Any("arguments", args))

	var rows *sql.Rows
	rows, err = r.database.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, utils.HandleError(span, err, base.ErrorCode_ERROR_CODE_EXECUTION)
	}
	defer rows.Close()

	var definitions []string
	for rows.Next() {
		sd := storage.SchemaDefinition{}
		err = rows.Scan(&sd.Name, &sd.SerializedDefinition, &sd.Version)
		if err != nil {
			return nil, utils.HandleError(span, err, base.ErrorCode_ERROR_CODE_SCAN)
		}
		definitions = append(definitions, sd.Serialized())
	}
	if err = rows.Err(); err != nil {
		return nil, utils.HandleError(span, err, base.ErrorCode_ERROR_CODE_SCAN)
	}

	slog.Info("Successfully retrieved", slog.Any("schema definitions", len(definitions)))

	sch, err = schema.NewSchemaFromStringDefinitions(false, definitions...)
	if err != nil {
		return nil, utils.HandleError(span, err, base.ErrorCode_ERROR_CODE_INTERNAL)
	}

	slog.Info("Successfully created schema.")

	return sch, err
}

// ReadEntityDefinition - Reads entity config from the repository.
func (r *SchemaReader) ReadEntityDefinition(ctx context.Context, tenantID, name, version string) (definition *base.EntityDefinition, v string, err error) {
	ctx, span := tracer.Start(ctx, "schema-reader.read-entity-definition")
	defer span.End()

	slog.Info("Reading entity definition: ", slog.Any("tenant_id", tenantID), slog.Any("version", version))

	builder := r.database.Builder.Select("name, serialized_definition, version").Where(squirrel.Eq{"name": name, "version": version, "tenant_id": tenantID}).From(SchemaDefinitionTable).Limit(1)

	var query string
	var args []interface{}

	query, args, err = builder.ToSql()
	if err != nil {
		return nil, "", utils.HandleError(span, err, base.ErrorCode_ERROR_CODE_SQL_BUILDER)
	}

	slog.Debug("Executing SQL query: ", slog.Any("query", query), slog.Any("arguments", args))

	var def storage.SchemaDefinition
	row := r.database.DB.QueryRowContext(ctx, query, args...)
	if err = row.Err(); err != nil {
		return nil, "", utils.HandleError(span, err, base.ErrorCode_ERROR_CODE_EXECUTION)
	}

	if err = row.Scan(&def.Name, &def.SerializedDefinition, &def.Version); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", utils.HandleError(span, err, base.ErrorCode_ERROR_CODE_SCHEMA_NOT_FOUND)
		}
		return nil, "", utils.HandleError(span, err, base.ErrorCode_ERROR_CODE_SCAN)
	}

	var sch *base.SchemaDefinition
	sch, err = schema.NewSchemaFromStringDefinitions(false, def.Serialized())
	if err != nil {
		return nil, "", utils.HandleError(span, err, base.ErrorCode_ERROR_CODE_INTERNAL)
	}

	definition, err = schema.GetEntityByName(sch, name)

	slog.Info("Successfully retrieved", slog.Any("schema definition", definition))

	return definition, def.Version, err
}

// ReadRuleDefinition - Reads rule config from the repository.
func (r *SchemaReader) ReadRuleDefinition(ctx context.Context, tenantID, name, version string) (definition *base.RuleDefinition, v string, err error) {
	ctx, span := tracer.Start(ctx, "schema-reader.read-rule-definition")
	defer span.End()

	slog.Info("Reading rule definition: ", slog.Any("tenant_id", tenantID), slog.Any("name", name), slog.Any("version", version))

	builder := r.database.Builder.Select("name, serialized_definition, version").Where(squirrel.Eq{"name": name, "version": version, "tenant_id": tenantID}).From(SchemaDefinitionTable).Limit(1)

	var query string
	var args []interface{}

	query, args, err = builder.ToSql()
	if err != nil {
		return nil, "", utils.HandleError(span, err, base.ErrorCode_ERROR_CODE_SQL_BUILDER)
	}

	slog.Debug("Executing SQL query: ", slog.Any("query", query), slog.Any("arguments", args))

	var def storage.SchemaDefinition
	row := r.database.DB.QueryRowContext(ctx, query, args...)
	if err = row.Err(); err != nil {
		return nil, "", utils.HandleError(span, err, base.ErrorCode_ERROR_CODE_EXECUTION)
	}

	if err = row.Scan(&def.Name, &def.SerializedDefinition, &def.Version); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", utils.HandleError(span, err, base.ErrorCode_ERROR_CODE_SCHEMA_NOT_FOUND)
		}
		return nil, "", utils.HandleError(span, err, base.ErrorCode_ERROR_CODE_SCAN)
	}

	slog.Info("Successfully retrieved rule definition for: ", slog.Any("name", name))

	var sch *base.SchemaDefinition
	sch, err = schema.NewSchemaFromStringDefinitions(false, def.Serialized())
	if err != nil {
		return nil, "", utils.HandleError(span, err, base.ErrorCode_ERROR_CODE_INTERNAL)
	}

	definition, err = schema.GetRuleByName(sch, name)
	slog.Info("Successfully created rule definition")

	return definition, def.Version, err
}

// HeadVersion - Finds the latest version of the schema.
func (r *SchemaReader) HeadVersion(ctx context.Context, tenantID string) (version string, err error) {
	ctx, span := tracer.Start(ctx, "schema-reader.head-version")
	defer span.End()

	slog.Info("Finding the latest version fo the schema for: ", slog.String("tenant_id", tenantID))

	var query string
	var args []interface{}
	query, args, err = r.database.Builder.
		Select("version").From(SchemaDefinitionTable).Where(squirrel.Eq{"tenant_id": tenantID}).OrderBy("version DESC").Limit(1).
		ToSql()
	if err != nil {
		return "", utils.HandleError(span, err, base.ErrorCode_ERROR_CODE_SQL_BUILDER)
	}

	slog.Debug("Executing SQL query: ", slog.Any("query", query), slog.Any("arguments", args))

	row := r.database.DB.QueryRowContext(ctx, query, args...)
	err = row.Scan(&version)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", utils.HandleError(span, err, base.ErrorCode_ERROR_CODE_SCHEMA_NOT_FOUND)
		}
		return "", utils.HandleError(span, err, base.ErrorCode_ERROR_CODE_SCAN)
	}

	slog.Info("Successfully found the latest schema version: ", slog.Any("version", version))

	return version, nil
}
