package servers

import (
	"context"
	"log/slog"

	otelCodes "go.opentelemetry.io/otel/codes"
	"google.golang.org/grpc/status"

	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/internal/validation"
	v1 "github.com/Permify/permify/pkg/pb/base/v1"
)

// BundleServer handles bundle operations.
type BundleServer struct {
	v1.UnimplementedBundleServer

	br storage.BundleReader
	bw storage.BundleWriter
}

func NewBundleServer(
	br storage.BundleReader,
	bw storage.BundleWriter,
) *BundleServer {
	return &BundleServer{
		br: br,
		bw: bw,
	}
}

// Write handles the writing of bundles.
func (r *BundleServer) Write(ctx context.Context, request *v1.BundleWriteRequest) (*v1.BundleWriteResponse, error) {
	ctx, span := tracer.Start(ctx, "bundle.write")
	defer span.End()

	v := request.Validate()
	if v != nil {
		return nil, v
	}

	for _, bundle := range request.GetBundles() {
		for _, operation := range bundle.GetOperations() {
			err := validation.ValidateBundleOperation(operation)
			if err != nil {
				return nil, err
			}
		}
	}

	names, err := r.bw.Write(ctx, request.GetTenantId(), request.GetBundles())
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		slog.Error(err.Error())
		return nil, status.Error(GetStatus(err), err.Error())
	}

	return &v1.BundleWriteResponse{
		Names: names,
	}, nil
}

// Read handles the reading of bundles.
func (r *BundleServer) Read(ctx context.Context, request *v1.BundleReadRequest) (*v1.BundleReadResponse, error) {
	ctx, span := tracer.Start(ctx, "bundle.read")
	defer span.End()

	v := request.Validate()
	if v != nil {
		return nil, v
	}

	bundle, err := r.br.Read(ctx, request.GetTenantId(), request.GetName())
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		slog.Error(err.Error())
		return nil, status.Error(GetStatus(err), err.Error())
	}

	return &v1.BundleReadResponse{
		Bundle: bundle,
	}, nil
}

// Delete handles the deletion of bundles.
func (r *BundleServer) Delete(ctx context.Context, request *v1.BundleDeleteRequest) (*v1.BundleDeleteResponse, error) {
	ctx, span := tracer.Start(ctx, "bundle.delete")
	defer span.End()

	v := request.Validate()
	if v != nil {
		return nil, v
	}

	err := r.bw.Delete(ctx, request.GetTenantId(), request.GetName())
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		slog.Error(err.Error())
		return nil, status.Error(GetStatus(err), err.Error())
	}

	return &v1.BundleDeleteResponse{
		Name: request.GetName(),
	}, nil
}
