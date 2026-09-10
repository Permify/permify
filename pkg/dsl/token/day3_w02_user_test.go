package token

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("User Token Keyword Specs", func() {
	Context("Token initialization for user identifiers", func() {
		It("should correctly identify USER token literal", func() {
			tok := New(IDENT, "user", PositionInfo{LinePosition: 2, ColumnPosition: 1})
			Expect(tok.Type).To(Equal(IDENT))
			Expect(tok.Literal).To(Equal("user"))
			Expect(tok.PositionInfo.LinePosition).To(Equal(2))
		})
	})
})
