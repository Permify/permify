package postgres

import (
	"context"
	"database/sql"
	"errors"

	otelCodes "go.opentelemetry.io/otel/codes"

	"github.com/Permify/permify/internal/repositories"
	db "github.com/Permify/permify/pkg/database/postgres"
	"github.com/Permify/permify/pkg/logger"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// SchemaWriter - Structure for SchemaWriter
type SchemaWriter struct {
	database *db.Postgres
	// options
	txOptions sql.TxOptions
	// logger
	logger logger.Interface
}

// NewSchemaWriter creates a new SchemaWriter
func NewSchemaWriter(database *db.Postgres, logger logger.Interface) *SchemaWriter {
	return &SchemaWriter{
		database:  database,
		txOptions: sql.TxOptions{Isolation: sql.LevelReadCommitted, ReadOnly: false},
		logger:    logger,
	}
}

// WriteSchema writes a schema to the database
func (w *SchemaWriter) WriteSchema(ctx context.Context, schemas []repositories.SchemaDefinition) (err error) {
	ctx, span := tracer.Start(ctx, "schema-writer.write-schema")
	defer span.End()

	insertBuilder := w.database.Builder.Insert(SchemaDefinitionTable).Columns("entity_type, serialized_definition, version, tenant_id")

	for _, schema := range schemas {
		insertBuilder = insertBuilder.Values(schema.EntityType, schema.SerializedDefinition, schema.Version, schema.TenantID)
	}

	var query string
	var args []interface{}

	query, args, err = insertBuilder.ToSql()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		return errors.New(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String())
	}

	_, err = w.database.DB.ExecContext(ctx, query, args...)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		return err
	}

	return nil
}
