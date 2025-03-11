package balancer

import (
	"errors"
	"testing"

	estats "google.golang.org/grpc/experimental/stats"

	"github.com/cespare/xxhash/v2"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/resolver"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Permify/permify/pkg/consistent"
)

// This is the entry point for the test suite for the "consistent" package.
// It registers a failure handler and runs the specifications (specs) for this package.
func TestBalancer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "balancer-suite")
}

type mockClientConn struct {
	updateState balancer.State
	newSubConn  balancer.SubConn
}

func (m *mockClientConn) MetricsRecorder() estats.MetricsRecorder {
	return nil
}

func (m *mockClientConn) UpdateState(state balancer.State) {
	m.updateState = state
}

func (m *mockClientConn) NewSubConn(addrs []resolver.Address, opts balancer.NewSubConnOptions) (balancer.SubConn, error) {
	if m.newSubConn != nil {
		return m.newSubConn, nil
	}
	return nil, errors.New("mock SubConn creation failed")
}

func (m *mockClientConn) RemoveSubConn(sc balancer.SubConn)                             {}
func (m *mockClientConn) UpdateAddresses(sc balancer.SubConn, addrs []resolver.Address) {}
func (m *mockClientConn) ResolveNow(options resolver.ResolveNowOptions)                 {}
func (m *mockClientConn) Target() string                                                { return "test-target" }

type mockSubConnWrapper struct {
	balancer.SubConn
}

func (m *mockSubConnWrapper) Connect()                           {}
func (m *mockSubConnWrapper) Shutdown()                          {}
func (m *mockSubConnWrapper) UpdateAddresses([]resolver.Address) {}
func (m *mockSubConnWrapper) GetOrBuildProducer(balancer.ProducerBuilder) (balancer.Producer, func()) {
	return nil, func() {}
}
func (m *mockSubConnWrapper) RegisterHealthListener(func(balancer.SubConnState)) {}

var _ = Describe("Balancer", func() {
	var (
		b          *Balancer
		clientConn *mockClientConn
		subConn    *mockSubConnWrapper
	)

	BeforeEach(func() {
		subConn = &mockSubConnWrapper{}
		clientConn = &mockClientConn{newSubConn: subConn}
		b = &Balancer{
			clientConn:            clientConn,
			subConnStates:         make(map[balancer.SubConn]connectivity.State),
			connectivityEvaluator: &balancer.ConnectivityStateEvaluator{},
			addressSubConns:       resolver.NewAddressMap(),
			config: &Config{
				PartitionCount:    100,
				ReplicationFactor: 3,
				Load:              1.25,
				PickerWidth:       2,
			},
			consistent: consistent.New(consistent.Config{
				PartitionCount:    100,
				ReplicationFactor: 3,
				Load:              1.25,
				Hasher:            xxhash.Sum64,
			}),
		}
	})

	It("should update client connection state", func() {
		state := balancer.ClientConnState{
			ResolverState: resolver.State{
				Addresses: []resolver.Address{
					{Addr: "127.0.0.1:50051"},
				},
			},
			BalancerConfig: &Config{
				PartitionCount:    100,
				ReplicationFactor: 3,
				Load:              1.25,
				PickerWidth:       2,
			},
		}
		err := b.UpdateClientConnState(state)
		Expect(err).ToNot(HaveOccurred())
	})

	It("should handle resolver errors correctly", func() {
		b.ResolverError(errors.New("resolver failure"))
		Expect(b.state).To(Equal(connectivity.TransientFailure))
	})

	It("should update SubConn state", func() {
		b.subConnStates[subConn] = connectivity.Idle
		state := balancer.SubConnState{ConnectivityState: connectivity.Connecting}
		b.UpdateSubConnState(subConn, state)
		Expect(b.subConnStates[subConn]).To(Equal(connectivity.Connecting))
	})
})
