package balancer

import (
	"context"
	"log/slog"
	"sync"

	"github.com/serialx/hashring"
	"google.golang.org/grpc/balancer"
)

// ConsistentHashPicker is a custom gRPC picker that uses consistent hashing
// to determine which backend server should handle the request.
type ConsistentHashPicker struct {
	subConns map[string]balancer.SubConn // Map of server addresses to their respective SubConns
	mu       sync.RWMutex                // Mutex to protect concurrent access to subConns
	hashRing *hashring.HashRing          // Hash ring used for consistent hashing
}

// PickResult represents the result of a pick operation.
// It contains the context and the selected SubConn for the request.
type PickResult struct {
	Ctx context.Context
	SC  balancer.SubConn
}

// NewConsistentHashPicker initializes and returns a new ConsistentHashPicker.
// It creates a hash ring from the provided set of backend server addresses.
func NewConsistentHashPicker(subConns map[string]balancer.SubConn) *ConsistentHashPicker {
	addrs := make([]string, 0, len(subConns))

	// Extract addresses from the subConns map
	for addr := range subConns {
		addrs = append(addrs, addr)
	}

	slog.Debug("consistent hash picker built", slog.Any("addresses", addrs))

	return &ConsistentHashPicker{
		subConns: subConns,
		hashRing: hashring.New(addrs),
	}
}

// Pick selects an appropriate backend server (SubConn) for the incoming request.
// If a custom key is provided in the context, it will be used for consistent hashing;
// otherwise, the full method name of the request will be used.
func (p *ConsistentHashPicker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	var ret balancer.PickResult
	key, ok := info.Ctx.Value(Key).(string)
	if !ok {
		key = info.FullMethodName
	}
	slog.Debug("pick for", slog.String("key", key))

	// Safely read from the subConns map using the read lock
	p.mu.RLock()
	if targetAddr, ok := p.hashRing.GetNode(key); ok {
		ret.SubConn = p.subConns[targetAddr]
	}
	p.mu.RUnlock()

	// If no valid SubConn was found, return an error
	if ret.SubConn == nil {
		return ret, balancer.ErrNoSubConnAvailable
	}
	return ret, nil
}
