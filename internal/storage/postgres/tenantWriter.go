package postgres

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"strings"
	"time"

	"go.opentelemetry.io/otel/codes"

	"github.com/Masterminds/squirrel"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/Permify/permify/internal/storage/postgres/utils"
	db "github.com/Permify/permify/pkg/database/postgres"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// TenantWriter - Structure for Tenant Writer
type TenantWriter struct {
	database *db.Postgres
	// options
	txOptions sql.TxOptions
}

// NewTenantWriter - Creates a new TenantWriter
func NewTenantWriter(database *db.Postgres) *TenantWriter {
	return &TenantWriter{
		database:  database,
		txOptions: sql.TxOptions{Isolation: sql.LevelReadCommitted, ReadOnly: false},
	}
}

// CreateTenant - Creates a new Tenant
func (w *TenantWriter) CreateTenant(ctx context.Context, id, name string) (result *base.Tenant, err error) {
	ctx, span := tracer.Start(ctx, "tenant-writer.create-tenant")
	defer span.End()

	slog.Debug("creating new tenant", slog.Any("id", id), slog.Any("name", name))

	var createdAt time.Time

	query := w.database.Builder.Insert(TenantsTable).Columns("id, name").Values(id, name).Suffix("RETURNING created_at").RunWith(w.database.DB)

	err = query.QueryRowContext(ctx).Scan(&createdAt)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value") {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			slog.Error("error encountered", slog.Any("error", err))
			return nil, errors.New(base.ErrorCode_ERROR_CODE_UNIQUE_CONSTRAINT.String())
		}
		return nil, utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_EXECUTION)
	}

	slog.Debug("successfully created Tenant", slog.Any("id", id), slog.Any("name", name), slog.Any("created_at", createdAt))

	return &base.Tenant{
		Id:        id,
		Name:      name,
		CreatedAt: timestamppb.New(createdAt),
	}, nil
}

// DeleteTenant - Deletes a Tenant
func (w *TenantWriter) DeleteTenant(ctx context.Context, tenantID string) (result *base.Tenant, err error) {
	ctx, span := tracer.Start(ctx, "tenant-writer.delete-tenant")
	defer span.End()

	slog.Debug("deleting tenant", slog.Any("tenant_id", tenantID))

	var name string
	var createdAt time.Time

	query := w.database.Builder.Delete(TenantsTable).Where(squirrel.Eq{"id": tenantID}).Suffix("RETURNING name, created_at").RunWith(w.database.DB)
	err = query.QueryRowContext(ctx).Scan(&name, &createdAt)
	if err != nil {
		return nil, utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_EXECUTION)
	}

	slog.Debug("successfully deleted tenant")

	return &base.Tenant{
		Id:        tenantID,
		Name:      name,
		CreatedAt: timestamppb.New(createdAt),
	}, nil
}
