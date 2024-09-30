package servers

import (
	"context"

	v1 "github.com/Permify/permify/pkg/pb/base/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// HealthServer - Structure for Health Server
type HealthServer struct {
	v1.UnimplementedHealthServer
}

// NewHealthServer - Creates new HealthServer Server
func NewHealthServer() *HealthServer {
	return &HealthServer{}
}

// Check - Return health check status response
func (s *HealthServer) Check(_ context.Context, _ *v1.HealthCheckRequest) (*v1.HealthCheckResponse, error) {
	return &v1.HealthCheckResponse{Status: v1.HealthCheckResponse_SERVING}, nil
}

// Watch - TO:DO
func (s *HealthServer) Watch(_ *v1.HealthCheckRequest, _ v1.Health_WatchServer) error {
	// Example of how to register both methods but only implement the Check method.
	return status.Error(codes.Unimplemented, "unimplemented")
}

// AuthFuncOverride is called instead of authn.
func (s *HealthServer) AuthFuncOverride(ctx context.Context, fullMethodName string) (context.Context, error) {
	return ctx, nil
}
