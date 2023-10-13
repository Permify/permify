package memory

import (
	"context"
	"errors"
	"time"

	"github.com/Permify/permify/internal/storage"
	db "github.com/Permify/permify/pkg/database/memory"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// TenantWriter - Structure for Tenant Writer
type TenantWriter struct {
	database *db.Memory
}

// NewTenantWriter creates a new TenantWriter
func NewTenantWriter(database *db.Memory) *TenantWriter {
	return &TenantWriter{
		database: database,
	}
}

// CreateTenant -
func (w *TenantWriter) CreateTenant(_ context.Context, id, name string) (result *base.Tenant, err error) {
	tenant := storage.Tenant{
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
func (w *TenantWriter) DeleteTenant(_ context.Context, tenantID string) (result *base.Tenant, err error) {
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
	return raw.(storage.Tenant).ToTenant(), nil
}
