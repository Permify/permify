package memory

import (
	"context"
	"errors"

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

func (b *BundleWriter) Write(_ context.Context, _ string, _ []*base.DataBundle) (names []string, err error) {
	return nil, errors.New(base.ErrorCode_ERROR_CODE_NOT_IMPLEMENTED.String())
}

func (b *BundleWriter) Delete(_ context.Context, _, _ string) (err error) {
	return errors.New(base.ErrorCode_ERROR_CODE_NOT_IMPLEMENTED.String())
}
