package lexer

import (
	"github.com/Permify/permify/pkg/dsl/token"
)

// Lexer - represents a lexical analyzer for the input source code.
type Lexer struct {
	// The input source code to be analyzed.
	input string
	// The current position in the input source code.
	position int
	// The next position to read in the input source code.
	readPosition int
	// The current line position in the input source code.
	linePosition int
	// The current column position in the input source code.
	columnPosition int
	// The current character being read from the input source code.
	ch byte
}

// NewLexer - creates a new Lexer instance with the given input source code.
func NewLexer(input string) (l *Lexer) {
	l = &Lexer{input: input, linePosition: 1, columnPosition: 1}
	l.readChar()
	return
}

// GetLinePosition - returns the current line position of the Lexer in the input source code.
func (l *Lexer) GetLinePosition() int {
	return l.linePosition
}

// GetColumnPosition - returns the current column position of the Lexer in the input source code.
func (l *Lexer) GetColumnPosition() int {
	return l.columnPosition
}

// readChar - reads the next character from the input source code and updates the Lexer's position and column position.
func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
	l.columnPosition++
}

// peekChar - peeks the next character from the input source code without advancing the Lexer's position.
func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

// NextToken returns the next token from the input string
func (l *Lexer) NextToken() (tok token.Token) {
	// switch statement to determine the type of token based on the current character
	switch l.ch {
	case '\t':
		tok = token.New(token.TAB, l.ch)
	case ' ':
		tok = token.New(token.SPACE, l.ch)
	case '\n':
		l.newLine()
		tok = token.New(token.NEWLINE, l.ch)
	case '\r':
		l.newLine()
		tok = token.New(token.NEWLINE, l.ch)
	case ';':
		tok = token.New(token.NEWLINE, l.ch)
	case '=':
		tok = token.New(token.ASSIGN, l.ch)
	case '@':
		tok = token.New(token.SIGN, l.ch)
	case '(':
		tok = token.New(token.LPAREN, l.ch)
	case ')':
		tok = token.New(token.RPAREN, l.ch)
	case '{':
		tok = token.New(token.LBRACE, l.ch)
	case '}':
		tok = token.New(token.RBRACE, l.ch)
	case ',':
		tok = token.New(token.COMMA, l.ch)
	case '#':
		tok = token.New(token.HASH, l.ch)
	case '.':
		tok = token.New(token.DOT, l.ch)
	case 0:
		tok = token.Token{Type: token.EOF, Literal: ""}
	default:
		// check if the character is a letter, and if so, lex the identifier and look up the keyword
		if isLetter(l.ch) {
			tok.Literal = l.lexIdent()
			tok.Type = token.LookupKeywords(tok.Literal)
			return
		}
		// check if the character is the start of a single-line comment
		if l.ch == '/' && l.peekChar() == '/' {
			tok.Literal = l.lexSingleLineComment()
			tok.Type = token.SINGLE_LINE_COMMENT
			return
			// check if the character is the start of a multi-line comment
		} else if l.ch == '/' && l.peekChar() == '*' {
			tok.Literal = l.lexMultiLineComment()
			tok.Type = token.MULTI_LINE_COMMENT
			return
		} else {
			// if none of the above cases match, create an illegal token with the current character
			tok = token.New(token.ILLEGAL, l.ch)
		}
	}
	// read the next character and return the token
	l.readChar()
	return
}

// newLine - increments the line position and resets the column position to 1.
func (l *Lexer) newLine() {
	l.linePosition++
	l.columnPosition = 1
}

// lexIdent - reads and returns an identifier.
// An identifier is a sequence of letters (upper and lowercase) and underscores.
func (l *Lexer) lexIdent() string {
	position := l.position
	for isLetter(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

// lexSingleLineComment - reads and returns a single line comment.
// A single line comment starts with "//" and ends at the end of the line.
func (l *Lexer) lexSingleLineComment() string {
	l.readChar()
	l.readChar()
	position := l.position
	for !isNewline(l.ch) {
		if l.ch == 0 {
			return l.input[position:l.position]
		}
		l.readChar()
	}
	return l.input[position:l.position]
}

// lexMultiLineComment - reads and returns a multi-line comment.
// A multi-line comment starts with "/" and ends with "/".
func (l *Lexer) lexMultiLineComment() string {
	l.readChar()
	l.readChar()
	position := l.position
	for !(l.ch == '*' && l.peekChar() == '/') {
		if l.ch == 0 {
			return l.input[position:l.position]
		}
		l.readChar()
	}
	l.readChar()
	l.readChar()
	return l.input[position : l.position-2]
}

// isNewline - returns true if the given byte is a newline character (\r or \n).
func isNewline(r byte) bool {
	return r == '\r' || r == '\n'
}

// isLetter - returns true if the given byte is a letter (upper or lowercase) or an underscore.
func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}
