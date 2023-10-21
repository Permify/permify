package balancer

import (
	"sync"

	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/resolver"
)

// NewConsistentHashBalancerBuilder returns a consistentHashBalancerBuilder.
func NewConsistentHashBalancerBuilder() balancer.Builder {
	return &consistentHashBalancerBuilder{}
}

// consistentHashBalancerBuilder is an empty struct with functions Build and Name, implemented from balancer.Builder
type consistentHashBalancerBuilder struct{}

// Build creates a consistentHashBalancer, and starts its scManager.
func (builder *consistentHashBalancerBuilder) Build(cc balancer.ClientConn, opts balancer.BuildOptions) balancer.Balancer {
	b := &consistentHashBalancer{
		clientConn:          cc,
		addressInfoMap:      make(map[string]resolver.Address),
		subConnectionMap:    make(map[string]balancer.SubConn),
		subConnInfoSyncMap:  sync.Map{},
		pickerResultChannel: make(chan PickResult),
		activePickResults:   NewQueue(),
		subConnPickCounts:   make(map[balancer.SubConn]*int32),
		subConnStatusMap:    make(map[balancer.SubConn]bool),
	}
	go b.manageSubConnections()
	return b
}

// Name returns the name of the consistentHashBalancer registering in grpc.
func (builder *consistentHashBalancerBuilder) Name() string {
	return Policy
}
