package postgres

import (
	"context"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5"

	"github.com/Masterminds/squirrel"
	"go.opentelemetry.io/otel/codes"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/Permify/permify/internal"
	"github.com/Permify/permify/internal/storage/postgres/utils"
	db "github.com/Permify/permify/pkg/database/postgres"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

type BundleReader struct {
	database  *db.Postgres
	txOptions pgx.TxOptions
}

func NewBundleReader(database *db.Postgres) *BundleReader {
	return &BundleReader{
		database:  database,
		txOptions: pgx.TxOptions{IsoLevel: pgx.ReadCommitted, AccessMode: pgx.ReadWrite},
	}
}

func (b *BundleReader) Read(ctx context.Context, tenantID, name string) (bundle *base.DataBundle, err error) {
	ctx, span := internal.Tracer.Start(ctx, "bundle-reader.read-bundle")
	defer span.End()

	slog.DebugContext(ctx, "reading bundle", slog.Any("tenant_id", tenantID), slog.Any("name", name))

	builder := b.database.Builder.Select("payload").From(BundlesTable).Where(squirrel.Eq{"name": name, "tenant_id": tenantID})

	var query string
	var args []interface{}

	query, args, err = builder.ToSql()
	if err != nil {
		return nil, utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_SQL_BUILDER)
	}

	slog.DebugContext(ctx, "executing sql query", slog.Any("query", query), slog.Any("arguments", args))

	row := b.database.ReadPool.QueryRow(ctx, query, args...)

	var jsonData string
	err = row.Scan(&jsonData)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New(base.ErrorCode_ERROR_CODE_BUNDLE_NOT_FOUND.String())
		}
		return nil, utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_SCAN)
	}

	bundle = &base.DataBundle{}
	err = protojson.Unmarshal([]byte(jsonData), bundle)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		slog.ErrorContext(ctx, "failed to convert the value to bundle", slog.Any("error", err))

		return nil, errors.New(base.ErrorCode_ERROR_CODE_INVALID_ARGUMENT.String())
	}

	return bundle, err
}
