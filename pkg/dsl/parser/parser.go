package parser

import (
	"errors"
	"fmt"

	"github.com/Permify/permify/pkg/dsl/ast"
	"github.com/Permify/permify/pkg/dsl/lexer"
	"github.com/Permify/permify/pkg/dsl/token"
	base "github.com/Permify/permify/pkg/pb/base/v1"
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

	// references
	entityReferences map[string]struct{}

	// relational references
	relationReferences map[string][]string
	actionReferences   map[string]struct{}

	// its contains all references
	// sample keys: entity_type#member, entity_type#read
	relationalReferences map[string]ast.RelationalReferenceType
}

type (
	prefixParseFn func() (ast.Expression, error)
	infixParseFn  func(ast.Expression) (ast.Expression, error)
)

// NewParser -
func NewParser(str string) (p *Parser) {
	p = &Parser{
		l:                    lexer.NewLexer(str),
		errors:               []string{},
		entityReferences:     map[string]struct{}{},
		relationReferences:   map[string][]string{},
		actionReferences:     map[string]struct{}{},
		relationalReferences: map[string]ast.RelationalReferenceType{},
	}

	p.prefixParseFns = make(map[token.Type]prefixParseFn)
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.NOT, p.parsePrefixExpression)

	p.infixParseFunc = make(map[token.Type]infixParseFn)
	p.registerInfix(token.AND, p.parseInfixExpression)
	p.registerInfix(token.OR, p.parseInfixExpression)

	return
}

// setEntityReference -
func (p *Parser) setEntityReference(key string) error {
	if p.entityReferences == nil {
		p.entityReferences = map[string]struct{}{}
	}
	if _, ok := p.entityReferences[key]; ok {
		p.errors = append(p.errors, base.ErrorCode_ERROR_CODE_DUPLICATED_ENTITY_REFERENCE.String())
		return errors.New(base.ErrorCode_ERROR_CODE_DUPLICATED_ENTITY_REFERENCE.String())
	}
	p.entityReferences[key] = struct{}{}
	return nil
}

// setRelationReference -
func (p *Parser) setRelationReference(key string, types []string) error {
	if p.relationReferences == nil {
		p.relationReferences = map[string][]string{}
	}
	if _, ok := p.relationReferences[key]; ok {
		p.errors = append(p.errors, base.ErrorCode_ERROR_CODE_DUPLICATED_RELATION_REFERENCE.String())
		return errors.New(base.ErrorCode_ERROR_CODE_DUPLICATED_RELATION_REFERENCE.String())
	}
	if _, ok := p.relationalReferences[key]; ok {
		p.errors = append(p.errors, base.ErrorCode_ERROR_CODE_DUPLICATED_RELATION_REFERENCE.String())
		return errors.New(base.ErrorCode_ERROR_CODE_DUPLICATED_RELATION_REFERENCE.String())
	}
	p.relationReferences[key] = types
	p.relationalReferences[key] = ast.RELATION
	return nil
}

// setActionReference -
func (p *Parser) setActionReference(key string) error {
	if p.actionReferences == nil {
		p.actionReferences = map[string]struct{}{}
	}
	if _, ok := p.actionReferences[key]; ok {
		p.errors = append(p.errors, base.ErrorCode_ERROR_CODE_DUPLICATED_ACTION_REFERENCE.String())
		return errors.New(base.ErrorCode_ERROR_CODE_DUPLICATED_ACTION_REFERENCE.String())
	}
	if _, ok := p.relationalReferences[key]; ok {
		p.errors = append(p.errors, base.ErrorCode_ERROR_CODE_DUPLICATED_RELATION_REFERENCE.String())
		return errors.New(base.ErrorCode_ERROR_CODE_DUPLICATED_RELATION_REFERENCE.String())
	}
	p.actionReferences[key] = struct{}{}
	p.relationalReferences[key] = ast.ACTION
	return nil
}

// next -
func (p *Parser) next() {
	for {
		peek := p.l.NextToken()
		if !token.IsIgnores(peek.Type) {
			p.currentToken = p.peekToken
			p.peekToken = peek
			break
		}
	}
}

// currentTokenIs -
func (p *Parser) currentTokenIs(tokens ...token.Type) bool {
	for _, t := range tokens {
		if p.currentToken.Type == t {
			return true
		}
	}
	return false
}

// peekTokenIs -
func (p *Parser) peekTokenIs(tokens ...token.Type) bool {
	for _, t := range tokens {
		if p.peekToken.Type == t {
			return true
		}
	}
	return false
}

// Error -
func (p *Parser) Error() error {
	if len(p.errors) == 0 {
		return nil
	}
	return errors.New(p.errors[0])
}

// Parse -
func (p *Parser) Parse() (*ast.Schema, error) {
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

	schema.SetEntityReferences(p.entityReferences)
	schema.SetRelationReferences(p.relationReferences)
	schema.SetActionReferences(p.actionReferences)

	schema.SetRelationalReferences(p.relationalReferences)
	return schema, nil
}

// parseStatement method based on defined token types
func (p *Parser) parseStatement() (ast.Statement, error) {
	switch p.currentToken.Type {
	case token.ENTITY:
		return p.parseEntityStatement()
	default:
		return nil, nil
	}
}

