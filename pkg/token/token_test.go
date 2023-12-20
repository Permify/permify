package token

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// TestToken -
func TestToken(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "token-suite")
}

var _ = Describe("token", func() {
	Context("Encode", func() {
		It("Case 1: Success", func() {
			tests := []struct {
				target   SnapToken
				expected string
			}{
				{NewNoopToken(), "noop"},
			}

			for _, tt := range tests {
				Expect(tt.target.Encode().String()).Should(Equal(tt.expected))
			}
		})
	})

	Context("Decode", func() {
		It("Case 1: Success", func() {
			tests := []struct {
				target   EncodedSnapToken
				expected SnapToken
			}{
				{NoopEncodedToken{Value: "noop"}, NewNoopToken()},
			}

			for _, tt := range tests {
				t, err := tt.target.Decode()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(t).Should(Equal(tt.expected))
			}
		})

		It("Case 2: Success", func() {
			tests := []struct {
				target1  SnapToken
				target2  SnapToken
				expected bool
			}{
				{NewNoopToken(), NewNoopToken(), true},
			}

			for _, tt := range tests {
				Expect(tt.target1.Eg(tt.target2)).Should(BeTrue())
				Expect(tt.target1.Gt(tt.target2)).Should(BeFalse())
				Expect(tt.target1.Lt(tt.target2)).Should(BeFalse())
			}
		})
	})
})
