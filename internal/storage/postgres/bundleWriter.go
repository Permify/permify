package postgres

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5"

	"github.com/Masterminds/squirrel"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/Permify/permify/internal"
	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/internal/storage/postgres/utils"
	db "github.com/Permify/permify/pkg/database/postgres"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

type BundleWriter struct {
	database *db.Postgres
	// options
	txOptions pgx.TxOptions
}

func NewBundleWriter(database *db.Postgres) *BundleWriter {
	return &BundleWriter{
		database:  database,
		txOptions: pgx.TxOptions{IsoLevel: pgx.ReadCommitted, AccessMode: pgx.ReadWrite},
	}
}

func (b *BundleWriter) Write(ctx context.Context, bundles []storage.Bundle) (names []string, err error) {
	ctx, span := internal.Tracer.Start(ctx, "bundle-writer.write-bundle")
	defer span.End()

	slog.DebugContext(ctx, "writing bundles to the database", slog.Any("number_of_bundles", len(bundles)))

	insertBuilder := b.database.Builder.Insert(BundlesTable).
		Columns("name, payload, tenant_id").
		Suffix("ON CONFLICT (name, tenant_id) DO UPDATE SET payload = EXCLUDED.payload")

	for _, bundle := range bundles {

		names = append(names, bundle.Name)

		jsonBytes, err := protojson.Marshal(bundle.DataBundle)
		if err != nil {
			return names, utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_INVALID_ARGUMENT)
		}
		jsonStr := string(jsonBytes)

		insertBuilder = insertBuilder.Values(bundle.Name, jsonStr, bundle.TenantID)
	}

	var query string
	var args []interface{}

	query, args, err = insertBuilder.ToSql()
	if err != nil {
		return names, utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_SQL_BUILDER)
	}

	slog.DebugContext(ctx, "executing sql insert query", slog.Any("query", query), slog.Any("arguments", args))

	_, err = b.database.WritePool.Exec(ctx, query, args...)
	if err != nil {
		return nil, utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_EXECUTION)
	}

	slog.DebugContext(ctx, "successfully wrote bundles to the database", slog.Any("number_of_bundles", len(bundles)))

	return
}

func (b *BundleWriter) Delete(ctx context.Context, tenantID, name string) (err error) {
	ctx, span := internal.Tracer.Start(ctx, "bundle-writer.delete-bundle")
	defer span.End()

	slog.DebugContext(ctx, "deleting bundle", slog.Any("bundle", name))

	deleteBuilder := b.database.Builder.Delete(BundlesTable).Where(squirrel.Eq{"name": name, "tenant_id": tenantID})

	var query string
	var args []interface{}

	query, args, err = deleteBuilder.ToSql()
	if err != nil {
		return utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_SQL_BUILDER)
	}

	_, err = b.database.WritePool.Exec(ctx, query, args...)
	if err != nil {
		return utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_EXECUTION)
	}

	slog.DebugContext(ctx, "bundle successfully deleted")

	return nil
}
