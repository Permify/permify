package postgres

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/Masterminds/squirrel"

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

	slog.Debug("listing tenants with pagination", slog.Any("pagination", pagination))

	builder := r.database.Builder.Select("id, name, created_at").From(TenantsTable)
	if pagination.Token() != "" {
		var t database.ContinuousToken
		t, err = utils.EncodedContinuousToken{Value: pagination.Token()}.Decode()
		if err != nil {
			return nil, nil, utils.HandleError(span, err, base.ErrorCode_ERROR_CODE_INTERNAL)
		}
		builder = builder.Where(squirrel.GtOrEq{"id": t.(utils.ContinuousToken).Value})
	}

	builder = builder.OrderBy("id").Limit(uint64(pagination.PageSize() + 1))

	var query string
	var args []interface{}

	query, args, err = builder.ToSql()
	if err != nil {
		return nil, nil, utils.HandleError(span, err, base.ErrorCode_ERROR_CODE_SQL_BUILDER)
	}

	slog.Debug("executing sql query", slog.Any("query", query), slog.Any("arguments", args))

	var rows *sql.Rows
	rows, err = r.database.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, nil, utils.HandleError(span, err, base.ErrorCode_ERROR_CODE_EXECUTION)
	}
	defer rows.Close()

	var lastID string
	tenants = make([]*base.Tenant, 0, pagination.PageSize()+1)
	for rows.Next() {
		sd := storage.Tenant{}
		err = rows.Scan(&sd.ID, &sd.Name, &sd.CreatedAt)
		if err != nil {
			return nil, nil, utils.HandleError(span, err, base.ErrorCode_ERROR_CODE_SCAN)
		}
		lastID = sd.ID
		tenants = append(tenants, sd.ToTenant())
	}
	if err = rows.Err(); err != nil {
		return nil, nil, utils.HandleError(span, err, base.ErrorCode_ERROR_CODE_INTERNAL)
	}

	slog.Debug("successfully listed tenants", slog.Any("number_of_tenants", len(tenants)))

	if len(tenants) > int(pagination.PageSize()) {
		return tenants[:pagination.PageSize()], utils.NewContinuousToken(lastID).Encode(), nil
	}

	return tenants, database.NewNoopContinuousToken().Encode(), nil
}
