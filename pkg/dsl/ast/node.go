package ast

import (
	"strings"

	"github.com/Permify/permify/pkg/dsl/token"
)

type (
	// ExpressionType defines the type of expression.
	ExpressionType string
	// RelationalReferenceType defines the type of relational reference.
	RelationalReferenceType string
	// Operator defines a logical operator.
	Operator string
)

// String returns a string representation of the operator.
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

// Statement defines an interface for a statement node.
type Statement interface {
	Node
	statementNode()
}

// EntityStatement represents a statement that refers to an entity.
type EntityStatement struct {
	Entity             token.Token // token.ENTITY
	Name               token.Token // token.IDENT
	RelationStatements []Statement // Statements that define relationships between entities
	ActionStatements   []Statement // Statements that define actions performed on the entity
}

// statementNode is a dummy method that satisfies the Statement interface.
func (ls *EntityStatement) statementNode() {}

// String returns a string representation of the EntityStatement.
func (ls *EntityStatement) String() string {
	var sb strings.Builder
	sb.WriteString("entity")
	sb.WriteString(" ")
	sb.WriteString(ls.Name.Literal)
	sb.WriteString(" {")
	sb.WriteString("\n")

	// Iterate over the relation statements and add them to the string builder.
	for _, rs := range ls.RelationStatements {
		sb.WriteString(rs.String())
		sb.WriteString("\n")
	}

	sb.WriteString("\n")

	// Iterate over the action statements and add them to the string builder.
	for _, rs := range ls.ActionStatements {
		sb.WriteString(rs.String())
		sb.WriteString("\n")
	}

	sb.WriteString("}")
	sb.WriteString(" ")
	sb.WriteString("\n")

	// Return the final string.
	return sb.String()
}

// RelationStatement represents a statement that defines a relationship between two entities.
type RelationStatement struct {
	Relation      token.Token             // token.RELATION
	Name          token.Token             // token.IDENT
	RelationTypes []RelationTypeStatement // Statements that define the types of the relationship
}

// statementNode is a dummy method that satisfies the Statement interface.
func (ls *RelationStatement) statementNode() {}

// String returns a string representation of the RelationStatement.
func (ls *RelationStatement) String() string {
	var sb strings.Builder
	sb.WriteString("\t")
	sb.WriteString("relation")
	sb.WriteString(" ")
	sb.WriteString(ls.Name.Literal)
	sb.WriteString(" ")

	// Iterate over the relation types and append them to the string builder.
	for _, rs := range ls.RelationTypes {
		sb.WriteString(rs.String())
		sb.WriteString(" ")
	}

	// Return the final string.
	return sb.String()
}

// RelationTypeStatement represents a statement that defines the type of a relationship.
type RelationTypeStatement struct {
	Sign     token.Token // token.SIGN
	Type     token.Token // token.IDENT
	Relation token.Token // token.IDENT
}

// String returns a string representation of the RelationTypeStatement.
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

// IsDirectEntityReference returns true if the RelationTypeStatement is a direct entity reference.
func IsDirectEntityReference(s RelationTypeStatement) bool {
	return s.Relation.Literal == ""
}

// Identifier represents an expression that identifies an entity, action or relation
type Identifier struct {
	Prefix token.Token   // Prefix is a token that negates the identifier
	Idents []token.Token // Idents is a slice of tokens that make up the identifier
}

// expressionNode is a marker method to differentiate Expression and Statement interfaces
func (ls *Identifier) expressionNode() {}

// String returns a string representation of the identifier expression
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

// IsPrefix returns true if the identifier has a negating prefix
func (ls *Identifier) IsPrefix() bool {
	return ls.Prefix.Literal != ""
}

// IsInfix returns false since an identifier is not an infix expression
func (ls *Identifier) IsInfix() bool {
	return false
}

// GetType returns the type of the expression which is Identifier
func (ls *Identifier) GetType() ExpressionType {
	return IDENTIFIER
}

// ActionStatement represents an action statement, which consists of an action name and an optional expression statement.
// It implements the Statement interface.
type ActionStatement struct {
	Action              token.Token // token.ACTION
	Name                token.Token // token.IDENT
	ExpressionStatement Statement
}

// statementNode is a marker method used to implement the Statement interface.
func (ls *ActionStatement) statementNode() {}

// String returns a string representation of the action statement.
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

// ExpressionStatement struct represents an expression statement
type ExpressionStatement struct {
	Expression Expression
}

// statementNode function is needed to mark the struct as a Statement node
func (es *ExpressionStatement) statementNode() {}

// String function returns a string representation of the ExpressionStatement
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	// If there is no expression, return an empty string
	return ""
}

// InfixExpression represents an expression with an operator between two sub-expressions.
type InfixExpression struct {
	Op       token.Token // The operator token, e.g. and, or.
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
