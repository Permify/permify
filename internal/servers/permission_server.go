package servers

import (
	"context"
	"log/slog"

	otelCodes "go.opentelemetry.io/otel/codes"
	"google.golang.org/grpc/status"

	"github.com/Permify/permify/internal"
	"github.com/Permify/permify/internal/invoke"
	v1 "github.com/Permify/permify/pkg/pb/base/v1"
)

// PermissionServer - Structure for Permission Server
type PermissionServer struct {
	v1.UnimplementedPermissionServer

	invoker invoke.Invoker
}

// NewPermissionServer - Creates new Permission Server
func NewPermissionServer(i invoke.Invoker) *PermissionServer {
	return &PermissionServer{
		invoker: i,
	}
}

// Check - Performs Authorization Check
func (r *PermissionServer) Check(ctx context.Context, request *v1.PermissionCheckRequest) (*v1.PermissionCheckResponse, error) {
	ctx, span := internal.Tracer.Start(ctx, "permissions.check")
	defer span.End()

	v := request.Validate()
	if v != nil {
		return nil, status.Error(GetStatus(v), v.Error()) // Return validation error
	}

	response, err := r.invoker.Check(ctx, request)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		slog.ErrorContext(ctx, err.Error())
		return nil, status.Error(GetStatus(err), err.Error())
	}

	return response, nil
}

// BulkCheck - Performs multiple authorization checks in a single request
func (r *PermissionServer) BulkCheck(ctx context.Context, request *v1.PermissionBulkCheckRequest) (*v1.PermissionBulkCheckResponse, error) {
	ctx, span := internal.Tracer.Start(ctx, "permissions.bulk-check")
	defer span.End()

	// Validate tenant_id
	if request.GetTenantId() == "" {
		err := status.Error(GetStatus(nil), "tenant_id is required")
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		return nil, err
	}

	// Validate number of requests
	if len(request.GetItems()) == 0 {
		err := status.Error(GetStatus(nil), "at least one item is required")
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		return nil, err
	}

	if len(request.GetItems()) > 100 {
		err := status.Error(GetStatus(nil), "maximum 100 items allowed")
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		return nil, err
	}

	// Process each check request
	results := make([]*v1.PermissionCheckResponse, len(request.GetItems()))
	for i, checkRequestItem := range request.GetItems() {

		// Validate individual request
		v := checkRequestItem.Validate()
		if v != nil {
			// Return error response for this check
			results[i] = &v1.PermissionCheckResponse{
				Can: v1.CheckResult_CHECK_RESULT_DENIED,
				Metadata: &v1.PermissionCheckResponseMetadata{
					CheckCount: 0,
				},
			}
			continue
		}

		// Perform the check using existing Check function
		checkRequest := &v1.PermissionCheckRequest{
			TenantId:      request.GetTenantId(),
			Subject:       checkRequestItem.GetSubject(),
			Entity:        checkRequestItem.GetEntity(),
			Permission:    checkRequestItem.GetPermission(),
			Metadata: 	   request.GetMetadata(),
			Context:       request.GetContext(),
			Arguments:     request.GetArguments(),
		}
		response, err := r.invoker.Check(ctx, checkRequest)
		if err != nil {
			// Log error but don't fail the entire bulk operation
			slog.ErrorContext(ctx, "check failed in bulk operation", "error", err.Error(), "index", i)
			results[i] = &v1.PermissionCheckResponse{
				Can: v1.CheckResult_CHECK_RESULT_DENIED,
				Metadata: &v1.PermissionCheckResponseMetadata{
					CheckCount: 0,
				},
			}
			continue
		}

		results[i] = response
	}

	return &v1.PermissionBulkCheckResponse{
		Results: results,
	}, nil
}

// Expand - Get schema actions in a tree structure
func (r *PermissionServer) Expand(ctx context.Context, request *v1.PermissionExpandRequest) (*v1.PermissionExpandResponse, error) {
	ctx, span := internal.Tracer.Start(ctx, "permissions.expand")
	defer span.End()

	v := request.Validate()
	if v != nil {
		return nil, status.Error(GetStatus(v), v.Error()) // Return validation error
	}

	response, err := r.invoker.Expand(ctx, request)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		slog.ErrorContext(ctx, err.Error())
		return nil, status.Error(GetStatus(err), err.Error())
	}

	return response, nil
}

// LookupEntity -
func (r *PermissionServer) LookupEntity(ctx context.Context, request *v1.PermissionLookupEntityRequest) (*v1.PermissionLookupEntityResponse, error) {
	ctx, span := internal.Tracer.Start(ctx, "permissions.lookup-entity")
	defer span.End()

	v := request.Validate()
	if v != nil {
		return nil, status.Error(GetStatus(v), v.Error()) // Return validation error
	}

	response, err := r.invoker.LookupEntity(ctx, request)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		slog.ErrorContext(ctx, err.Error())
		return nil, status.Error(GetStatus(err), err.Error())
	}

	return response, nil
}

// LookupEntityStream -
func (r *PermissionServer) LookupEntityStream(request *v1.PermissionLookupEntityRequest, server v1.Permission_LookupEntityStreamServer) error {
	ctx, span := internal.Tracer.Start(server.Context(), "permissions.lookup-entity-stream")
	defer span.End()

	v := request.Validate()
	if v != nil {
		return v
	}

	err := r.invoker.LookupEntityStream(ctx, request, server)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		slog.ErrorContext(ctx, err.Error())
		return status.Error(GetStatus(err), err.Error())
	}

	return nil
}

// LookupSubject -
func (r *PermissionServer) LookupSubject(ctx context.Context, request *v1.PermissionLookupSubjectRequest) (*v1.PermissionLookupSubjectResponse, error) {
	ctx, span := internal.Tracer.Start(ctx, "permissions.lookup-subject")
	defer span.End()

	v := request.Validate()
	if v != nil {
		return nil, status.Error(GetStatus(v), v.Error()) // Return validation error
	}

	response, err := r.invoker.LookupSubject(ctx, request)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		slog.ErrorContext(ctx, err.Error())
		return nil, status.Error(GetStatus(err), err.Error())
	}

	return response, nil
}

// SubjectPermission -
func (r *PermissionServer) SubjectPermission(ctx context.Context, request *v1.PermissionSubjectPermissionRequest) (*v1.PermissionSubjectPermissionResponse, error) {
	ctx, span := internal.Tracer.Start(ctx, "permissions.subject-permission")
	defer span.End()

	v := request.Validate()
	if v != nil {
		return nil, status.Error(GetStatus(v), v.Error()) // Return validation error
	}

	response, err := r.invoker.SubjectPermission(ctx, request)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		slog.ErrorContext(ctx, err.Error())
		return nil, status.Error(GetStatus(err), err.Error())
	}

	return response, nil
}
