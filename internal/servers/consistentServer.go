package servers

import (
	"errors"
	"github.com/Permify/permify/pkg/cache"
	"google.golang.org/grpc/status"

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
		Value: response.(string),
	}, nil
}

// Set - Set created Key
func (r *ConsistentServer) Set(ctx context.Context, request *v1.ConsistentSetRequest) (*v1.ConsistentSetResponse, error) {
	ctx, span := tracer.Start(ctx, "consistent.set")
	defer span.End()

	r.cacheService.Set(request.GetKey(), 100, 1)

	return &v1.ConsistentSetResponse{
		Value: "OK",
	}, nil
}
