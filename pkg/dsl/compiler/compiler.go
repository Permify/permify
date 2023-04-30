package compiler

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Permify/permify/pkg/dsl/ast"
	"github.com/Permify/permify/pkg/dsl/token"
	"github.com/Permify/permify/pkg/dsl/utils"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// Compiler compiles an AST schema into a list of entity definitions.
type Compiler struct {
	// The AST schema to be compiled
	schema *ast.Schema
	// Whether to skip reference validation during compilation
	withoutReferenceValidation bool
}

// NewCompiler returns a new Compiler instance with the given schema and reference validation flag.
func NewCompiler(w bool, sch *ast.Schema) *Compiler {
	return &Compiler{
		withoutReferenceValidation: w,
		schema:                     sch,
	}
}

// Compile compiles the schema into a list of entity definitions.
// Returns a slice of EntityDefinition pointers and an error, if any.
func (t *Compiler) Compile() ([]*base.EntityDefinition, error) {
	// If withoutReferenceValidation is not set to true, validate the schema for reference errors.
	if !t.withoutReferenceValidation {
		err := t.schema.Validate()
		if err != nil {
			return nil, err
		}
	}

	// Create an empty slice to hold the entity definitions.
	entities := make([]*base.EntityDefinition, 0, len(t.schema.Statements))

	// Loop through each statement in the schema.
	for _, statement := range t.schema.Statements {
		// Check if the statement is an EntityStatement.
		entityStatement, ok := statement.(*ast.EntityStatement)
		if !ok {
			// If the statement is not an EntityStatement, return a compile error.
			return nil, compileError(entityStatement.Entity.PositionInfo, base.ErrorCode_ERROR_CODE_SCHEMA_COMPILE.String())
		}

		// Compile the EntityStatement into an EntityDefinition.
		entityDef, err := t.compile(entityStatement)
		if err != nil {
			return nil, err
		}

		// Append the EntityDefinition to the slice of entity definitions.
		entities = append(entities, entityDef)
	}

	return entities, nil
}

// compile - compiles an EntityStatement into an EntityDefinition
func (t *Compiler) compile(sc *ast.EntityStatement) (*base.EntityDefinition, error) {
	// Initialize the entity definition
	entityDefinition := &base.EntityDefinition{
		Name:        sc.Name.Literal,
		Relations:   map[string]*base.RelationDefinition{},
		Permissions: map[string]*base.PermissionDefinition{},
		References:  map[string]base.EntityDefinition_RelationalReference{},
	}

	// Compile relations
	for _, rs := range sc.RelationStatements {
		// Cast the relation statement
		relationSt, okRs := rs.(*ast.RelationStatement)
		if !okRs {
			return nil, compileError(relationSt.Relation.PositionInfo, base.ErrorCode_ERROR_CODE_SCHEMA_COMPILE.String())
		}

		// Initialize the relation definition
		relationDefinition := &base.RelationDefinition{
			Name:               relationSt.Name.Literal,
			RelationReferences: []*base.RelationReference{},
		}

		// Compile the relation types
		for _, rts := range relationSt.RelationTypes {
			relationDefinition.RelationReferences = append(relationDefinition.RelationReferences, &base.RelationReference{
				Type:     rts.Type.Literal,
				Relation: rts.Relation.Literal,
			})
		}

		// Add the relation definition and reference
		entityDefinition.Relations[relationDefinition.GetName()] = relationDefinition
		entityDefinition.References[relationDefinition.GetName()] = base.EntityDefinition_RELATIONAL_REFERENCE_RELATION
	}

	// Compile permissions
	for _, as := range sc.PermissionStatements {
		// Cast the permission statement
		st, okAs := as.(*ast.PermissionStatement)
		if !okAs {
			return nil, compileError(st.Permission.PositionInfo, base.ErrorCode_ERROR_CODE_SCHEMA_COMPILE.String())
		}

		// Compile the child expression
		ch, err := t.compileExpressionStatement(entityDefinition.GetName(), st.ExpressionStatement.(*ast.ExpressionStatement))
		if err != nil {
			return nil, err
		}

		// Initialize the permission definition and reference
		permissionDefinition := &base.PermissionDefinition{
			Name:  st.Name.Literal,
			Child: ch,
		}
		entityDefinition.Permissions[permissionDefinition.GetName()] = permissionDefinition
		entityDefinition.References[permissionDefinition.GetName()] = base.EntityDefinition_RELATIONAL_REFERENCE_PERMISSION
	}

	return entityDefinition, nil
}

// compileExpressionStatement compiles an ExpressionStatement into a Child node that can be used to construct an PermissionDefinition.
// It calls compileChildren to compile the expression into Child node(s).
// entityName is passed as an argument to the function to use it as a reference to the parent entity.
// Returns a pointer to a Child and an error if the compilation process fails.
func (t *Compiler) compileExpressionStatement(entityName string, expression *ast.ExpressionStatement) (*base.Child, error) {
	return t.compileChildren(entityName, expression.Expression)
}

// compileChildren - compiles the child nodes of an expression and returns a Child struct that represents them.
func (t *Compiler) compileChildren(entityName string, expression ast.Expression) (*base.Child, error) {
	if expression.IsInfix() {
		return t.compileRewrite(entityName, expression.(*ast.InfixExpression))
	}
	return t.compileLeaf(entityName, expression)
}

