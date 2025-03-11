package balancer

import (
	"errors"
	"fmt"
	"log/slog"

	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/resolver"

	"github.com/Permify/permify/pkg/consistent"
)

type Balancer struct {
	// Current overall connectivity state of the balancer.
	state connectivity.State

	// The ClientConn to communicate with the gRPC client.
	clientConn ClientConnWrapper

	// Current picker used to select SubConns for requests.
	picker balancer.Picker

	// Evaluates connectivity state transitions for SubConns.
	connectivityEvaluator *balancer.ConnectivityStateEvaluator

	// Map of resolver addresses to SubConns.
	addressSubConns *resolver.AddressMap

	// Tracks the connectivity state of each SubConn.
	subConnStates map[balancer.SubConn]connectivity.State

	// Configuration for consistent hashing and replication.
	config *Config

	// Consistent hashing mechanism to distribute requests.
	consistent *consistent.Consistent

	// Hasher used by the consistent hashing mechanism.
	hasher consistent.Hasher

	// Stores the last resolver error encountered.
	lastResolverError error

	// Stores the last connection error encountered.
	lastConnectionError error
}

func (b *Balancer) ResolverError(err error) {
	b.lastResolverError = err
	if b.addressSubConns.Len() == 0 {
		b.state = connectivity.TransientFailure
		b.picker = base.NewErrPicker(errors.Join(b.lastConnectionError, b.lastResolverError))
	}

	if b.state != connectivity.TransientFailure {
		return
	}

	// Update the balancer state and picker.
	b.clientConn.UpdateState(balancer.State{
		ConnectivityState: b.state,
		Picker:            b.picker,
	})
}

func (b *Balancer) UpdateClientConnState(s balancer.ClientConnState) error {
	// Log the new ClientConn state.
	slog.Info("Received new ClientConn state",
		slog.Any("state", s),
	)

	// Reset any existing resolver error.
	b.lastResolverError = nil

	// Handle changes to the balancer configuration.
	if s.BalancerConfig != nil {
		svcConfig := s.BalancerConfig.(*Config)
		if b.config == nil || svcConfig.ReplicationFactor != b.config.ReplicationFactor {
			slog.Info("Updating consistent hashing configuration",
				slog.Int("partition_count", svcConfig.PartitionCount),
				slog.Int("replication_factor", svcConfig.ReplicationFactor),
				slog.Float64("load", svcConfig.Load),
				slog.Int("picker_width", svcConfig.PickerWidth),
			)
			b.consistent = consistent.New(consistent.Config{
				PartitionCount:    svcConfig.PartitionCount,
				ReplicationFactor: svcConfig.ReplicationFactor,
				Load:              svcConfig.Load,
				PickerWidth:       svcConfig.PickerWidth,
				Hasher:            b.hasher,
			})
			b.config = svcConfig
		}
	}

	// Check if the consistent hashing configuration exists.
	if b.consistent == nil {
		slog.Error("No consistent hashing configuration found")
		b.picker = base.NewErrPicker(errors.Join(b.lastConnectionError, b.lastResolverError))
		b.clientConn.UpdateState(balancer.State{ConnectivityState: b.state, Picker: b.picker})
		return fmt.Errorf("no consistent hashing configuration found")
	}

	// Maintain a set of addresses provided by the resolver.
	addrsSet := resolver.NewAddressMap()
	for _, addr := range s.ResolverState.Addresses {
		addrsSet.Set(addr, nil)

		// Add new SubConns for addresses that are not already tracked.
		if _, ok := b.addressSubConns.Get(addr); !ok {
			sc, err := b.clientConn.NewSubConn([]resolver.Address{addr}, balancer.NewSubConnOptions{HealthCheckEnabled: false})
			if err != nil {
				slog.Warn("Failed to create new SubConn",
					slog.String("address", addr.Addr),
					slog.String("server_name", addr.ServerName),
					slog.String("error", err.Error()),
				)
				continue
			}

			b.addressSubConns.Set(addr, sc)
			b.subConnStates[sc] = connectivity.Idle
			b.connectivityEvaluator.RecordTransition(connectivity.Shutdown, connectivity.Idle)
			sc.Connect()

			b.consistent.Add(ConsistentMember{
				SubConn: sc,
				name:    fmt.Sprintf("%s|%s", addr.ServerName, addr.Addr),
			})
		}
	}

	// Remove SubConns that are no longer part of the resolved addresses.
	for _, addr := range b.addressSubConns.Keys() {
		sci, _ := b.addressSubConns.Get(addr)
		sc := sci.(balancer.SubConn)
		if _, ok := addrsSet.Get(addr); !ok {
			slog.Info("Removing SubConn",
				slog.String("address", addr.Addr),
				slog.String("server_name", addr.ServerName),
			)
			b.clientConn.RemoveSubConn(sc)
			b.addressSubConns.Delete(addr)
			b.consistent.Remove(ConsistentMember{
				SubConn: sc,
				name:    fmt.Sprintf("%s|%s", addr.ServerName, addr.Addr),
			}.String())
		}
	}

	// Log the current members in the consistent hashing ring.
	slog.Info("Current consistent members",
		slog.Int("member_count", len(b.consistent.Members())),
	)
	for _, m := range b.consistent.Members() {
		slog.Info("Consistent member", slog.String("member", m.String()))
	}

	// Handle the case where the resolver produces zero addresses.
	if len(s.ResolverState.Addresses) == 0 {
		err := errors.New("resolver produced zero addresses")
		b.ResolverError(err)
		slog.Error("Resolver produced zero addresses")
		return balancer.ErrBadResolverState
	}

	// Update the picker based on the current balancer state.
	if b.state == connectivity.TransientFailure {
		slog.Warn("Transient failure detected, using error picker")
		b.picker = base.NewErrPicker(errors.Join(b.lastConnectionError, b.lastResolverError))
	} else {
		width := b.config.PickerWidth
		if width < 1 {
			width = 1
		}
		slog.Info("Creating new picker",
			slog.Int("width", width),
		)
		b.picker = &picker{
			consistent: b.consistent,
			width:      width,
		}
	}

	// Update the ClientConn state with the new picker.
	slog.Info("Updating ClientConn state",
		slog.String("connectivity_state", b.state.String()),
	)
	b.clientConn.UpdateState(balancer.State{ConnectivityState: b.state, Picker: b.picker})

	return nil
}

