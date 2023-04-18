package servers

import (
	"errors"
	"fmt"
	"github.com/Permify/permify/pkg/cache"
	"github.com/Permify/permify/pkg/tuple"
	"github.com/cespare/xxhash/v2"
	"google.golang.org/grpc/status"

	base "github.com/Permify/permify/pkg/pb/base/v1"
	otelCodes "go.opentelemetry.io/otel/codes"
	"golang.org/x/net/context"

	"github.com/Permify/permify/pkg/logger"
	v1 "github.com/Permify/permify/pkg/pb/base/v1"
)

// ConsistentServer - Structure for Consistent Server
type ConsistentServer struct {
	v1.UnimplementedConsistentServer

	cacheService cache.Cache
	logger       logger.Interface
}

// NewConsistentServer - Creates new Consistent Server
func NewConsistentServer(s cache.Cache, l logger.Interface) *ConsistentServer {
	return &ConsistentServer{
		cacheService: s,
		logger:       l,
	}
}

// Get - Get created Key
func (r *ConsistentServer) Get(ctx context.Context, request *v1.ConsistentGetRequest) (*v1.ConsistentGetResponse, error) {
	ctx, span := tracer.Start(ctx, "consistent.get")
	defer span.End()

	var err error
	response, ok := r.cacheService.Get(request.GetKey())
	if !ok {
		err = errors.New("key not found" + request.GetKey())
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		r.logger.Error(err.Error())
		return nil, status.Error(GetStatus(err), err.Error())
	}

	return &v1.ConsistentGetResponse{
		PermissionCheckResponse: response.(*base.PermissionCheckResponse),
	}, nil
}

// Set - Set created Key
func (r *ConsistentServer) Set(ctx context.Context, request *v1.ConsistentSetRequest) (*v1.ConsistentSetResponse, error) {
	ctx, span := tracer.Start(ctx, "consistent.set")
	defer span.End()

	checkKey := fmt.Sprintf("check_%s_%s:%s:%s@%s", request.PermissionCheckRequest.GetTenantId(), request.PermissionCheckRequest.GetMetadata().GetSchemaVersion(), request.PermissionCheckRequest.GetMetadata().GetSnapToken(), tuple.EntityAndRelationToString(&base.EntityAndRelation{
		Entity:   request.PermissionCheckRequest.GetEntity(),
		Relation: request.PermissionCheckRequest.GetPermission(),
	}), tuple.SubjectToString(request.PermissionCheckRequest.GetSubject()))

	h := xxhash.New()

	// Write the checkKey string to the hash object
	size, err := h.Write([]byte(checkKey))
	if err != nil {
		// If there's an error, return false
		return nil, nil
	}

	r.cacheService.Set(request.GetKey(), request.PermissionCheckRequest, int64(size))

	return &v1.ConsistentSetResponse{
		Value: "OK",
	}, nil
}
