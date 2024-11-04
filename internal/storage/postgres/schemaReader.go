package postgres

import (
	"context"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5"

	"github.com/Masterminds/squirrel"
	"github.com/rs/xid"

	"github.com/Permify/permify/internal"
	"github.com/Permify/permify/internal/schema"
	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/internal/storage/postgres/utils"
	"github.com/Permify/permify/pkg/database"
	db "github.com/Permify/permify/pkg/database/postgres"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// SchemaReader - Structure for SchemaReader
type SchemaReader struct {
	database *db.Postgres
	// options
	txOptions pgx.TxOptions
}

// NewSchemaReader - Creates a new SchemaReader
func NewSchemaReader(database *db.Postgres) *SchemaReader {
	return &SchemaReader{
		database:  database,
		txOptions: pgx.TxOptions{IsoLevel: pgx.ReadCommitted, AccessMode: pgx.ReadOnly},
	}
}

// ReadSchema returns the schema definition for a specific tenant and version as a structured object.
func (r *SchemaReader) ReadSchema(ctx context.Context, tenantID, version string) (sch *base.SchemaDefinition, err error) {
	ctx, span := internal.Tracer.Start(ctx, "schema-reader.read-schema")
	defer span.End()

	slog.DebugContext(ctx, "reading schema", slog.Any("tenant_id", tenantID), slog.Any("version", version))

	builder := r.database.Builder.Select("name, serialized_definition, version").From(SchemaDefinitionTable).Where(squirrel.Eq{"version": version, "tenant_id": tenantID})

	var query string
	var args []interface{}

	query, args, err = builder.ToSql()
	if err != nil {
		return nil, utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_SQL_BUILDER)
	}

	slog.DebugContext(ctx, "executing sql query", slog.Any("query", query), slog.Any("arguments", args))

	var rows pgx.Rows
	rows, err = r.database.ReadPool.Query(ctx, query, args...)
	if err != nil {
		return nil, utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_EXECUTION)
	}
	defer rows.Close()

	var definitions []string
	for rows.Next() {
		sd := storage.SchemaDefinition{}
		err = rows.Scan(&sd.Name, &sd.SerializedDefinition, &sd.Version)
		if err != nil {
			return nil, utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_SCAN)
		}
		definitions = append(definitions, sd.Serialized())
	}
	if err = rows.Err(); err != nil {
		return nil, utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_SCAN)
	}

	slog.DebugContext(ctx, "successfully retrieved", slog.Any("schema definitions", len(definitions)))

	sch, err = schema.NewSchemaFromStringDefinitions(false, definitions...)
	if err != nil {
		return nil, utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_INTERNAL)
	}

	return sch, err
}

// ReadSchemaString returns the schema definition for a specific tenant and version as a string.
func (r *SchemaReader) ReadSchemaString(ctx context.Context, tenantID, version string) (definitions []string, err error) {
	ctx, span := internal.Tracer.Start(ctx, "schema-reader.read-schema-string")
	defer span.End()

	slog.DebugContext(ctx, "reading schema", slog.Any("tenant_id", tenantID), slog.Any("version", version))

	builder := r.database.Builder.Select("name, serialized_definition, version").From(SchemaDefinitionTable).Where(squirrel.Eq{"version": version, "tenant_id": tenantID})

	var query string
	var args []interface{}

	query, args, err = builder.ToSql()
	if err != nil {
		return []string{}, utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_SQL_BUILDER)
	}

	slog.DebugContext(ctx, "executing sql query", slog.Any("query", query), slog.Any("arguments", args))

	var rows pgx.Rows
	rows, err = r.database.ReadPool.Query(ctx, query, args...)
	if err != nil {
		return []string{}, utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_EXECUTION)
	}
	defer rows.Close()

	for rows.Next() {
		sd := storage.SchemaDefinition{}
		err = rows.Scan(&sd.Name, &sd.SerializedDefinition, &sd.Version)
		if err != nil {
			return []string{}, utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_SCAN)
		}
		definitions = append(definitions, sd.Serialized())
	}
	if err = rows.Err(); err != nil {
		return []string{}, utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_SCAN)
	}

	slog.DebugContext(ctx, "successfully retrieved", slog.Any("schema definitions", len(definitions)))

	return definitions, err
}

// ReadEntityDefinition - Reads entity config from the repository.
func (r *SchemaReader) ReadEntityDefinition(ctx context.Context, tenantID, name, version string) (definition *base.EntityDefinition, v string, err error) {
	ctx, span := internal.Tracer.Start(ctx, "schema-reader.read-entity-definition")
	defer span.End()

	slog.DebugContext(ctx, "reading entity definition", slog.Any("tenant_id", tenantID), slog.Any("version", version))

	builder := r.database.Builder.Select("name, serialized_definition, version").Where(squirrel.Eq{"name": name, "version": version, "tenant_id": tenantID}).From(SchemaDefinitionTable).Limit(1)

	var query string
	var args []interface{}

	query, args, err = builder.ToSql()
	if err != nil {
		return nil, "", utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_SQL_BUILDER)
	}

	slog.DebugContext(ctx, "executing sql query", slog.Any("query", query), slog.Any("arguments", args))

	var def storage.SchemaDefinition
	row := r.database.ReadPool.QueryRow(ctx, query, args...)
	if err = row.Scan(&def.Name, &def.SerializedDefinition, &def.Version); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, "", utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_SCHEMA_NOT_FOUND)
		}
		return nil, "", utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_SCAN)
	}

	var sch *base.SchemaDefinition
	sch, err = schema.NewSchemaFromStringDefinitions(false, def.Serialized())
	if err != nil {
		return nil, "", utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_INTERNAL)
	}

	definition, err = schema.GetEntityByName(sch, name)

	slog.DebugContext(ctx, "successfully retrieved", slog.Any("schema definition", definition))

	return definition, def.Version, err
}

