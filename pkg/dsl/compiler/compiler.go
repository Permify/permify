package compiler

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Permify/permify/pkg/dsl/ast"
	"github.com/Permify/permify/pkg/dsl/parser"
	"github.com/Permify/permify/pkg/dsl/schema"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/tuple"
)

// Compiler -
type Compiler struct {
	schema                     *ast.Schema
	withoutReferenceValidation bool
}

// NewCompiler -
func NewCompiler(w bool, sch *ast.Schema) *Compiler {
	return &Compiler{
		withoutReferenceValidation: w,
		schema:                     sch,
	}
}

// Compile -
func (t *Compiler) Compile() (sch *base.Schema, err error) {
	var entities []*base.EntityDefinition

	if !t.withoutReferenceValidation {
		err = t.schema.ValidateReferences()
		if err != nil {
			return nil, err
		}
	}

	for _, sc := range t.schema.Statements {
		var en *base.EntityDefinition
		es, ok := sc.(*ast.EntityStatement)
		if !ok {
			return nil, errors.New(base.ErrorCode_schema_compile.String())
		}
		en, err = t.compile(es)
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
		Name:      sc.Name.Literal,
		Option:    map[string]string{},
		Relations: map[string]*base.RelationDefinition{},
		Actions:   map[string]*base.ActionDefinition{},
	}

	if sc.Option.Literal != "" {
		options := strings.Split(sc.Option.Literal, "|")
		for _, option := range options {
			op := strings.Split(option, ":")
			if len(op) == 2 {
				entityDefinition.Option[op[0]] = op[1]
			}
		}
	}

	// relations
	for _, rs := range sc.RelationStatements {
		relationSt := rs.(*ast.RelationStatement)
		relationDefinition := &base.RelationDefinition{
			Name:               relationSt.Name.Literal,
			Option:             map[string]string{},
			RelationReferences: []*base.RelationReference{},
		}

		for _, rts := range relationSt.RelationTypes {
			relationTypeSt := rts.(*ast.RelationTypeStatement)
			if relationTypeSt.IsEntityReference() {
				relationDefinition.EntityReference = &base.RelationReference{Name: relationTypeSt.Token.Literal}
			}
			relationDefinition.RelationReferences = append(relationDefinition.RelationReferences, &base.RelationReference{Name: relationTypeSt.Token.Literal})
		}

		if relationSt.Option.Literal != "" {
			options := strings.Split(relationSt.Option.Literal, "|")
			for _, option := range options {
				op := strings.Split(option, ":")
				if len(op) == 2 {
					relationDefinition.Option[op[0]] = op[1]
				}
			}
		}

		entityDefinition.Relations[relationDefinition.GetName()] = relationDefinition
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
		entityDefinition.Actions[actionDefinition.GetName()] = actionDefinition
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
			if !t.withoutReferenceValidation {
				exist := t.schema.IsRelationReferenceExist(fmt.Sprintf("%v#%v", entityName, s[0]))
				if !exist {
					return nil, errors.New(base.ErrorCode_undefined_relation_reference.String())
				}
			}
			leaf.Type = &base.Leaf_ComputedUserSet{ComputedUserSet: computedUserSet}
		} else if len(s) == 2 {
			tupleToUserSet := &base.TupleToUserSet{}
			tupleToUserSet.Relation = exp.GetValue()

			if !t.withoutReferenceValidation {
				value, exist := t.schema.GetRelationReferenceIfExist(fmt.Sprintf("%v#%v", entityName, s[0]))
				if !exist {
					return nil, errors.New(base.ErrorCode_undefined_relation_reference.String())
				}

				exist = t.schema.IsRelationReferenceExist(fmt.Sprintf("%v#%v", ast.RelationTypeStatements(value).GetEntityReference(), s[1]))
				if !exist {
					return nil, errors.New(base.ErrorCode_undefined_relation_reference.String())
				}
			}

			leaf.Type = &base.Leaf_TupleToUserSet{TupleToUserSet: tupleToUserSet}
		} else {
			return nil, errors.New(base.ErrorCode_not_supported_relation_walk.String())
		}

		child.Type = &base.Child_Leaf{Leaf: leaf}
		return child, nil
	}
}

// NewSchema -
func NewSchema(schema ...string) (*base.Schema, error) {
	sch, err := parser.NewParser(strings.Join(schema, "\n")).Parse()
	if err != nil {
		return nil, err
	}
	var s *base.Schema
	s, err = NewCompiler(false, sch).Compile()
	if err != nil {
		return nil, err
	}
	return s, err
}

// NewSchemaWithoutReferenceValidation -
func NewSchemaWithoutReferenceValidation(schema ...string) (*base.Schema, error) {
	sch, err := parser.NewParser(strings.Join(schema, "\n")).Parse()
	if err != nil {
		return nil, err
	}
	var s *base.Schema
	s, err = NewCompiler(true, sch).Compile()
	if err != nil {
		return nil, err
	}
	return s, err
}
