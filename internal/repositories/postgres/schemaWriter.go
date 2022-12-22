package postgres

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/rs/xid"
	otelCodes "go.opentelemetry.io/otel/codes"

	"github.com/Permify/permify/internal/repositories"
	db "github.com/Permify/permify/pkg/database/postgres"
)

// SchemaWriter - Structure for SchemaWriter
type SchemaWriter struct {
	database  *db.Postgres
	txOptions pgx.TxOptions
}

// NewSchemaWriter creates a new SchemaWriter
func NewSchemaWriter(database *db.Postgres) *SchemaWriter {
	return &SchemaWriter{
		database:  database,
		txOptions: pgx.TxOptions{IsoLevel: pgx.ReadCommitted, AccessMode: pgx.ReadWrite},
	}
}

// WriteSchema writes a schema to the database
func (w *SchemaWriter) WriteSchema(ctx context.Context, schemas []repositories.SchemaDefinition) (string, error) {
	ctx, span := tracer.Start(ctx, "schemas.write.save")
	defer span.End()

	id := xid.New()
	tx, bErr := w.database.Pool.BeginTx(ctx, w.txOptions)
	if bErr != nil {
		span.RecordError(bErr)
		span.SetStatus(otelCodes.Error, bErr.Error())

		return "", bErr
	}

	batch := &pgx.Batch{}

	for _, schema := range schemas {
		query, args, err := w.database.Builder.
			Insert(SchemaDefinitionTable).
			Columns("entity_type, serialized_definition, version").Values(schema.EntityType, schema.SerializedDefinition, id.String()).ToSql()
		if err != nil {
			return "", err
		}
		batch.Queue(query, args...)
	}

	results := tx.SendBatch(ctx, batch)
	if err := results.Close(); err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())

		_ = tx.Rollback(ctx)
		return "", err
	}

	if err := tx.Commit(ctx); err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())

		return "", err
	}

	return id.String(), nil
}
