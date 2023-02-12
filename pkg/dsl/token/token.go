package token

// Type -
type Type string

// String -
func (t Type) String() string {
	return string(t)
}

// WithIgnores -
type WithIgnores struct {
	Token   Token
	Ignores []Token
}

// Token -
type Token struct {
	Type    Type
	Literal string
}

// New -
func New(typ Type, ch byte) Token {
	return Token{Type: typ, Literal: string(ch)}
}

// keywords -
var keywords = map[string]Type{
	"entity":   ENTITY,
	"relation": RELATION,
	"action":   ACTION,
	"and":      AND,
	"or":       OR,
	"not":      NOT,
}

// ignores -
var ignores = map[Type]struct{}{
	SINGLE_LINE_COMMENT: {},
	MULTI_LINE_COMMENT:  {},
	SPACE:               {},
	TAB:                 {},
	NEWLINE:             {},
}

const (

	//
	// Special Types
	//

	EOF     = "EOF"
	ILLEGAL = "ILLEGAL"

	//
	// Identifiers & Literals
	//

	IDENT = "IDENT"

	//
	// Delimiters
	//

	COMMA = "COMMA"

	LBRACE = "LBRACE"
	RBRACE = "RBRACE"

	LPAREN = "LPAREN"
	RPAREN = "RPAREN"

	ASSIGN = "ASSIGN"
	SIGN   = "SIGN"

	NEWLINE = "NEWLINE"

	//
	// Keywords
	//

	ENTITY   = "ENTITY"
	RELATION = "RELATION"
	ACTION   = "ACTION"

	//
	// Prefix
	//

	NOT = "NOT"

	//
	// Logical
	//

	AND = "AND"
	OR  = "OR"

	//
	// Comments
	//

	SINGLE_LINE_COMMENT = "SINGLE_LINE_COMMENT"
	MULTI_LINE_COMMENT  = "MULTI_LINE_COMMENT"
	SPACE               = "SPACE"
	TAB                 = "TAB"
)

// LookupKeywords -
func LookupKeywords(ident string) Type {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}

// IsIgnores -
func IsIgnores(typ Type) bool {
	if _, ok := ignores[typ]; ok {
		return true
	}
	return false
}
