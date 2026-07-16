package servers

import (
	"context"
	"errors"
	"log/slog"
	"sync"

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

	response, err := r.invoker.Check(ctx, invoke.NewBatchCheckRequest(request))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		slog.ErrorContext(ctx, err.Error())
		return nil, status.Error(GetStatus(err), err.Error())
	}

	return response.ToPermissionCheckResponse(), nil
}

// BulkCheck - Performs multiple authorization checks in a single request
func (r *PermissionServer) BulkCheck(ctx context.Context, request *v1.PermissionBulkCheckRequest) (*v1.PermissionBulkCheckResponse, error) {
	// emptyResp is a default, empty response that we will return in case of an error or when the context is cancelled.
	emptyResp := &v1.PermissionBulkCheckResponse{
		Results: make([]*v1.PermissionCheckResponse, 0),
	}

	ctx, span := internal.Tracer.Start(ctx, "permissions.bulk-check")
	defer span.End()

	// Validate tenant_id
	if request.GetTenantId() == "" {
		err := status.Error(GetStatus(nil), "tenant_id is required")
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		return nil, err
	}

	checkItems := request.GetItems()

	// Validate number of requests
	if len(checkItems) == 0 {
		err := status.Error(GetStatus(nil), "at least one item is required")
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		return nil, err
	}

	if len(checkItems) > 100 {
		err := status.Error(GetStatus(nil), "maximum 100 items allowed")
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		return nil, err
	}

	// Group items by (entityType, permission, subject) for batch processing.
	// Items in the same group share a single BatchCheckRequest.
	type groupKey struct {
		entityType, permission, subjectType, subjectID, subjectRelation string
	}
	type groupEntry struct {
		entityIDs []string
		indices   []int
		subject   *v1.Subject
	}

	results := make([]*v1.PermissionCheckResponse, len(checkItems))
	groups := map[groupKey]*groupEntry{}

	for i, item := range checkItems {
		if v := item.Validate(); v != nil {
			results[i] = &v1.PermissionCheckResponse{
				Can:      v1.CheckResult_CHECK_RESULT_DENIED,
				Metadata: &v1.PermissionCheckResponseMetadata{},
			}
			continue
		}

		subj := item.GetSubject()
		key := groupKey{
			entityType:      item.GetEntity().GetType(),
			permission:      item.GetPermission(),
			subjectType:     subj.GetType(),
			subjectID:       subj.GetId(),
			subjectRelation: subj.GetRelation(),
		}
		g, ok := groups[key]
		if !ok {
			g = &groupEntry{subject: subj}
			groups[key] = g
		}
		g.entityIDs = append(g.entityIDs, item.GetEntity().GetId())
		g.indices = append(g.indices, i)
	}

	// Execute each group as a batch check, concurrently.
	var wg sync.WaitGroup
	for key, g := range groups {
		wg.Add(1)
		go func(key groupKey, g *groupEntry) {
			defer wg.Done()

			batchReq := &invoke.BatchCheckRequest{
				TenantID:   request.GetTenantId(),
				EntityType: key.entityType,
				EntityIDs:  g.entityIDs,
				Permission: key.permission,
				Subject:    g.subject,
				Metadata:   request.GetMetadata(),
				Context:    request.GetContext(),
				Arguments:  request.GetArguments(),
			}

			resp, err := r.invoker.Check(ctx, batchReq)
			if err != nil {
				slog.ErrorContext(ctx, "batch check failed in bulk operation", "error", err.Error())
				for _, idx := range g.indices {
					results[idx] = &v1.PermissionCheckResponse{
						Can:      v1.CheckResult_CHECK_RESULT_DENIED,
						Metadata: &v1.PermissionCheckResponseMetadata{CheckCount: 1},
					}
				}
				return
			}

			// Map batch results back to original indices.
			for j, entityID := range g.entityIDs {
				result := v1.CheckResult_CHECK_RESULT_DENIED
				if r, ok := resp.Results[entityID]; ok {
					result = r
				}
				results[g.indices[j]] = &v1.PermissionCheckResponse{
					Can:      result,
					Metadata: resp.Metadata,
				}
			}
		}(key, g)
	}
	wg.Wait()

	// Check for context cancellation.
	if ctx.Err() != nil {
		return emptyResp, errors.New(v1.ErrorCode_ERROR_CODE_CANCELLED.String())
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