// compileRewrite - Compiles an InfixExpression node of type OR or AND to a base.Child struct with a base.Rewrite struct
// representing the logical operation of the expression. Recursively calls compileChildren to compile the child nodes.
// Parameters:
// - entityName: The name of the entity being compiled
// - exp: The InfixExpression node being compiled
// Returns:
// - *base.Child: A pointer to a base.Child struct representing the expression
// - error: An error if one occurred during compilation
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

// compileLeaf compiles a leaf expression into a child object. If the leaf expression is an identifier,
// it checks whether it is a valid reference to a relational reference, and creates a leaf object accordingly.
// If the identifier has one segment, it is treated as a reference to a relational reference.
// If the identifier has two segments, it is treated as a reference to a tuple and its corresponding user set.
// If the identifier has more than two segments, it is not supported and an error is returned.
// The created child object will have a Leaf field, which will be a computed user set identifier for the reference.
func (t *Compiler) compileLeaf(entityName string, expression ast.Expression) (*base.Child, error) {
	child := &base.Child{}

	var ident *ast.Identifier
	if expression.GetType() == ast.IDENTIFIER {
		ident = expression.(*ast.Identifier)
	} else {
		return nil, compileError(token.PositionInfo{
			LinePosition:   1,
			ColumnPosition: 1,
		}, base.ErrorCode_ERROR_CODE_RELATION_DEFINITION_NOT_FOUND.String())
	}

	// If the identifier has more than two segments, it is not supported
	if len(ident.Idents) == 0 {
		return nil, compileError(token.PositionInfo{
			LinePosition:   1,
			ColumnPosition: 1,
		}, base.ErrorCode_ERROR_CODE_SCHEMA_COMPILE.String())
	}

	// If the identifier has one segment, it is treated as a reference to a relational reference.
	if len(ident.Idents) == 1 {
		if !t.withoutReferenceValidation {
			if !t.schema.IsRelationalReferenceExist(utils.Key(entityName, ident.Idents[0].Literal)) {
				return nil, compileError(ident.Idents[0].PositionInfo, base.ErrorCode_ERROR_CODE_UNDEFINED_RELATION_REFERENCE.String())
			}
		}

		leaf, err := t.compileComputedUserSetIdentifier(ident.Idents[0].Literal)
		if err != nil {
			return nil, compileError(ident.Idents[0].PositionInfo, base.ErrorCode_ERROR_CODE_SCHEMA_COMPILE.String())
		}

		leaf.Exclusion = ident.IsPrefix()
		child.Type = &base.Child_Leaf{Leaf: leaf}
		return child, nil
	}

	// If the identifier has two segments, it is treated as a reference to a tuple and its corresponding user set.
	if len(ident.Idents) == 2 {
		if !t.withoutReferenceValidation {
			types, exist := t.schema.GetRelationReferenceIfExist(utils.Key(entityName, ident.Idents[0].Literal))
			if !exist {
				return nil, compileError(ident.Idents[0].PositionInfo, base.ErrorCode_ERROR_CODE_UNDEFINED_RELATION_REFERENCE.String())
			}
			if !t.schema.IsRelationalReferenceExist(utils.Key(utils.GetBaseEntityRelationTypeStatement(types).Type.Literal, ident.Idents[1].Literal)) {
				return nil, compileError(ident.Idents[1].PositionInfo, base.ErrorCode_ERROR_CODE_UNDEFINED_RELATION_REFERENCE.String())
			}
		}

		leaf, err := t.compileTupleToUserSetIdentifier(ident.Idents[0].Literal, ident.Idents[1].Literal)
		if err != nil {
			return nil, compileError(ident.Idents[0].PositionInfo, base.ErrorCode_ERROR_CODE_SCHEMA_COMPILE.String())
		}

		leaf.Exclusion = ident.IsPrefix()
		child.Type = &base.Child_Leaf{Leaf: leaf}
		return child, nil
	}

	return nil, compileError(ident.Idents[2].PositionInfo, base.ErrorCode_ERROR_CODE_NOT_SUPPORTED_RELATION_WALK.String())
}

// compileComputedUserSetIdentifier - compiles the computed user set identifier by creating a leaf with a ComputedUserSet type.
func (t *Compiler) compileComputedUserSetIdentifier(r string) (l *base.Leaf, err error) {
	leaf := &base.Leaf{}
	computedUserSet := &base.ComputedUserSet{
		Relation: r,
	}
	leaf.Type = &base.Leaf_ComputedUserSet{ComputedUserSet: computedUserSet}
	return leaf, nil
}

// compileTupleToUserSetIdentifier compiles a tuple to user set identifier to a leaf node in the IR tree.
// The resulting leaf node is used in the child node of an permission definition in the final compiled schema.
// It takes in the parameters p and r, which represent the parent and relation of the tuple, respectively.
// It returns a pointer to a leaf node and an error.
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

// compileError creates an error with the given message and position information.
func compileError(info token.PositionInfo, message string) error {
	msg := fmt.Sprintf("%v:%v: %s", info.LinePosition, info.ColumnPosition, strings.ToLower(strings.Replace(strings.Replace(message, "ERROR_CODE_", "", -1), "_", " ", -1)))
	return errors.New(msg)
}
