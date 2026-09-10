package token

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Action Keyword Token Specs", func() {
	Context("Token initialization for action keywords", func() {
		It("should correctly identify ACTION token", func() {
			tok := New(ACTION, "action", PositionInfo{LinePosition: 3, ColumnPosition: 4})
			Expect(tok.Type).To(Equal(ACTION))
			Expect(tok.Literal).To(Equal("action"))
			Expect(tok.PositionInfo.LinePosition).To(Equal(3))
		})
	})
})
