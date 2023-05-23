package servers

import (
	otelCodes "go.opentelemetry.io/otel/codes"
	"golang.org/x/net/context"
	"google.golang.org/grpc/status"

	"github.com/Permify/permify/internal/invoke"
	"github.com/Permify/permify/pkg/logger"
	v1 "github.com/Permify/permify/pkg/pb/base/v1"
)

// PermissionServer - Structure for Permission Server
type PermissionServer struct {
	v1.UnimplementedPermissionServer

	invoker invoke.Invoker
	logger  logger.Interface
}

// NewPermissionServer - Creates new Permission Server
func NewPermissionServer(i invoke.Invoker, l logger.Interface) *PermissionServer {
	return &PermissionServer{
		invoker: i,
		logger:  l,
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

	response, err := r.invoker.Check(ctx, request)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		r.logger.Error(err.Error())
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

	response, err := r.invoker.Expand(ctx, request)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		r.logger.Error(err.Error())
		return nil, status.Error(GetStatus(err), err.Error())
	}

	return response, nil
}

// LookupEntity -
func (r *PermissionServer) LookupEntity(ctx context.Context, request *v1.PermissionLookupEntityRequest) (*v1.PermissionLookupEntityResponse, error) {
	ctx, span := tracer.Start(ctx, "permissions.lookup-entity")
	defer span.End()

	v := request.Validate()
	if v != nil {
		return nil, v
	}

	response, err := r.invoker.LookupEntity(ctx, request)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		r.logger.Error(err.Error())
		return nil, status.Error(GetStatus(err), err.Error())
	}

	return response, nil
}

// LookupEntityStream -
func (r *PermissionServer) LookupEntityStream(request *v1.PermissionLookupEntityRequest, server v1.Permission_LookupEntityStreamServer) error {
	ctx, span := tracer.Start(context.Background(), "permissions.lookup-entity-stream")
	defer span.End()

	v := request.Validate()
	if v != nil {
		return v
	}

	err := r.invoker.LookupEntityStream(ctx, request, server)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		r.logger.Error(err.Error())
		return status.Error(GetStatus(err), err.Error())
	}

	return nil
}

// LookupSubject -
func (r *PermissionServer) LookupSubject(ctx context.Context, request *v1.PermissionLookupSubjectRequest) (*v1.PermissionLookupSubjectResponse, error) {
	ctx, span := tracer.Start(ctx, "permissions.lookup-subject")
	defer span.End()

	v := request.Validate()
	if v != nil {
		return nil, v
	}

	response, err := r.invoker.LookupSubject(ctx, request)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		r.logger.Error(err.Error())
		return nil, status.Error(GetStatus(err), err.Error())
	}

	return response, nil
}
