package balancer

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc/balancer"
)

var _ = Describe("consistentHashBalancerBuilder", func() {
	var builder balancer.Builder
	var mockClientConn balancer.ClientConn
	var buildOpts balancer.BuildOptions

	BeforeEach(func() {
		builder = NewConsistentHashBalancerBuilder()
		// You'll want to mock or create a real balancer.ClientConn and balancer.BuildOptions here.
		// For now, we'll keep them nil for simplicity.
		mockClientConn = nil
		buildOpts = balancer.BuildOptions{}
	})

	Describe("Name", func() {
		It("should return the expected balancer name", func() {
			name := builder.Name()
			Expect(name).To(Equal(Policy))
		})
	})

	Describe("Build", func() {
		It("should return a consistentHashBalancer", func() {
			b := builder.Build(mockClientConn, buildOpts)
			Expect(b).To(Not(BeNil()))

			// Further assertions can be made depending on the properties and
			// behaviors of the consistentHashBalancer. For example:
			chb, ok := b.(*consistentHashBalancer)
			Expect(ok).To(BeTrue())
			Expect(chb.clientConn).To(BeNil())
			Expect(chb.addressInfoMap).To(Not(BeNil()))
			Expect(chb.subConnectionMap).To(Not(BeNil()))
			Expect(chb.activePickResults).To(Not(BeNil()))
		})
	})
})
