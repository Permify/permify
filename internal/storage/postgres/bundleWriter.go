package postgres

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/Masterminds/squirrel"
	"github.com/golang/protobuf/jsonpb"
	"github.com/pkg/errors"
	otelCodes "go.opentelemetry.io/otel/codes"

	db "github.com/Permify/permify/pkg/database/postgres"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

type BundleWriter struct {
	database *db.Postgres
	// options
	txOptions sql.TxOptions
}

func NewBundleWriter(database *db.Postgres) *BundleWriter {
	return &BundleWriter{
		database:  database,
		txOptions: sql.TxOptions{Isolation: sql.LevelReadCommitted, ReadOnly: false},
	}
}

func (b *BundleWriter) Write(ctx context.Context, tenantID string, bundles []*base.DataBundle) (names []string, err error) {
	ctx, span := tracer.Start(ctx, "bundle-writer.write-bundle")
	defer span.End()

	slog.Info("Writing bundles to the database", slog.Any("number_of_bundles", len(bundles)))

	insertBuilder := b.database.Builder.Insert(BundlesTable).
		Columns("name, payload, tenant_id").
		Suffix("ON CONFLICT (name, tenant_id) DO UPDATE SET payload = EXCLUDED.payload")

	for _, bundle := range bundles {

		names = append(names, bundle.Name)

		m := jsonpb.Marshaler{}
		jsonStr, err := m.MarshalToString(bundle)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(otelCodes.Error, err.Error())

			slog.Error("Failed to convert the value to string: ", slog.Any("error", err))

			return names, errors.New(base.ErrorCode_ERROR_CODE_INVALID_ARGUMENT.String())
		}

		insertBuilder = insertBuilder.Values(bundle.Name, jsonStr, tenantID)
	}

	var query string
	var args []interface{}

	query, args, err = insertBuilder.ToSql()
	if err != nil {

		slog.Error("Error while building SQL query: ", slog.Any("error", err))

		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		return names, errors.New(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String())
	}

	slog.Debug("Executing SQL insert query: ", slog.Any("query", query), slog.Any("arguments", args))

	_, err = b.database.DB.ExecContext(ctx, query, args...)
	if err != nil {

		slog.Error("Failed to execute insert query: ", slog.Any("error", err))

		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		return
	}

	slog.Info("Successfully wrote bundles to the database. ", slog.Any("number_of_bundles", len(bundles)))

	return
}

func (b *BundleWriter) Delete(ctx context.Context, tenantID, name string) (err error) {
	ctx, span := tracer.Start(ctx, "bundle-writer.delete-bundle")
	defer span.End()

	slog.Info("Deleting bundle: ", slog.Any("bundle", name))

	deleteBuilder := b.database.Builder.Delete(BundlesTable).Where(squirrel.Eq{"name": name, "tenant_id": tenantID})

	var query string
	var args []interface{}

	query, args, err = deleteBuilder.ToSql()
	if err != nil {

		slog.Error("Error while building SQL query: ", slog.Any("error", err))

		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		return errors.New(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String())
	}

	_, err = b.database.DB.ExecContext(ctx, query, args...)
	if err != nil {

		slog.Error("Failed to execute insert query: ", slog.Any("error", err))

		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		return
	}

	slog.Info("Successfully deleted Bundle")

	return nil
}
