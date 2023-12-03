package memory

import (
	"context"
	"log/slog"

	"github.com/Permify/permify/internal/storage/postgres/utils"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/hashicorp/go-memdb"
)

type BundleWriter struct {
	database *memdb.MemDB
}

func NewBundleWriter(database *memdb.MemDB) *BundleWriter {
	return &BundleWriter{
		database: database,
	}
}

func (b *BundleWriter) Write(ctx context.Context, tenantID string, bundles []*base.DataBundle) (names []string, err error) {
	ctx, span := tracer.Start(ctx, "bundle-writer.write-bundle")
	defer span.End()

	slog.Info("Writing bundles to the database", slog.Any("number_of_bundles", len(bundles)))

	txn := b.database.Txn(true)

	for _, bundle := range bundles {

		if err := txn.Insert("bundle", bundle); err != nil {
			txn.Abort()
			return names, utils.HandleError(span, err, base.ErrorCode_ERROR_CODE_INTERNAL)
		}
	}

	txn.Commit()

	slog.Info("Successfully wrote bundles to the database. ", slog.Any("number_of_bundles", len(bundles)))

	return
}

func (b *BundleWriter) Delete(ctx context.Context, tenantID string) (err error) {
	ctx, span := tracer.Start(ctx, "bundle-writer.delete-bundle")
	defer span.End()

	slog.Info("Deleting bundle: ", slog.Any("bundle", tenantID))

	txn := b.database.Txn(true)

	raw, err := txn.First("bundle", "TenantID", tenantID)
	if err != nil {
		txn.Abort()
		return utils.HandleError(span, err, base.ErrorCode_ERROR_CODE_INTERNAL)
	}

	if err := txn.Delete("bundle", &raw); err != nil {
		txn.Abort()
		return utils.HandleError(span, err, base.ErrorCode_ERROR_CODE_INTERNAL)
	}

	txn.Commit()

	slog.Info("Successfully deleted Bundle")

	return nil
}
