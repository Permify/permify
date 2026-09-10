package token

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Relational Token Specs", func() {
	Context("Token initialization for greater than operators", func() {
		It("should correctly identify GT (>) token", func() {
			tok := New(GT, ">", PositionInfo{LinePosition: 18, ColumnPosition: 4})
			Expect(tok.Type).To(Equal(GT))
			Expect(tok.Literal).To(Equal(">"))
			Expect(tok.PositionInfo.LinePosition).To(Equal(18))
		})

		It("should correctly identify GTE (>=) token", func() {
			tok := New(GTE, ">=", PositionInfo{LinePosition: 18, ColumnPosition: 10})
			Expect(tok.Type).To(Equal(GTE))
			Expect(tok.Literal).To(Equal(">="))
			Expect(tok.PositionInfo.ColumnPosition).To(Equal(10))
		})
	})
})
