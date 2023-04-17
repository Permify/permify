package servers

import (
	"errors"
	hash "github.com/Permify/permify/pkg/consistent"
	"google.golang.org/grpc/status"

	otelCodes "go.opentelemetry.io/otel/codes"
	"golang.org/x/net/context"

	"github.com/Permify/permify/pkg/logger"
	v1 "github.com/Permify/permify/pkg/pb/base/v1"
)

// ConsistentServer - Structure for Consistent Server
type ConsistentServer struct {
	v1.UnimplementedConsistentServer

	consistentService *hash.ConsistentHash
	logger            logger.Interface
}

// NewConsistentServer - Creates new Consistent Server
func NewConsistentServer(s *hash.ConsistentHash, l logger.Interface) *ConsistentServer {
	return &ConsistentServer{
		consistentService: s,
		logger:            l,
	}
}

// Get - Get created Key
func (r *ConsistentServer) Get(ctx context.Context, request *v1.ConsistentGetRequest) (*v1.ConsistentGetResponse, error) {
	ctx, span := tracer.Start(ctx, "consistent.get")
	defer span.End()

	var err error
	response, ok := r.consistentService.Get(request.GetKey())
	if !ok {
		err = errors.New("key not found" + request.GetKey())
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		r.logger.Error(err.Error())
		return nil, status.Error(GetStatus(err), err.Error())
	}

	return &v1.ConsistentGetResponse{
		Value: response,
	}, nil
}

// Set - Set created Key
func (r *ConsistentServer) Set(ctx context.Context, request *v1.ConsistentSetRequest) (*v1.ConsistentSetResponse, error) {
	ctx, span := tracer.Start(ctx, "consistent.set")
	defer span.End()

	r.consistentService.AddWithWeight(request.GetKey(), 100)

	return &v1.ConsistentSetResponse{
		Value: "OK",
	}, nil
}
