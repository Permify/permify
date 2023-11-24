package memory

import (
	"context"
	"errors"

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

func (b *BundleReader) Read(_ context.Context, _, _ string) (bundle *base.DataBundle, err error) {
	return nil, errors.New(base.ErrorCode_ERROR_CODE_NOT_IMPLEMENTED.String())
}
