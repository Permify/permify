package servers

import (
	"errors"
	"github.com/Permify/permify/internal/keys"
	"google.golang.org/grpc/status"

	otelCodes "go.opentelemetry.io/otel/codes"
	"golang.org/x/net/context"

	"github.com/Permify/permify/pkg/logger"
	v1 "github.com/Permify/permify/pkg/pb/base/v1"
)

// ConsistentServer - Structure for Consistent Server
type ConsistentServer struct {
	v1.UnimplementedConsistentServer

	cacheService keys.EngineKeyManager
	logger       logger.Interface
}

// NewConsistentServer - Creates new Consistent Server
func NewConsistentServer(s keys.EngineKeyManager, l logger.Interface) *ConsistentServer {
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
	response, ok := r.cacheService.GetCheckKey(request.PermissionCheckRequest)
	if !ok {
		err = errors.New("key not found" + request.PermissionCheckRequest.TenantId)
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		r.logger.Error("ConsistentServer.Get: " + err.Error())
		return nil, status.Error(GetStatus(err), err.Error())
	}

	return &v1.ConsistentGetResponse{
		PermissionCheckResponse: response,
		PermissionCheckRequest:  request.GetPermissionCheckRequest(),
	}, nil
}

// Set - Set created Key
func (r *ConsistentServer) Set(ctx context.Context, request *v1.ConsistentSetRequest) (*v1.ConsistentSetResponse, error) {
	ctx, span := tracer.Start(ctx, "consistent.set")
	defer span.End()

	ok := r.cacheService.SetCheckKey(request.PermissionCheckRequest, request.PermissionCheckResponse)
	if !ok {
		err := errors.New("key not set" + request.PermissionCheckRequest.TenantId)
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		r.logger.Error("ConsistentServer.Set: " + err.Error())
		return nil, status.Error(GetStatus(err), err.Error())
	}

	return &v1.ConsistentSetResponse{
		Value: "OK",
	}, nil
}
