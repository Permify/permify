package parser

import (
	`fmt`

	`github.com/Permify/permify/pkg/dsl/ast`
	`github.com/Permify/permify/pkg/dsl/lexer`
	`github.com/Permify/permify/pkg/dsl/token`
)

const (
	_ int = iota

	LOWEST
	LOGIC
)

var precedences = map[token.Type]int{
	token.AND: LOGIC,
	token.OR:  LOGIC,
}

// Parser -
type Parser struct {
	l              *lexer.Lexer
	currentToken   token.Token
	peekToken      token.Token
	errors         []string
	prefixParseFns map[token.Type]prefixParseFn
	infixParseFunc map[token.Type]infixParseFn
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

// NewParser -
func NewParser(str string) (p *Parser) {
	p = &Parser{
		l:      lexer.NewLexer(str),
		errors: []string{},
	}

	p.prefixParseFns = make(map[token.Type]prefixParseFn)
	p.registerPrefix(token.IDENT, p.parseIdentifier)

	p.infixParseFunc = make(map[token.Type]infixParseFn)
	p.registerInfix(token.AND, p.parseInfixExpression)
	p.registerInfix(token.OR, p.parseInfixExpression)

	p.next()
	p.next()
	return
}

// nextToken -
func (p *Parser) next() {
	p.currentToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

// currentTokenIs -
func (p *Parser) currentTokenIs(t token.Type) bool {
	return p.currentToken.Type == t
}

// peekTokenIs -
func (p *Parser) peekTokenIs(t token.Type) bool {
	return p.peekToken.Type == t
}

// Errors -
func (p *Parser) Errors() []string {
	return p.errors
}

// Parse -
func (p *Parser) Parse() *ast.Schema {
	schema := &ast.Schema{}
	schema.Statements = []ast.Statement{}

	for !p.currentTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			schema.Statements = append(schema.Statements, stmt)
		}
		p.next()
	}

	return schema
}

// parseStatement method based on defined token types
func (p *Parser) parseStatement() ast.Statement {
	switch p.currentToken.Type {
	case token.ENTITY:
		return p.parseEntityStatement()
	default:
		return nil
	}
}

// parseEntityStatement returns a LET Statement AST Node
func (p *Parser) parseEntityStatement() *ast.EntityStatement {
	stmt := &ast.EntityStatement{Token: p.currentToken}
	if !p.expectAndNext(token.IDENT) {
		return nil
	}
	stmt.Name = p.currentToken
	for !p.currentTokenIs(token.RBRACE) {
		switch p.currentToken.Type {
		case token.RELATION:
			stmt.RelationStatements = append(stmt.RelationStatements, p.parseRelationStatement())
			break
		case token.ACTION:
			stmt.ActionStatements = append(stmt.ActionStatements, p.parseActionStatement())
			break
		default:
			break
		}
		p.next()
	}

	if p.expectAndNext(token.OPTION) {
		stmt.Option = p.currentToken
	}

	return stmt
}

// parseRelationStatement -
func (p *Parser) parseRelationStatement() *ast.RelationStatement {
	stmt := &ast.RelationStatement{Token: p.currentToken}
	if !p.expectAndNext(token.IDENT) {
		return nil
	}
	stmt.Name = p.currentToken

	if !p.expectAndNext(token.SIGN) {
		return nil
	}

	stmt.Sign = p.currentToken

	if !p.expectAndNext(token.IDENT) {
		return nil
	}

	stmt.Type = p.currentToken

	if p.expectAndNext(token.OPTION) {
		stmt.Option = p.currentToken
	}

	return stmt
}

// parseActionStatement -
func (p *Parser) parseActionStatement() ast.Statement {
	stmt := &ast.ActionStatement{Token: p.currentToken}

	if !p.expectAndNext(token.IDENT) {
		return nil
	}

	stmt.Name = p.currentToken

	if !p.expectAndNext(token.ASSIGN) {
		return nil
	}

	p.next()

	stmt.ExpressionStatement = p.parseExpressionStatement()

	return stmt
}

// parseRewrite -
func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {

	stmt := &ast.ExpressionStatement{}
	stmt.Expression = p.parseExpression(LOWEST)

	if p.currentTokenIs(token.NEWLINE) {
		p.next()
	}

	return stmt
}

// expectPeek -
func (p *Parser) expectAndNext(t token.Type) bool {
	if p.peekTokenIs(t) {
		p.next()
		return true
	}
	if t == token.OPTION {
		return false
	}
	p.peekError(t)
	return false
}

// parseExpression -
func (p *Parser) parseExpression(precedence int) ast.Expression {

	if p.currentTokenIs(token.LPAREN) {
		p.next()
		return p.parseInnerParen()
	}

	prefix := p.prefixParseFns[p.currentToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.currentToken.Type)
		return nil
	}
	left := prefix()

	for !p.peekTokenIs(token.NEWLINE) && precedence < p.peekPrecedence() {
		infix := p.infixParseFunc[p.peekToken.Type]
		if infix == nil {
			return left
		}
		p.next()
		left = infix(left)
	}

	return left
}

// parseInnerParen -
func (p *Parser) parseInnerParen() ast.Expression {

	if p.currentTokenIs(token.LPAREN) {
		return p.parseExpression(LOWEST)
	}

	prefix := p.prefixParseFns[p.currentToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.currentToken.Type)
		return nil
	}
	left := prefix()

	for !p.currentTokenIs(token.RPAREN) {
		infix := p.infixParseFunc[p.peekToken.Type]
		if infix == nil {
			return left
		}
		p.next()
		left = infix(left)
	}

	return left
}

// parseInfixExpression
func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token:    p.currentToken, // and, or
		Left:     left,
		Operator: p.currentToken.Literal,
	}
	precedence := p.currentPrecedence()
	p.next()
	expression.Right = p.parseExpression(precedence)
	return expression
}

// peekPrecedence -
func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

// peekPrecedence -
func (p *Parser) currentPrecedence() int {
	if p, ok := precedences[p.currentToken.Type]; ok {
		return p
	}
	return LOWEST
}

// parseIdentifier
func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.currentToken, Value: p.currentToken.Literal}
}

// registerPrefix
func (p *Parser) registerPrefix(tokenType token.Type, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

// registerInfix
func (p *Parser) registerInfix(tokenType token.Type, fn infixParseFn) {
	p.infixParseFunc[tokenType] = fn
}

// noPrefixParseFnError -
func (p *Parser) noPrefixParseFnError(t token.Type) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.errors = append(p.errors, msg)
}

// peekError -
func (p *Parser) peekError(t token.Type) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead", t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}
