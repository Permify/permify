package token

// Type -
type Type string

// String -
func (t Type) String() string {
	return string(t)
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

	QUOTE  = "QUOTE"
	OPTION = "OPTION"
)

// LookupKeywords -
func LookupKeywords(ident string) Type {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}
