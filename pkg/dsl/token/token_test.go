package token

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// TestToken -
func TestToken(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "token-suite")
}

var _ = Describe("token", func() {
	Context("LookupKeywords", func() {
		It("Case 1", func() {
			tests := []struct {
				target   string
				expected Type
			}{
				{target: "entity", expected: ENTITY},
				{target: "relation", expected: RELATION},
				{target: "action", expected: PERMISSION},
				{target: "or", expected: OR},
				{target: "not", expected: NOT},
				{target: "no t", expected: IDENT},
				{target: " no t", expected: IDENT},
				{target: "test", expected: IDENT},
				{target: "and", expected: AND},
				{target: "permission", expected: PERMISSION},
				{target: "attribute", expected: ATTRIBUTE},
				{target: "rule", expected: RULE},
			}

			for _, tt := range tests {
				Expect(LookupKeywords(tt.target)).Should(Equal(tt.expected))
			}
		})
	})

	Context("LookupKeywords", func() {
		It("Case 1", func() {
			tests := []struct {
				target   Type
				expected bool
			}{
				{target: MULTI_LINE_COMMENT, expected: true},
				{target: PERMISSION, expected: false},
				{target: OR, expected: false},
				{target: NEWLINE, expected: false},
				{target: SINGLE_LINE_COMMENT, expected: true},
				{target: ENTITY, expected: false},
				{target: SPACE, expected: true},
				{target: TAB, expected: true},
				{target: "test", expected: false},
				{target: RULE, expected: false},
			}

			for _, tt := range tests {
				Expect(IsIgnores(tt.target)).Should(Equal(tt.expected))
			}
		})
	})

	Context("String", func() {
		It("Case 1", func() {
			tests := []struct {
				target   Type
				expected string
			}{
				{target: MULTI_LINE_COMMENT, expected: "MULTI_LINE_COMMENT"},
				{target: SPACE, expected: "SPACE"},
				{target: TAB, expected: "TAB"},
				{target: RULE, expected: "RULE"},
			}

			for _, tt := range tests {
				Expect(tt.target.String()).Should(Equal(tt.expected))
			}
		})
	})

	Context("New", func() {
		It("Case 1", func() {
			tests := []struct {
				positionInfo PositionInfo
				typ          Type
				ch           byte
				result       Token
			}{
				{
					positionInfo: PositionInfo{
						LinePosition:   1,
						ColumnPosition: 2,
					}, typ: RP, ch: ')',
					result: Token{
						PositionInfo: PositionInfo{
							LinePosition:   1,
							ColumnPosition: 2,
						},
						Type:    RP,
						Literal: ")",
					},
				},
				{
					positionInfo: PositionInfo{
						LinePosition:   1,
						ColumnPosition: 5,
					}, typ: AMPERSAND, ch: '\'',
					result: Token{
						PositionInfo: PositionInfo{
							LinePosition:   1,
							ColumnPosition: 5,
						},
						Type:    AMPERSAND,
						Literal: "'",
					},
				},
			}

			for _, tt := range tests {
				Expect(New(tt.positionInfo, tt.typ, tt.ch)).Should(Equal(tt.result))
			}
		})
	})
})
