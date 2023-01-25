package memory

import (
	"context"
	"errors"
	"time"

	"github.com/Permify/permify/internal/repositories"
	db "github.com/Permify/permify/pkg/database/memory"
	"github.com/Permify/permify/pkg/logger"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// TenantWriter - Structure for Tenant Writer
type TenantWriter struct {
	database *db.Memory
	// logger
	logger logger.Interface
}

// NewTenantWriter creates a new TenantWriter
func NewTenantWriter(database *db.Memory, logger logger.Interface) *TenantWriter {
	return &TenantWriter{
		database: database,
		logger:   logger,
	}
}

// CreateTenant -
func (w *TenantWriter) CreateTenant(ctx context.Context, id, name string) (result *base.Tenant, err error) {
	tenant := repositories.Tenant{
		ID:        id,
		Name:      name,
		CreatedAt: time.Now(),
	}
	txn := w.database.DB.Txn(true)
	defer txn.Abort()
	if err = txn.Insert(TenantsTable, tenant); err != nil {
		return nil, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
	}
	txn.Commit()
	return tenant.ToTenant(), nil
}

// DeleteTenant -
func (w *TenantWriter) DeleteTenant(ctx context.Context, tenantID string) (result *base.Tenant, err error) {
	txn := w.database.DB.Txn(true)
	defer txn.Abort()
	var raw interface{}
	raw, err = txn.First(TenantsTable, "id", tenantID)
	if err != nil {
		return nil, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
	}
	if _, err = txn.DeleteAll(TenantsTable, "id", tenantID); err != nil {
		return nil, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
	}
	txn.Commit()
	return raw.(repositories.Tenant).ToTenant(), nil
}
