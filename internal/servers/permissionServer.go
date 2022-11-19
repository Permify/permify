package servers

import (
	"fmt"

	otelCodes "go.opentelemetry.io/otel/codes"
	"golang.org/x/net/context"
	"google.golang.org/grpc/status"

	"github.com/Permify/permify/internal/services"
	"github.com/Permify/permify/pkg/logger"
	v1 "github.com/Permify/permify/pkg/pb/base/v1"
)

// PermissionServer - Structure for Permission Server
type PermissionServer struct {
	v1.UnimplementedPermissionServer

	permissionService services.IPermissionService
	logger            logger.Interface
}

// NewPermissionServer - Creates new Permission Server
func NewPermissionServer(p services.IPermissionService, l logger.Interface) *PermissionServer {
	return &PermissionServer{
		permissionService: p,
		logger:            l,
	}
}

// Check - Performs Authorization Check
func (r *PermissionServer) Check(ctx context.Context, request *v1.PermissionCheckRequest) (*v1.PermissionCheckResponse, error) {
	ctx, span := tracer.Start(ctx, "permissions.check")
	defer span.End()

	v := request.Validate()
	if v != nil {
		return nil, v
	}

	var err error
	var response *v1.PermissionCheckResponse
	response, err = r.permissionService.CheckPermissions(ctx, request)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		r.logger.Error(fmt.Sprintf(err.Error()))
		return nil, status.Error(GetStatus(err), err.Error())
	}

	return response, nil
}

// Expand - Get schema actions in a tree structure
func (r *PermissionServer) Expand(ctx context.Context, request *v1.PermissionExpandRequest) (*v1.PermissionExpandResponse, error) {
	ctx, span := tracer.Start(ctx, "permissions.expand")
	defer span.End()

	v := request.Validate()
	if v != nil {
		return nil, v
	}

	var err error
	var response *v1.PermissionExpandResponse
	response, err = r.permissionService.ExpandPermissions(ctx, request)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		r.logger.Error(fmt.Sprintf(err.Error()))
		return nil, status.Error(GetStatus(err), err.Error())
	}

	return response, nil
}

// LookupSchema -
func (r *PermissionServer) LookupSchema(ctx context.Context, request *v1.PermissionLookupSchemaRequest) (*v1.PermissionLookupSchemaResponse, error) {
	ctx, span := tracer.Start(ctx, "permissions.lookupSchema")
	defer span.End()

	v := request.Validate()
	if v != nil {
		return nil, v
	}

	var err error
	var response *v1.PermissionLookupSchemaResponse
	response, err = r.permissionService.LookupSchema(ctx, request)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		r.logger.Error(fmt.Sprintf(err.Error()))
		return nil, status.Error(GetStatus(err), err.Error())
	}

	return response, nil
}

// LookupEntity -
func (r *PermissionServer) LookupEntity(ctx context.Context, request *v1.PermissionLookupEntityRequest) (*v1.PermissionLookupEntityResponse, error) {
	ctx, span := tracer.Start(ctx, "permissions.lookupEntity")
	defer span.End()

	v := request.Validate()
	if v != nil {
		return nil, v
	}

	var err error
	var response *v1.PermissionLookupEntityResponse
	response, err = r.permissionService.LookupEntity(ctx, request)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		r.logger.Error(fmt.Sprintf(err.Error()))
		return nil, status.Error(GetStatus(err), err.Error())
	}

	return response, nil
}
