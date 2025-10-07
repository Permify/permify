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
func (w *TenantWriter) DeleteTenant(ctx context.Context, tenantID string) (err error) {
	ctx, span := internal.Tracer.Start(ctx, "tenant-writer.delete-tenant")
	defer span.End()

	slog.DebugContext(ctx, "deleting tenant", slog.Any("tenant_id", tenantID))

	tx, err := w.database.WritePool.Begin(ctx)
	if err != nil {
		return utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_EXECUTION)
	}
	defer tx.Rollback(ctx)

	// Check if tenant exists first
	var exists bool
	err = tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM "+TenantsTable+" WHERE id = $1)", tenantID).Scan(&exists)
	if err != nil {
		return utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_EXECUTION)
	}
	if !exists {
		return utils.HandleError(ctx, span, errors.New("tenant not found"), base.ErrorCode_ERROR_CODE_NOT_FOUND)
	}

	// Prepare batch operations for deleting tenant-related records from multiple tables
	tables := []string{BundlesTable, RelationTuplesTable, AttributesTable, SchemaDefinitionTable, TransactionsTable}
	batch := &pgx.Batch{}
	for _, table := range tables {
		query := fmt.Sprintf(utils.DeleteAllByTenantTemplate, table)
		batch.Queue(query, tenantID)
	}
	batch.Queue(utils.DeleteTenantTemplate, tenantID)

	// Execute the batch of delete queries
	br := tx.SendBatch(ctx, batch)

	for range tables {
		_, err := br.Exec()
		if err != nil {
			originalErr := err
			closeErr := br.Close()
			if closeErr != nil {
				return closeErr
			}
			// Don't commit on error, let defer tx.Rollback() handle it
			return utils.HandleError(ctx, span, originalErr, base.ErrorCode_ERROR_CODE_EXECUTION)
		}
	}

	// Execute the tenant deletion
	_, err = br.Exec()
	if err != nil {
		originalErr := err
		closeErr := br.Close()
		if closeErr != nil {
			return closeErr
		}
		// Don't commit on error, let defer tx.Rollback() handle it
		return utils.HandleError(ctx, span, originalErr, base.ErrorCode_ERROR_CODE_EXECUTION)
	}

	err = br.Close()
	if err != nil {
		return utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_EXECUTION)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return utils.HandleError(ctx, span, err, base.ErrorCode_ERROR_CODE_EXECUTION)
	}

	return nil
}
