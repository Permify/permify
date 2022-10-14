package compiler

import (
	"errors"
	"fmt"
	"strings"

	"google.golang.org/protobuf/types/known/anypb"

	"github.com/Permify/permify/pkg/dsl/ast"
	"github.com/Permify/permify/pkg/dsl/parser"
	"github.com/Permify/permify/pkg/dsl/schema"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/tuple"
)

// Compiler -
type Compiler struct {
	schema *ast.Schema
}

// NewCompiler -
func NewCompiler(sch *ast.Schema) *Compiler {
	return &Compiler{
		schema: sch,
	}
}

// Compile -
func (t *Compiler) Compile() (sch *base.Schema, err error) {
	var entities []*base.EntityDefinition

	err = t.schema.ValidateReferences()
	if err != nil {
		return nil, err
	}

	for _, sc := range t.schema.Statements {
		var en *base.EntityDefinition
		en, err = t.compile(sc.(*ast.EntityStatement))
		if err != nil {
			return nil, err
		}
		entities = append(entities, en)
	}

	return schema.NewSchema(entities...), err
}

// translateToEntity -
func (t *Compiler) compile(sc *ast.EntityStatement) (*base.EntityDefinition, error) {
	entityDefinition := &base.EntityDefinition{
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
		relationDefinition := &base.RelationDefinition{
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
		ch, err := t.parseChild(entityDefinition.GetName(), st.ExpressionStatement.(*ast.ExpressionStatement))
		if err != nil {
			return nil, err
		}
		actionDefinition := &base.ActionDefinition{
			Name:  st.Name.Literal,
			Child: ch,
		}
		entityDefinition.Actions = append(entityDefinition.Actions, actionDefinition)
	}

	return entityDefinition, nil
}

// parseChild -
func (t *Compiler) parseChild(entityName string, expression *ast.ExpressionStatement) (*base.Child, error) {
	return t.parseChildren(entityName, expression.Expression.(ast.Expression))
}

// parseChildren -
func (t *Compiler) parseChildren(entityName string, expression ast.Expression) (children *base.Child, err error) {
	if expression.IsInfix() {
		exp := expression.(*ast.InfixExpression)
		child := &base.Child{}
		rewrite := &base.Rewrite{}

		switch exp.Operator {
		case ast.OR:
			rewrite.RewriteOperation = base.Rewrite_UNION
			break
		case ast.AND:
			rewrite.RewriteOperation = base.Rewrite_INTERSECTION
			break
		default:
			rewrite.RewriteOperation = base.Rewrite_INVALID
			break
		}

		var ch []*base.Child

		var leftChild *base.Child
		leftChild, err = t.parseChildren(entityName, exp.Left)
		if err != nil {
			return nil, err
		}

		var rightChild *base.Child
		rightChild, err = t.parseChildren(entityName, exp.Right)
		if err != nil {
			return nil, err
		}

		ch = append(ch, []*base.Child{leftChild, rightChild}...)

		rewrite.Children = ch
		child.Type = &base.Child_Rewrite{Rewrite: rewrite}
		child.GetRewrite().Children = ch
		return child, nil
	} else {
		child := &base.Child{}

		leaf := &base.Leaf{}
		var exp ast.Expression
		switch expression.GetType() {
		case ast.IDENTIFIER:
			exp = expression.(*ast.Identifier)
			leaf.Exclusion = false
		case ast.PREFIX:
			exp = expression.(*ast.PrefixExpression)
			leaf.Exclusion = true
		default:
			exp = expression.(*ast.Identifier)
			leaf.Exclusion = false
		}

		s := strings.Split(expression.GetValue(), tuple.SEPARATOR)

		if len(s) == 1 {
			computedUserSet := &base.ComputedUserSet{}
			computedUserSet.Relation = exp.GetValue()
			exist := t.schema.IsRelationReferenceExist(fmt.Sprintf("%v#%v", entityName, s[0]))
			if !exist {
				return nil, errors.New(base.ErrorCode_undefined_relation_reference.String())
			}
			leaf.Type = &base.Leaf_ComputedUserSet{ComputedUserSet: computedUserSet}
		} else if len(s) == 2 {
			tupleToUserSet := &base.TupleToUserSet{}
			tupleToUserSet.Relation = exp.GetValue()

			value, exist := t.schema.GetRelationReferenceIfExist(fmt.Sprintf("%v#%v", entityName, s[0]))
			if !exist {
				return nil, errors.New(base.ErrorCode_undefined_relation_reference.String())
			}

			exist = t.schema.IsRelationReferenceExist(fmt.Sprintf("%v#%v", ast.RelationTypeStatements(value).GetEntityReference(), s[1]))
			if !exist {
				return nil, errors.New(base.ErrorCode_undefined_relation_reference.String())
			}

			leaf.Type = &base.Leaf_TupleToUserSet{TupleToUserSet: tupleToUserSet}
		} else {
			return nil, errors.New(base.ErrorCode_not_supported_relation_walk.String())
		}

		child.Type = &base.Child_Leaf{Leaf: leaf}
		return child, nil
	}
}

// StringToSchema -
func StringToSchema(configs ...string) (*base.Schema, error) {
	sch, err := parser.NewParser(strings.Join(configs, "\n")).Parse()
	if err != nil {
		return nil, err
	}
	var s *base.Schema
	s, err = NewCompiler(sch).Compile()
	if err != nil {
		return nil, err
	}
	return s, err
}
