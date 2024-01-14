package memory

import (
	"context"
	"errors"

	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/internal/storage/memory/constants"

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
	txn := b.database.DB.Txn(false)
	defer txn.Abort()
	var raw interface{}
	raw, err = txn.First(constants.BundlesTable, "id", tenantID, name)
	if err != nil {
		return bundle, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
	}

	bun, ok := raw.(storage.Bundle)
	if ok {
		return bun.DataBundle, err
	}

	return nil, errors.New(base.ErrorCode_ERROR_CODE_BUNDLE_NOT_FOUND.String())
}
