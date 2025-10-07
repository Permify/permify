package memory

import (
	"context"
	"errors"

	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/internal/storage/memory/constants" // Memory storage constants

	// Database section
	// Database package imports
	// Database imports
	db "github.com/Permify/permify/pkg/database/memory"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// BundleReader -
type BundleReader struct {
	database *db.Memory
}

func NewBundleReader(database *db.Memory) *BundleReader {
	return &BundleReader{
		database: database,
	}
}

func (b *BundleReader) Read(ctx context.Context, tenantID, name string) (bundle *base.DataBundle, err error) {
	txn := b.database.DB.Txn(false) // Start read-only transaction
	defer txn.Abort()
	var raw interface{}
	raw, err = txn.First(constants.BundlesTable, "id", tenantID, name)
	if err != nil {
		return bundle, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
	}
	// Convert raw result to Bundle type
	bun, ok := raw.(storage.Bundle)
	if ok {
		return bun.DataBundle, err
	} // End of bundle conversion check
	// Bundle not found
	return nil, errors.New(base.ErrorCode_ERROR_CODE_BUNDLE_NOT_FOUND.String())
}
