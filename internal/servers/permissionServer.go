package servers

import (
	"log/slog"

	otelCodes "go.opentelemetry.io/otel/codes"
	"golang.org/x/net/context"
	"google.golang.org/grpc/status"

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
	ctx, span := tracer.Start(ctx, "permissions.check")
	defer span.End()

	v := request.Validate()
	if v != nil {
		return nil, status.Error(GetStatus(v), v.Error())
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

// Bulk Check - Performs Bulk Authorization Checks
func (r *PermissionServer) BulkCheck(ctx context.Context, request *v1.BulkPermissionCheckRequest) (*v1.BulkPermissionCheckResponse, error) {
	ctx, span := tracer.Start(ctx, "permissions.bulk-check")
	defer span.End()

	// Validate the incoming request
	v := request.Validate()
	if v != nil {
		return nil, status.Error(GetStatus(v), v.Error())
	}

	checks := request.Checks
	results := make([]*v1.SinglePermissionCheckResponse, len(checks))

	for i, check := range checks {
		// Create individual PermissionCheckRequest for each item
		singleRequest := &v1.PermissionCheckRequest{
			TenantId:   request.TenantId,
			Metadata:   check.Metadata,
			Entity:     check.Entity,
			Permission: check.Permission,
			Subject:    check.Subject,
			Context:    check.Context,
			Arguments:  check.Arguments,
		}

		// Perform the permission check
		response, err := r.invoker.Check(ctx, singleRequest)
		if err != nil {
			// Log and record the error for each failed check
			span.RecordError(err)
			span.SetStatus(otelCodes.Error, err.Error())
			slog.ErrorContext(ctx, err.Error())

			// Add the failure response with index
			results[i] = &v1.SinglePermissionCheckResponse{
				Can:      v1.CheckResult_CHECK_RESULT_DENIED,
				Metadata: &v1.PermissionCheckResponseMetadata{},
				Index:    int32(i),
			}
		} else {
			// Successful check response, attach the index
			results[i] = &v1.SinglePermissionCheckResponse{
				Can:      response.Can,
				Metadata: response.Metadata,
				Index:    int32(i),
			}
		}
	}

	// Return the bulk response
	return &v1.BulkPermissionCheckResponse{
		Results: results,
	}, nil
}

// Expand - Get schema actions in a tree structure
func (r *PermissionServer) Expand(ctx context.Context, request *v1.PermissionExpandRequest) (*v1.PermissionExpandResponse, error) {
	ctx, span := tracer.Start(ctx, "permissions.expand")
	defer span.End()

	v := request.Validate()
	if v != nil {
		return nil, status.Error(GetStatus(v), v.Error())
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
	ctx, span := tracer.Start(ctx, "permissions.lookup-entity")
	defer span.End()

	v := request.Validate()
	if v != nil {
		return nil, status.Error(GetStatus(v), v.Error())
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
	ctx, span := tracer.Start(server.Context(), "permissions.lookup-entity-stream")
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
	ctx, span := tracer.Start(ctx, "permissions.lookup-subject")
	defer span.End()

	v := request.Validate()
	if v != nil {
		return nil, status.Error(GetStatus(v), v.Error())
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
	ctx, span := tracer.Start(ctx, "permissions.subject-permission")
	defer span.End()

	v := request.Validate()
	if v != nil {
		return nil, status.Error(GetStatus(v), v.Error())
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
