package lexer

import (
	`strconv`
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
		It("Case 1", func() {
			str :=
				`
entity user {}
entity organization {
	relation admin @user
	relation banned @user
	relation member @user

	action create_repository = admin and not banned 
	action delete = admin
}
`

			tests := []struct {
				expectedType    token.Type
				expectedLiteral string
			}{
				{token.NEWLINE, "\n"},
				{token.ENTITY, "entity"},
				{token.SPACE, " "},
				{token.IDENT, "user"},
				{token.SPACE, " "},
				{token.LBRACE, "{"},
				{token.RBRACE, "}"},
				{token.NEWLINE, "\n"},
				{token.ENTITY, "entity"},
				{token.SPACE, " "},
				// --
				{token.IDENT, "organization"},
				{token.SPACE, " "},
				{token.LBRACE, "{"},
				{token.NEWLINE, "\n"},
				{token.TAB, "\t"},
				{token.RELATION, "relation"},
				{token.SPACE, " "},
				{token.IDENT, "admin"},
				{token.SPACE, " "},
				{token.SIGN, "@"},
				// --
				{token.IDENT, "user"},
				{token.NEWLINE, "\n"},
				{token.TAB, "\t"},
				{token.RELATION, "relation"},
				{token.SPACE, " "},
				{token.IDENT, "banned"},
				{token.SPACE, " "},
				{token.SIGN, "@"},
				{token.IDENT, "user"},
				{token.NEWLINE, "\n"},
				// --
				{token.TAB, "\t"},
				{token.RELATION, "relation"},
				{token.SPACE, " "},
				{token.IDENT, "member"},
				{token.SPACE, " "},
				{token.SIGN, "@"},
				{token.IDENT, "user"},
				{token.NEWLINE, "\n"},
				{token.NEWLINE, "\n"},
				{token.TAB, "\t"},
				// --
				{token.ACTION, "action"},
				{token.SPACE, " "},
				{token.IDENT, "create_repository"},
				{token.SPACE, " "},
				{token.ASSIGN, "="},
				{token.SPACE, " "},
				{token.IDENT, "admin"},
				{token.SPACE, " "},
				{token.AND, "and"},
				{token.SPACE, " "},
				// --
				{token.NOT, "not"},
				{token.SPACE, " "},
				{token.IDENT, "banned"},
				{token.SPACE, " "},
				{token.NEWLINE, "\n"},
				{token.TAB, "\t"},
				{token.ACTION, "action"},
				{token.SPACE, " "},
				{token.IDENT, "delete"},
				{token.SPACE, " "},
				// --
				{token.ASSIGN, "="},
				{token.SPACE, " "},
				{token.IDENT, "admin"},
				{token.NEWLINE, "\n"},
				{token.RBRACE, "}"},
				{token.NEWLINE, "\n"},
				{token.EOF, ""},
			}

			l := NewLexer(str)

			for i, tt := range tests {
				lexeme := l.NextToken()
				index := strconv.Itoa(i) + ": "
				Expect(index + lexeme.Type.String()).Should(Equal(index + tt.expectedType.String()))
				Expect(index + lexeme.Literal).Should(Equal(index + tt.expectedLiteral))
			}
		})

		It("Case 2", func() {
			str :=
				`
entity user {}

entity organization {
	relation admin @user
    relation member @user
	action create_repository = admin or member
	action delete = admin
}`

			tests := []struct {
				expectedType    token.Type
				expectedLiteral string
			}{
				{token.NEWLINE, "\n"},
				{token.ENTITY, "entity"},
				{token.SPACE, " "},
				{token.IDENT, "user"},
				{token.SPACE, " "},
				{token.LBRACE, "{"},
				{token.RBRACE, "}"},
				{token.NEWLINE, "\n"},
				{token.NEWLINE, "\n"},
				{token.ENTITY, "entity"},
				// --
				{token.SPACE, " "},
				{token.IDENT, "organization"},
				{token.SPACE, " "},
				{token.LBRACE, "{"},
				{token.NEWLINE, "\n"},
				{token.TAB, "\t"},
				{token.RELATION, "relation"},
				{token.SPACE, " "},
				{token.IDENT, "admin"},
				{token.SPACE, " "},
				// --
				{token.SIGN, "@"},
				{token.IDENT, "user"},
				{token.NEWLINE, "\n"},
				{token.SPACE, " "},
				{token.SPACE, " "},
				{token.SPACE, " "},
				{token.SPACE, " "},
				{token.RELATION, "relation"},
				{token.SPACE, " "},
				{token.IDENT, "member"},
				// --
				{token.SPACE, " "},
				{token.SIGN, "@"},
				{token.IDENT, "user"},
				{token.NEWLINE, "\n"},
				{token.TAB, "\t"},
				{token.ACTION, "action"},
				{token.SPACE, " "},
				{token.IDENT, "create_repository"},
				{token.SPACE, " "},
				{token.ASSIGN, "="},
				// --
				{token.SPACE, " "},
				{token.IDENT, "admin"},
				{token.SPACE, " "},
				{token.OR, "or"},
				{token.SPACE, " "},
				{token.IDENT, "member"},
				{token.NEWLINE, "\n"},
				{token.TAB, "\t"},
				{token.ACTION, "action"},
				{token.SPACE, " "},
				// --
				{token.IDENT, "delete"},
				{token.SPACE, " "},
				{token.ASSIGN, "="},
				{token.SPACE, " "},
				{token.IDENT, "admin"},
				{token.NEWLINE, "\n"},
				{token.RBRACE, "}"},
				{token.EOF, ""},
			}

			l := NewLexer(str)

			for i, tt := range tests {
				lexeme := l.NextToken()
				index := strconv.Itoa(i) + ": "
				Expect(index + lexeme.Type.String()).Should(Equal(index + tt.expectedType.String()))
				Expect(index + lexeme.Literal).Should(Equal(index + tt.expectedLiteral))
			}
		})

		It("Case 3", func() {
			str := `
entity user {}

entity organization {
	//comment line
	relation admin @user
	relation member @user
	action create_repository = admin or member
	action delete = admin
}

entity repository {
	/*
	comment line 1
	comment line 2
	*/
	relation parent @organization
	relation member @user @organization#member
	action update = parent.delete or (member and parent.admin)
}`

			tests := []struct {
				expectedType    token.Type
				expectedLiteral string
			}{
				{token.NEWLINE, "\n"},
				{token.ENTITY, "entity"},
				{token.SPACE, " "},
				{token.IDENT, "user"},
				{token.SPACE, " "},
				{token.LBRACE, "{"},
				{token.RBRACE, "}"},
				{token.NEWLINE, "\n"},
				{token.NEWLINE, "\n"},
				{token.ENTITY, "entity"},
				// --
				{token.SPACE, " "},
				{token.IDENT, "organization"},
				{token.SPACE, " "},
				{token.LBRACE, "{"},
				{token.NEWLINE, "\n"},
				{token.TAB, "\t"},
				{token.SINGLE_LINE_COMMENT, "comment line"},
				{token.NEWLINE, "\n"},
				{token.TAB, "\t"},
				{token.RELATION, "relation"},
				// --
				{token.SPACE, " "},
				{token.IDENT, "admin"},
				{token.SPACE, " "},
				{token.SIGN, "@"},
				{token.IDENT, "user"},
				{token.NEWLINE, "\n"},
				{token.TAB, "\t"},
				{token.RELATION, "relation"},
				{token.SPACE, " "},
				{token.IDENT, "member"},
				// --
				{token.SPACE, " "},
				{token.SIGN, "@"},
				{token.IDENT, "user"},
				{token.NEWLINE, "\n"},
				{token.TAB, "\t"},
				{token.ACTION, "action"},
				{token.SPACE, " "},
				{token.IDENT, "create_repository"},
				{token.SPACE, " "},
				{token.ASSIGN, "="},
				// --
				{token.SPACE, " "},
				{token.IDENT, "admin"},
				{token.SPACE, " "},
				{token.OR, "or"},
				{token.SPACE, " "},
				{token.IDENT, "member"},
				{token.NEWLINE, "\n"},
				{token.TAB, "\t"},
				{token.ACTION, "action"},
				{token.SPACE, " "},
				// --
				{token.IDENT, "delete"},
				{token.SPACE, " "},
				{token.ASSIGN, "="},
				{token.SPACE, " "},
				{token.IDENT, "admin"},
				{token.NEWLINE, "\n"},
				{token.RBRACE, "}"},
				{token.NEWLINE, "\n"},
				{token.NEWLINE, "\n"},
				{token.ENTITY, "entity"},
				// --
				{token.SPACE, " "},
				{token.IDENT, "repository"},
				{token.SPACE, " "},
				{token.LBRACE, "{"},
				{token.NEWLINE, "\n"},
				{token.TAB, "\t"},
				{token.MULTI_LINE_COMMENT, "\n\tcomment line 1\n\tcomment line 2\n\t"},
				{token.NEWLINE, "\n"},
				{token.TAB, "\t"},
				{token.RELATION, "relation"},
				// --
				{token.SPACE, " "},
				{token.IDENT, "parent"},
				{token.SPACE, " "},
				{token.SIGN, "@"},
				{token.IDENT, "organization"},
				{token.NEWLINE, "\n"},
				{token.TAB, "\t"},
				{token.RELATION, "relation"},
				{token.SPACE, " "},
				{token.IDENT, "member"},
				// --
				{token.SPACE, " "},
				{token.SIGN, "@"},
				{token.IDENT, "user"},
				{token.SPACE, " "},
				{token.SIGN, "@"},
				{token.IDENT, "organization#member"},
				{token.NEWLINE, "\n"},
				{token.TAB, "\t"},
				{token.ACTION, "action"},
				{token.SPACE, " "},
				// --
				{token.IDENT, "update"},
				{token.SPACE, " "},
				{token.ASSIGN, "="},
				{token.SPACE, " "},
				{token.IDENT, "parent.delete"},
				{token.SPACE, " "},
				{token.OR, "or"},
				{token.SPACE, " "},
				{token.LPAREN, "("},
				{token.IDENT, "member"},
				// --
				{token.SPACE, " "},
				{token.AND, "and"},
				{token.SPACE, " "},
				{token.IDENT, "parent.admin"},
				{token.RPAREN, ")"},
				{token.NEWLINE, "\n"},
				{token.RBRACE, "}"},
				{token.EOF, ""},
			}

			l := NewLexer(str)

			for i, tt := range tests {
				lexeme := l.NextToken()
				index := strconv.Itoa(i) + ": "
				Expect(index + lexeme.Type.String()).Should(Equal(index + tt.expectedType.String()))
				Expect(index + lexeme.Literal).Should(Equal(index + tt.expectedLiteral))
			}
		})

		It("Case 4", func() {
			str := `
entity user {}
/*
entity organization {
	relation admin @user
	relation member @user
	action create_repository = admin or member
	action delete = admin
}
*/
`

			tests := []struct {
				expectedType    token.Type
				expectedLiteral string
			}{
				{token.NEWLINE, "\n"},
				{token.ENTITY, "entity"},
				{token.SPACE, " "},
				{token.IDENT, "user"},
				{token.SPACE, " "},
				{token.LBRACE, "{"},
				{token.RBRACE, "}"},
				{token.NEWLINE, "\n"},
				{token.MULTI_LINE_COMMENT, "\nentity organization {\n\trelation admin @user\n\trelation member @user\n\taction create_repository = admin or member\n\taction delete = admin\n}\n"},
				{token.NEWLINE, "\n"},
				// --
				{token.EOF, ""},
			}

			l := NewLexer(str)

			for i, tt := range tests {
				lexeme := l.NextToken()
				index := strconv.Itoa(i) + ": "
				Expect(index + lexeme.Type.String()).Should(Equal(index + tt.expectedType.String()))
				Expect(index + lexeme.Literal).Should(Equal(index + tt.expectedLiteral))
			}
		})

		It("Case 5", func() {
			str := `
entity user {}
/*
entity organization {
	relation admin @user
	relation member @user
	action create_repository = admin or member
	action delete = admin
}
`

			tests := []struct {
				expectedType    token.Type
				expectedLiteral string
			}{
				{token.NEWLINE, "\n"},
				{token.ENTITY, "entity"},
				{token.SPACE, " "},
				{token.IDENT, "user"},
				{token.SPACE, " "},
				{token.LBRACE, "{"},
				{token.RBRACE, "}"},
				{token.NEWLINE, "\n"},
				{token.MULTI_LINE_COMMENT, "\nentity organization {\n\trelation admin @user\n\trelation member @user\n\taction create_repository = admin or member\n\taction delete = admin\n}\n"},
				{token.EOF, ""},
			}

			l := NewLexer(str)

			for i, tt := range tests {
				lexeme := l.NextToken()
				index := strconv.Itoa(i) + ": "
				Expect(index + lexeme.Type.String()).Should(Equal(index + tt.expectedType.String()))
				Expect(index + lexeme.Literal).Should(Equal(index + tt.expectedLiteral))
			}
		})
	})
})
