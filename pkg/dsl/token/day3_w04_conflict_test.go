package token

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Conflict and Override Identifier Token Specs", func() {
	Context("Token initialization for rule conflict handling", func() {
		It("should correctly identify OVERRIDE identifier literal", func() {
			tok := New(IDENT, "override", PositionInfo{LinePosition: 4, ColumnPosition: 1})
			Expect(tok.Type).To(Equal(IDENT))
			Expect(tok.Literal).To(Equal("override"))
			Expect(tok.PositionInfo.LinePosition).To(Equal(4))
		})
	})
})
