package servers

import (
	"context"

	"google.golang.org/grpc/codes"
	health "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

// HealthServer - Structure for Health Server
type HealthServer struct {
	health.UnimplementedHealthServer
}

// NewHealthServer - Creates new HealthServer Server
func NewHealthServer() *HealthServer {
	return &HealthServer{}
}

// Check - Return health check status response
func (s *HealthServer) Check(_ context.Context, _ *health.HealthCheckRequest) (*health.HealthCheckResponse, error) {
	return &health.HealthCheckResponse{Status: health.HealthCheckResponse_SERVING}, nil
}

// Watch - TO:DO
func (s *HealthServer) Watch(_ *health.HealthCheckRequest, _ health.Health_WatchServer) error {
	// Example of how to register both methods but only implement the Check method.
	return status.Error(codes.Unimplemented, "unimplemented")
}
