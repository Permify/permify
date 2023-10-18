package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Masterminds/squirrel"
	"go.opentelemetry.io/otel/codes"

	"github.com/Permify/permify/internal/schema"
	"github.com/Permify/permify/internal/storage"
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

	builder := r.database.Builder.Select("name, serialized_definition, version").From(SchemaDefinitionTable).Where(squirrel.Eq{"version": version, "tenant_id": tenantID})

	var query string
	var args []interface{}

	query, args, err = builder.ToSql()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, errors.New(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String())
	}

	var rows *sql.Rows
	rows, err = r.database.DB.QueryContext(ctx, query, args...)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
	}
	defer rows.Close()

	var definitions []string
	for rows.Next() {
		sd := storage.SchemaDefinition{}
		err = rows.Scan(&sd.Name, &sd.SerializedDefinition, &sd.Version)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return nil, err
		}
		definitions = append(definitions, sd.Serialized())
	}
	if err = rows.Err(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	sch, err = schema.NewSchemaFromStringDefinitions(false, definitions...)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	return sch, err
}

// ReadEntityDefinition - Reads entity config from the repository.
func (r *SchemaReader) ReadEntityDefinition(ctx context.Context, tenantID, name, version string) (definition *base.EntityDefinition, v string, err error) {
	ctx, span := tracer.Start(ctx, "schema-reader.read-entity-definition")
	defer span.End()

	builder := r.database.Builder.Select("name, serialized_definition, version").Where(squirrel.Eq{"name": name, "version": version, "tenant_id": tenantID}).From(SchemaDefinitionTable).Limit(1)

	var query string
	var args []interface{}

	query, args, err = builder.ToSql()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, "", errors.New(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String())
	}

	var def storage.SchemaDefinition
	row := r.database.DB.QueryRowContext(ctx, query, args...)
	if err = row.Err(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, "", errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
	}

	if err = row.Scan(&def.Name, &def.SerializedDefinition, &def.Version); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", errors.New(base.ErrorCode_ERROR_CODE_SCHEMA_NOT_FOUND.String())
		}
		return nil, "", errors.New(base.ErrorCode_ERROR_CODE_SCAN.String())
	}

	var sch *base.SchemaDefinition
	sch, err = schema.NewSchemaFromStringDefinitions(false, def.Serialized())
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, "", err
	}

	definition, err = schema.GetEntityByName(sch, name)
	return definition, def.Version, err
}

// ReadRuleDefinition - Reads rule config from the repository.
func (r *SchemaReader) ReadRuleDefinition(ctx context.Context, tenantID, name, version string) (definition *base.RuleDefinition, v string, err error) {
	ctx, span := tracer.Start(ctx, "schema-reader.read-rule-definition")
	defer span.End()

	builder := r.database.Builder.Select("name, serialized_definition, version").Where(squirrel.Eq{"name": name, "version": version, "tenant_id": tenantID}).From(SchemaDefinitionTable).Limit(1)

	var query string
	var args []interface{}

	query, args, err = builder.ToSql()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, "", errors.New(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String())
	}

	var def storage.SchemaDefinition
	row := r.database.DB.QueryRowContext(ctx, query, args...)
	if err = row.Err(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, "", errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
	}

	if err = row.Scan(&def.Name, &def.SerializedDefinition, &def.Version); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", errors.New(base.ErrorCode_ERROR_CODE_SCHEMA_NOT_FOUND.String())
		}
		return nil, "", errors.New(base.ErrorCode_ERROR_CODE_SCAN.String())
	}

	var sch *base.SchemaDefinition
	sch, err = schema.NewSchemaFromStringDefinitions(false, def.Serialized())
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, "", err
	}

	definition, err = schema.GetRuleByName(sch, name)
	return definition, def.Version, err
}

// HeadVersion - Finds the latest version of the schema.
func (r *SchemaReader) HeadVersion(ctx context.Context, tenantID string) (version string, err error) {
	ctx, span := tracer.Start(ctx, "schema-reader.head-version")
	defer span.End()

	var query string
	var args []interface{}
	query, args, err = r.database.Builder.
		Select("version").From(SchemaDefinitionTable).Where(squirrel.Eq{"tenant_id": tenantID}).OrderBy("version DESC").Limit(1).
		ToSql()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return "", errors.New(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String())
	}
	row := r.database.DB.QueryRowContext(ctx, query, args...)
	err = row.Scan(&version)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		if errors.Is(err, sql.ErrNoRows) {
			return "", errors.New(base.ErrorCode_ERROR_CODE_SCHEMA_NOT_FOUND.String())
		}
		return "", err
	}

	return version, nil
}
