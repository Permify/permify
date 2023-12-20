package utils_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Permify/permify/internal/storage/postgres/utils"
)

var _ = Describe("Pagination", func() {
	Context("TestContinuousToken", func() {
		It("Case 1", func() {
			tokenValue := "test_token"
			token := utils.NewContinuousToken(tokenValue)

			// Test Encode
			encodedToken := token.Encode()
			Expect(encodedToken.String()).ShouldNot(Equal(""))

			// Test Decode
			decodedToken, err := encodedToken.Decode()
			Expect(err).ShouldNot(HaveOccurred())
			Expect(tokenValue).Should(Equal(decodedToken.(utils.ContinuousToken).Value))
		})
	})
})
