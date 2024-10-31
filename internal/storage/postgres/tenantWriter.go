package postgres

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"go.opentelemetry.io/otel/codes"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/Permify/permify/internal"
	"github.com/Permify/permify/internal/storage/postgres/utils"
	db "github.com/Permify/permify/pkg/database/postgres"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// TenantWriter - Structure for Tenant Writer
type TenantWriter struct {
	database *db.Postgres
	// options
	txOptions pgx.TxOptions
}

// NewTenantWriter - Creates a new TenantWriter
func NewTenantWriter(database *db.Postgres) *TenantWriter {
	return &TenantWriter{
		database:  database,
		txOptions: pgx.TxOptions{IsoLevel: pgx.ReadCommitted, AccessMode: pgx.ReadWrite},
	}
}

// CreateTenant - Creates a new Tenant
func (w *TenantWriter) CreateTenant(ctx context.Context, id, name string) (result *base.Tenant, err error) {
	ctx, span := internal.Tracer.Start(ctx, "tenant-writer.create-tenant")
	defer span.End()

	slog.DebugContext(ctx, "creating new tenant", slog.Any("id", id), slog.Any("name", name))

	var createdAt time.Time
	err = w.database.WritePool.QueryRow(ctx, utils.InsertTenantTemplate, id, name).Scan(&createdAt)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value") {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			slog.ErrorContext(ctx, "error encountered", slog.Any("error", err))
			return nil, errors.New(base.ErrorCode_ERROR_CODE_UNIQUE_CONSTRAINT.String())
		}
		return nil, utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_EXECUTION)
	}

	slog.DebugContext(ctx, "successfully created Tenant", slog.Any("id", id), slog.Any("name", name), slog.Any("created_at", createdAt))

	return &base.Tenant{
		Id:        id,
		Name:      name,
		CreatedAt: timestamppb.New(createdAt),
	}, nil
}

// DeleteTenant - Deletes a Tenant
func (w *TenantWriter) DeleteTenant(ctx context.Context, tenantID string) (result *base.Tenant, err error) {
	ctx, span := internal.Tracer.Start(ctx, "tenant-writer.delete-tenant")
	defer span.End()

	slog.DebugContext(ctx, "deleting tenant", slog.Any("tenant_id", tenantID))

	tx, err := w.database.WritePool.Begin(ctx)
	if err != nil {
		return nil, utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_EXECUTION)
	}
	defer tx.Rollback(ctx)

	// Prepare batch operations for deleting tenant-related records from multiple tables
	tables := []string{"bundles", "relation_tuples", "attributes", "schema_definitions", "transactions"}
	batch := &pgx.Batch{}
	var totalDeleted int64
	for _, table := range tables {
		query := fmt.Sprintf(utils.DeleteAllByTenantTemplate, table)
		batch.Queue(query, tenantID)
	}
	batch.Queue(utils.DeleteTenantTemplate, tenantID)

	// Execute the batch of delete queries
	br := tx.SendBatch(ctx, batch)

	for i := 0; i < len(tables); i++ {
		tag, err := br.Exec()
		if err != nil {
			err = br.Close()
			if err != nil {
				return nil, err
			}
			err = tx.Commit(ctx)
			if err != nil {
				return nil, utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_EXECUTION)
			}
			return nil, utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_EXECUTION)
		} else {
			totalDeleted += tag.RowsAffected()
		}
	}

	// Retrieve the tenant details after deletion
	var name string
	var createdAt time.Time
	err = br.QueryRow().Scan(&name, &createdAt)
	if err != nil {
		if totalDeleted > 0 {
			name = fmt.Sprintf("Affected rows: %d", totalDeleted)
		} else {
			return nil, utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_EXECUTION)
		}
	}

	err = br.Close()
	if err != nil {
		return nil, utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_EXECUTION)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_EXECUTION)
	}

	// Return the deleted tenant information
	return &base.Tenant{
		Id:        tenantID,
		Name:      name,
		CreatedAt: timestamppb.New(createdAt),
	}, nil
}
