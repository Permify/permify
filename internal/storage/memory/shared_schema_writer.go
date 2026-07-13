package memory

import (
	"context"
	"errors"

	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/internal/storage/memory/constants"
	db "github.com/Permify/permify/pkg/database/memory"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// SharedSchemaWriter - Structure for Shared Schema Writer
type SharedSchemaWriter struct {
	database *db.Memory
}

// NewSharedSchemaWriter creates a new SharedSchemaWriter
func NewSharedSchemaWriter(database *db.Memory) *SharedSchemaWriter {
	return &SharedSchemaWriter{
		database: database,
	}
}

// WriteSharedSchema writes a shared schema to repository
func (w *SharedSchemaWriter) WriteSharedSchema(_ context.Context, definitions []storage.SharedSchemaDefinition) error {
	txn := w.database.DB.Txn(true)
	defer txn.Abort()

	var sharedSchemaID string
	var version string

	for _, definition := range definitions {
		if err := txn.Insert(constants.SharedSchemaDefinitionsTable, definition); err != nil {
			return errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
		}
		sharedSchemaID = definition.SharedSchemaID
		version = definition.Version
	}
	txn.Commit()

	sharedHeadVersionMu.Lock()
	sharedHeadVersion[sharedSchemaID] = version
	sharedHeadVersionMu.Unlock()

	return nil
}

// AssignSharedSchema sets shared_schema_id on the given tenants.
func (w *SharedSchemaWriter) AssignSharedSchema(_ context.Context, sharedSchemaID string, tenantIDs []string) error {
	txn := w.database.DB.Txn(true)
	defer txn.Abort()

	for _, tenantID := range tenantIDs {
		raw, err := txn.First(constants.TenantsTable, "id", tenantID)
		if err != nil {
			return errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
		}
		if raw == nil {
			continue
		}
		tenant := raw.(storage.Tenant)
		tenant.SharedSchemaID = sharedSchemaID
		if err := txn.Insert(constants.TenantsTable, tenant); err != nil {
			return errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
		}
	}

	txn.Commit()
	return nil
}
