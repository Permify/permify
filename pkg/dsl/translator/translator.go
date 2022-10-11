package translator

import (
	`google.golang.org/protobuf/types/known/anypb`
	"strings"

	"github.com/Permify/permify/pkg/dsl/ast"
	`github.com/Permify/permify/pkg/dsl/parser`
	"github.com/Permify/permify/pkg/dsl/schema"
	"github.com/Permify/permify/pkg/errors"
	base `github.com/Permify/permify/pkg/pb/base/v1`
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
func (t *SchemaTranslator) Translate() (sch *base.Schema) {
	var entities []*base.EntityDefinition

	for _, sc := range t.schema.Statements {
		entities = append(entities, t.translateToEntity(sc.(*ast.EntityStatement)))
	}

	return schema.NewSchema(entities...)
}

// translateToEntity -
func (t *SchemaTranslator) translateToEntity(sc *ast.EntityStatement) *base.EntityDefinition {
	var entityDefinition = &base.EntityDefinition{
		Name:   sc.Name.Literal,
		Option: map[string]*anypb.Any{},
	}

	if sc.Option.Literal != "" {
		options := strings.Split(sc.Option.Literal, "|")
		for _, option := range options {
			op := strings.Split(option, ":")
			if len(op) == 2 {
				entityDefinition.Option[op[0]] = &anypb.Any{
					Value: []byte(op[1]),
				}
			}
		}
	}

	// relations
	for _, rs := range sc.RelationStatements {
		relationSt := rs.(*ast.RelationStatement)
		var relationDefinition = &base.RelationDefinition{
			Name:   relationSt.Name.Literal,
			Option: map[string]*anypb.Any{},
		}

		for _, rts := range relationSt.RelationTypes {
			relationTypeSt := rts.(*ast.RelationTypeStatement)
			relationDefinition.Types = append(relationDefinition.Types, &base.RelationType{Name: relationTypeSt.Token.Literal})
		}

		if relationSt.Option.Literal != "" {
			options := strings.Split(relationSt.Option.Literal, "|")
			for _, option := range options {
				op := strings.Split(option, ":")
				if len(op) == 2 {
					relationDefinition.Option[op[0]] = &anypb.Any{
						Value: []byte(op[1]),
					}
				}
			}
		}

		entityDefinition.Relations = append(entityDefinition.Relations, relationDefinition)
	}

	// actions
	for _, as := range sc.ActionStatements {
		st := as.(*ast.ActionStatement)
		var actionDefinition = &base.ActionDefinition{
			Name:  st.Name.Literal,
			Child: parseChild(st.ExpressionStatement.(*ast.ExpressionStatement)),
		}
		entityDefinition.Actions = append(entityDefinition.Actions, actionDefinition)
	}

	return entityDefinition
}

// parseChild -
func parseChild(expression *ast.ExpressionStatement) *base.Child {
	return parseChildren(expression.Expression.(ast.Expression))
}

// parseChildren -
func parseChildren(expression ast.Expression) (children *base.Child) {
	if expression.IsInfix() {
		exp := expression.(*ast.InfixExpression)
		var child = &base.Child{}
		var rewrite = &base.Rewrite{}

		switch exp.Operator {
		case "or":
			rewrite.RewriteOperation = base.Rewrite_UNION
			break
		case "and":
			rewrite.RewriteOperation = base.Rewrite_INTERSECTION
			break
		default:
			rewrite.RewriteOperation = base.Rewrite_INVALID
			break
		}

		var ch []*base.Child
		ch = append(ch, parseChildren(exp.Left))
		ch = append(ch, parseChildren(exp.Right))

		rewrite.Children = ch
		child.Type = &base.Child_Rewrite{Rewrite: rewrite}
		child.GetRewrite().Children = ch
		return child
	} else {
		var child = &base.Child{}
		var s = strings.Split(expression.String(), ".")

		var leaf = &base.Leaf{}
		var exp ast.Expression
		switch expression.Type() {
		case "identifier":
			exp = expression.(*ast.Identifier)
			leaf.Exclusion = false
		case "prefix":
			exp = expression.(*ast.PrefixExpression)
			leaf.Exclusion = true
		default:
			exp = expression.(*ast.Identifier)
			leaf.Exclusion = false
		}

		if len(s) > 1 {
			var tupleToUserSet = &base.TupleToUserSet{}
			tupleToUserSet.Relation = exp.GetValue()
			leaf.Type = &base.Leaf_TupleToUserSet{TupleToUserSet: tupleToUserSet}
		} else {
			var computedUserSet = &base.ComputedUserSet{}
			computedUserSet.Relation = exp.GetValue()
			leaf.Type = &base.Leaf_ComputedUserSet{ComputedUserSet: computedUserSet}
		}

		child.Type = &base.Child_Leaf{Leaf: leaf}
		return child
	}
}

// StringToSchema -
func StringToSchema(configs ...string) (*base.Schema, errors.Error) {
	var err error
	pr, pErr := parser.NewParser(strings.Join(configs, "\n")).Parse()
	if pErr != nil {
		return nil, pErr
	}
	var s *SchemaTranslator
	s, err = NewSchemaTranslator(pr)
	if err != nil {
		return nil, errors.ValidationError.AddParam("schema", err.Error())
	}
	return s.Translate(), nil
}
