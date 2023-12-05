package memory

import (
	"context"
	"encoding/json"
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

func (b *BundleWriter) Write(ctx context.Context, tenantID string, bundles []*base.DataBundle) (names []string, err error) {
	for _, bundle := range bundles {
		names = append(names, bundle.Name)

		jsonStr, err := json.Marshal(bundle)
		if err != nil {
			return names, errors.New(base.ErrorCode_ERROR_CODE_INVALID_ARGUMENT.String())
		}

		err = b.database.Set(BundlesTable, bundle.Name, jsonStr)
		if err != nil {
			return names, errors.New(base.ErrorCode_ERROR_CODE_INTERNAL.String())
		}
	}

	return names, nil
}

func (b *BundleWriter) Delete(ctx context.Context, tenantID, name string) (err error) {
	err = b.database.Delete(BundlesTable, name)
	if err != nil {
		return errors.New(base.ErrorCode_ERROR_CODE_NOT_FOUND.String())
	}

	return nil
}
