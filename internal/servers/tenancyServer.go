package servers

import (
	"context"
	"log/slog"

	otelCodes "go.opentelemetry.io/otel/codes"
	"google.golang.org/grpc/status"

	"github.com/Permify/permify/internal"
	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/pkg/database"
	v1 "github.com/Permify/permify/pkg/pb/base/v1"
)

// TenancyServer - Structure for Tenancy Server
type TenancyServer struct {
	v1.UnimplementedTenancyServer

	tr storage.TenantReader
	tw storage.TenantWriter
}

// NewTenancyServer - Creates new Tenancy Server
func NewTenancyServer(tr storage.TenantReader, tw storage.TenantWriter) *TenancyServer {
	return &TenancyServer{
		tr: tr,
		tw: tw,
	}
}

// Create - Create new Tenant
func (t *TenancyServer) Create(ctx context.Context, request *v1.TenantCreateRequest) (*v1.TenantCreateResponse, error) {
	ctx, span := internal.Tracer.Start(ctx, "tenant.create")
	defer span.End()

	tenant, err := t.tw.CreateTenant(ctx, request.GetId(), request.GetName())
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		slog.ErrorContext(ctx, err.Error())
		return nil, status.Error(GetStatus(err), err.Error())
	}

	return &v1.TenantCreateResponse{
		Tenant: tenant,
	}, nil
}

// Delete - Delete a Tenant
func (t *TenancyServer) Delete(ctx context.Context, request *v1.TenantDeleteRequest) (*v1.TenantDeleteResponse, error) {
	ctx, span := internal.Tracer.Start(ctx, "tenant.delete")
	defer span.End()

	tenant, err := t.tw.DeleteTenant(ctx, request.GetId())
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		slog.ErrorContext(ctx, err.Error())
		return nil, status.Error(GetStatus(err), err.Error())
	}

	return &v1.TenantDeleteResponse{
		Tenant: tenant,
	}, nil
}

// List - List Tenants
func (t *TenancyServer) List(ctx context.Context, request *v1.TenantListRequest) (*v1.TenantListResponse, error) {
	ctx, span := internal.Tracer.Start(ctx, "tenant.list")
	defer span.End()

	tenants, ct, err := t.tr.ListTenants(ctx, database.NewPagination(database.Size(request.GetPageSize()), database.Token(request.GetContinuousToken())))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		slog.ErrorContext(ctx, err.Error())
		return nil, status.Error(GetStatus(err), err.Error())
	}

	return &v1.TenantListResponse{
		Tenants:         tenants,
		ContinuousToken: ct.String(),
	}, nil
}
