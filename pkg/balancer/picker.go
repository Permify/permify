package balancer

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"

	"google.golang.org/grpc/balancer"

	"github.com/Permify/permify/pkg/consistent"
)

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

func (p *picker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	// Safely extract the key from the context
	keyValue := info.Ctx.Value(Key)
	if keyValue == nil {
		return balancer.PickResult{}, fmt.Errorf("context key missing")
	}
	key, ok := keyValue.([]byte)
	if !ok {
		return balancer.PickResult{}, fmt.Errorf("context key is not of type []byte")
	}

	// Retrieve the closest N members
	members, err := p.consistent.ClosestN(key, p.width)
	if err != nil {
		return balancer.PickResult{}, fmt.Errorf("failed to get closest members: %v", err)
	}
	if len(members) == 0 {
		return balancer.PickResult{}, fmt.Errorf("no available members")
	}

	// Randomly pick one member if width > 1
	index := 0
	if p.width > 1 {
		index = randomIndex(p.width)
	}

	// Assert the member type
	chosen, ok := members[index].(ConsistentMember)
	if !ok {
		return balancer.PickResult{}, fmt.Errorf("invalid member type: expected subConnMember")
	}

	// Return the chosen connection
	return balancer.PickResult{SubConn: chosen.SubConn}, nil
}
