package servers

import (
	"fmt"

	otelCodes "go.opentelemetry.io/otel/codes"
	"golang.org/x/net/context"
	"google.golang.org/grpc/status"

	"github.com/Permify/permify/internal/commands"
	"github.com/Permify/permify/internal/services"
	"github.com/Permify/permify/pkg/logger"
	v1 "github.com/Permify/permify/pkg/pb/base/v1"
)

// PermissionServer -
type PermissionServer struct {
	v1.UnimplementedPermissionServer

	permissionService services.IPermissionService
	l                 logger.Interface
}

// NewPermissionServer -
func NewPermissionServer(p services.IPermissionService, l logger.Interface) *PermissionServer {
	return &PermissionServer{
		permissionService: p,
		l:                 l,
	}
}

// Check -
func (r *PermissionServer) Check(ctx context.Context, request *v1.PermissionCheckRequest) (*v1.PermissionCheckResponse, error) {
	ctx, span := tracer.Start(ctx, "permissions.check")
	defer span.End()

	var depth int32 = 20
	if request.Depth != nil {
		depth = request.Depth.Value
	}

	v := request.Validate()
	if v != nil {
		return nil, v
	}

	var err error
	var response commands.CheckResponse
	response, err = r.permissionService.CheckPermissions(ctx, request.GetSubject(), request.GetAction(), request.GetEntity(), request.GetSchemaVersion(), request.GetSnapToken(), depth)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		r.l.Error(fmt.Sprintf(err.Error()))
		return nil, status.Error(GetStatus(err), err.Error())
	}

	return &v1.PermissionCheckResponse{
		Can:            response.Can,
		RemainingDepth: response.RemainingDepth,
	}, nil
}

// Expand -
func (r *PermissionServer) Expand(ctx context.Context, request *v1.PermissionExpandRequest) (*v1.PermissionExpandResponse, error) {
	ctx, span := tracer.Start(ctx, "permissions.expand")
	defer span.End()

	v := request.Validate()
	if v != nil {
		return nil, v
	}

	var err error
	var response commands.ExpandResponse
	response, err = r.permissionService.ExpandPermissions(ctx, request.GetEntity(), request.GetAction(), request.GetSchemaVersion(), request.GetSnapToken())
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		r.l.Error(fmt.Sprintf(err.Error()))
		return nil, status.Error(GetStatus(err), err.Error())
	}

	return &v1.PermissionExpandResponse{Tree: response.Tree}, nil
}

// LookupQuery -
func (r *PermissionServer) LookupQuery(ctx context.Context, request *v1.PermissionLookupQueryRequest) (*v1.PermissionLookupQueryResponse, error) {
	ctx, span := tracer.Start(ctx, "permissions.lookupQuery")
	defer span.End()

	v := request.Validate()
	if v != nil {
		return nil, v
	}

	var err error
	var response commands.LookupQueryResponse
	response, err = r.permissionService.LookupQueryPermissions(ctx, request.GetEntityType(), request.GetSubject(), request.GetAction(), request.GetSchemaVersion())
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		r.l.Error(fmt.Sprintf(err.Error()))
		return nil, status.Error(GetStatus(err), err.Error())
	}

	return &v1.PermissionLookupQueryResponse{
		Query: response.Query,
	}, nil
}
