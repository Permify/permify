package token

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Permission Keyword Token Specs", func() {
	Context("Token initialization for permission keywords", func() {
		It("should correctly identify PERMISSION token", func() {
			tok := New(PERMISSION, "permission", PositionInfo{LinePosition: 4, ColumnPosition: 4})
			Expect(tok.Type).To(Equal(PERMISSION))
			Expect(tok.Literal).To(Equal("permission"))
			Expect(tok.PositionInfo.LinePosition).To(Equal(4))
		})
	})
})
