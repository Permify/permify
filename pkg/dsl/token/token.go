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
	DOT    = "DOT"

	COLON     = "COLON"
	SEMICOLON = "SEMICOLON"

	NEWLINE = "NEWLINE"

	//
	// Keywords
	//

	ENTITY   = "ENTITY"
	RELATION = "RELATION"
	ACTION   = "ACTION"

	//
	// Logical
	//

	AND = "AND"
	OR  = "OR"

	QUOTE  = "QUOTE"
	OPTION = "OPTION"
)

// Lookup -
func Lookup(ident string) Type {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}
