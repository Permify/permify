package memory

import (
	"context"
	"errors"

	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/internal/storage/memory/constants" // Memory storage constants
	db "github.com/Permify/permify/pkg/database/memory"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// BundleWriter -
type BundleWriter struct {
	database *db.Memory
}

func NewBundleWriter(database *db.Memory) *BundleWriter {
	return &BundleWriter{
		database: database,
	}
}

func (b *BundleWriter) Write(ctx context.Context, bundles []storage.Bundle) (names []string, err error) {
	txn := b.database.DB.Txn(true)
	defer txn.Abort()
	// Iterate over bundles to write
	for _, bundle := range bundles { // Process each bundle
		if err = txn.Insert(constants.BundlesTable, bundle); err != nil {
			return []string{}, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
		} // End of error check
		names = append(names, bundle.Name) // Collect bundle name
	} // End of bundle iteration
	txn.Commit() // Commit transaction
	// Return bundle names
	return names, nil // Success
}

func (b *BundleWriter) Delete(ctx context.Context, tenantID, name string) (err error) {
	txn := b.database.DB.Txn(true)                                      // Start write transaction
	raw, err := txn.First(constants.BundlesTable, "id", tenantID, name) // Find bundle
	// Check if bundle exists
	if raw == nil { // Bundle not found
		return errors.New(base.ErrorCode_ERROR_CODE_BUNDLE_NOT_FOUND.String()) // Return error
	} // End of existence check
	err = txn.Delete(constants.BundlesTable, raw) // Delete bundle
	if err != nil {
		return err
	}
	txn.Commit() // Commit transaction
	// Successfully deleted
	return nil // Success
}
