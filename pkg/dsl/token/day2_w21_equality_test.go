package token

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Equality Operator Token Specs", func() {
	Context("Token initialization for equality literals", func() {
		It("should correctly identify EQUAL (==) token", func() {
			tok := New(EQUAL, "==", PositionInfo{LinePosition: 15, ColumnPosition: 6})
			Expect(tok.Type).To(Equal(EQUAL))
			Expect(tok.Literal).To(Equal("=="))
			Expect(tok.PositionInfo.LinePosition).To(Equal(15))
		})

		It("should correctly identify NOTEQUAL (!=) token", func() {
			tok := New(NOTEQUAL, "!=", PositionInfo{LinePosition: 15, ColumnPosition: 12})
			Expect(tok.Type).To(Equal(NOTEQUAL))
			Expect(tok.Literal).To(Equal("!="))
			Expect(tok.PositionInfo.ColumnPosition).To(Equal(12))
		})
	})
})
