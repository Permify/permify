package token

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Or and In Keyword Token Specs", func() {
	Context("Token initialization for set and disjunction keywords", func() {
		It("should correctly identify OR keyword token", func() {
			tok := New(OR, "or", PositionInfo{LinePosition: 12, ColumnPosition: 4})
			Expect(tok.Type).To(Equal(OR))
			Expect(tok.Literal).To(Equal("or"))
			Expect(tok.PositionInfo.LinePosition).To(Equal(12))
		})

		It("should correctly identify IN keyword token", func() {
			tok := New(IN, "in", PositionInfo{LinePosition: 12, ColumnPosition: 15})
			Expect(tok.Type).To(Equal(IN))
			Expect(tok.Literal).To(Equal("in"))
			Expect(tok.PositionInfo.ColumnPosition).To(Equal(15))
		})
	})
})
