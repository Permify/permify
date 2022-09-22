package parser

import (
	"fmt"
	"strings"

	"github.com/Permify/permify/pkg/dsl/ast"
	"github.com/Permify/permify/pkg/dsl/lexer"
	"github.com/Permify/permify/pkg/dsl/token"
	"github.com/Permify/permify/pkg/errors"
	`github.com/Permify/permify/pkg/helper`
)

const (
	_ int = iota

	LOWEST
	LOGIC
	PREFIX // not IDENT
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
	prefixParseFn func() (ast.Expression, errors.Error)
	infixParseFn  func(ast.Expression) (ast.Expression, errors.Error)
)

// NewParser -
func NewParser(str string) (p *Parser) {
	p = &Parser{
		l:      lexer.NewLexer(str),
		errors: []string{},
	}

	p.prefixParseFns = make(map[token.Type]prefixParseFn)
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.NOT, p.parsePrefixExpression)

	p.infixParseFunc = make(map[token.Type]infixParseFn)
	p.registerInfix(token.AND, p.parseInfixExpression)
	p.registerInfix(token.OR, p.parseInfixExpression)

	p.next()
	p.next()
	return
}

// next -
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

// Error -
func (p *Parser) Error() errors.Error {
	if len(p.errors) == 0 {
		return nil
	}
	return errors.NewError(errors.Validation).SetParams(map[string]interface{}{
		"schema": strings.Join(p.errors, ","),
	})
}

// Parse -
func (p *Parser) Parse() (*ast.Schema, errors.Error) {
	schema := &ast.Schema{}
	schema.Statements = []ast.Statement{}

	for !p.currentTokenIs(token.EOF) {
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, p.Error()
		}
		if stmt != nil {
			schema.Statements = append(schema.Statements, stmt)
		}
		p.next()
	}

	return schema, nil
}

// parseStatement method based on defined token types
func (p *Parser) parseStatement() (ast.Statement, errors.Error) {
	switch p.currentToken.Type {
	case token.ENTITY:
		return p.parseEntityStatement()
	default:
		return nil, nil
	}
}

// parseEntityStatement returns a LET Statement AST Node
func (p *Parser) parseEntityStatement() (*ast.EntityStatement, errors.Error) {
	stmt := &ast.EntityStatement{Token: p.currentToken}
	if !p.expectAndNext(token.IDENT) {
		return nil, p.Error()
	}
	stmt.Name = p.currentToken
	if !p.expectAndNext(token.LBRACE) {
		return nil, p.Error()
	}

	for !p.currentTokenIs(token.RBRACE) {
		if p.currentTokenIs(token.EOF) {
			p.currentError(token.RBRACE)
			return nil, p.Error()
		}
		switch p.currentToken.Type {
		case token.RELATION:
			relation, err := p.parseRelationStatement()
			if err != nil {
				return nil, p.Error()
			}
			stmt.RelationStatements = append(stmt.RelationStatements, relation)
			break
		case token.ACTION:
			action, err := p.parseActionStatement()
			if err != nil {
				return nil, p.Error()
			}
			stmt.ActionStatements = append(stmt.ActionStatements, action)
			break
		default:
			//if !p.currentTokenIs(token.NEWLINE) && !p.currentTokenIs(token.LBRACE) && !p.currentTokenIs(token.RBRACE) {
			//	p.currentError(token.RELATION, token.ACTION)
			//	return nil, p.Error()
			//}
			break
		}
		p.next()
	}

	if p.peekTokenIs(token.OPTION) {
		p.next()
		stmt.Option = p.currentToken
	}

	return stmt, nil
}

// parseRelationStatement -
func (p *Parser) parseRelationStatement() (*ast.RelationStatement, errors.Error) {
	stmt := &ast.RelationStatement{Token: p.currentToken}
	if !p.expectAndNext(token.IDENT) {
		return nil, p.Error()
	}
	stmt.Name = p.currentToken

	if !p.expect(token.SIGN) {
		return nil, p.Error()
	}

	for p.peekTokenIs(token.SIGN) && !p.peekTokenIs(token.OPTION) {
		relSt, err := p.parseRelationTypeStatement()
		if err != nil {
			return nil, p.Error()
		}
		stmt.RelationTypes = append(stmt.RelationTypes, relSt)
	}

	if p.peekTokenIs(token.OPTION) {
		p.next()
		stmt.Option = p.currentToken
	}

	return stmt, nil
}

// parseRelationTypeStatement -
func (p *Parser) parseRelationTypeStatement() (*ast.RelationTypeStatement, errors.Error) {
	if !p.expectAndNext(token.SIGN) {
		return nil, p.Error()
	}
	stmt := &ast.RelationTypeStatement{Sign: p.currentToken}
	if !p.expectAndNext(token.IDENT) {
		return nil, p.Error()
	}
	stmt.Token = p.currentToken
	return stmt, nil
}

