package ast

import (
	"strings"

	"github.com/Permify/permify/pkg/dsl/token"
)

type (
	ExpressionType          string
	RelationalReferenceType string
	Operator                string
)

// String -
func (o Operator) String() string {
	return string(o)
}

const (
	IDENTIFIER ExpressionType = "identifier"
	INFLIX     ExpressionType = "inflix"

	AND Operator = "and"
	OR  Operator = "or"

	ACTION   RelationalReferenceType = "action"
	RELATION RelationalReferenceType = "relation"
)

// Node -
type Node interface {
	String() string
}

// Expression -
type Expression interface {
	Node
	expressionNode()
	IsInfix() bool
	GetType() ExpressionType
}

// Statement -
type Statement interface {
	Node
	statementNode()
}

// EntityStatement -
type EntityStatement struct {
	Entity             token.Token // token.ENTITY
	Name               token.Token // token.IDENT
	RelationStatements []Statement
	ActionStatements   []Statement
}

// statementNode -
func (ls *EntityStatement) statementNode() {}

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
	sb.WriteString("\n")
	return sb.String()
}

// RelationStatement -
type RelationStatement struct {
	Relation      token.Token // token.RELATION
	Name          token.Token // token.IDENT
	RelationTypes []RelationTypeStatement
}

// statementNode -
func (ls *RelationStatement) statementNode() {}

// String -
func (ls *RelationStatement) String() string {
	var sb strings.Builder
	sb.WriteString("\t")
	sb.WriteString("relation")
	sb.WriteString(" ")
	sb.WriteString(ls.Name.Literal)
	sb.WriteString(" ")

	for _, rs := range ls.RelationTypes {
		sb.WriteString(rs.String())
		sb.WriteString(" ")
	}

	return sb.String()
}

// RelationTypeStatement -
type RelationTypeStatement struct {
	Sign     token.Token // token.SIGN
	Type     token.Token // token.IDENT
	Relation token.Token // token.IDENT
}

// String -
func (ls *RelationTypeStatement) String() string {
	var sb strings.Builder
	sb.WriteString("@")
	sb.WriteString(ls.Type.Literal)
	if ls.Relation.Literal != "" {
		sb.WriteString("#")
		sb.WriteString(ls.Relation.Literal)
	}
	return sb.String()
}

// IsDirectEntityReference -
func IsDirectEntityReference(s RelationTypeStatement) bool {
	return s.Relation.Literal == ""
}

// Identifier -
type Identifier struct {
	Prefix token.Token
	Idents []token.Token // token.IDENT
}

// expressionNode -
func (ls *Identifier) expressionNode() {}

// String -
func (ls *Identifier) String() string {
	var sb strings.Builder
	if ls.Prefix.Literal != "" {
		sb.WriteString("not")
		sb.WriteString(" ")
	}
	for _, ident := range ls.Idents[:len(ls.Idents)-1] {
		sb.WriteString(ident.Literal)
		sb.WriteString(".")
	}
	sb.WriteString(ls.Idents[len(ls.Idents)-1].Literal)
	return sb.String()
}

// IsPrefix -
func (ls *Identifier) IsPrefix() bool {
	return ls.Prefix.Literal != ""
}

// IsInfix -
func (ls *Identifier) IsInfix() bool {
	return false
}

// GetType -
func (ls *Identifier) GetType() ExpressionType {
	return IDENTIFIER
}

// ActionStatement -
type ActionStatement struct {
	Action              token.Token // token.ACTION
	Name                token.Token // token.IDENT
	ExpressionStatement Statement
}

// statementNode -
func (ls *ActionStatement) statementNode() {}

// String -
func (ls *ActionStatement) String() string {
	var sb strings.Builder
	sb.WriteString("\t")
	sb.WriteString("action")
	sb.WriteString(" ")
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

// String -
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}

// InfixExpression -
type InfixExpression struct {
	Op       token.Token // The operator token, e.g. and, or
	Left     Expression
	Operator Operator
	Right    Expression
}

// expressionNode -
func (ie *InfixExpression) expressionNode() {}

// String -
func (ie *InfixExpression) String() string {
	var sb strings.Builder
	sb.WriteString("(")
	sb.WriteString(ie.Left.String())
	sb.WriteString(" ")
	sb.WriteString(ie.Operator.String())
	sb.WriteString(" ")
	sb.WriteString(ie.Right.String())
	sb.WriteString(")")
	return sb.String()
}

// IsInfix -
func (ie *InfixExpression) IsInfix() bool {
	return true
}

// GetType -
func (ie *InfixExpression) GetType() ExpressionType {
	return INFLIX
}
