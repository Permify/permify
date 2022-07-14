package ast

import (
	`strings`

	`github.com/Permify/permify/pkg/dsl/token`
)

// Node -
type Node interface {
	TokenLiteral() string
	String() string
}

// Expression -
type Expression interface {
	Node
	expressionNode()
	IsInfix() bool
}

// Statement -
type Statement interface {
	Node
	statementNode()
}

// Schema -
type Schema struct {
	Statements []Statement
}

// EntityStatement -
type EntityStatement struct {
	Token              token.Token // token.ENTITY
	Name               token.Token // token.IDENT
	RelationStatements []Statement
	ActionStatements   []Statement
	Option             token.Token // token.OPTION
}

// statementNode -
func (ls *EntityStatement) statementNode() {}

// TokenLiteral -
func (ls *EntityStatement) TokenLiteral() string {
	return ls.Token.Literal
}

// String -
func (ls *EntityStatement) String() string {
	var sb strings.Builder
	sb.WriteString("entity")
	sb.WriteString(" ")
	sb.WriteString(ls.Name.Literal)
	sb.WriteString(" {")
	sb.WriteString("\n")

	for _, rs := range ls.RelationStatements {
		sb.WriteString(rs.String())
		sb.WriteString("\n")
	}

	sb.WriteString("\n")

	for _, rs := range ls.ActionStatements {
		sb.WriteString(rs.String())
		sb.WriteString("\n")
	}

	sb.WriteString("}")
	sb.WriteString(" ")
	sb.WriteString("`")
	sb.WriteString(ls.Option.Literal)
	sb.WriteString("`")
	sb.WriteString("\n")
	return sb.String()
}

// RelationStatement -
type RelationStatement struct {
	Token  token.Token // token.RELATION
	Name   token.Token // token.IDENT
	Sign   token.Token // token.SIGN
	Type   token.Token // token.IDENT
	Option token.Token // token.OPTION
}

// statementNode -
func (ls *RelationStatement) statementNode() {}

// TokenLiteral -
func (ls *RelationStatement) TokenLiteral() string {
	return ls.Token.Literal
}

// String -
func (ls *RelationStatement) String() string {
	var sb strings.Builder
	sb.WriteString("\trelation")
	sb.WriteString(" ")
	sb.WriteString(ls.Name.Literal)
	sb.WriteString(" ")
	sb.WriteString(ls.Sign.Literal)
	sb.WriteString(ls.Type.Literal)
	sb.WriteString(" ")
	sb.WriteString("`")
	sb.WriteString(ls.Option.Literal)
	sb.WriteString("`")
	return sb.String()
}

// Identifier -
type Identifier struct {
	Token token.Token // token.IDENT
	Value string
}

// statementNode -
func (ls *Identifier) expressionNode() {}

// TokenLiteral -
func (ls *Identifier) TokenLiteral() string {
	return ls.Token.Literal
}

// String -
func (ls *Identifier) String() string {
	return ls.Value
}

// IsInfix -
func (ls *Identifier) IsInfix() bool {
	return false
}

// ActionStatement -
type ActionStatement struct {
	Token               token.Token // token.ACTION
	Name                token.Token // token.IDENT
	ExpressionStatement Statement
}

// statementNode -
func (ls *ActionStatement) statementNode() {}

// TokenLiteral -
func (ls *ActionStatement) TokenLiteral() string {
	return ls.Token.Literal
}

// String -
func (ls *ActionStatement) String() string {
	var sb strings.Builder
	sb.WriteString("\t" + ls.TokenLiteral() + " ")
	sb.WriteString(ls.Name.Literal)
	sb.WriteString(" = ")
	if ls.ExpressionStatement != nil {
		sb.WriteString(ls.ExpressionStatement.String())
	}
	return sb.String()
}

// ExpressionStatement struct
type ExpressionStatement struct {
	Expression Expression
}

// statementNode function on ExpressionStatement
func (es *ExpressionStatement) statementNode() {}

// TokenLiteral function on ExpressionStatement
func (es *ExpressionStatement) TokenLiteral() string {
	return "start"
}

func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}

// InfixExpression -
type InfixExpression struct {
	Token    token.Token // The operator token, e.g. and, or
	Left     Expression
	Operator string
	Right    Expression
}

// expressionNode -
func (ie *InfixExpression) expressionNode() {}

// TokenLiteral -
func (ie *InfixExpression) TokenLiteral() string {
	return ie.Token.Literal
}

// String -
func (ie *InfixExpression) String() string {
	var sb strings.Builder
	sb.WriteString("(")
	sb.WriteString(ie.Left.String())
	sb.WriteString(" " + ie.Operator)
	sb.WriteString(" ")
	sb.WriteString(ie.Right.String())
	sb.WriteString(")")
	return sb.String()
}

// IsInfix -
func (ie *InfixExpression) IsInfix() bool {
	return true
}
