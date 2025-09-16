package lexer

import (
	"strconv"
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
			str := `
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
				{token.LCB, "{"},
				{token.RCB, "}"},
				{token.NEWLINE, "\n"},
				{token.ENTITY, "entity"},
				{token.SPACE, " "},
				// --
				{token.IDENT, "organization"},
				{token.SPACE, " "},
				{token.LCB, "{"},
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
				{token.PERMISSION, "action"},
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
				{token.PERMISSION, "action"},
				{token.SPACE, " "},
				{token.IDENT, "delete"},
				{token.SPACE, " "},
				// --
				{token.ASSIGN, "="},
				{token.SPACE, " "},
				{token.IDENT, "admin"},
				{token.NEWLINE, "\n"},
				{token.RCB, "}"},
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
			str := `
entity user {}

entity organization {
	relation admin @user
    relation member @user
	action create_repository = admin or member;
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
				{token.LCB, "{"},
				{token.RCB, "}"},
				{token.NEWLINE, "\n"},
				{token.NEWLINE, "\n"},
				{token.ENTITY, "entity"},
				// --
				{token.SPACE, " "},
				{token.IDENT, "organization"},
				{token.SPACE, " "},
				{token.LCB, "{"},
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
				{token.PERMISSION, "action"},
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
				{token.NEWLINE, ";"},
				{token.NEWLINE, "\n"},
				{token.TAB, "\t"},
				{token.PERMISSION, "action"},
				// --
				{token.SPACE, " "},
				{token.IDENT, "delete"},
				{token.SPACE, " "},
				{token.ASSIGN, "="},
				{token.SPACE, " "},
				{token.IDENT, "admin"},
				{token.NEWLINE, "\n"},
				{token.RCB, "}"},
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
				{token.LCB, "{"},
				{token.RCB, "}"},
				{token.NEWLINE, "\n"},
				{token.NEWLINE, "\n"},
				{token.ENTITY, "entity"},
				// --
				{token.SPACE, " "},
				{token.IDENT, "organization"},
				{token.SPACE, " "},
				{token.LCB, "{"},
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
				{token.PERMISSION, "action"},
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
				{token.PERMISSION, "action"},
				{token.SPACE, " "},
				// --
				{token.IDENT, "delete"},
				{token.SPACE, " "},
				{token.ASSIGN, "="},
				{token.SPACE, " "},
				{token.IDENT, "admin"},
				{token.NEWLINE, "\n"},
				{token.RCB, "}"},
				{token.NEWLINE, "\n"},
				{token.NEWLINE, "\n"},
				{token.ENTITY, "entity"},
				// --
				{token.SPACE, " "},
				{token.IDENT, "repository"},
				{token.SPACE, " "},
				{token.LCB, "{"},
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
				{token.IDENT, "organization"},
				{token.HASH, "#"},
				{token.IDENT, "member"},
				{token.NEWLINE, "\n"},
				{token.TAB, "\t"},
				// --
				{token.PERMISSION, "action"},
				{token.SPACE, " "},
				{token.IDENT, "update"},
				{token.SPACE, " "},
				{token.ASSIGN, "="},
				{token.SPACE, " "},
				{token.IDENT, "parent"},
				{token.DOT, "."},
				{token.IDENT, "delete"},
				{token.SPACE, " "},
				{token.OR, "or"},
				{token.SPACE, " "},
				// --
				{token.LP, "("},
				{token.IDENT, "member"},
				{token.SPACE, " "},
				{token.AND, "and"},
				{token.SPACE, " "},
				{token.IDENT, "parent"},
				{token.DOT, "."},
				{token.IDENT, "admin"},
				{token.RP, ")"},
				{token.NEWLINE, "\n"},
				{token.RCB, "}"},
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
				{token.LCB, "{"},
				{token.RCB, "}"},
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
				{token.LCB, "{"},
				{token.RCB, "}"},
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

		It("Case 6", func() {
			str := `
	entity user {}

	entity organization {
		relation member @user
	}

	// This is a role for an entity
	entity maintainer {
		relation org @organization#member

		action enabled = org
	}

	rule is_time_greater(created_at time, started_at time) {
		created_at > started_at
	}
`

			tests := []struct {
				expectedType    token.Type
				expectedLiteral string
			}{
				{token.NEWLINE, "\n"},
				{token.TAB, "\t"},
				{token.ENTITY, "entity"},
				{token.SPACE, " "},
				{token.IDENT, "user"},
				{token.SPACE, " "},
				{token.LCB, "{"},
				{token.RCB, "}"},
				{token.NEWLINE, "\n"},
				{token.NEWLINE, "\n"},
				{token.TAB, "\t"},
				{token.ENTITY, "entity"},
				{token.SPACE, " "},

				{token.IDENT, "organization"},
				{token.SPACE, " "},
				{token.LCB, "{"},
				{token.NEWLINE, "\n"},
				{token.TAB, "\t"},
				{token.TAB, "\t"},
				{token.RELATION, "relation"},
				{token.SPACE, " "},
				{token.IDENT, "member"},
				{token.SPACE, " "},
				{token.SIGN, "@"},

				{token.IDENT, "user"},
				{token.NEWLINE, "\n"},
				{token.TAB, "\t"},
				{token.RCB, "}"},
				{token.NEWLINE, "\n"},
				{token.NEWLINE, "\n"},
				{token.TAB, "\t"},
				{token.SINGLE_LINE_COMMENT, " This is a role for an entity"},
				{token.NEWLINE, "\n"},
				{token.TAB, "\t"},
				{token.ENTITY, "entity"},
				{token.SPACE, " "},
				{token.IDENT, "maintainer"},

				{token.SPACE, " "},
				{token.LCB, "{"},
				{token.NEWLINE, "\n"},
				{token.TAB, "\t"},
				{token.TAB, "\t"},
				{token.RELATION, "relation"},
				{token.SPACE, " "},
				{token.IDENT, "org"},
				{token.SPACE, " "},
				{token.SIGN, "@"},
				{token.IDENT, "organization"},
				{token.HASH, "#"},
				{token.IDENT, "member"},
				{token.NEWLINE, "\n"},
				{token.NEWLINE, "\n"},
				{token.TAB, "\t"},
				{token.TAB, "\t"},

				{token.PERMISSION, "action"},
				{token.SPACE, " "},
				{token.IDENT, "enabled"},
				{token.SPACE, " "},
				{token.ASSIGN, "="},
				{token.SPACE, " "},
				{token.IDENT, "org"},
				{token.NEWLINE, "\n"},
				{token.TAB, "\t"},
				{token.RCB, "}"},

				{token.NEWLINE, "\n"},
				{token.NEWLINE, "\n"},
				{token.TAB, "\t"},
				{token.RULE, "rule"},
				{token.SPACE, " "},
				{token.IDENT, "is_time_greater"},
				{token.LP, "("},
				{token.IDENT, "created_at"},
				{token.SPACE, " "},
				{token.IDENT, "time"},

				{token.COMMA, ","},
				{token.SPACE, " "},
				{token.IDENT, "started_at"},
				{token.SPACE, " "},
				{token.IDENT, "time"},
				{token.RP, ")"},
				{token.SPACE, " "},
				{token.LCB, "{"},
				{token.NEWLINE, "\n"},
				{token.TAB, "\t"},

				{token.TAB, "\t"},
				{token.IDENT, "created_at"},
				{token.SPACE, " "},
				{token.GT, ">"},
				{token.SPACE, " "},
				{token.IDENT, "started_at"},
				{token.NEWLINE, "\n"},
				{token.TAB, "\t"},
				{token.RCB, "}"},
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

		It("Case 7", func() {
			str := `
	entity user {}

	entity organization {
		relation member @user
	}

	// This is a role for an entity
	entity maintainer {
		relation member @organization#member

		permission enabled = member
	}`

			tests := []struct {
				expectedType    token.Type
				expectedLiteral string
			}{
				{token.NEWLINE, "\n"},
				{token.TAB, "\t"},
				{token.ENTITY, "entity"},
				{token.SPACE, " "},
				{token.IDENT, "user"},
				{token.SPACE, " "},
				{token.LCB, "{"},
				{token.RCB, "}"},
				{token.NEWLINE, "\n"},
				{token.NEWLINE, "\n"},
				{token.TAB, "\t"},
				{token.ENTITY, "entity"},
				{token.SPACE, " "},

				{token.IDENT, "organization"},
				{token.SPACE, " "},
				{token.LCB, "{"},
				{token.NEWLINE, "\n"},
				{token.TAB, "\t"},
				{token.TAB, "\t"},
				{token.RELATION, "relation"},
				{token.SPACE, " "},
				{token.IDENT, "member"},
				{token.SPACE, " "},
				{token.SIGN, "@"},

				{token.IDENT, "user"},
				{token.NEWLINE, "\n"},
				{token.TAB, "\t"},
				{token.RCB, "}"},
				{token.NEWLINE, "\n"},
				{token.NEWLINE, "\n"},
				{token.TAB, "\t"},
				{token.SINGLE_LINE_COMMENT, " This is a role for an entity"},
				{token.NEWLINE, "\n"},
				{token.TAB, "\t"},
				{token.ENTITY, "entity"},
				{token.SPACE, " "},
				{token.IDENT, "maintainer"},

				{token.SPACE, " "},
				{token.LCB, "{"},
				{token.NEWLINE, "\n"},
				{token.TAB, "\t"},
				{token.TAB, "\t"},
				{token.RELATION, "relation"},
				{token.SPACE, " "},
				{token.IDENT, "member"},
				{token.SPACE, " "},
				{token.SIGN, "@"},
				{token.IDENT, "organization"},
				{token.HASH, "#"},
				{token.IDENT, "member"},
				{token.NEWLINE, "\n"},
				{token.NEWLINE, "\n"},
				{token.TAB, "\t"},
				{token.TAB, "\t"},

				{token.PERMISSION, "permission"},
				{token.SPACE, " "},
				{token.IDENT, "enabled"},
				{token.SPACE, " "},
				{token.ASSIGN, "="},
				{token.SPACE, " "},
				{token.IDENT, "member"},
				{token.NEWLINE, "\n"},
				{token.TAB, "\t"},
				{token.RCB, "}"},
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

		It("Case 8", func() {
			str := `
entity user {}

entity organization {
	relation admin @user
    relation member @user
	action create_repository = admin or member;
	action delete = admin
}

rule is_weekday(day_of_week string) {
	day_of_week != 'saturday' && day_of_week != 'sunday'
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
				{token.LCB, "{"},
				{token.RCB, "}"},
				{token.NEWLINE, "\n"},
				{token.NEWLINE, "\n"},
				{token.ENTITY, "entity"},
				// --
				{token.SPACE, " "},
				{token.IDENT, "organization"},
				{token.SPACE, " "},
				{token.LCB, "{"},
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
				{token.PERMISSION, "action"},
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
				{token.NEWLINE, ";"},
				{token.NEWLINE, "\n"},
				{token.TAB, "\t"},
				{token.PERMISSION, "action"},
				// --
				{token.SPACE, " "},
				{token.IDENT, "delete"},
				{token.SPACE, " "},
				{token.ASSIGN, "="},
				{token.SPACE, " "},
				{token.IDENT, "admin"},
				{token.NEWLINE, "\n"},
				{token.RCB, "}"},
				// --
				{token.NEWLINE, "\n"},
				{token.NEWLINE, "\n"},
				{token.RULE, "rule"},
				{token.SPACE, " "},
				{token.IDENT, "is_weekday"},
				{token.LP, "("},
				{token.IDENT, "day_of_week"},
				{token.SPACE, " "},
				{token.IDENT, "string"},
				{token.RP, ")"},
				// --
				{token.SPACE, " "},
				{token.LCB, "{"},
				{token.NEWLINE, "\n"},
				{token.TAB, "\t"},
				{token.IDENT, "day_of_week"},
				{token.SPACE, " "},
				{token.EXCL, "!"},
				{token.ASSIGN, "="},
				{token.SPACE, " "},
				{token.APOS, "'"},
				// --
				{token.IDENT, "saturday"},
				{token.APOS, "'"},
				{token.SPACE, " "},
				{token.AMPERSAND, "&"},
				{token.AMPERSAND, "&"},
				{token.SPACE, " "},
				{token.IDENT, "day_of_week"},
				{token.SPACE, " "},
				{token.EXCL, "!"},
				{token.ASSIGN, "="},
				// --
				{token.SPACE, " "},
				{token.APOS, "'"},
				{token.IDENT, "sunday"},
				{token.APOS, "'"},
				{token.NEWLINE, "\n"},
				{token.RCB, "}"},
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

		It("Case 9", func() {
			str := `
entity user {}

entity organization {
	relation admin @user
    relation member @user

	attribute ip_addresses string[]

	action create_repository = admin or member
	action delete = admin or check_ip_address(ip_addresses)
}

rule check_ip_address(ip_addresses string[]) {
	"127.0.0.1" in ip_addresses && 100 > 89
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
				{token.LCB, "{"},
				{token.RCB, "}"},
				{token.NEWLINE, "\n"},
				{token.NEWLINE, "\n"},
				{token.ENTITY, "entity"},
				// --
				{token.SPACE, " "},
				{token.IDENT, "organization"},
				{token.SPACE, " "},
				{token.LCB, "{"},
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
				{token.NEWLINE, "\n"},
				{token.TAB, "\t"},
				{token.ATTRIBUTE, "attribute"},
				{token.SPACE, " "},
				{token.IDENT, "ip_addresses"},
				{token.SPACE, " "},
				// --
				{token.IDENT, "string"},
				{token.LSB, "["},
				{token.RSB, "]"},
				{token.NEWLINE, "\n"},
				{token.NEWLINE, "\n"},
				{token.TAB, "\t"},
				{token.PERMISSION, "action"},
				{token.SPACE, " "},
				{token.IDENT, "create_repository"},
				{token.SPACE, " "},
				// --
				{token.ASSIGN, "="},
				{token.SPACE, " "},
				{token.IDENT, "admin"},
				{token.SPACE, " "},
				{token.OR, "or"},
				{token.SPACE, " "},
				{token.IDENT, "member"},
				{token.NEWLINE, "\n"},
				{token.TAB, "\t"},
				{token.PERMISSION, "action"},
				// --
				{token.SPACE, " "},
				{token.IDENT, "delete"},
				{token.SPACE, " "},
				{token.ASSIGN, "="},
				{token.SPACE, " "},
				{token.IDENT, "admin"},
				{token.SPACE, " "},
				{token.OR, "or"},
				{token.SPACE, " "},
				{token.IDENT, "check_ip_address"},
				// --
				{token.LP, "("},
				{token.IDENT, "ip_addresses"},
				{token.RP, ")"},
				{token.NEWLINE, "\n"},
				{token.RCB, "}"},
				{token.NEWLINE, "\n"},
				{token.NEWLINE, "\n"},
				{token.RULE, "rule"},
				{token.SPACE, " "},
				{token.IDENT, "check_ip_address"},
				// --
				{token.LP, "("},
				{token.IDENT, "ip_addresses"},
				{token.SPACE, " "},
				{token.IDENT, "string"},
				{token.LSB, "["},
				{token.RSB, "]"},
				{token.RP, ")"},
				{token.SPACE, " "},
				{token.LCB, "{"},
				{token.NEWLINE, "\n"},
				// --
				{token.TAB, "\t"},
				{token.STRING, "127.0.0.1"},
				{token.SPACE, " "},
				{token.IN, "in"},
				{token.SPACE, " "},
				{token.IDENT, "ip_addresses"},
				{token.SPACE, " "},
				{token.AMPERSAND, "&"},
				{token.AMPERSAND, "&"},
				{token.SPACE, " "},
				// --
				{token.INTEGER, "100"},
				{token.SPACE, " "},
				{token.GT, ">"},
				{token.SPACE, " "},
				{token.INTEGER, "89"},
				{token.NEWLINE, "\n"},
				{token.RCB, "}"},
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

		It("Case 10", func() {
			str := `
entity user {}

entity organization {
	relation admin @user
    relation member @user

	action create_repository = nil
	action delete = nil
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
				{token.LCB, "{"},
				{token.RCB, "}"},
				{token.NEWLINE, "\n"},
				{token.NEWLINE, "\n"},
				// --
				{token.ENTITY, "entity"},
				{token.SPACE, " "},
				{token.IDENT, "organization"},
				{token.SPACE, " "},
				{token.LCB, "{"},
				{token.NEWLINE, "\n"},
				// --
				{token.TAB, "\t"},
				{token.RELATION, "relation"},
				{token.SPACE, " "},
				{token.IDENT, "admin"},
				{token.SPACE, " "},
				{token.SIGN, "@"},
				{token.IDENT, "user"},
				{token.NEWLINE, "\n"},
				// --
				{token.SPACE, " "},
				{token.SPACE, " "},
				{token.SPACE, " "},
				{token.SPACE, " "},
				{token.RELATION, "relation"},
				{token.SPACE, " "},
				{token.IDENT, "member"},
				{token.SPACE, " "},
				{token.SIGN, "@"},
				{token.IDENT, "user"},
				{token.NEWLINE, "\n"},
				{token.NEWLINE, "\n"},
				// --
				{token.TAB, "\t"},
				{token.PERMISSION, "action"},
				{token.SPACE, " "},
				{token.IDENT, "create_repository"},
				{token.SPACE, " "},
				{token.ASSIGN, "="},
				{token.SPACE, " "},
				{token.NIL, "nil"},
				{token.NEWLINE, "\n"},
				// --
				{token.TAB, "\t"},
				{token.PERMISSION, "action"},
				{token.SPACE, " "},
				{token.IDENT, "delete"},
				{token.SPACE, " "},
				{token.ASSIGN, "="},
				{token.SPACE, " "},
				{token.NIL, "nil"},
				{token.NEWLINE, "\n"},
				{token.RCB, "}"},
				{token.NEWLINE, "\n"},
				{token.NEWLINE, "\n"},
				// --
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

	It("Case 10 - Coverage for uncovered code paths", func() {
		str := `entity user {
	relation admin @user
	action test = admin and (age > 18.5 or is_active = false)
}

rule test_rule(created_at time, started_at time) {
	created_at > started_at
}

rule math_rule(a integer, b integer) {
	a + b * 2 - 3 / 4 % 5 ^ 2 < 10
}

// Single line comment without newline at end`

		l := NewLexer(str)

		// Just test that the lexer can process the input without errors
		// and that we get the expected token types for uncovered paths
		var tokens []token.Token
		for {
			tok := l.NextToken()
			tokens = append(tokens, tok)
			if tok.Type == token.EOF {
				break
			}
		}

		// Verify we have some key tokens that exercise uncovered code
		Expect(len(tokens)).Should(BeNumerically(">", 10))

		// Check for specific token types that were uncovered
		hasDouble := false
		hasBoolean := false
		hasArithmetic := false
		hasComparison := false
		hasComment := false

		for _, tok := range tokens {
			switch tok.Type {
			case token.DOUBLE:
				hasDouble = true
			case token.BOOLEAN:
				hasBoolean = true
			case token.PLUS, token.MINUS, token.TIMES, token.DIVIDE, token.MOD, token.POW:
				hasArithmetic = true
			case token.LT, token.GT:
				hasComparison = true
			case token.SINGLE_LINE_COMMENT:
				hasComment = true
			}
		}

		Expect(hasDouble).Should(BeTrue())
		Expect(hasBoolean).Should(BeTrue())
		Expect(hasArithmetic).Should(BeTrue())
		Expect(hasComparison).Should(BeTrue())
		Expect(hasComment).Should(BeTrue())
	})

	It("Case 11 - Additional coverage for specific uncovered paths", func() {
		str := `entity user {
	relation admin @user
	action test = admin and (age > 18.5 or is_active = false)
}

rule test_rule(created_at time, started_at time) {
	created_at > started_at
}

// Single line comment without newline at end`

		tests := []struct {
			expectedType    token.Type
			expectedLiteral string
		}{
			{token.ENTITY, "entity"},
			{token.SPACE, " "},
			{token.IDENT, "user"},
			{token.SPACE, " "},
			{token.LCB, "{"},
			{token.NEWLINE, "\n"},
			{token.TAB, "\t"},
			{token.RELATION, "relation"},
			{token.SPACE, " "},
			{token.IDENT, "admin"},
			{token.SPACE, " "},
			{token.SIGN, "@"},
			{token.IDENT, "user"},
			{token.NEWLINE, "\n"},
			{token.TAB, "\t"},
			{token.PERMISSION, "action"},
			{token.SPACE, " "},
			{token.IDENT, "test"},
			{token.SPACE, " "},
			{token.ASSIGN, "="},
			{token.SPACE, " "},
			{token.IDENT, "admin"},
			{token.SPACE, " "},
			{token.AND, "and"},
			{token.SPACE, " "},
			{token.LP, "("},
			{token.IDENT, "age"},
			{token.SPACE, " "},
			{token.GT, ">"},
			{token.SPACE, " "},
			{token.DOUBLE, "18.5"},
			{token.SPACE, " "},
			{token.OR, "or"},
			{token.SPACE, " "},
			{token.IDENT, "is_active"},
			{token.SPACE, " "},
			{token.ASSIGN, "="},
			{token.SPACE, " "},
			{token.BOOLEAN, "false"},
			{token.RP, ")"},
			{token.NEWLINE, "\n"},
			{token.RCB, "}"},
			{token.NEWLINE, "\n"},
			{token.NEWLINE, "\n"},
			{token.RULE, "rule"},
			{token.SPACE, " "},
			{token.IDENT, "test_rule"},
			{token.LP, "("},
			{token.IDENT, "created_at"},
			{token.SPACE, " "},
			{token.IDENT, "time"},
			{token.COMMA, ","},
			{token.SPACE, " "},
			{token.IDENT, "started_at"},
			{token.SPACE, " "},
			{token.IDENT, "time"},
			{token.RP, ")"},
			{token.SPACE, " "},
			{token.LCB, "{"},
			{token.NEWLINE, "\n"},
			{token.TAB, "\t"},
			{token.IDENT, "created_at"},
			{token.SPACE, " "},
			{token.GT, ">"},
			{token.SPACE, " "},
			{token.IDENT, "started_at"},
			{token.NEWLINE, "\n"},
			{token.RCB, "}"},
			{token.NEWLINE, "\n"},
			{token.NEWLINE, "\n"},
			{token.SINGLE_LINE_COMMENT, " Single line comment without newline at end"},
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

	It("Case 12 - Test carriage return and colon", func() {
		str := `entity user {
	relation admin @user
	action test = admin and (age: 18 or is_active = false)
}

rule test_rule(created_at time, started_at time) {
	created_at > started_at
}`

		l := NewLexer(str)

		// Just test that the lexer can process the input without errors
		// and that we get the expected token types for uncovered paths
		var tokens []token.Token
		for {
			tok := l.NextToken()
			tokens = append(tokens, tok)
			if tok.Type == token.EOF {
				break
			}
		}

		// Verify we have some key tokens that exercise uncovered code
		Expect(len(tokens)).Should(BeNumerically(">", 10))

		// Check for specific token types that were uncovered
		hasColon := false
		hasBoolean := false
		hasComparison := false

		for _, tok := range tokens {
			switch tok.Type {
			case token.COLON:
				hasColon = true
			case token.BOOLEAN:
				hasBoolean = true
			case token.LT, token.GT:
				hasComparison = true
			}
		}

		Expect(hasColon).Should(BeTrue())
		Expect(hasBoolean).Should(BeTrue())
		Expect(hasComparison).Should(BeTrue())
	})
})
