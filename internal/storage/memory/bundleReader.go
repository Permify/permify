package memory

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/Permify/permify/internal/storage/memory/utils"
	db "github.com/Permify/permify/pkg/database/memory"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/golang/protobuf/jsonpb"
	"go.opentelemetry.io/otel/codes"
)

type BundleReader struct {
	database *db.Memory
}

func NewBundleReader(database *db.Memory) *BundleReader {
	return &BundleReader{
		database: database,
	}
}

func (b *BundleReader) Read(ctx context.Context, tenantID, name string) (bundle *base.DataBundle, err error) {
	ctx, span := tracer.Start(ctx, "bundle-reader.read-bundle")
	defer span.End()

	slog.Info("Reading bundle: ", slog.Any("tenant_id", tenantID), slog.Any("name", name))

	txn := b.database.DB.Txn(false)
	defer txn.Abort()

	raw, err := txn.First("bundle", "id", tenantID, name)
	if err != nil {
		return nil, utils.HandleError(span, err, base.ErrorCode_ERROR_CODE_SQL_BUILDER)
	}

	if raw == nil {
		return nil, errors.New(base.ErrorCode_ERROR_CODE_BUNDLE_NOT_FOUND.String())
	}

	jsonData := raw.(*base.DataBundle).String()

	m := jsonpb.Unmarshaler{}
	bundle = &base.DataBundle{}
	err = m.Unmarshal(strings.NewReader(jsonData), bundle)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		slog.Error("Failed to convert the value to bundle: ", slog.Any("error", err))

		return nil, errors.New(base.ErrorCode_ERROR_CODE_INVALID_ARGUMENT.String())
	}

	return bundle, err
}
