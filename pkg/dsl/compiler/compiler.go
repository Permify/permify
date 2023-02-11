package compiler

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Permify/permify/pkg/dsl/ast"
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
func (t *Compiler) Compile() (sch []*base.EntityDefinition, err error) {
	if !t.withoutReferenceValidation {
		err = t.schema.ValidateReferences()
		if err != nil {
			return nil, err
		}
	}

	entities := make([]*base.EntityDefinition, 0, len(t.schema.Statements))
	for _, sc := range t.schema.Statements {
		var en *base.EntityDefinition
		es, ok := sc.(*ast.EntityStatement)
		if !ok {
			return nil, errors.New(base.ErrorCode_ERROR_CODE_SCHEMA_COMPILE.String())
		}
		en, err = t.compile(es)
		if err != nil {
			return nil, err
		}
		entities = append(entities, en)
	}

	return entities, err
}

// translateToEntity -
func (t *Compiler) compile(sc *ast.EntityStatement) (*base.EntityDefinition, error) {
	entityDefinition := &base.EntityDefinition{
		Name:       sc.Name.Literal,
		Option:     map[string]string{},
		Relations:  map[string]*base.RelationDefinition{},
		Actions:    map[string]*base.ActionDefinition{},
		References: map[string]base.EntityDefinition_RelationalReference{},
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
		relationSt, okRs := rs.(*ast.RelationStatement)
		if !okRs {
			return nil, errors.New(base.ErrorCode_ERROR_CODE_SCHEMA_COMPILE.String())
		}
		relationDefinition := &base.RelationDefinition{
			Name:               relationSt.Name.Literal,
			Option:             map[string]string{},
			RelationReferences: []*base.RelationReference{},
		}

		for _, rts := range relationSt.RelationTypes {
			relationTypeSt, okRt := rts.(*ast.RelationTypeStatement)
			if !okRt {
				return nil, errors.New(base.ErrorCode_ERROR_CODE_SCHEMA_COMPILE.String())
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
		entityDefinition.References[relationDefinition.GetName()] = base.EntityDefinition_RELATIONAL_REFERENCE_RELATION
	}

	// actions
	for _, as := range sc.ActionStatements {
		st, okAs := as.(*ast.ActionStatement)
		if !okAs {
			return nil, errors.New(base.ErrorCode_ERROR_CODE_SCHEMA_COMPILE.String())
		}
		ch, err := t.compileExpressionStatement(entityDefinition.GetName(), st.ExpressionStatement.(*ast.ExpressionStatement))
		if err != nil {
			return nil, err
		}
		actionDefinition := &base.ActionDefinition{
			Name:  st.Name.Literal,
			Child: ch,
		}
		entityDefinition.Actions[actionDefinition.GetName()] = actionDefinition
		entityDefinition.References[actionDefinition.GetName()] = base.EntityDefinition_RELATIONAL_REFERENCE_ACTION
	}

	return entityDefinition, nil
}

// compileExpressionStatement -
func (t *Compiler) compileExpressionStatement(entityName string, expression *ast.ExpressionStatement) (*base.Child, error) {
	return t.compileChildren(entityName, expression.Expression)
}

// compileChildren -
func (t *Compiler) compileChildren(entityName string, expression ast.Expression) (children *base.Child, err error) {
	if expression.IsInfix() {
		return t.compileRewrite(entityName, expression.(*ast.InfixExpression))
	}
	return t.compileLeaf(entityName, expression)
}

// compileRewrite -
func (t *Compiler) compileRewrite(entityName string, exp *ast.InfixExpression) (children *base.Child, err error) {
	child := &base.Child{}
	rewrite := &base.Rewrite{}

	switch exp.Operator {
	case ast.OR:
		rewrite.RewriteOperation = base.Rewrite_OPERATION_UNION
	case ast.AND:
		rewrite.RewriteOperation = base.Rewrite_OPERATION_INTERSECTION
	default:
		rewrite.RewriteOperation = base.Rewrite_OPERATION_UNSPECIFIED
	}

	var ch []*base.Child

	var leftChild *base.Child
	leftChild, err = t.compileChildren(entityName, exp.Left)
	if err != nil {
		return nil, err
	}

	var rightChild *base.Child
	rightChild, err = t.compileChildren(entityName, exp.Right)
	if err != nil {
		return nil, err
	}

	ch = append(ch, []*base.Child{leftChild, rightChild}...)

	rewrite.Children = ch
	child.Type = &base.Child_Rewrite{Rewrite: rewrite}
	child.GetRewrite().Children = ch
	return child, nil
}

// compileLeaf -
func (t *Compiler) compileLeaf(entityName string, expression ast.Expression) (children *base.Child, err error) {
	child := &base.Child{}
	leaf := &base.Leaf{}
	switch expression.GetType() {
	case ast.IDENTIFIER:
		s := strings.Split(expression.GetValue(), tuple.SEPARATOR)
		if len(s) == 1 {
			if !t.withoutReferenceValidation {
				_, exist := t.schema.GetRelationalReferenceTypeIfExist(fmt.Sprintf("%v#%v", entityName, s[0]))
				if !exist {
					return nil, errors.New(base.ErrorCode_ERROR_CODE_UNDEFINED_RELATION_REFERENCE.String())
				}
			}
			leaf, err = t.compileComputedUserSetIdentifier(s[0])
			if err != nil {
				return nil, errors.New(base.ErrorCode_ERROR_CODE_SCHEMA_COMPILE.String())
			}
			leaf.Exclusion = false
		} else if len(s) == 2 {
			if !t.withoutReferenceValidation {
				value, exist := t.schema.GetRelationReferenceIfExist(fmt.Sprintf("%v#%v", entityName, s[0]))
				if !exist {
					return nil, errors.New(base.ErrorCode_ERROR_CODE_UNDEFINED_RELATION_REFERENCE.String())
				}
				_, exist = t.schema.GetRelationalReferenceTypeIfExist(fmt.Sprintf("%v#%v", ast.RelationTypeStatements(value).GetEntityReference(), s[1]))
				if !exist {
					return nil, errors.New(base.ErrorCode_ERROR_CODE_UNDEFINED_RELATION_REFERENCE.String())
				}
			}
			leaf.Exclusion = false
			leaf, err = t.compileTupleToUserSetIdentifier(s[0], s[1])
			if err != nil {
				return nil, errors.New(base.ErrorCode_ERROR_CODE_SCHEMA_COMPILE.String())
			}
		} else {
			return nil, errors.New(base.ErrorCode_ERROR_CODE_NOT_SUPPORTED_RELATION_WALK.String())
		}
	case ast.PREFIX:
		s := strings.Split(expression.GetValue(), tuple.SEPARATOR)
		if len(s) == 1 {
			if !t.withoutReferenceValidation {
				if !t.schema.IsRelationReferenceExist(fmt.Sprintf("%v#%v", entityName, s[0])) {
					return nil, errors.New(base.ErrorCode_ERROR_CODE_UNDEFINED_RELATION_REFERENCE.String())
				}
			}
			leaf, err = t.compileComputedUserSetIdentifier(s[0])
			if err != nil {
				return nil, errors.New(base.ErrorCode_ERROR_CODE_SCHEMA_COMPILE.String())
			}
		} else if len(s) == 2 {
			if !t.withoutReferenceValidation {
				value, exist := t.schema.GetRelationReferenceIfExist(fmt.Sprintf("%v#%v", entityName, s[0]))
				if !exist {
					return nil, errors.New(base.ErrorCode_ERROR_CODE_UNDEFINED_RELATION_REFERENCE.String())
				}
				_, exist = t.schema.GetRelationalReferenceTypeIfExist(fmt.Sprintf("%v#%v", ast.RelationTypeStatements(value).GetEntityReference(), s[1]))
				if !exist {
					return nil, errors.New(base.ErrorCode_ERROR_CODE_UNDEFINED_RELATION_REFERENCE.String())
				}
			}
			leaf, err = t.compileTupleToUserSetIdentifier(s[0], s[1])
			if err != nil {
				return nil, errors.New(base.ErrorCode_ERROR_CODE_SCHEMA_COMPILE.String())
			}
		} else {
			return nil, errors.New(base.ErrorCode_ERROR_CODE_NOT_SUPPORTED_RELATION_WALK.String())
		}
		leaf.Exclusion = true
	default:
		return nil, errors.New(base.ErrorCode_ERROR_CODE_RELATION_DEFINITION_NOT_FOUND.String())
	}
	child.Type = &base.Child_Leaf{Leaf: leaf}
	return child, nil
}

// compileComputedUserSetIdentifier -
func (t *Compiler) compileComputedUserSetIdentifier(r string) (l *base.Leaf, err error) {
	leaf := &base.Leaf{}
	computedUserSet := &base.ComputedUserSet{
		Relation: r,
	}
	leaf.Type = &base.Leaf_ComputedUserSet{ComputedUserSet: computedUserSet}
	return leaf, nil
}

// compileTupleToUserSetIdentifier -
func (t *Compiler) compileTupleToUserSetIdentifier(p, r string) (l *base.Leaf, err error) {
	leaf := &base.Leaf{}
	computedUserSet := &base.ComputedUserSet{
		Relation: r,
	}
	tupleToUserSet := &base.TupleToUserSet{
		TupleSet: &base.TupleSet{
			Relation: p,
		},
		Computed: computedUserSet,
	}
	leaf.Type = &base.Leaf_TupleToUserSet{TupleToUserSet: tupleToUserSet}
	return leaf, nil
}
