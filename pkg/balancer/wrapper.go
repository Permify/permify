package balancer

import (
	"google.golang.org/grpc/balancer"
	estats "google.golang.org/grpc/experimental/stats"
	"google.golang.org/grpc/resolver"
)

// ClientConnWrapper is an interface that wraps the gRPC ClientConn interface.
type ClientConnWrapper interface {
	// NewSubConn creates a new SubConn with the specified addresses and options.
	NewSubConn([]resolver.Address, balancer.NewSubConnOptions) (balancer.SubConn, error)
	// RemoveSubConn removes the specified SubConn.
	RemoveSubConn(balancer.SubConn)
	// UpdateAddresses updates the addresses for the specified SubConn.
	UpdateAddresses(balancer.SubConn, []resolver.Address)
	// UpdateState updates the overall connectivity state of the balancer.
	UpdateState(balancer.State)
	// ResolveNow triggers an immediate resolution of the target.
	ResolveNow(resolver.ResolveNowOptions)
	// Target returns the target URI of the connection.
	Target() string
	// MetricsRecorder returns the metrics recorder for the connection.
	MetricsRecorder() estats.MetricsRecorder
}
