package memory

import (
	"context"
	"errors"

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

func (b *BundleReader) Read(ctx context.Context, tenantID, tenantName string) (bundle *base.DataBundle, err error) {
	b.database.RLock()
	defer b.database.RUnlock()
	txn := b.database.DB.Txn(false)
	raw, _ := txn.First(constants.BundlesTable, "id", tenantName)
	bundle, ok := raw.(*base.DataBundle)
	if !ok {

		return nil, errors.New(base.ErrorCode_ERROR_CODE_BUNDLE_NOT_FOUND.String())
	}
	return bundle, nil
}
