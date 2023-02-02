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
	Context("LookupKeywords", func() {
		It("Case 1", func() {
			tests := []struct {
				target   string
				expected Type
			}{
				{target: "entity", expected: ENTITY},
				{target: "relation", expected: RELATION},
				{target: "action", expected: ACTION},
				{target: "or", expected: OR},
				{target: "not", expected: NOT},
				{target: "no t", expected: IDENT},
				{target: " no t", expected: IDENT},
				{target: "test", expected: IDENT},
				{target: "and", expected: AND},
			}

			for _, tt := range tests {
				Expect(LookupKeywords(tt.target)).Should(Equal(tt.expected))
			}
		})
	})
})
