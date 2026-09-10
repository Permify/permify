package token

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Schema Keyword Token Specs", func() {
	Context("Token initialization for entity keywords", func() {
		It("should correctly identify ENTITY token", func() {
			tok := New(ENTITY, "entity", PositionInfo{LinePosition: 1, ColumnPosition: 1})
			Expect(tok.Type).To(Equal(ENTITY))
			Expect(tok.Literal).To(Equal("entity"))
			Expect(tok.PositionInfo.LinePosition).To(Equal(1))
		})
	})
})
