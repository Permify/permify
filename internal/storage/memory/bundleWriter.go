package memory

import (
	"context"
	"errors"

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

func (b *BundleWriter) Write(ctx context.Context, tenantID string, bundles []*base.DataBundle) (names []string, err error) {
	for _, bundle := range bundles {
		names = append(names, bundle.Name)
		b.database.Lock()

		txn := b.database.DB.Txn(true)
		err = txn.Insert(constants.BundlesTable, bundle)
		if err != nil {
			b.database.Unlock()
			return names, errors.New(err.Error())
		}
		txn.Commit()
		b.database.Unlock()
	}

	return names, nil
}

func (b *BundleWriter) Delete(ctx context.Context, tenantID, tenantName string) (err error) {
	txn := b.database.DB.Txn(true)
	existing, _ := txn.First(constants.BundlesTable, "id", tenantName)

	if existing == nil {
		return errors.New(base.ErrorCode_ERROR_CODE_BUNDLE_NOT_FOUND.String())
	}
	txn.Delete(constants.BundlesTable, existing)
	txn.Commit()

	return nil
}
