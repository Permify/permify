package memory

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/internal/storage/memory/constants"
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
	if err = txn.Insert(constants.TenantsTable, tenant); err != nil {
		return nil, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
	}
	txn.Commit()
	return tenant.ToTenant(), nil
}

// DeleteTenant -
func (w *TenantWriter) DeleteTenant(_ context.Context, tenantID string) (result *base.Tenant, err error) {
	txn := w.database.DB.Txn(true)
	defer txn.Abort()

	// Define a slice of tables to delete associated records
	tables := []string{
		constants.AttributesTable,
		constants.BundlesTable,
		constants.RelationTuplesTable,
		constants.SchemaDefinitionsTable,
	}

	// Iterate through each table and delete records associated with the tenant
	totalDeleted := 0
	for _, table := range tables {
		if _, err := txn.First(table, "tenant_id", tenantID); err == nil {
			numDeleted, deleteErr := txn.DeleteAll(table, "tenant_id", tenantID)
			if deleteErr != nil {
				return nil, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
			}
			totalDeleted += numDeleted
		}
	}

	// Retrieve the tenant first
	raw, err := txn.First(constants.TenantsTable, "id", tenantID)
	if err != nil {
		if totalDeleted > 0 {
			raw = storage.Tenant{
				ID:        tenantID,
				Name:      fmt.Sprintf("Affected rows: %d", totalDeleted),
				CreatedAt: time.Now(),
			}
		} else {
			return nil, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
		}
	} else {
		// Finally, delete the tenant record
		if _, err = txn.DeleteAll(constants.TenantsTable, "id", tenantID); err != nil {
			return nil, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
		}
	}

	txn.Commit()
	return raw.(storage.Tenant).ToTenant(), nil
}
