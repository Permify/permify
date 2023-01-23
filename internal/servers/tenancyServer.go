package servers

import (
	"context"

	otelCodes "go.opentelemetry.io/otel/codes"
	"google.golang.org/grpc/status"

	"github.com/Permify/permify/internal/services"
	"github.com/Permify/permify/pkg/logger"
	v1 "github.com/Permify/permify/pkg/pb/base/v1"
)

// TenancyServer - Structure for Tenancy Server
type TenancyServer struct {
	v1.UnimplementedTenancyServer

	tenancyService services.ITenancyService
	logger         logger.Interface
}

// NewTenancyServer - Creates new Tenancy Server
func NewTenancyServer(s services.ITenancyService, l logger.Interface) *TenancyServer {
	return &TenancyServer{
		tenancyService: s,
		logger:         l,
	}
}

// Create - Create new Tenant
func (t *TenancyServer) Create(ctx context.Context, request *v1.TenantCreateRequest) (*v1.TenantCreateResponse, error) {
	ctx, span := tracer.Start(ctx, "tenant.create")
	defer span.End()

	tenant, err := t.tenancyService.CreateTenant(ctx, request.GetName())
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		t.logger.Error(err.Error())
		return nil, status.Error(GetStatus(err), err.Error())
	}

	return &v1.TenantCreateResponse{
		Tenant: tenant,
	}, nil
}

// Delete - Delete a Tenant
func (t *TenancyServer) Delete(ctx context.Context, request *v1.TenantDeleteRequest) (*v1.TenantDeleteResponse, error) {
	ctx, span := tracer.Start(ctx, "tenant.delete")
	defer span.End()

	tenant, err := t.tenancyService.DeleteTenant(ctx, request.GetId())
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		t.logger.Error(err.Error())
		return nil, status.Error(GetStatus(err), err.Error())
	}

	return &v1.TenantDeleteResponse{
		Tenant: tenant,
	}, nil
}

// List - List Tenants
func (t *TenancyServer) List(ctx context.Context, request *v1.TenantListRequest) (*v1.TenantListResponse, error) {
	ctx, span := tracer.Start(ctx, "tenant.list")
	defer span.End()

	tenants, ct, err := t.tenancyService.ListTenants(ctx, request.GetPageSize(), request.GetContinuousToken())
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		t.logger.Error(err.Error())
		return nil, status.Error(GetStatus(err), err.Error())
	}

	return &v1.TenantListResponse{
		Tenants:         tenants,
		ContinuousToken: ct.String(),
	}, nil
}
