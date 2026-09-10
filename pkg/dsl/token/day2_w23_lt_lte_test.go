package token

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Less Than Relational Token Specs", func() {
	Context("Token initialization for less than operators", func() {
		It("should correctly identify LT (<) token", func() {
			tok := New(LT, "<", PositionInfo{LinePosition: 20, ColumnPosition: 5})
			Expect(tok.Type).To(Equal(LT))
			Expect(tok.Literal).To(Equal("<"))
			Expect(tok.PositionInfo.LinePosition).To(Equal(20))
		})

		It("should correctly identify LTE (<=) token", func() {
			tok := New(LTE, "<=", PositionInfo{LinePosition: 20, ColumnPosition: 12})
			Expect(tok.Type).To(Equal(LTE))
			Expect(tok.Literal).To(Equal("<="))
			Expect(tok.PositionInfo.ColumnPosition).To(Equal(12))
		})
	})
})
