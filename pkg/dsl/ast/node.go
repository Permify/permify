package ast

import (
	"strings"

	"github.com/Permify/permify/pkg/dsl/token"
)

type (
	// ExpressionType defines the type of expression.
	ExpressionType string
	// Operator defines a logical operator.
	Operator string
)

// String returns a string representation of the operator.
func (o Operator) String() string {
	return string(o)
}

const (
	IDENTIFIER ExpressionType = "identifier"
	CALL       ExpressionType = "call"
	INFLIX     ExpressionType = "inflix"

	AND Operator = "and"
	OR  Operator = "or"
	NOT Operator = "not"
)

// Node defines an interface for a tree node.
type Node interface {
	String() string
}

// Expression defines an interface for an expression node.
type Expression interface {
	Node
	expressionNode()
	IsInfix() bool
	GetType() ExpressionType
}

type Call struct {
	Name      token.Token  // Rule Name token
	Arguments []Identifier // Idents is a slice of tokens that make up the identifier
}

// expressionNode is a marker method to differentiate Expression and Statement interfaces
func (ls *Call) expressionNode() {}

// String returns a string representation of the identifier expression
func (ls *Call) String() string {
	var sb strings.Builder
	sb.WriteString(ls.Name.Literal)
	sb.WriteString("(")
	if len(ls.Arguments) > 0 {
		for _, ident := range ls.Arguments[:len(ls.Arguments)-1] {
			sb.WriteString(ident.String())
			sb.WriteString(",")
			sb.WriteString(" ")
		}
		sb.WriteString(ls.Arguments[len(ls.Arguments)-1].String())
	}
	sb.WriteString(")")
	return sb.String()
}

// IsInfix returns false since an identifier is not an infix expression
func (ls *Call) IsInfix() bool {
	return false
}

// GetType returns the type of the expression which is Identifier
func (ls *Call) GetType() ExpressionType {
	return CALL
}

// Identifier represents an expression that identifies an entity, permission or relation
type Identifier struct {
	Idents []token.Token // Idents is a slice of tokens that make up the identifier
}

// expressionNode is a marker method to differentiate Expression and Statement interfaces
func (ls *Identifier) expressionNode() {}

// String returns a string representation of the identifier expression
func (ls *Identifier) String() string {
	var sb strings.Builder
	for _, ident := range ls.Idents[:len(ls.Idents)-1] {
		sb.WriteString(ident.Literal)
		sb.WriteString(".")
	}
	sb.WriteString(ls.Idents[len(ls.Idents)-1].Literal)
	return sb.String()
}

// IsInfix returns false since an identifier is not an infix expression
func (ls *Identifier) IsInfix() bool {
	return false
}

// GetType returns the type of the expression which is Identifier
func (ls *Identifier) GetType() ExpressionType {
	return IDENTIFIER
}

// InfixExpression represents an expression with an operator between two sub-expressions.
type InfixExpression struct {
	Op       token.Token // The operator token, e.g. and, or, not.
	Left     Expression  // The left-hand side sub-expression.
	Operator Operator    // The operator as a string.
	Right    Expression  // The right-hand side sub-expression.
}

// expressionNode function on InfixExpression.
func (ie *InfixExpression) expressionNode() {}

// String returns the string representation of the infix expression.
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

// IsInfix returns true because it's an infix expression.
func (ie *InfixExpression) IsInfix() bool {
	return true
}

// GetType returns the type of the expression, which is infix.
func (ie *InfixExpression) GetType() ExpressionType {
	return INFLIX
}
