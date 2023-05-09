package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Masterminds/squirrel"
	"go.opentelemetry.io/otel/codes"

	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/internal/storage/postgres/utils"
	"github.com/Permify/permify/pkg/database"
	db "github.com/Permify/permify/pkg/database/postgres"
	"github.com/Permify/permify/pkg/logger"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

type TenantReader struct {
	database *db.Postgres
	// options
	txOptions sql.TxOptions
	// logger
	logger logger.Interface
}

// NewTenantReader - Creates a new TenantReader
func NewTenantReader(database *db.Postgres, logger logger.Interface) *TenantReader {
	return &TenantReader{
		database:  database,
		txOptions: sql.TxOptions{Isolation: sql.LevelReadCommitted, ReadOnly: true},
		logger:    logger,
	}
}

// ListTenants - Lists all Tenants
func (r *TenantReader) ListTenants(ctx context.Context, pagination database.Pagination) (tenants []*base.Tenant, ct database.EncodedContinuousToken, err error) {
	ctx, span := tracer.Start(ctx, "tenant-reader.list-tenants")
	defer span.End()

	builder := r.database.Builder.Select("id, name, created_at").From(TenantsTable)
	if pagination.Token() != "" {
		var t database.ContinuousToken
		t, err = utils.EncodedContinuousToken{Value: pagination.Token()}.Decode()
		if err != nil {
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
		return nil, nil, errors.New(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String())
	}

	var rows *sql.Rows
	rows, err = r.database.DB.QueryContext(ctx, query, args...)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
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
			return nil, nil, err
		}
		lastID = sd.ID
		tenants = append(tenants, sd.ToTenant())
	}
	if err = rows.Err(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, nil, err
	}

	if len(tenants) > int(pagination.PageSize()) {
		return tenants[:pagination.PageSize()], utils.NewContinuousToken(lastID).Encode(), nil
	}

	return tenants, utils.NewNoopContinuousToken().Encode(), nil
}
