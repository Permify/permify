package parser

import (
	"strings"

	"github.com/Permify/permify/pkg/dsl/ast"
	"github.com/Permify/permify/pkg/dsl/schema"
	"github.com/Permify/permify/pkg/errors"
)

// SchemaTranslator -
type SchemaTranslator struct {
	schema *ast.Schema
}

// NewSchemaTranslator -
func NewSchemaTranslator(sch *ast.Schema) (*SchemaTranslator, errors.Error) {
	return &SchemaTranslator{
		schema: sch,
	}, nil
}

// Translate -
func (t *SchemaTranslator) Translate() (sch schema.Schema) {
	var entities []schema.Entity

	for _, sc := range t.schema.Statements {
		entities = append(entities, t.translateToEntity(sc.(*ast.EntityStatement)))
	}

	return schema.NewSchema(entities...)
}

// translateToEntity -
func (t *SchemaTranslator) translateToEntity(sc *ast.EntityStatement) (entity schema.Entity) {
	entity.Name = sc.Name.Literal

	// relations
	for _, rs := range sc.RelationStatements {
		relationSt := rs.(*ast.RelationStatement)
		var relation schema.Relation
		relation.Name = relationSt.Name.Literal

		for _, rts := range relationSt.RelationTypes {
			relationTypeSt := rts.(*ast.RelationTypeStatement)
			relation.Types = append(relation.Types, relationTypeSt.Token.Literal)
		}

		entity.Relations = append(entity.Relations, relation)
	}

	// actions
	for _, as := range sc.ActionStatements {
		st := as.(*ast.ActionStatement)
		var action schema.Action
		action.Name = st.Name.Literal
		action.Child = parseChild(st.ExpressionStatement.(*ast.ExpressionStatement))
		entity.Actions = append(entity.Actions, action)
	}

	return
}

// parseChild -
func parseChild(expression *ast.ExpressionStatement) (re schema.Child) {
	return parseChildren(expression.Expression.(ast.Expression))
}

// parseChildren -
func parseChildren(expression ast.Expression) (children schema.Child) {
	if expression.IsInfix() {
		exp := expression.(*ast.InfixExpression)
		var child schema.Rewrite

		switch exp.Operator {
		case "or":
			child.Type = schema.Union
			break
		case "and":
			child.Type = schema.Intersection
			break
		default:
			child.Type = schema.Union
			break
		}

		var ch []schema.Child
		ch = append(ch, parseChildren(exp.Left))
		ch = append(ch, parseChildren(exp.Right))

		child.Children = ch
		return child
	} else {

		var child schema.Leaf
		var s []string

		switch expression.Type() {
		case "identifier":
			exp := expression.(*ast.Identifier)
			s = strings.Split(expression.String(), ".")
			child.Exclusion = false
			child.Value = exp.Value
		case "prefix":
			exp := expression.(*ast.PrefixExpression)
			s = strings.Split(expression.String(), ".")
			child.Value = exp.Value
			child.Exclusion = true
		default:
			exp := expression.(*ast.Identifier)
			s = strings.Split(expression.String(), ".")
			child.Exclusion = false
			child.Value = exp.Value
		}

		if len(s) > 1 {
			child.Type = schema.TupleToUserSetType
		} else {
			child.Type = schema.ComputedUserSetType
		}

		return child
	}
}
