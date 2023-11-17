package postgres

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"github.com/Masterminds/squirrel"
	"go.opentelemetry.io/otel/codes"

	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/internal/storage/postgres/utils"
	"github.com/Permify/permify/pkg/database"
	db "github.com/Permify/permify/pkg/database/postgres"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

type TenantReader struct {
	database *db.Postgres
	// options
	txOptions sql.TxOptions
}

// NewTenantReader - Creates a new TenantReader
func NewTenantReader(database *db.Postgres) *TenantReader {
	return &TenantReader{
		database:  database,
		txOptions: sql.TxOptions{Isolation: sql.LevelReadCommitted, ReadOnly: true},
	}
}

// ListTenants - Lists all Tenants
func (r *TenantReader) ListTenants(ctx context.Context, pagination database.Pagination) (tenants []*base.Tenant, ct database.EncodedContinuousToken, err error) {
	ctx, span := tracer.Start(ctx, "tenant-reader.list-tenants")
	defer span.End()

	slog.Info("Listing tenants with pagination: ", slog.Any("pagination", pagination))

	builder := r.database.Builder.Select("id, name, created_at").From(TenantsTable)
	if pagination.Token() != "" {
		var t database.ContinuousToken
		t, err = utils.EncodedContinuousToken{Value: pagination.Token()}.Decode()
		if err != nil {

			slog.Error("Failed to decode pagination token. ", slog.Any("error", err))

			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return nil, nil, err
		}
		builder = builder.Where(squirrel.GtOrEq{"id": t.(utils.ContinuousToken).Value})
	}

	builder = builder.OrderBy("id").Limit(uint64(pagination.PageSize() + 1))

	var query string
	var args []interface{}

	query, args, err = builder.ToSql()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		slog.Error("Error while building SQL query: ", slog.Any("error", err))

		return nil, nil, errors.New(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String())
	}

	slog.Debug("Executing SQL query: ", slog.Any("query", query), slog.Any("arguments", args))

	var rows *sql.Rows
	rows, err = r.database.DB.QueryContext(ctx, query, args...)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		slog.Error("Failed to execute query: ", slog.Any("error", err))

		return nil, nil, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
	}
	defer rows.Close()

	var lastID string
	tenants = make([]*base.Tenant, 0, pagination.PageSize()+1)
	for rows.Next() {
		sd := storage.Tenant{}
		err = rows.Scan(&sd.ID, &sd.Name, &sd.CreatedAt)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())

			slog.Error("Failed to scan rows: ", slog.Any("error", err))

			return nil, nil, err
		}
		lastID = sd.ID
		tenants = append(tenants, sd.ToTenant())
	}
	if err = rows.Err(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		slog.Error("Error iterating over rows: ", slog.Any("error", err))

		return nil, nil, err
	}

	slog.Info("Successfully listed tenants. ", slog.Any("number_of_tenants", len(tenants)))

	if len(tenants) > int(pagination.PageSize()) {

		slog.Info("Returning tenants with a continuous token. ", slog.Any("page_size", pagination.PageSize()))
		return tenants[:pagination.PageSize()], utils.NewContinuousToken(lastID).Encode(), nil
	}

	slog.Info("Returning all tenants with no continuous token.")

	return tenants, database.NewNoopContinuousToken().Encode(), nil
}
