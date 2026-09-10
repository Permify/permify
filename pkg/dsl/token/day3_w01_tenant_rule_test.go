package token

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Tenant and Rule Token Specs", func() {
	Context("Token initialization for rule keywords", func() {
		It("should correctly identify RULE keyword token", func() {
			tok := New(RULE, "rule", PositionInfo{LinePosition: 1, ColumnPosition: 1})
			Expect(tok.Type).To(Equal(RULE))
			Expect(tok.Literal).To(Equal("rule"))
			Expect(tok.PositionInfo.LinePosition).To(Equal(1))
		})
	})
})
