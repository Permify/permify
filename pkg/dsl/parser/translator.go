package parser

import (
	"strings"

	"github.com/Permify/permify/pkg/dsl/ast"
	"github.com/Permify/permify/pkg/dsl/schema"
)

// SchemaTranslator -
type SchemaTranslator struct {
	schema *ast.Schema
}

// NewSchemaTranslator -
func NewSchemaTranslator(sch *ast.Schema) (*SchemaTranslator, error) {
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

	if sc.Option.Literal != "" {
		entity.EntityOption = parseEntityOption(sc.Option.Literal)
	}

	if entity.EntityOption.Table == "" {
		entity.EntityOption.Table = "-"
	}

	if entity.EntityOption.Identifier == "" {
		entity.EntityOption.Identifier = "id"
	}

	// relations
	for _, rs := range sc.RelationStatements {
		relationSt := rs.(*ast.RelationStatement)
		var relation schema.Relation
		relation.Name = relationSt.Name.Literal
		relation.Type = relationSt.Type.Literal

		if sc.Option.Literal != "" {
			relation.RelationOption = parseRelationOption(relationSt.Option.Literal)
		}

		if relation.RelationOption.Rel == "" {
			relation.RelationOption.Rel = schema.Custom
		}

		if relation.RelationOption.Rel == schema.ManyToMany {
			if relation.RelationOption.Table == "" {
				relation.RelationOption.Table = relation.Name
			}
			if len(relation.RelationOption.Cols) < 2 {
				relation.RelationOption.Cols = append(relation.RelationOption.Cols, entity.Name+"_id", relation.Name+"_id")
			}
		}

		if relation.RelationOption.Rel == schema.BelongsTo {
			if len(relation.RelationOption.Cols) == 0 {
				relation.RelationOption.Cols = append(relation.RelationOption.Cols, relation.Name+"_id")
			}
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

// parseEntityOption -
func parseEntityOption(str string) (opt schema.EntityOption) {
	split := strings.Split(str, "|")
	for _, s := range split {
		spt := strings.Split(s, ":")
		if len(spt) < 2 {
			break
		}
		switch spt[0] {
		case "table":
			opt.Table = spt[1]
			break
		case "identifier":
			opt.Identifier = spt[1]
			break
		default:
			break
		}
	}
	return
}

// parseRelationOption -
func parseRelationOption(str string) (opt schema.RelationOption) {
	split := strings.Split(str, "|")

	for _, s := range split {
		spt := strings.Split(s, ":")
		if len(spt) < 2 {
			break
		}
		switch spt[0] {
		case "rel":
			opt.Rel = schema.RelationType(spt[1])
			break
		case "table":
			opt.Table = spt[1]
			break
		case "cols":
			opt.Cols = strings.Split(spt[1], ",")
			break
		default:
			break
		}
	}
	return
}
