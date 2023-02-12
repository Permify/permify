package lexer

import (
	"github.com/Permify/permify/pkg/dsl/token"
)

// Lexer -
type Lexer struct {
	input          string
	position       int
	readPosition   int
	linePosition   int
	columnPosition int
	ch             byte
}

// NewLexer -
func NewLexer(input string) (l *Lexer) {
	l = &Lexer{input: input, linePosition: 1, columnPosition: 1}
	l.readChar()
	return
}

// GetLinePosition -
func (l *Lexer) GetLinePosition() int {
	return l.linePosition
}

// GetColumnPosition -
func (l *Lexer) GetColumnPosition() int {
	return l.columnPosition
}

// readChar -
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

// peekChar -
func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

// NextToken -
func (l *Lexer) NextToken() (tok token.Token) {
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
	case 0:
		tok = token.Token{Type: token.EOF, Literal: ""}
	default:
		if isLetter(l.ch) {
			tok.Literal = l.lexIdent()
			tok.Type = token.LookupKeywords(tok.Literal)
			return
		}
		if l.ch == '/' && l.peekChar() == '/' {
			tok.Literal = l.lexSingleLineComment()
			tok.Type = token.SINGLE_LINE_COMMENT
			return
		} else if l.ch == '/' && l.peekChar() == '*' {
			tok.Literal = l.lexMultiLineComment()
			tok.Type = token.MULTI_LINE_COMMENT
			return
		} else {
			tok = token.New(token.ILLEGAL, l.ch)
		}
	}
	l.readChar()
	return
}

// newLine -
func (l *Lexer) newLine() {
	l.linePosition++
	l.columnPosition = 1
}

// lexIdent -
func (l *Lexer) lexIdent() string {
	position := l.position
	for isLetter(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

// lexSingleLineComment -
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

// lexMultiLineComment -
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

// isNewline -
func isNewline(r byte) bool {
	return r == '\r' || r == '\n'
}

// isLetter -
func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_' || ch == '.' || ch == '#'
}
