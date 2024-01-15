package postgres

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/internal/storage/postgres/utils"
	db "github.com/Permify/permify/pkg/database/postgres"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// SchemaWriter - Structure for SchemaWriter
type SchemaWriter struct {
	database *db.Postgres
	// options
	txOptions sql.TxOptions
}

// NewSchemaWriter creates a new SchemaWriter
func NewSchemaWriter(database *db.Postgres) *SchemaWriter {
	return &SchemaWriter{
		database:  database,
		txOptions: sql.TxOptions{Isolation: sql.LevelReadCommitted, ReadOnly: false},
	}
}

// WriteSchema writes a schema to the database
func (w *SchemaWriter) WriteSchema(ctx context.Context, schemas []storage.SchemaDefinition) (err error) {
	ctx, span := tracer.Start(ctx, "schema-writer.write-schema")
	defer span.End()

	slog.Debug("writing schemas to the database", slog.Any("number_of_schemas", len(schemas)))

	insertBuilder := w.database.Builder.Insert(SchemaDefinitionTable).Columns("name, serialized_definition, version, tenant_id")

	for _, schema := range schemas {
		insertBuilder = insertBuilder.Values(schema.Name, schema.SerializedDefinition, schema.Version, schema.TenantID)
	}

	var query string
	var args []interface{}

	query, args, err = insertBuilder.ToSql()
	if err != nil {
		return utils.HandleError(span, err, base.ErrorCode_ERROR_CODE_SQL_BUILDER)
	}

	slog.Debug("executing sql insert query", slog.Any("query", query), slog.Any("arguments", args))

	_, err = w.database.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return utils.HandleError(span, err, base.ErrorCode_ERROR_CODE_EXECUTION)
	}

	slog.Debug("successfully wrote schemas to the database", slog.Any("number_of_schemas", len(schemas)))

	return nil
}
