package token

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Punctuation Token Specs", func() {
	Context("Token initialization for punctuation literals", func() {
		It("should correctly identify COMMA token", func() {
			tok := New(COMMA, ",", PositionInfo{LinePosition: 3, ColumnPosition: 12})
			Expect(tok.Type).To(Equal(COMMA))
			Expect(tok.Literal).To(Equal(","))
			Expect(tok.PositionInfo.LinePosition).To(Equal(3))
		})

		It("should correctly identify SEMICOLON token", func() {
			tok := New(SEMICOLON, ";", PositionInfo{LinePosition: 4, ColumnPosition: 1})
			Expect(tok.Type).To(Equal(SEMICOLON))
			Expect(tok.Literal).To(Equal(";"))
			Expect(tok.PositionInfo.ColumnPosition).To(Equal(1))
		})
	})
})
