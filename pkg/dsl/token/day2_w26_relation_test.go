package token

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Relation Keyword Token Specs", func() {
	Context("Token initialization for relation keywords", func() {
		It("should correctly identify RELATION token", func() {
			tok := New(RELATION, "relation", PositionInfo{LinePosition: 2, ColumnPosition: 4})
			Expect(tok.Type).To(Equal(RELATION))
			Expect(tok.Literal).To(Equal("relation"))
			Expect(tok.PositionInfo.LinePosition).To(Equal(2))
		})
	})
})
