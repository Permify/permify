package postgres

import (
	"context"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5"

	"github.com/Permify/permify/internal"
	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/internal/storage/postgres/utils"
	db "github.com/Permify/permify/pkg/database/postgres"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// SharedSchemaWriter - Structure for SharedSchemaWriter
type SharedSchemaWriter struct {
	database *db.Postgres
	// options
	txOptions pgx.TxOptions
}

// NewSharedSchemaWriter creates a new SharedSchemaWriter
func NewSharedSchemaWriter(database *db.Postgres) *SharedSchemaWriter {
	return &SharedSchemaWriter{
		database:  database,
		txOptions: pgx.TxOptions{IsoLevel: pgx.ReadCommitted, AccessMode: pgx.ReadWrite},
	}
}

// WriteSharedSchema writes a shared schema to the database
func (w *SharedSchemaWriter) WriteSharedSchema(ctx context.Context, definitions []storage.SharedSchemaDefinition) (err error) {
	ctx, span := internal.Tracer.Start(ctx, "shared-schema-writer.write-shared-schema")
	defer span.End()
	slog.DebugContext(ctx, "writing shared schemas to the database", slog.Any("number_of_definitions", len(definitions)))

	insertBuilder := w.database.Builder.Insert(SharedSchemaDefinitionTable).Columns("name, serialized_definition, version, shared_schema_id")
	for _, def := range definitions {
		insertBuilder = insertBuilder.Values(def.Name, def.SerializedDefinition, def.Version, def.SharedSchemaID)
	}

	var query string
	var args []interface{}

	query, args, err = insertBuilder.ToSql()
	if err != nil {
		return utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_SQL_BUILDER)
	}

	slog.DebugContext(ctx, "executing sql insert query", slog.Any("query", query), slog.Any("arguments", args))
	_, err = w.database.WritePool.Exec(ctx, query, args...)
	if err != nil {
		return utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_EXECUTION)
	}

	slog.DebugContext(ctx, "successfully wrote shared schemas to the database", slog.Any("number_of_definitions", len(definitions)))
	return nil
}

// AssignSharedSchema sets shared_schema_id on the given tenants.
func (w *SharedSchemaWriter) AssignSharedSchema(ctx context.Context, sharedSchemaID string, tenantIDs []string) (err error) {
	ctx, span := internal.Tracer.Start(ctx, "shared-schema-writer.assign-shared-schema")
	defer span.End()
	slog.DebugContext(ctx, "assigning shared schema to tenants", slog.String("shared_schema_id", sharedSchemaID), slog.Any("tenant_ids", tenantIDs))

	tx, err := w.database.WritePool.BeginTx(ctx, w.txOptions)
	if err != nil {
		return utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_EXECUTION)
	}
	defer tx.Rollback(ctx)

	// Validate shared schema exists
	var exists bool
	err = tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM "+SharedSchemaDefinitionTable+" WHERE shared_schema_id = $1 LIMIT 1)", sharedSchemaID).Scan(&exists)
	if err != nil {
		return utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_EXECUTION)
	}
	if !exists {
		return utils.HandleError(ctx, span, errors.New("shared schema not found"), base.ErrorCode_ERROR_CODE_NOT_FOUND)
	}

	// Assign to all tenants
	_, err = tx.Exec(ctx, "UPDATE "+TenantsTable+" SET shared_schema_id = $1 WHERE id = ANY($2)", sharedSchemaID, tenantIDs)
	if err != nil {
		return utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_EXECUTION)
	}

	if err = tx.Commit(ctx); err != nil {
		return utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_EXECUTION)
	}

	slog.DebugContext(ctx, "successfully assigned shared schema to tenants", slog.String("shared_schema_id", sharedSchemaID), slog.Any("tenant_ids", tenantIDs))
	return nil
}
