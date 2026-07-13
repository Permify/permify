package postgres

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/rs/xid"

	"github.com/Permify/permify/internal"
	"github.com/Permify/permify/internal/schema"
	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/internal/storage/postgres/utils"
	"github.com/Permify/permify/pkg/database"
	db "github.com/Permify/permify/pkg/database/postgres"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// SharedSchemaReader - Structure for SharedSchemaReader
type SharedSchemaReader struct {
	database *db.Postgres
	// options
	txOptions pgx.TxOptions
}

// NewSharedSchemaReader - Creates a new SharedSchemaReader
func NewSharedSchemaReader(database *db.Postgres) *SharedSchemaReader {
	return &SharedSchemaReader{
		database:  database,
		txOptions: pgx.TxOptions{IsoLevel: pgx.ReadCommitted, AccessMode: pgx.ReadOnly},
	}
}

// ReadSharedSchema returns the shared schema definition for a specific shared_schema_id and version.
func (r *SharedSchemaReader) ReadSharedSchema(ctx context.Context, sharedSchemaID, version string) (sch *base.SchemaDefinition, err error) {
	ctx, span := internal.Tracer.Start(ctx, "shared-schema-reader.read-shared-schema")
	defer span.End()
	slog.DebugContext(ctx, "reading shared schema", slog.Any("shared_schema_id", sharedSchemaID), slog.Any("version", version))

	builder := r.database.Builder.Select("name, serialized_definition, version").From(SharedSchemaDefinitionTable).Where(squirrel.Eq{"version": version, "shared_schema_id": sharedSchemaID})

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
		sd := storage.SharedSchemaDefinition{}
		err = rows.Scan(&sd.Name, &sd.SerializedDefinition, &sd.Version)
		if err != nil {
			return nil, utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_SCAN)
		}
		definitions = append(definitions, sd.Serialized())
	}
	if err = rows.Err(); err != nil {
		return nil, utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_SCAN)
	}

	slog.DebugContext(ctx, "successfully retrieved", slog.Any("shared schema definitions", len(definitions)))
	sch, err = schema.NewSchemaFromStringDefinitions(false, definitions...)
	if err != nil {
		return nil, utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_INTERNAL)
	}

	return sch, err
}

// ReadSharedSchemaString returns the shared schema definition as string definitions.
func (r *SharedSchemaReader) ReadSharedSchemaString(ctx context.Context, sharedSchemaID, version string) (definitions []string, err error) {
	ctx, span := internal.Tracer.Start(ctx, "shared-schema-reader.read-shared-schema-string")
	defer span.End()
	slog.DebugContext(ctx, "reading shared schema string", slog.Any("shared_schema_id", sharedSchemaID), slog.Any("version", version))

	builder := r.database.Builder.Select("name, serialized_definition, version").From(SharedSchemaDefinitionTable).Where(squirrel.Eq{"version": version, "shared_schema_id": sharedSchemaID})

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
		sd := storage.SharedSchemaDefinition{}
		err = rows.Scan(&sd.Name, &sd.SerializedDefinition, &sd.Version)
		if err != nil {
			return []string{}, utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_SCAN)
		}
		definitions = append(definitions, sd.Serialized())
	}
	if err = rows.Err(); err != nil {
		return []string{}, utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_SCAN)
	}

	slog.DebugContext(ctx, "successfully retrieved", slog.Any("shared schema definitions", len(definitions)))
	return definitions, err
}

// ReadSharedEntityDefinition reads a single entity definition from a shared schema.
func (r *SharedSchemaReader) ReadSharedEntityDefinition(ctx context.Context, sharedSchemaID, name, version string) (definition *base.EntityDefinition, v string, err error) {
	ctx, span := internal.Tracer.Start(ctx, "shared-schema-reader.read-shared-entity-definition")
	defer span.End()
	slog.DebugContext(ctx, "reading shared entity definition", slog.Any("shared_schema_id", sharedSchemaID), slog.Any("version", version))

	builder := r.database.Builder.Select("name, serialized_definition, version").Where(squirrel.Eq{"name": name, "version": version, "shared_schema_id": sharedSchemaID}).From(SharedSchemaDefinitionTable).Limit(1)

	var query string
	var args []interface{}

	query, args, err = builder.ToSql()
	if err != nil {
		return nil, "", utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_SQL_BUILDER)
	}

	slog.DebugContext(ctx, "executing sql query", slog.Any("query", query), slog.Any("arguments", args))
	var def storage.SharedSchemaDefinition
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
	slog.DebugContext(ctx, "successfully retrieved", slog.Any("shared schema definition", definition))
	return definition, def.Version, err
}

