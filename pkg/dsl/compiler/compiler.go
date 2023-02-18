package compiler

import (
	"errors"

	"github.com/Permify/permify/pkg/dsl/ast"
	"github.com/Permify/permify/pkg/dsl/utils"
	base "github.com/Permify/permify/pkg/pb/base/v1"
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
		err = t.schema.Validate()
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
		Relations:  map[string]*base.RelationDefinition{},
		Actions:    map[string]*base.ActionDefinition{},
		References: map[string]base.EntityDefinition_RelationalReference{},
	}

	// relations
	for _, rs := range sc.RelationStatements {
		relationSt, okRs := rs.(*ast.RelationStatement)
		if !okRs {
			return nil, errors.New(base.ErrorCode_ERROR_CODE_SCHEMA_COMPILE.String())
		}
		relationDefinition := &base.RelationDefinition{
			Name:               relationSt.Name.Literal,
			RelationReferences: []*base.RelationReference{},
		}

		for _, rts := range relationSt.RelationTypes {
			relationDefinition.RelationReferences = append(relationDefinition.RelationReferences, &base.RelationReference{
				Type:     rts.Type.Literal,
				Relation: rts.Relation.Literal,
			})
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
func (t *Compiler) compileChildren(entityName string, expression ast.Expression) (*base.Child, error) {
	if expression.IsInfix() {
		return t.compileRewrite(entityName, expression.(*ast.InfixExpression))
	}
	return t.compileLeaf(entityName, expression)
}

// compileRewrite -
func (t *Compiler) compileRewrite(entityName string, exp *ast.InfixExpression) (*base.Child, error) {
	var err error

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
func (t *Compiler) compileLeaf(entityName string, expression ast.Expression) (*base.Child, error) {
	child := &base.Child{}

	var ident *ast.Identifier
	if expression.GetType() == ast.IDENTIFIER {
		ident = expression.(*ast.Identifier)
	} else {
		return nil, errors.New(base.ErrorCode_ERROR_CODE_RELATION_DEFINITION_NOT_FOUND.String())
	}

	if len(ident.Idents) == 1 {
		if !t.withoutReferenceValidation {
			if !t.schema.IsRelationalReferenceExist(utils.Key(entityName, ident.Idents[0].Literal)) {
				return nil, errors.New(base.ErrorCode_ERROR_CODE_UNDEFINED_RELATION_REFERENCE.String())
			}
		}

		leaf, err := t.compileComputedUserSetIdentifier(ident.Idents[0].Literal)
		if err != nil {
			return nil, errors.New(base.ErrorCode_ERROR_CODE_SCHEMA_COMPILE.String())
		}

		leaf.Exclusion = ident.IsPrefix()
		child.Type = &base.Child_Leaf{Leaf: leaf}
		return child, nil
	}

	if len(ident.Idents) == 2 {
		if !t.withoutReferenceValidation {
			types, exist := t.schema.GetRelationReferenceIfExist(utils.Key(entityName, ident.Idents[0].Literal))
			if !exist {
				return nil, errors.New(base.ErrorCode_ERROR_CODE_UNDEFINED_RELATION_REFERENCE.String())
			}
			if !t.schema.IsRelationalReferenceExist(utils.Key(utils.GetBaseEntityRelationTypeStatement(types).Type.Literal, ident.Idents[1].Literal)) {
				return nil, errors.New(base.ErrorCode_ERROR_CODE_UNDEFINED_RELATION_REFERENCE.String())
			}
		}

		leaf, err := t.compileTupleToUserSetIdentifier(ident.Idents[0].Literal, ident.Idents[1].Literal)
		if err != nil {
			return nil, errors.New(base.ErrorCode_ERROR_CODE_SCHEMA_COMPILE.String())
		}

		leaf.Exclusion = ident.IsPrefix()
		child.Type = &base.Child_Leaf{Leaf: leaf}
		return child, nil
	}

	return nil, errors.New(base.ErrorCode_ERROR_CODE_NOT_SUPPORTED_RELATION_WALK.String())
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