// parseActionStatement -
func (p *Parser) parseActionStatement() (ast.Statement, errors.Error) {
	stmt := &ast.ActionStatement{Token: p.currentToken}

	if !p.expectAndNext(token.IDENT) {
		return nil, p.Error()
	}

	stmt.Name = p.currentToken

	if !p.expectAndNext(token.ASSIGN) {
		return nil, p.Error()
	}

	p.next()

	ex, err := p.parseExpressionStatement()
	if err != nil {
		return nil, p.Error()
	}
	stmt.ExpressionStatement = ex

	helper.Pre(stmt)
	return stmt, nil
}

// parseExpressionStatement -
func (p *Parser) parseExpressionStatement() (*ast.ExpressionStatement, errors.Error) {
	stmt := &ast.ExpressionStatement{}
	var err errors.Error
	stmt.Expression, err = p.parseExpression(LOWEST)
	if err != nil {
		return nil, p.Error()
	}

	if p.peekTokenIs(token.RPAREN) {
		p.next()
		for p.currentTokenIs(token.RPAREN) {
			p.next()
		}
	}

	return stmt, nil
}

// expectAndNext -
func (p *Parser) expectAndNext(t token.Type) bool {
	if p.peekTokenIs(t) {
		p.next()
		return true
	}
	p.peekError(t)
	return false
}

// expect -
func (p *Parser) expect(t token.Type) bool {
	if p.peekTokenIs(t) {
		return true
	}
	p.peekError(t)
	return false
}

// parseExpression -
func (p *Parser) parseExpression(precedence int) (ast.Expression, errors.Error) {
	if p.currentTokenIs(token.LPAREN) {
		p.next()
		return p.parseInnerParen()
	}

	prefix := p.prefixParseFns[p.currentToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.currentToken.Type)
		return nil, p.Error()
	}
	exp, err := prefix()
	if err != nil {
		return nil, p.Error()
	}

	for !p.peekTokenIs(token.NEWLINE) && precedence < p.peekPrecedence() {
		infix := p.infixParseFunc[p.peekToken.Type]
		if infix == nil {
			return exp, nil
		}
		p.next()
		exp, err = infix(exp)
		if err != nil {
			return nil, p.Error()
		}
	}

	return exp, nil
}

// parseInnerParen -
func (p *Parser) parseInnerParen() (ast.Expression, errors.Error) {
	if p.currentTokenIs(token.LPAREN) {
		return p.parseExpression(LOWEST)
	}

	prefix := p.prefixParseFns[p.currentToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.currentToken.Type)
		return nil, p.Error()
	}
	exp, err := prefix()
	if err != nil {
		return nil, p.Error()
	}

	for !p.currentTokenIs(token.RPAREN) {
		if p.peekTokenIs(token.RPAREN) {
			p.next()
		}
		infix := p.infixParseFunc[p.peekToken.Type]
		if infix == nil {
			return exp, nil
		}
		p.next()
		exp, err = infix(exp)
		if err != nil {
			return nil, p.Error()
		}
	}

	return exp, nil
}

// parsePrefixExpression -
func (p *Parser) parsePrefixExpression() (ast.Expression, errors.Error) {
	expression := &ast.PrefixExpression{
		Token:    p.currentToken,
		Operator: p.currentToken.Literal,
	}
	p.next()
	expression.Value = p.currentToken.Literal
	return expression, nil
}

// parseInfixExpression
func (p *Parser) parseInfixExpression(left ast.Expression) (ast.Expression, errors.Error) {
	expression := &ast.InfixExpression{
		Token:    p.currentToken, // and, or
		Left:     left,
		Operator: p.currentToken.Literal,
	}
	precedence := p.currentPrecedence()
	p.next()
	ex, err := p.parseExpression(precedence)
	if err != nil {
		return nil, p.Error()
	}
	expression.Right = ex
	return expression, nil
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
	if pr, ok := precedences[p.currentToken.Type]; ok {
		return pr
	}
	return LOWEST
}

// parseIdentifier
func (p *Parser) parseIdentifier() (ast.Expression, errors.Error) {
	return &ast.Identifier{Token: p.currentToken, Value: p.currentToken.Literal}, nil
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
	msg := fmt.Sprintf("%v:%v:no prefix parse function for %s found", p.l.GetLinePosition(), p.l.GetColumnPosition(), t)
	p.errors = append(p.errors, msg)
}

// noInfixParseFnError -
func (p *Parser) noInfixParseFnError(t token.Type) {
	msg := fmt.Sprintf("%v:%v:no infix parse function for %s found", p.l.GetLinePosition(), p.l.GetColumnPosition(), t)
	p.errors = append(p.errors, msg)
}

// illegal -
func (p *Parser) illegal() {
	msg := fmt.Sprintf("%v:%v:illegal token found", p.l.GetLinePosition(), p.l.GetColumnPosition())
	p.errors = append(p.errors, msg)
}

// peekError -
func (p *Parser) peekError(t ...token.Type) {
	msg := fmt.Sprintf("%v:%v:expected next token to be %s, got %s instead", p.l.GetLinePosition(), p.l.GetColumnPosition(), t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

// currentError -
func (p *Parser) currentError(t ...token.Type) {
	msg := fmt.Sprintf("%v:%v:expected token to be %s, got %s instead", p.l.GetLinePosition(), p.l.GetColumnPosition(), t, p.currentToken.Type)
	p.errors = append(p.errors, msg)
}
