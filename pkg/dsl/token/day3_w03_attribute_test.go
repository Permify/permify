package token

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Attribute Keyword Token Specs", func() {
	Context("Token initialization for schema attribute keywords", func() {
		It("should correctly identify ATTRIBUTE keyword token", func() {
			tok := New(ATTRIBUTE, "attribute", PositionInfo{LinePosition: 3, ColumnPosition: 2})
			Expect(tok.Type).To(Equal(ATTRIBUTE))
			Expect(tok.Literal).To(Equal("attribute"))
			Expect(tok.PositionInfo.LinePosition).To(Equal(3))
		})
	})
})
