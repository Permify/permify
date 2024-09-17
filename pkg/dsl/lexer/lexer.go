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
		tok = token.New(positionInfo(l.linePosition, l.columnPosition), token.TAB, l.ch)
	case ' ':
		tok = token.New(positionInfo(l.linePosition, l.columnPosition), token.SPACE, l.ch)
	case '\n':
		l.newLine()
		tok = token.New(positionInfo(l.linePosition, l.columnPosition), token.NEWLINE, l.ch)
	case '\r':
		l.newLine()
		tok = token.New(positionInfo(l.linePosition, l.columnPosition), token.NEWLINE, l.ch)
	case ';':
		tok = token.New(positionInfo(l.linePosition, l.columnPosition), token.NEWLINE, l.ch)
	case ':':
		tok = token.New(positionInfo(l.linePosition, l.columnPosition), token.COLON, l.ch)
	case '=':
		tok = token.New(positionInfo(l.linePosition, l.columnPosition), token.ASSIGN, l.ch)
	case '@':
		tok = token.New(positionInfo(l.linePosition, l.columnPosition), token.SIGN, l.ch)
	case '(':
		tok = token.New(positionInfo(l.linePosition, l.columnPosition), token.LP, l.ch)
	case ')':
		tok = token.New(positionInfo(l.linePosition, l.columnPosition), token.RP, l.ch)
	case '{':
		tok = token.New(positionInfo(l.linePosition, l.columnPosition), token.LCB, l.ch)
	case '}':
		tok = token.New(positionInfo(l.linePosition, l.columnPosition), token.RCB, l.ch)
	case '[':
		tok = token.New(positionInfo(l.linePosition, l.columnPosition), token.LSB, l.ch)
	case ']':
		tok = token.New(positionInfo(l.linePosition, l.columnPosition), token.RSB, l.ch)
	case '+':
		tok = token.New(positionInfo(l.linePosition, l.columnPosition), token.PLUS, l.ch)
	case '-':
		tok = token.New(positionInfo(l.linePosition, l.columnPosition), token.MINUS, l.ch)
	case '*':
		tok = token.New(positionInfo(l.linePosition, l.columnPosition), token.TIMES, l.ch)
	case '%':
		tok = token.New(positionInfo(l.linePosition, l.columnPosition), token.MOD, l.ch)
	case '^':
		tok = token.New(positionInfo(l.linePosition, l.columnPosition), token.POW, l.ch)
	case '>':
		tok = token.New(positionInfo(l.linePosition, l.columnPosition), token.GT, l.ch)
	case '<':
		tok = token.New(positionInfo(l.linePosition, l.columnPosition), token.LT, l.ch)
	case '!':
		tok = token.New(positionInfo(l.linePosition, l.columnPosition), token.EXCL, l.ch)
	case '?':
		tok = token.New(positionInfo(l.linePosition, l.columnPosition), token.QM, l.ch)
	case ',':
		tok = token.New(positionInfo(l.linePosition, l.columnPosition), token.COMMA, l.ch)
	case '#':
		tok = token.New(positionInfo(l.linePosition, l.columnPosition), token.HASH, l.ch)
	case '.':
		tok = token.New(positionInfo(l.linePosition, l.columnPosition), token.DOT, l.ch)
	case '\'':
		tok = token.New(positionInfo(l.linePosition, l.columnPosition), token.APOS, l.ch)
	case '&':
		tok = token.New(positionInfo(l.linePosition, l.columnPosition), token.AMPERSAND, l.ch)
	case 0:
		tok = token.Token{PositionInfo: positionInfo(l.linePosition, l.columnPosition), Type: token.EOF, Literal: ""}
	case '/':
		switch l.peekChar() {
		case '/':
			tok.PositionInfo = positionInfo(l.linePosition, l.columnPosition)
			tok.Literal = l.lexSingleLineComment()
			tok.Type = token.SINGLE_LINE_COMMENT
			return
		case '*':
			tok.PositionInfo = positionInfo(l.linePosition, l.columnPosition)
			tok.Literal = l.lexMultiLineComment()
			tok.Type = token.MULTI_LINE_COMMENT
			return
		default:
			tok = token.New(positionInfo(l.linePosition, l.columnPosition), token.DIVIDE, l.ch)
		}
	case '"':
		// check if the character is a double quote, indicating a string
		tok.PositionInfo = positionInfo(l.linePosition, l.columnPosition)
		tok.Literal = l.lexString()
		tok.Type = token.STRING
		return
	default:
		// check if the character is a letter, and if so, lex the identifier and look up the keyword
		if isLetter(l.ch) {
			tok.PositionInfo = positionInfo(l.linePosition, l.columnPosition)
			tok.Literal = l.lexIdent()
			if tok.Literal == "true" || tok.Literal == "false" {
				tok.Type = token.BOOLEAN
				return
			}
			tok.Type = token.LookupKeywords(tok.Literal)
			return
		} else if isDigit(l.ch) {
			var isDouble bool
			tok.PositionInfo = positionInfo(l.linePosition, l.columnPosition)
			tok.Literal, isDouble = l.lexNumber()
			if isDouble {
				tok.Type = token.DOUBLE
			} else {
				tok.Type = token.INTEGER
			}
			return
		} else {
			// if none of the above cases match, create an illegal token with the current character
			tok = token.New(positionInfo(l.linePosition, l.columnPosition), token.ILLEGAL, l.ch)
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

// lexNumber - reads and returns a number.
func (l *Lexer) lexNumber() (string, bool) {
	position := l.position
	seenDot := false
	for isDigit(l.ch) || (!seenDot && l.ch == '.') {
		if l.ch == '.' {
			seenDot = true
		}
		l.readChar()
	}
	return l.input[position:l.position], seenDot
}

// lexString lex a string literal. It does not support escape sequences or multi-line strings.
func (l *Lexer) lexString() string {
	// Skip the initial quotation mark.
	l.readChar()
	position := l.position
	var str string
	for {
		if l.ch == '\\' {
			str += l.input[position:l.position]
			l.readChar() // Skip the backslash
			switch l.ch {
			case 'n':
				str += "\n"
			case 't':
				str += "\t"
			case '"':
				str += "\""
			case '\\':
				str += "\\"
			}
			position = l.position + 1
		} else if l.ch == '"' || l.ch == 0 {
			break
		}
		l.readChar()
	}
	str += l.input[position:l.position]
	if l.ch == '"' {
		l.readChar()
	}
	return str
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

// isDigit - returns true if the given byte is a digit.
func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

// positionInfo - returns a token.PositionInfo struct with the current line and column position.
func positionInfo(line, column int) token.PositionInfo {
	return token.PositionInfo{
		LinePosition:   line,
		ColumnPosition: column,
	}
}
