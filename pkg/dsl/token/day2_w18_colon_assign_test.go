package token

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Colon and Assign Token Specs", func() {
	Context("Token initialization for assignment literals", func() {
		It("should correctly identify COLON token", func() {
			tok := New(COLON, ":", PositionInfo{LinePosition: 5, ColumnPosition: 8})
			Expect(tok.Type).To(Equal(COLON))
			Expect(tok.Literal).To(Equal(":"))
			Expect(tok.PositionInfo.LinePosition).To(Equal(5))
		})

		It("should correctly identify ASSIGN token", func() {
			tok := New(ASSIGN, "=", PositionInfo{LinePosition: 5, ColumnPosition: 10})
			Expect(tok.Type).To(Equal(ASSIGN))
			Expect(tok.Literal).To(Equal("="))
			Expect(tok.PositionInfo.ColumnPosition).To(Equal(10))
		})
	})
})
