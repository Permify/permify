package ast

import (
	"strings"

	"github.com/Permify/permify/pkg/dsl/token"
)

type (
	StatementType string
)

const (
	PERMISSION_STATEMENT     StatementType = "permission"
	RELATION_STATEMENT       StatementType = "relation"
	ATTRIBUTE_STATEMENT      StatementType = "attribute"
	ENTITY_STATEMENT         StatementType = "entity"
	RULE_STATEMENT           StatementType = "rule"
	EXPRESSION_STATEMENT     StatementType = "expression"
	RELATION_TYPE_STATEMENT  StatementType = "relation_type"
	ATTRIBUTE_TYPE_STATEMENT StatementType = "attribute_type"
)

// Statement defines an interface for a statement node.
type Statement interface {
	Node
	statementNode()
	GetName() string
	StatementType() StatementType
}

// EntityStatement represents a statement that refers to an entity.
type EntityStatement struct {
	Entity               token.Token // token.ENTITY
	Name                 token.Token // token.IDENT
	RelationStatements   []Statement // Statements that define relationships between entities
	AttributeStatements  []Statement // Statements that define attributes of the entity
	PermissionStatements []Statement // Statements that define permissions performed on the entity
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

	// Iterate over the attribute statements and add them to the string builder.
	for _, as := range ls.AttributeStatements {
		sb.WriteString(as.String())
		sb.WriteString("\n")
	}

	sb.WriteString("\n")

	// Iterate over the permission statements and add them to the string builder.
	for _, ps := range ls.PermissionStatements {
		sb.WriteString(ps.String())
		sb.WriteString("\n")
	}

	sb.WriteString("}")
	sb.WriteString(" ")
	sb.WriteString("\n")

	// Return the final string.
	return sb.String()
}

func (ls *EntityStatement) GetName() string {
	return ls.Name.Literal
}

func (ls *EntityStatement) StatementType() StatementType {
	return ENTITY_STATEMENT
}

// AttributeStatement represents a statement that defines an attribute of an entity.
type AttributeStatement struct {
	Attribute     token.Token // token.ATTRIBUTE
	Name          token.Token // token.IDENT
	AttributeType AttributeTypeStatement
}

// statementNode is a dummy method that satisfies the Statement interface.
func (as *AttributeStatement) statementNode() {}

// String returns a string representation of the AttributeStatement.
func (as *AttributeStatement) String() string {
	var sb strings.Builder
	sb.WriteString("\t")
	sb.WriteString("attribute")
	sb.WriteString(" ")
	sb.WriteString(as.Name.Literal)
	sb.WriteString(" ")
	sb.WriteString(as.AttributeType.String())

	// Return the final string.
	return sb.String()
}

func (as *AttributeStatement) GetName() string {
	return as.Name.Literal
}

func (as *AttributeStatement) StatementType() StatementType {
	return ATTRIBUTE_STATEMENT
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

func (ls *RelationStatement) GetName() string {
	return ls.Name.Literal
}

func (ls *RelationStatement) StatementType() StatementType {
	return RELATION_STATEMENT
}

// AttributeTypeStatement represents a statement that defines the type of a relationship.
type AttributeTypeStatement struct {
	Type    token.Token // token.IDENT
	IsArray bool
}

// String returns a string representation of the RelationTypeStatement.
func (as *AttributeTypeStatement) String() string {
	var sb strings.Builder
	sb.WriteString(as.Type.Literal)
	if as.IsArray {
		sb.WriteString("[]")
	}
	return sb.String()
}

func (as *AttributeTypeStatement) GetName() string {
	return ""
}

func (as *AttributeTypeStatement) StatementType() StatementType {
	return ATTRIBUTE_TYPE_STATEMENT
}

func (as *AttributeTypeStatement) statementNode() {}

// RelationTypeStatement represents a statement that defines the type of relationship.
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

func (ls *RelationTypeStatement) GetName() string {
	return ""
}

func (ls *RelationTypeStatement) StatementType() StatementType {
	return RELATION_TYPE_STATEMENT
}

func (ls *RelationTypeStatement) statementNode() {}

// IsDirectEntityReference returns true if the RelationTypeStatement is a direct entity reference.
func IsDirectEntityReference(s RelationTypeStatement) bool {
	return s.Relation.Literal == ""
}

// PermissionStatement represents an permission statement, which consists of an permission name and an optional expression statement.
// It implements the Statement interface.
type PermissionStatement struct {
	Permission          token.Token // token.PERMISSION
	Name                token.Token // token.IDENT
	ExpressionStatement Statement
}

// statementNode is a marker method used to implement the Statement interface.
func (ls *PermissionStatement) statementNode() {}

// String returns a string representation of the permission statement.
func (ls *PermissionStatement) String() string {
	var sb strings.Builder
	sb.WriteString("\t")
	sb.WriteString("permission")
	sb.WriteString(" ")
	sb.WriteString(ls.Name.Literal)
	sb.WriteString(" = ")
	if ls.ExpressionStatement != nil {
		sb.WriteString(ls.ExpressionStatement.String())
	}
	return sb.String()
}

func (ls *PermissionStatement) GetName() string {
	return ls.Name.Literal
}

func (ls *PermissionStatement) StatementType() StatementType {
	return PERMISSION_STATEMENT
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

func (es *ExpressionStatement) GetName() string {
	return ""
}

func (es *ExpressionStatement) StatementType() StatementType {
	return EXPRESSION_STATEMENT
}

// RuleStatement represents a rule statement, which consists of a rule name, a list of parameters and a body.
type RuleStatement struct {
	Rule       token.Token // token.RULE
	Name       token.Token // token.IDENT
	Arguments  map[token.Token]AttributeTypeStatement
	Expression string
}

// statementNode is a marker method used to implement the Statement interface.
func (rs *RuleStatement) statementNode() {}

// String returns a string representation of the permission statement.
func (rs *RuleStatement) String() string {
	var sb strings.Builder
	sb.WriteString("rule")
	sb.WriteString(" ")
	sb.WriteString(rs.Name.Literal)
	sb.WriteString("(")

	var literals []string
	for param, typ := range rs.Arguments {
		var pb strings.Builder
		pb.WriteString(param.Literal)
		pb.WriteString(" ")
		pb.WriteString(typ.Type.Literal)
		if typ.IsArray {
			pb.WriteString("[]")
		}
		literals = append(literals, pb.String())
	}

	sb.WriteString(strings.Join(literals, ", "))

	sb.WriteString(")")
	sb.WriteString(" ")
	sb.WriteString("{")

	sb.WriteString("\n")
	sb.WriteString("\t")
	sb.WriteString(rs.Expression)
	sb.WriteString("\n")

	sb.WriteString("}")
	return sb.String()
}

func (rs *RuleStatement) GetName() string {
	return rs.Name.Literal
}

func (rs *RuleStatement) StatementType() StatementType {
	return RULE_STATEMENT
}
