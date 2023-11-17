package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"log/slog"

	"github.com/Masterminds/squirrel"
	otelCodes "go.opentelemetry.io/otel/codes"
	"google.golang.org/protobuf/types/known/timestamppb"

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

	slog.Info("Creating new Tenant: ", slog.Any("id", id), slog.Any("name", name))

	var createdAt time.Time

	query := w.database.Builder.Insert(TenantsTable).Columns("id, name").Values(id, name).Suffix("RETURNING created_at").RunWith(w.database.DB)

	err = query.QueryRowContext(ctx).Scan(&createdAt)
	if err != nil {

		slog.Error("Error while creating tenant: ", slog.Any("error", err))

		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		if strings.Contains(err.Error(), "duplicate key value") {

			slog.Error("Duplicate key violation: Tenant with ID already exists", slog.Any("id", id))

			return nil, errors.New(base.ErrorCode_ERROR_CODE_UNIQUE_CONSTRAINT.String())
		}

		slog.Error("Error executing query: ", slog.Any("error", err))

		return nil, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
	}

	slog.Info("Successfully created Tenant", slog.Any("id", id), slog.Any("name", name), slog.Any("createdAt", createdAt))

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

	slog.Info("Deleting Tenant: ", slog.Any("tenant_id", tenantID))

	var name string
	var createdAt time.Time

	query := w.database.Builder.Delete(TenantsTable).Where(squirrel.Eq{"id": tenantID}).Suffix("RETURNING name, created_at").RunWith(w.database.DB)
	err = query.QueryRowContext(ctx).Scan(&name, &createdAt)
	if err != nil {
		slog.Error("Error while deleting tenant: ", slog.Any("error", err))
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		return nil, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
	}

	slog.Info("Successfully deleted Tenant")

	return &base.Tenant{
		Id:        tenantID,
		Name:      name,
		CreatedAt: timestamppb.New(createdAt),
	}, nil
}
