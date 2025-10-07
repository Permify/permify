package memory

import (
	"context"
	"errors"

	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/internal/storage/memory/constants"
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

	for _, bundle := range bundles {
		if err = txn.Insert(constants.BundlesTable, bundle); err != nil {
			return []string{}, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
		}
		names = append(names, bundle.Name)
	}
	txn.Commit()

	return names, nil
}

func (b *BundleWriter) Delete(ctx context.Context, tenantID, name string) (err error) {
	txn := b.database.DB.Txn(true)
	raw, err := txn.First(constants.BundlesTable, "id", tenantID, name)

	if raw == nil {
		return errors.New(base.ErrorCode_ERROR_CODE_BUNDLE_NOT_FOUND.String())
	}
	err = txn.Delete(constants.BundlesTable, raw)
	if err != nil {
		return err
	}
	txn.Commit()

	return nil
}