func (b *Balancer) UpdateSubConnState(sc balancer.SubConn, state balancer.SubConnState) {
	s := state.ConnectivityState
	slog.Info("Received SubConn state change",
		slog.String("connectivity_state", s.String()),
		slog.String("sub_conn", fmt.Sprintf("%p", sc)),
	)

	oldS, ok := b.subConnStates[sc]
	if !ok {
		slog.Warn("State change for unknown SubConn",
			slog.String("connectivity_state", s.String()),
			slog.String("sub_conn", fmt.Sprintf("%p", sc)),
		)
		return
	}

	if oldS == connectivity.TransientFailure && (s == connectivity.Connecting || s == connectivity.Idle) {
		if s == connectivity.Idle {
			slog.Info("Transitioning SubConn to connecting state",
				slog.String("sub_conn", fmt.Sprintf("%p", sc)),
			)
			sc.Connect()
		}
		return
	}

	b.subConnStates[sc] = s
	switch s {
	case connectivity.Idle:
		slog.Info("SubConn is idle, initiating connection",
			slog.String("sub_conn", fmt.Sprintf("%p", sc)),
		)
		sc.Connect()
	case connectivity.Shutdown:
		slog.Info("Removing shutdown SubConn",
			slog.String("sub_conn", fmt.Sprintf("%p", sc)),
		)
		delete(b.subConnStates, sc)
	case connectivity.TransientFailure:
		slog.Warn("SubConn in transient failure",
			slog.String("sub_conn", fmt.Sprintf("%p", sc)),
			slog.String("error", state.ConnectionError.Error()),
		)
		b.lastConnectionError = state.ConnectionError
	}

	b.state = b.connectivityEvaluator.RecordTransition(oldS, s)
	slog.Info("Updating ClientConn state",
		slog.String("connectivity_state", b.state.String()),
	)
	b.clientConn.UpdateState(balancer.State{ConnectivityState: b.state, Picker: b.picker})
}

func (b *Balancer) Close() {}
