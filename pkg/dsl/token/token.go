package token

// PositionInfo - represents the current position in the input source code.
type PositionInfo struct {
	// The current line position in the input source code.
	LinePosition int
	// The current column position in the input source code.
	ColumnPosition int
}

// Type - defines a custom type for tokens.
type Type string

// String - converts the Type to a string.
func (t Type) String() string {
	return string(t)
}

// WithIgnores - is a helper struct that includes a token and a list of ignored tokens.
type WithIgnores struct {
	Token   Token
	Ignores []Token
}

// Token - represents a lexical token in the input source code.
type Token struct {
	// The current position in the input source code.
	PositionInfo PositionInfo
	// The type of the token.
	Type Type
	// The literal value of the token.
	Literal string
}

// New - creates a new Token with the given type and literal value.
func New(positionInfo PositionInfo, typ Type, ch byte) Token {
	return Token{PositionInfo: positionInfo, Type: typ, Literal: string(ch)}
}

// keywords - maps string keywords to their corresponding Type.
var keywords = map[string]Type{
	"entity":     ENTITY,
	"relation":   RELATION,
	"action":     PERMISSION,
	"permission": PERMISSION,
	"rule":       RULE,
	"attribute":  ATTRIBUTE,
	"and":        AND,
	"or":         OR,
	"not":        NOT,
	"in":         IN,
}

// ignores - maps ignored token types to an empty struct.
var ignores = map[Type]struct{}{
	SINGLE_LINE_COMMENT: {},
	MULTI_LINE_COMMENT:  {},
	SPACE:               {},
	TAB:                 {},
}

const (

	/*
		Special Types
	*/
	EOF     = "EOF"
	ILLEGAL = "ILLEGAL"
	NEWLINE = "NEWLINE"

	/*
		Identifiers & Literals
	*/
	IDENT   = "IDENT"
	STRING  = "STRING"
	INTEGER = "INTEGER"
	DOUBLE  = "DOUBLE"
	BOOLEAN = "BOOLEAN"

	/*
		Symbols
	*/
	COMMA     = "COMMA"
	LCB       = "LCB"
	RCB       = "RCB"
	LP        = "LP"
	RP        = "RP"
	ASSIGN    = "ASSIGN"
	SIGN      = "SIGN"
	COLON     = "COLON"
	HASH      = "HASH"
	QM        = "QM"
	DOT       = "DOT"
	LSB       = "LSB"
	RSB       = "RSB"
	EXCL      = "EXCL"
	PLUS      = "PLUS"
	MINUS     = "MINUS"
	TIMES     = "TIMES"
	DIVIDE    = "DIVIDE"
	MOD       = "MOD"
	POW       = "POW"
	GT        = "GT"
	LT        = "LT"
	APOS      = "APOSTROPHE"
	AMPERSAND = "AMPERSAND"

	/*
		Keywords
	*/
	ENTITY     = "ENTITY"
	RELATION   = "RELATION"
	PERMISSION = "PERMISSION"
	ATTRIBUTE  = "ATTRIBUTE"
	RULE       = "RULE"
	AND        = "AND"
	OR         = "OR"
	NOT        = "NOT"
	IN         = "IN"

	/*
		Comments
	*/
	SINGLE_LINE_COMMENT = "SINGLE_LINE_COMMENT"
	MULTI_LINE_COMMENT  = "MULTI_LINE_COMMENT"

	/*
		Whitespace
	*/
	SPACE = "SPACE"
	TAB   = "TAB"
)

// LookupKeywords - looks up a keyword in the keywords map and returns its corresponding Type.
func LookupKeywords(ident string) Type {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}

// IsIgnores - checks if the given Type is an ignored token type.
func IsIgnores(typ Type) bool {
	if _, ok := ignores[typ]; ok {
		return true
	}
	return false
}