// ReadSharedRuleDefinition reads a single rule definition from a shared schema.
func (r *SharedSchemaReader) ReadSharedRuleDefinition(ctx context.Context, sharedSchemaID, name, version string) (definition *base.RuleDefinition, v string, err error) {
	ctx, span := internal.Tracer.Start(ctx, "shared-schema-reader.read-shared-rule-definition")
	defer span.End()
	slog.DebugContext(ctx, "reading shared rule definition", slog.Any("shared_schema_id", sharedSchemaID), slog.Any("name", name), slog.Any("version", version))

	builder := r.database.Builder.Select("name, serialized_definition, version").Where(squirrel.Eq{"name": name, "version": version, "shared_schema_id": sharedSchemaID}).From(SharedSchemaDefinitionTable).Limit(1)

	var query string
	var args []interface{}

	query, args, err = builder.ToSql()
	if err != nil {
		return nil, "", utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_SQL_BUILDER)
	}

	slog.DebugContext(ctx, "executing sql query", slog.Any("query", query), slog.Any("arguments", args))
	var def storage.SharedSchemaDefinition
	row := r.database.ReadPool.QueryRow(ctx, query, args...)
	if err = row.Scan(&def.Name, &def.SerializedDefinition, &def.Version); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, "", utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_SCHEMA_NOT_FOUND)
		}
		return nil, "", utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_SCAN)
	}

	slog.DebugContext(ctx, "successfully retrieved shared rule definition for", slog.Any("name", name))
	var sch *base.SchemaDefinition
	sch, err = schema.NewSchemaFromStringDefinitions(false, def.Serialized())
	if err != nil {
		return nil, "", utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_INTERNAL)
	}

	definition, err = schema.GetRuleByName(sch, name)
	slog.DebugContext(ctx, "successfully created shared rule definition")
	return definition, def.Version, err
}

// SharedHeadVersion returns the latest version for a shared schema.
func (r *SharedSchemaReader) SharedHeadVersion(ctx context.Context, sharedSchemaID string) (version string, err error) {
	ctx, span := internal.Tracer.Start(ctx, "shared-schema-reader.shared-head-version")
	defer span.End()
	slog.DebugContext(ctx, "finding the latest version of shared schema", slog.String("shared_schema_id", sharedSchemaID))

	var query string
	var args []interface{}
	query, args, err = r.database.Builder.
		Select("version").From(SharedSchemaDefinitionTable).Where(squirrel.Eq{"shared_schema_id": sharedSchemaID}).OrderBy("version DESC").Limit(1).
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

	slog.DebugContext(ctx, "successfully found shared schema head version", slog.Any("version", version))
	return version, nil
}

// ListSharedSchemas lists all shared schema IDs with pagination.
func (r *SharedSchemaReader) ListSharedSchemas(ctx context.Context, pagination database.Pagination) (schemas []*base.SharedSchemaListItem, ct database.EncodedContinuousToken, err error) {
	ctx, span := internal.Tracer.Start(ctx, "shared-schema-reader.list-shared-schemas")
	defer span.End()
	slog.DebugContext(ctx, "listing shared schemas with pagination", slog.Any("pagination", pagination))

	// Get distinct shared_schema_ids with their latest version
	builder := r.database.Builder.
		Select("shared_schema_id, MAX(version) as head_version").
		From(SharedSchemaDefinitionTable).
		GroupBy("shared_schema_id")

	if pagination.Token() != "" {
		var t database.ContinuousToken
		t, err = utils.EncodedContinuousToken{Value: pagination.Token()}.Decode()
		if err != nil {
			return nil, nil, utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_INVALID_CONTINUOUS_TOKEN)
		}
		builder = builder.Having(squirrel.GtOrEq{"shared_schema_id": t.(utils.ContinuousToken).Value})
	}

	builder = builder.OrderBy("shared_schema_id").Limit(uint64(pagination.PageSize() + 1))

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

	var lastID string
	schemas = make([]*base.SharedSchemaListItem, 0, pagination.PageSize()+1)
	for rows.Next() {
		item := &base.SharedSchemaListItem{}
		err = rows.Scan(&item.SharedSchemaId, &item.HeadVersion)
		if err != nil {
			return nil, nil, utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_SCAN)
		}
		id, parseErr := xid.FromString(item.HeadVersion)
		if parseErr != nil {
			return nil, nil, utils.HandleError(ctx, span, parseErr, base.ErrorCode_ERROR_CODE_SCAN)
		}
		item.CreatedAt = id.Time().String()
		lastID = item.SharedSchemaId
		schemas = append(schemas, item)
	}
	if err = rows.Err(); err != nil {
		return nil, nil, utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_INTERNAL)
	}

	slog.DebugContext(ctx, "successfully listed shared schemas", slog.Any("count", len(schemas)))

	if len(schemas) > int(pagination.PageSize()) {
		return schemas[:pagination.PageSize()], utils.NewContinuousToken(lastID).Encode(), nil
	}
	return schemas, database.NewNoopContinuousToken().Encode(), nil
}

// GetTenantSharedSchemaID returns the shared_schema_id for a tenant, or empty string if none.
func (r *SharedSchemaReader) GetTenantSharedSchemaID(ctx context.Context, tenantID string) (sharedSchemaID string, err error) {
	ctx, span := internal.Tracer.Start(ctx, "shared-schema-reader.get-tenant-shared-schema-id")
	defer span.End()
	slog.DebugContext(ctx, "getting shared schema id for tenant", slog.String("tenant_id", tenantID))

	var query string
	var args []interface{}
	query, args, err = r.database.Builder.
		Select("shared_schema_id").From(TenantsTable).Where(squirrel.Eq{"id": tenantID}).Limit(1).
		ToSql()
	if err != nil {
		return "", utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_SQL_BUILDER)
	}

	var nullableID sql.NullString
	row := r.database.ReadPool.QueryRow(ctx, query, args...)
	err = row.Scan(&nullableID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", nil
		}
		return "", utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_SCAN)
	}

	if nullableID.Valid {
		return nullableID.String, nil
	}
	return "", nil
}
