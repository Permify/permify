package memory

import (
	"context"
	"encoding/json"
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

func (b *BundleReader) Read(ctx context.Context, tenantID, name string) (bundle *base.DataBundle, err error) {
	data, err := b.database.Get(BundlesTable, name)
	if err != nil {
		return nil, errors.New(base.ErrorCode_ERROR_CODE_NOT_FOUND.String())
	}

	bundle = &base.DataBundle{}

	dataBytes, ok := data.([]byte)
	if !ok {
		return nil, errors.New("data is not of type []byte")
	}

	err = json.Unmarshal(dataBytes, bundle)
	if err != nil {
		return nil, errors.New(base.ErrorCode_ERROR_CODE_INVALID_ARGUMENT.String())
	}

	return bundle, nil
}
