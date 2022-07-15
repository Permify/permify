package lexer

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Permify/permify/pkg/dsl/token"
)

// TestLexer -
func TestLexer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "lexer-suite")
}

var _ = Describe("lexer", func() {

	Context("NextToken", func() {

		It("Success", func() {

			str := "entity user {} `table:\"users\",identifier:\"id\"`\n entity organization {\n relation admin @user `rel:\"custom\"`\n relation member @user `rel:\"many-to-many\", table:\"org_members\", cols:\"org_id,user_id\"`\n    action create_repository = admin or member\n    action delete = admin \n} `table:\"organizations\", identifier:\"id\"`\n"

			tests := []struct {
				expectedType    token.Type
				expectedLiteral string
			}{
				{token.ENTITY, "entity"},
				{token.IDENT, "user"},
				{token.LBRACE, "{"},
				{token.RBRACE, "}"},
				{token.OPTION, "table:\"users\",identifier:\"id\""},
				{token.NEWLINE, "\n"},
				{token.ENTITY, "entity"},
				{token.IDENT, "organization"},
				{token.LBRACE, "{"},
				{token.NEWLINE, "\n"},
				{token.RELATION, "relation"},
				{token.IDENT, "admin"},
				{token.SIGN, "@"},
				{token.IDENT, "user"},
				{token.OPTION, "rel:\"custom\""},
				{token.NEWLINE, "\n"},
				{token.RELATION, "relation"},
				{token.IDENT, "member"},
				{token.SIGN, "@"},
				{token.IDENT, "user"},
				{token.OPTION, "rel:\"many-to-many\", table:\"org_members\", cols:\"org_id,user_id\""},
				{token.NEWLINE, "\n"},
				{token.ACTION, "action"},
				{token.IDENT, "create_repository"},
				{token.ASSIGN, "="},
				{token.IDENT, "admin"},
				{token.OR, "or"},
				{token.IDENT, "member"},
				{token.ACTION, "action"},
				{token.IDENT, "delete"},
				{token.ASSIGN, "="},
				{token.IDENT, "admin"},
				{token.NEWLINE, "\n"},
				{token.RBRACE, "}"},
				{token.OPTION, "table:\"organizations\", identifier:\"id\""},
				{token.NEWLINE, "\n"},
				{token.EOF, ""},
			}

			l := NewLexer(str)

			for _, tt := range tests {
				lexeme := l.NextToken()
				Expect(lexeme.Type).Should(Equal(tt.expectedType))
				Expect(lexeme.Literal).Should(Equal(tt.expectedLiteral))
			}
		})
	})
})
