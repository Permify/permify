package ast

import (
	"bytes"
	"strings"

	"github.com/pkg/errors"

	"github.com/Permify/permify/pkg/dsl/token"
	base `github.com/Permify/permify/pkg/pb/base/v1`
)

type (
	ExpressionType string
	Operator       string
)

// String -
func (o Operator) String() string {
	return string(o)
}

const (
	IDENTIFIER ExpressionType = "identifier"
	PREFIX     ExpressionType = "prefix"
	INFLIX     ExpressionType = "inflix"

	AND Operator = "and"
	OR  Operator = "or"
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
	GetType() ExpressionType
	GetValue() string
}

// Statement -
type Statement interface {
	Node
	statementNode()
}

// Schema -
type Schema struct {
	Statements []Statement

	entityReferences   map[string]struct{}
	relationReferences map[string][]RelationTypeStatement
}

// ValidateReferences -
func (sch *Schema) ValidateReferences() error {
	for _, st := range sch.relationReferences {
		entityReferenceCount := 0
		for _, s := range st {
			if s.IsEntityReference() {
				if !sch.IsEntityReferenceExist(s.Token.Literal) {
					return errors.New(base.ErrorCode_relation_reference_not_found_in_entity_references.String())
				}
				entityReferenceCount++
			}
			if entityReferenceCount > 1 {
				return errors.New(base.ErrorCode_relation_reference_must_have_one_entity_reference.String())
			}
		}
	}
	return nil
}

// SetEntityReferences -
func (sch *Schema) SetEntityReferences(r map[string]struct{}) {
	sch.entityReferences = r
}

// SetRelationReferences -
func (sch *Schema) SetRelationReferences(r map[string][]RelationTypeStatement) {
	sch.relationReferences = r
}

// IsEntityReferenceExist -
func (sch *Schema) IsEntityReferenceExist(name string) bool {
	if _, ok := sch.entityReferences[name]; ok {
		return ok
	}
	return false
}

// IsRelationReferenceExist -
func (sch *Schema) IsRelationReferenceExist(name string) bool {
	if _, ok := sch.relationReferences[name]; ok {
		return true
	}
	return false
}

// GetRelationReferenceIfExist -
func (sch *Schema) GetRelationReferenceIfExist(name string) ([]RelationTypeStatement, bool) {
	if _, ok := sch.relationReferences[name]; ok {
		return sch.relationReferences[name], true
	}
	return nil, false
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

	if ls.Option.Literal != "" {
		sb.WriteString("`")
		sb.WriteString(ls.Option.Literal)
		sb.WriteString("`")
	}

	sb.WriteString("\n")
	return sb.String()
}

// RelationStatement -
type RelationStatement struct {
	Token         token.Token // token.RELATION
	Name          token.Token // token.IDENT
	RelationTypes []Statement
	Option        token.Token // token.OPTION
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

	for _, rs := range ls.RelationTypes {
		sb.WriteString(rs.String())
		sb.WriteString(" ")
	}

	sb.WriteString(" ")

	if ls.Option.Literal != "" {
		sb.WriteString("`")
		sb.WriteString(ls.Option.Literal)
		sb.WriteString("`")
	}

	return sb.String()
}

// RelationTypeStatement -
type RelationTypeStatement struct {
	Sign  token.Token // token.sign
	Token token.Token // token.IDENT
}

// statementNode -
func (ls *RelationTypeStatement) statementNode() {}

// TokenLiteral -
func (ls *RelationTypeStatement) TokenLiteral() string {
	return ls.Token.Literal
}

// String -
func (ls *RelationTypeStatement) String() string {
	var sb strings.Builder
	sb.WriteString(ls.Sign.Literal)
	sb.WriteString(ls.Token.Literal)
	return sb.String()
}

// IsEntityReference -
func (ls *RelationTypeStatement) IsEntityReference() bool {
	if !strings.Contains(ls.Token.Literal, "#") {
		return true
	}
	return false
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

// GetType -
func (ls *Identifier) GetType() ExpressionType {
	return IDENTIFIER
}

// GetValue -
func (ls *Identifier) GetValue() string {
	return ls.Value
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
	Operator Operator
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
	sb.WriteString(" " + ie.Operator.String())
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

// GetValue -
func (ie *InfixExpression) GetValue() string {
	return ie.Token.Literal
}

// PrefixExpression -
type PrefixExpression struct {
	Token    token.Token // not
	Operator string
	Value    string
}

// TokenLiteral -
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }

// String -
func (pe *PrefixExpression) String() string {
	var sb bytes.Buffer
	sb.WriteString(pe.Operator)
	sb.WriteString(" ")
	sb.WriteString(pe.Value)
	return sb.String()
}

// expressionNode -
func (pe *PrefixExpression) expressionNode() {}

// IsInfix -
func (pe *PrefixExpression) IsInfix() bool {
	return false
}

// GetType -
func (pe *PrefixExpression) GetType() ExpressionType {
	return PREFIX
}

// GetValue -
func (pe *PrefixExpression) GetValue() string {
	return pe.Value
}

// RelationTypeStatements -
type RelationTypeStatements []RelationTypeStatement

// GetEntityReference -
func (st RelationTypeStatements) GetEntityReference() string {
	for _, rt := range st {
		if rt.IsEntityReference() {
			return rt.TokenLiteral()
		}
	}
	return ""
}