// ReadRuleDefinition - Reads rule config from the repository.
func (r *SchemaReader) ReadRuleDefinition(ctx context.Context, tenantID, name, version string) (definition *base.RuleDefinition, v string, err error) {
	ctx, span := internal.Tracer.Start(ctx, "schema-reader.read-rule-definition")
	defer span.End()

	slog.DebugContext(ctx, "reading rule definition", slog.Any("tenant_id", tenantID), slog.Any("name", name), slog.Any("version", version))

	builder := r.database.Builder.Select("name, serialized_definition, version").Where(squirrel.Eq{"name": name, "version": version, "tenant_id": tenantID}).From(SchemaDefinitionTable).Limit(1)

	var query string
	var args []interface{}

	query, args, err = builder.ToSql()
	if err != nil {
		return nil, "", utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_SQL_BUILDER)
	}

	slog.DebugContext(ctx, "executing sql query", slog.Any("query", query), slog.Any("arguments", args))

	var def storage.SchemaDefinition
	row := r.database.ReadPool.QueryRow(ctx, query, args...)
	if err = row.Scan(&def.Name, &def.SerializedDefinition, &def.Version); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, "", utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_SCHEMA_NOT_FOUND)
		}
		return nil, "", utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_SCAN)
	}

	slog.DebugContext(ctx, "successfully retrieved rule definition for", slog.Any("name", name))

	var sch *base.SchemaDefinition
	sch, err = schema.NewSchemaFromStringDefinitions(false, def.Serialized())
	if err != nil {
		return nil, "", utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_INTERNAL)
	}

	definition, err = schema.GetRuleByName(sch, name)

	slog.DebugContext(ctx, "successfully created rule definition")

	return definition, def.Version, err
}

// HeadVersion - Finds the latest version of the schema.
func (r *SchemaReader) HeadVersion(ctx context.Context, tenantID string) (version string, err error) {
	ctx, span := internal.Tracer.Start(ctx, "schema-reader.head-version")
	defer span.End()

	slog.DebugContext(ctx, "finding the latest version fo the schema for", slog.String("tenant_id", tenantID))

	var query string
	var args []interface{}
	query, args, err = r.database.Builder.
		Select("version").From(SchemaDefinitionTable).Where(squirrel.Eq{"tenant_id": tenantID}).OrderBy("version DESC").Limit(1).
		ToSql()
	if err != nil {
		return "", utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_SQL_BUILDER)
	}

	slog.DebugContext(ctx, "executing sql query", slog.Any("query", query), slog.Any("arguments", args))

	row := r.database.ReadPool.QueryRow(ctx, query, args...)
	err = row.Scan(&version)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_SCHEMA_NOT_FOUND)
		}
		return "", utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_SCAN)
	}

	slog.DebugContext(ctx, "successfully found the latest schema version", slog.Any("version", version))

	return version, nil
}

// ListSchemas - List all Schemas
func (r *SchemaReader) ListSchemas(ctx context.Context, tenantID string, pagination database.Pagination) (schemas []*base.SchemaList, ct database.EncodedContinuousToken, err error) {
	ctx, span := internal.Tracer.Start(ctx, "tenant-reader.list-tenants")
	defer span.End()

	slog.DebugContext(ctx, "listing schemas with pagination", slog.Any("pagination", pagination))

	builder := r.database.Builder.Select("DISTINCT version").From(SchemaDefinitionTable).Where(squirrel.Eq{"tenant_id": tenantID})
	if pagination.Token() != "" {
		var t database.ContinuousToken
		t, err = utils.EncodedContinuousToken{Value: pagination.Token()}.Decode()
		if err != nil {
			return nil, nil, utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_INVALID_CONTINUOUS_TOKEN)
		}
		builder = builder.Where(squirrel.LtOrEq{"version": t.(utils.ContinuousToken).Value})
	}

	builder = builder.OrderBy("version DESC").Limit(uint64(pagination.PageSize() + 1))

	var query string
	var args []interface{}

	query, args, err = builder.ToSql()
	if err != nil {
		return nil, nil, utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_SQL_BUILDER)
	}

	slog.DebugContext(ctx, "executing sql query", slog.Any("query", query), slog.Any("arguments", args))

	var rows pgx.Rows
	rows, err = r.database.ReadPool.Query(ctx, query, args...)
	if err != nil {
		return nil, nil, utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_EXECUTION)
	}
	defer rows.Close()

	var lastVersion string
	schemas = make([]*base.SchemaList, 0, pagination.PageSize()+1)
	for rows.Next() {
		sch := &base.SchemaList{}
		err = rows.Scan(&sch.Version)
		if err != nil {
			return nil, nil, utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_SCAN)
		}
		id, err := xid.FromString(sch.Version)
		if err != nil {
			return nil, nil, utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_SCAN)
		}
		sch.CreatedAt = id.Time().String()
		lastVersion = sch.Version
		schemas = append(schemas, sch)
	}
	if err = rows.Err(); err != nil {
		return nil, nil, utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_INTERNAL)
	}

	slog.DebugContext(ctx, "successfully listed schemas", slog.Any("number_of_schemas", len(schemas)))

	if len(schemas) > int(pagination.PageSize()) {
		return schemas[:pagination.PageSize()], utils.NewContinuousToken(lastVersion).Encode(), nil
	}
	return schemas, database.NewNoopContinuousToken().Encode(), nil
}
