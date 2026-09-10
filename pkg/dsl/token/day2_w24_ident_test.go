package token

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Identifier Token Specs", func() {
	Context("Token initialization for custom identifier literals", func() {
		It("should correctly identify custom IDENT token", func() {
			tok := New(IDENT, "custom_permission_rule", PositionInfo{LinePosition: 25, ColumnPosition: 1})
			Expect(tok.Type).To(Equal(IDENT))
			Expect(tok.Literal).To(Equal("custom_permission_rule"))
			Expect(tok.PositionInfo.LinePosition).To(Equal(25))
		})
	})
})
