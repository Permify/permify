package balancer

import (
	"context"
	"sync"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/resolver"
)

var _ = Describe("consistentHashBalancer", func() {
	var b *consistentHashBalancer

	BeforeEach(func() {
		b = &consistentHashBalancer{
			clientConn:          nil, // Mock this if needed
			connectionState:     connectivity.Ready,
			addressInfoMap:      make(map[string]resolver.Address),
			subConnectionMap:    make(map[string]balancer.SubConn),
			subConnInfoSyncMap:  sync.Map{},
			currentPicker:       nil, // Mock this if needed
			lastResolverError:   nil,
			lastConnectionError: nil,
			pickerResultChannel: make(chan PickResult, 10),
			activePickResults:   NewQueue(),
			subConnPickCounts:   make(map[balancer.SubConn]*int32),
			subConnStatusMap:    make(map[balancer.SubConn]bool),
			balancerLock:        sync.Mutex{},
		}
	})

	Describe("manageSubConnections", func() {
		It("should start without panicking", func() {
			go b.manageSubConnections()
		})
	})

	Describe("enqueueAndTrackPickResult", func() {
		It("should enqueue and track pick results", func() {
			mockSC := balancer.SubConn(nil)
			pr := PickResult{
				Ctx: context.Background(),
				SC:  mockSC,
			}

			b.enqueueAndTrackPickResult(pr)

			Expect(b.activePickResults.Len()).Should(Equal(2))
			cnt, ok := b.subConnPickCounts[mockSC]
			Expect(ok).Should(BeTrue())
			Expect(*cnt).Should(Equal(int32(2)))
		})
	})

	Describe("handleDequeuedPickResult", func() {
		It("should re-enqueue pick result when context isn't done", func() {
			mockSC := balancer.SubConn(nil)
			pr := PickResult{
				Ctx: context.Background(),
				SC:  mockSC,
			}

			b.activePickResults.EnQueue(pr)
			b.handleDequeuedPickResult(pr)
			Expect(b.activePickResults.Len()).Should(Equal(2))
		})

		It("should decrease count and possibly reset SubConn when context is done", func() {
			mockSC := balancer.SubConn(nil)
			ctx, cancel := context.WithCancel(context.Background())
			pr := PickResult{
				Ctx: ctx,
				SC:  mockSC,
			}

			b.subConnPickCounts[mockSC] = new(int32)
			*b.subConnPickCounts[mockSC] = 1

			cancel()
			b.handleDequeuedPickResult(pr)

			_, ok := b.subConnPickCounts[mockSC]
			Expect(ok).Should(BeFalse())
		})
	})
})
