package balancer

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"google.golang.org/grpc/balancer"
)

var _ = Describe("ConsistentHashPicker", func() {
	var (
		testSubConns = map[string]balancer.SubConn{
			"addr1": nil, // Normally, you would use mock SubConn objects. For simplicity, we use nil here.
			"addr2": nil,
		}
		picker *ConsistentHashPicker
	)

	BeforeEach(func() {
		picker = NewConsistentHashPicker(testSubConns)
	})

	Describe("Initialization", func() {
		It("should initialize with provided subConns", func() {
			Expect(picker.subConns).To(Equal(testSubConns))
			Expect(picker.hashRing).ShouldNot(BeNil())
		})
	})

	Describe("Pick", func() {
		var pickInfo balancer.PickInfo

		Context("with custom key in context", func() {
			BeforeEach(func() {
				pickInfo = balancer.PickInfo{
					Ctx: context.WithValue(context.Background(), Key, "customKey"),
				}
			})

			It("should return an ErrNoSubConnAvailable error", func() {
				_, err := picker.Pick(pickInfo)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("no SubConn is available"))
			})
		})

		Context("without custom key in context", func() {
			BeforeEach(func() {
				pickInfo = balancer.PickInfo{
					FullMethodName: "testMethod",
					Ctx:            context.Background(),
				}
			})

			It("should return an ErrNoSubConnAvailable error", func() {
				_, err := picker.Pick(pickInfo)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("no SubConn is available"))
			})
		})

		Context("with unavailable SubConn", func() {
			BeforeEach(func() {
				picker.subConns = make(map[string]balancer.SubConn) // Empty the subConns map
				pickInfo = balancer.PickInfo{
					Ctx: context.WithValue(context.Background(), Key, "unavailableKey"),
				}
			})

			It("should return an ErrNoSubConnAvailable error", func() {
				_, err := picker.Pick(pickInfo)
				Expect(err).To(Equal(balancer.ErrNoSubConnAvailable))
			})
		})
	})
})
