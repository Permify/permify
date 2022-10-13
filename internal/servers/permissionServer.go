package servers

import (
	"fmt"

	otelCodes "go.opentelemetry.io/otel/codes"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Permify/permify/internal/commands"
	"github.com/Permify/permify/internal/services"
	"github.com/Permify/permify/pkg/database"
	"github.com/Permify/permify/pkg/errors"
	"github.com/Permify/permify/pkg/logger"
	"github.com/Permify/permify/pkg/pb/base/v1"
)

// PermissionServer -
type PermissionServer struct {
	v1.UnimplementedPermissionAPIServer

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
func (r *PermissionServer) Check(ctx context.Context, request *v1.CheckRequest) (*v1.CheckResponse, error) {
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

	var err errors.Error
	var response commands.CheckResponse
	response, err = r.permissionService.Check(ctx, request.GetSubject(), request.GetAction(), request.GetEntity(), request.GetSchemaVersion(), depth)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		r.l.Error(fmt.Sprintf(err.Error()))
		switch err.Kind() {
		case errors.Database:
			return nil, status.Error(database.GetKindToGRPCStatus(err.SubKind()), err.Error())
		case errors.Validation:
			return nil, status.Error(codes.InvalidArgument, err.Error())
		case errors.Service:
			return nil, status.Error(codes.Internal, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &v1.CheckResponse{
		Can: response.Can,
		// Decisions:      response.Visits,
		RemainingDepth: response.RemainingDepth,
	}, nil
}

// Expand -
func (r *PermissionServer) Expand(ctx context.Context, request *v1.ExpandRequest) (*v1.ExpandResponse, error) {
	ctx, span := tracer.Start(ctx, "permissions.expand")
	defer span.End()

	v := request.Validate()
	if v != nil {
		return nil, v
	}

	var err errors.Error
	var response commands.ExpandResponse
	response, err = r.permissionService.Expand(ctx, request.GetEntity(), request.GetAction(), request.GetSchemaVersion())
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		r.l.Error(fmt.Sprintf(err.Error()))
		switch err.Kind() {
		case errors.Database:
			return nil, status.Error(database.GetKindToGRPCStatus(err.SubKind()), err.Error())
		case errors.Validation:
			return nil, status.Error(codes.InvalidArgument, err.Error())
		case errors.Service:
			return nil, status.Error(codes.Internal, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &v1.ExpandResponse{Tree: response.Tree}, nil
}

// LookupQuery -
func (r *PermissionServer) LookupQuery(ctx context.Context, request *v1.LookupQueryRequest) (*v1.LookupQueryResponse, error) {
	ctx, span := tracer.Start(ctx, "permissions.lookupQuery")
	defer span.End()

	v := request.Validate()
	if v != nil {
		return nil, v
	}

	var err errors.Error
	var response commands.LookupQueryResponse
	response, err = r.permissionService.LookupQuery(ctx, request.GetEntityType(), request.GetSubject(), request.GetAction(), request.GetSchemaVersion())
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		r.l.Error(fmt.Sprintf(err.Error()))
		switch err.Kind() {
		case errors.Database:
			return nil, status.Error(database.GetKindToGRPCStatus(err.SubKind()), err.Error())
		case errors.Validation:
			return nil, status.Error(codes.InvalidArgument, err.Error())
		case errors.Service:
			return nil, status.Error(codes.Internal, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &v1.LookupQueryResponse{
		Query: response.Query,
		Args:  response.Args,
	}, nil
}
