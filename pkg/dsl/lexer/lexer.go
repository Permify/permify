package lexer

import (
	"github.com/Permify/permify/pkg/dsl/token"
)

// Lexer -
type Lexer struct {
	input        string
	position     int
	readPosition int
	ch           byte
}

// NewLexer -
func NewLexer(input string) (l *Lexer) {
	l = &Lexer{input: input}
	l.readChar()
	return
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
}

// peekChar -
func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	} else {
		return l.input[l.readPosition]
	}
}

// NextToken -
func (l *Lexer) NextToken() (tok token.Token) {
	l.skipWhitespace()
	switch l.ch {
	case '\n':
		tok = token.New(token.NEWLINE, l.ch)
	case '\r':
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
	case '`':
		tok.Type = token.OPTION
		tok.Literal = l.lexBacktick()
	case '"':
		tok = token.New(token.QUOTE, l.ch)
	case ',':
		tok = token.New(token.COMMA, l.ch)
	case 0:
		tok = token.Token{Type: token.EOF, Literal: ""}
	default:
		if isLetter(l.ch) {
			tok.Literal = l.lexIdent()
			tok.Type = token.Lookup(tok.Literal)
		} else {
			tok = token.New(token.ILLEGAL, l.ch)
		}
	}
	l.readChar()
	return
}

// lexBacktick -
func (l *Lexer) lexBacktick() (lit string) {
	l.readChar()
	position := l.position
	for !isBacktick(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

// lexIdent -
func (l *Lexer) lexIdent() string {
	position := l.position
	for isLetter(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

// skipWhitespace -
func (l *Lexer) skipWhitespace() {
	for isSpace(l.ch) {
		l.readChar()
	}
}

// isBacktick -
func isBacktick(r byte) bool {
	return r == '`'
}

// isSpace -
func isSpace(r byte) bool {
	return r == ' ' || r == '\t'
}

// isNewline -
func isNewline(r byte) bool {
	return r == '\r' || r == '\n'
}

// isLetter -
func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_' || ch == '.'
}