// parseEntityStatement returns a LET Statement AST Node
func (p *Parser) parseEntityStatement() (*ast.EntityStatement, error) {
	stmt := &ast.EntityStatement{Entity: p.currentToken}
	if !p.expectAndNext(token.IDENT) {
		return nil, p.Error()
	}

	stmt.Name = p.currentToken

	entityName := stmt.Name.Literal
	err := p.setEntityReference(entityName)
	if err != nil {
		return nil, err
	}

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
			relation, err := p.parseRelationStatement(entityName)
			if err != nil {
				return nil, p.Error()
			}
			stmt.RelationStatements = append(stmt.RelationStatements, relation)
		case token.ACTION:
			action, err := p.parseActionStatement(entityName)
			if err != nil {
				return nil, p.Error()
			}
			stmt.ActionStatements = append(stmt.ActionStatements, action)
		default:
			if !p.currentTokenIs(token.LBRACE) && !p.currentTokenIs(token.RBRACE) {
				p.currentError(token.RELATION, token.ACTION)
				return nil, p.Error()
			}
		}
		p.next()
	}

	return stmt, nil
}

// parseRelationStatement -
func (p *Parser) parseRelationStatement(entityName string) (*ast.RelationStatement, error) {
	stmt := &ast.RelationStatement{Relation: p.currentToken}
	if !p.expectAndNext(token.IDENT) {
		return nil, p.Error()
	}

	var relationTypeStatements []string

	stmt.Name = p.currentToken
	relationName := stmt.Name.Literal

	if !p.expect(token.SIGN) {
		return nil, p.Error()
	}

	for p.peekTokenIs(token.SIGN) {
		relSt, err := p.parseRelationTypeStatement()
		if err != nil {
			return nil, p.Error()
		}
		stmt.RelationTypes = append(stmt.RelationTypes, relSt)
		relationTypeStatements = append(relationTypeStatements, relSt.Ident.Literal)
	}

	err := p.setRelationReference(fmt.Sprintf("%v#%v", entityName, relationName), relationTypeStatements)
	if err != nil {
		return nil, err
	}

	return stmt, nil
}

// parseRelationTypeStatement -
func (p *Parser) parseRelationTypeStatement() (*ast.RelationTypeStatement, error) {
	if !p.expectAndNext(token.SIGN) {
		return nil, p.Error()
	}
	stmt := &ast.RelationTypeStatement{Sign: p.currentToken}
	if !p.expectAndNext(token.IDENT) {
		return nil, p.Error()
	}
	stmt.Ident = p.currentToken
	return stmt, nil
}

// parseActionStatement -
func (p *Parser) parseActionStatement(entityName string) (ast.Statement, error) {
	stmt := &ast.ActionStatement{Action: p.currentToken}

	if !p.expectAndNext(token.IDENT) {
		return nil, p.Error()
	}

	stmt.Name = p.currentToken
	err := p.setActionReference(fmt.Sprintf("%v#%v", entityName, stmt.Name.Literal))
	if err != nil {
		return nil, err
	}

	if !p.expectAndNext(token.ASSIGN) {
		return nil, p.Error()
	}

	p.next()

	ex, err := p.parseExpressionStatement()
	if err != nil {
		return nil, p.Error()
	}
	stmt.ExpressionStatement = ex

	return stmt, nil
}

// parseExpressionStatement -
func (p *Parser) parseExpressionStatement() (*ast.ExpressionStatement, error) {
	stmt := &ast.ExpressionStatement{}
	var err error
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
func (p *Parser) parseExpression(precedence int) (ast.Expression, error) {
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

	for precedence < p.peekPrecedence() {
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
func (p *Parser) parseInnerParen() (ast.Expression, error) {
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
func (p *Parser) parsePrefixExpression() (ast.Expression, error) {
	expression := &ast.PrefixExpression{
		Not:      p.currentToken,
		Operator: p.currentToken.Literal,
	}
	p.next()
	expression.Ident = p.currentToken
	expression.Value = p.currentToken.Literal
	return expression, nil
}

// parseInfixExpression
func (p *Parser) parseInfixExpression(left ast.Expression) (ast.Expression, error) {
	expression := &ast.InfixExpression{
		Op:       p.currentToken, // and, or
		Left:     left,
		Operator: ast.Operator(p.currentToken.Literal),
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
	if pr, ok := precedences[p.peekToken.Type]; ok {
		return pr
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
func (p *Parser) parseIdentifier() (ast.Expression, error) {
	return &ast.Identifier{Ident: p.currentToken, Value: p.currentToken.Literal}, nil
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

//// noInfixParseFnError -
//func (p *Parser) noInfixParseFnError(t token.Type) {
//	msg := fmt.Sprintf("%v:%v:no infix parse function for %s found", p.l.GetLinePosition(), p.l.GetColumnPosition(), t)
//	p.errors = append(p.errors, msg)
//}

//// illegal -
//func (p *Parser) illegal() {
//	msg := fmt.Sprintf("%v:%v:illegal token found", p.l.GetLinePosition(), p.l.GetColumnPosition())
//	p.errors = append(p.errors, msg)
//}

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
