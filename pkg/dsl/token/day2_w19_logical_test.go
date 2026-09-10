package token

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Logical Token Operator Specs", func() {
	Context("Token initialization for boolean operators", func() {
		It("should correctly identify BANG (!) token", func() {
			tok := New(BANG, "!", PositionInfo{LinePosition: 8, ColumnPosition: 2})
			Expect(tok.Type).To(Equal(BANG))
			Expect(tok.Literal).To(Equal("!"))
			Expect(tok.PositionInfo.LinePosition).To(Equal(8))
		})

		It("should correctly identify AND (and) keyword token", func() {
			tok := New(AND, "and", PositionInfo{LinePosition: 8, ColumnPosition: 10})
			Expect(tok.Type).To(Equal(AND))
			Expect(tok.Literal).To(Equal("and"))
			Expect(tok.PositionInfo.ColumnPosition).To(Equal(10))
		})
	})
})
