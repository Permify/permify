package postgres

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/golang/protobuf/jsonpb"
	"go.opentelemetry.io/otel/codes"

	"github.com/Permify/permify/internal/storage/postgres/utils"
	db "github.com/Permify/permify/pkg/database/postgres"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

type BundleReader struct {
	database  *db.Postgres
	txOptions sql.TxOptions
}

func NewBundleReader(database *db.Postgres) *BundleReader {
	return &BundleReader{
		database:  database,
		txOptions: sql.TxOptions{Isolation: sql.LevelReadCommitted, ReadOnly: false},
	}
}

func (b *BundleReader) Read(ctx context.Context, tenantID, name string) (bundle *base.DataBundle, err error) {
	ctx, span := tracer.Start(ctx, "bundle-reader.read-bundle")
	defer span.End()

	slog.Info("Reading bundle: ", slog.Any("tenant_id", tenantID), slog.Any("name", name))

	builder := b.database.Builder.Select("payload").From(BundlesTable).Where(squirrel.Eq{"name": name, "tenant_id": tenantID})

	var query string
	var args []interface{}

	query, args, err = builder.ToSql()
	if err != nil {
		return nil, utils.HandleError(span, err, base.ErrorCode_ERROR_CODE_SQL_BUILDER)
	}

	slog.Debug("Executing SQL query: ", slog.Any("query", query), slog.Any("arguments", args))

	var row *sql.Row
	row = b.database.DB.QueryRowContext(ctx, query, args...)

	var jsonData string
	err = row.Scan(&jsonData)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New(base.ErrorCode_ERROR_CODE_BUNDLE_NOT_FOUND.String())
		}
		return nil, utils.HandleError(span, err, base.ErrorCode_ERROR_CODE_SCAN)
	}

	m := jsonpb.Unmarshaler{}
	bundle = &base.DataBundle{}
	err = m.Unmarshal(strings.NewReader(jsonData), bundle)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		slog.Error("Failed to convert the value to bundle: ", slog.Any("error", err))

		return nil, errors.New(base.ErrorCode_ERROR_CODE_INVALID_ARGUMENT.String())
	}

	return bundle, err
}
