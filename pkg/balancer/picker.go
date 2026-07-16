package balancer

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"

	"google.golang.org/grpc/balancer"

	"github.com/Permify/permify/pkg/consistent"
)

// subConnPicker is a trivial gRPC picker: reads a pre-computed SubConn from context.
// Installed once in gRPC's balancer state; never needs to be replaced.
type subConnPicker struct{}

func (p *subConnPicker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	sc, ok := info.Ctx.Value(SubConnKey).(balancer.SubConn)
	if !ok || sc == nil {
		return balancer.PickResult{}, fmt.Errorf("no SubConn in context")
	}
	return balancer.PickResult{SubConn: sc}, nil
}

// NodePicker resolves a routing key to a target SubConn.
type NodePicker interface {
	Pick(key []byte) (balancer.SubConn, error)
}

// picker implements NodePicker using consistent hashing.
type picker struct {
	consistent *consistent.Consistent
	width      int
}

// Generate a cryptographically secure random index function with resilient error handling
var randomIndex = func(max int) int {
	// Ensure max > 0 to avoid issues
	if max <= 0 {
		log.Println("randomIndex: max value is less than or equal to 0, returning 0 as fallback")
		return 0
	}

	// Use crypto/rand to generate a random index
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		// Log the error and return a deterministic fallback value (e.g., 0)
		log.Printf("randomIndex: failed to generate a secure random number, returning 0 as fallback: %v", err)
		return 0
	}

	return int(n.Int64())
}

// Pick computes the target SubConn for a routing key using consistent hashing.
func (p *picker) Pick(key []byte) (balancer.SubConn, error) {
	members, err := p.consistent.ClosestN(key, p.width)
	if err != nil {
		return nil, fmt.Errorf("failed to get closest members: %w", err)
	}
	if len(members) == 0 {
		return nil, fmt.Errorf("no available members")
	}

	// Randomly pick one member if width > 1
	index := 0
	if p.width > 1 {
		index = randomIndex(p.width)
	}

	// Assert the member type
	chosen, ok := members[index].(ConsistentMember)
	if !ok {
		return nil, fmt.Errorf("invalid member type: expected ConsistentMember")
	}

	// Return the chosen connection
	return chosen.SubConn, nil
}
