package compiler

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/cel-go/cel"

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
	withReferenceValidation bool
}

// NewCompiler returns a new Compiler instance with the given schema and reference validation flag.
func NewCompiler(w bool, sch *ast.Schema) *Compiler {
	return &Compiler{
		withReferenceValidation: w,
		schema:                  sch,
	}
}

// Compile compiles the schema into a list of entity definitions.
// Returns a slice of EntityDefinition pointers and an error, if any.
func (t *Compiler) Compile() ([]*base.EntityDefinition, []*base.RuleDefinition, error) {
	// If withoutReferenceValidation is not set to true, validate the schema for reference errors.
	if t.withReferenceValidation {
		err := t.schema.Validate()
		if err != nil {
			return nil, nil, err
		}
	}

	// Create an empty slice to hold the entity definitions.
	entities := make([]*base.EntityDefinition, 0, len(t.schema.Statements))
	rules := make([]*base.RuleDefinition, 0, len(t.schema.Statements))

	// Loop through each statement in the schema.
	for _, statement := range t.schema.Statements {
		switch v := statement.(type) {
		case *ast.EntityStatement:
			// Compile the EntityStatement into an EntityDefinition.
			entityDef, err := t.compileEntity(v)
			if err != nil {
				return nil, nil, err
			}

			// Append the EntityDefinition to the slice of entity definitions.
			entities = append(entities, entityDef)
		case *ast.RuleStatement:
			// Compile the RuleStatement into a RuleDefinition.
			ruleDef, err := t.compileRule(v)
			if err != nil {
				return nil, nil, err
			}

			// Append the RuleDefinition to the slice of rule definitions.
			rules = append(rules, ruleDef)
		default:
			return nil, nil, errors.New("invalid statement")
		}
	}

	return entities, rules, nil
}

// compile - compiles an EntityStatement into an EntityDefinition
func (t *Compiler) compileEntity(sc *ast.EntityStatement) (*base.EntityDefinition, error) {
	// Initialize the entity definition
	entityDefinition := &base.EntityDefinition{
		Name:        sc.Name.Literal,
		Relations:   map[string]*base.RelationDefinition{},
		Attributes:  map[string]*base.AttributeDefinition{},
		Permissions: map[string]*base.PermissionDefinition{},
		References:  map[string]base.EntityDefinition_Reference{},
	}

	// Compile relations
	for _, rs := range sc.RelationStatements {
		// Cast the relation statement
		st, okRs := rs.(*ast.RelationStatement)
		if !okRs {
			return nil, compileError(token.PositionInfo{}, base.ErrorCode_ERROR_CODE_SCHEMA_COMPILE.String())
		}

		// Initialize the relation definition
		relationDefinition := &base.RelationDefinition{
			Name:               st.Name.Literal,
			RelationReferences: []*base.RelationReference{},
		}

		// Compile the relation types
		for _, rts := range st.RelationTypes {
			relationDefinition.RelationReferences = append(relationDefinition.RelationReferences, &base.RelationReference{
				Type:     rts.Type.Literal,
				Relation: rts.Relation.Literal,
			})
		}

		// Add the relation definition and reference
		entityDefinition.Relations[relationDefinition.GetName()] = relationDefinition
		entityDefinition.References[relationDefinition.GetName()] = base.EntityDefinition_REFERENCE_RELATION
	}

	for _, as := range sc.AttributeStatements {
		st, okAs := as.(*ast.AttributeStatement)
		if !okAs {
			return nil, compileError(token.PositionInfo{}, base.ErrorCode_ERROR_CODE_SCHEMA_COMPILE.String())
		}

		typ, err := getArgumentTypeIfExist(st.AttributeType)
		if err != nil {
			return nil, err
		}

		attributeDefinition := &base.AttributeDefinition{
			Name: st.Name.Literal,
			Type: typ,
		}

		entityDefinition.Attributes[attributeDefinition.GetName()] = attributeDefinition
		entityDefinition.References[attributeDefinition.GetName()] = base.EntityDefinition_REFERENCE_ATTRIBUTE
	}

	// Compile permissions
	for _, ps := range sc.PermissionStatements {
		// Cast the permission statement
		st, okAs := ps.(*ast.PermissionStatement)
		if !okAs {
			return nil, compileError(token.PositionInfo{}, base.ErrorCode_ERROR_CODE_SCHEMA_COMPILE.String())
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
		entityDefinition.References[permissionDefinition.GetName()] = base.EntityDefinition_REFERENCE_PERMISSION
	}

	return entityDefinition, nil
}

// compileRule compiles an ast.RuleStatement into a base.RuleDefinition object.
// It takes an *ast.RuleStatement as input, processes its arguments, and
// returns a *base.RuleDefinition or an error.
func (t *Compiler) compileRule(sc *ast.RuleStatement) (*base.RuleDefinition, error) {
	// Initialize a new base.RuleDefinition with the name and body from the rule statement.
	// The Arguments field is initialized as an empty map.
	ruleDefinition := &base.RuleDefinition{
		Name:      sc.Name.Literal,
		Arguments: map[string]base.AttributeType{},
	}

	var envOptions []cel.EnvOption
	envOptions = append(envOptions, cel.Variable("context", cel.DynType))

	// Iterate over the arguments in the rule statement.
	for name, ty := range sc.Arguments {
		// For each argument, use the getArgumentTypeIfExist function to determine the attribute type.
		typ, err := getArgumentTypeIfExist(ty)
		// If the attribute type is not recognized, return an error.
		if err != nil {
			return nil, err
		}

		cType, err := utils.GetCelType(typ)
		if err != nil {
			return nil, err
		}

		// Add the argument name and its corresponding attribute type to the Arguments map in the rule definition.
		ruleDefinition.Arguments[name.Literal] = typ
		envOptions = append(envOptions, cel.Variable(name.Literal, cType))
	}

	// Variables used within this expression environment.
	env, err := cel.NewEnv(envOptions...)
	if err != nil {
		return nil, err
	}

	// Compile and type-check the expression.
	compiledExp, issues := env.Compile(sc.Expression)
	if issues != nil && issues.Err() != nil {
		pi := sc.Name.PositionInfo
		pi.LinePosition++
		return nil, compileError(pi, issues.Err().Error())
	}

	if compiledExp.OutputType() != cel.BoolType {
		return nil, compileError(sc.Name.PositionInfo, fmt.Sprintf("rule expression must result in a boolean type not %s", compiledExp.OutputType().String()))
	}

	expr, err := cel.AstToCheckedExpr(compiledExp)
	if err != nil {
		return nil, err
	}

	ruleDefinition.Expression = expr

	// Return the completed rule definition and no error.
	return ruleDefinition, nil
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
	case ast.NOT:
		rewrite.RewriteOperation = base.Rewrite_OPERATION_EXCLUSION
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

// compileLeaf is responsible for compiling a given AST (Abstract Syntax Tree) expression into a base.Child node.
// It does this based on the type of the provided expression.
// It expects either an identifier (a variable, a constant, etc.) or a function call.
// If the expression is neither of these types, it returns an error indicating that the relation definition was not found.
func (t *Compiler) compileLeaf(entityName string, expression ast.Expression) (*base.Child, error) {
	// Switch on the type of the expression.
	switch expression.GetType() {

	// Case when the expression is an Identifier (a variable, a constant, etc.).
	case ast.IDENTIFIER:
		// Type assertion to get the underlying Identifier.
		ident := expression.(*ast.Identifier)

		// Compile the identifier and return the result.
		return t.compileIdentifier(entityName, ident)

	// Case when the expression is a Call (a function call).
	case ast.CALL:
		// Type assertion to get the underlying Call.
		call := expression.(*ast.Call)

		// Compile the call and return the result.
		return t.compileCall(entityName, call)

	// Default case when the expression type is neither an Identifier nor a Call.
	default:
		// Return a nil Child and an error indicating that the relation definition was not found.
		return nil, compileError(token.PositionInfo{}, base.ErrorCode_ERROR_CODE_RELATION_DEFINITION_NOT_FOUND.String())
	}
}

// compileIdentifier compiles an ast.Identifier into a base.Child object.
// Depending on the length of the identifier and its type, it returns different types of Child object.
func (t *Compiler) compileIdentifier(entityName string, ident *ast.Identifier) (*base.Child, error) {
	// Initialize a new base.Child
	child := &base.Child{}

	// If the identifier has no segments, return an error
	if len(ident.Idents) == 0 {
		return nil, compileError(token.PositionInfo{
			LinePosition:   1,
			ColumnPosition: 1,
		}, base.ErrorCode_ERROR_CODE_SCHEMA_COMPILE.String())
	}

	// If the identifier has one segment
	if len(ident.Idents) == 1 {
		// Check the type of the reference from the schema
		typ, exist := t.schema.GetReferences().GetReferenceType(utils.Key(entityName, ident.Idents[0].Literal))

		// If reference validation is enabled and the reference does not exist, return an error
		if t.withReferenceValidation {
			if !exist {
				return nil, compileError(ident.Idents[0].PositionInfo, base.ErrorCode_ERROR_CODE_UNDEFINED_RELATION_REFERENCE.String())
			}
		}

		// If the reference type is an attribute
		if typ == ast.ATTRIBUTE {
			// Get the attribute reference type from the schema
			at, exist := t.schema.GetReferences().GetAttributeReferenceTypeIfExist(utils.Key(entityName, ident.Idents[0].Literal))
			// If the attribute reference type does not exist or is not boolean, return an error
			if !exist || at.String() != "boolean" {
				return nil, compileError(ident.Idents[0].PositionInfo, base.ErrorCode_ERROR_CODE_SCHEMA_COMPILE.String())
			}

			// Compile the identifier into a ComputedAttributeIdentifier
			leaf, err := t.compileComputedAttributeIdentifier(ident.Idents[0].Literal)
			if err != nil {
				return nil, compileError(ident.Idents[0].PositionInfo, base.ErrorCode_ERROR_CODE_SCHEMA_COMPILE.String())
			}

			// Set the Type of the Child to the compiled Leaf
			child.Type = &base.Child_Leaf{Leaf: leaf}
			return child, nil
		} else { // The reference type is a user set
			// Compile the identifier into a ComputedUserSetIdentifier
			leaf, err := t.compileComputedUserSetIdentifier(ident.Idents[0].Literal)
			if err != nil {
				return nil, compileError(ident.Idents[0].PositionInfo, base.ErrorCode_ERROR_CODE_SCHEMA_COMPILE.String())
			}

			// Set the Type of the Child to the compiled Leaf
			child.Type = &base.Child_Leaf{Leaf: leaf}
			return child, nil
		}
	}

	// If the identifier has two segments
	if len(ident.Idents) == 2 {
		// If reference validation is enabled, validate the tuple to user set reference
		if t.withReferenceValidation {
			err := t.validateTupleToUserSetReference(entityName, ident)
			if err != nil {
				return nil, err
			}
		}

		// Compile the identifier into a TupleToUserSetIdentifier
		leaf, err := t.compileTupleToUserSetIdentifier(ident.Idents[0].Literal, ident.Idents[1].Literal)
		if err != nil {
			return nil, compileError(ident.Idents[0].PositionInfo, base.ErrorCode_ERROR_CODE_SCHEMA_COMPILE.String())
		}

		// Set the Type of the Child to the compiled Leaf
		child.Type = &base.Child_Leaf{Leaf: leaf}
		return child, nil
	}

	// If the identifier has more than two segments, return an error
	return nil, compileError(ident.Idents[2].PositionInfo, base.ErrorCode_ERROR_CODE_NOT_SUPPORTED_RELATION_WALK.String())
}

// compileCall compiles a function call within the Compiler.
// It takes the entityName and a pointer to an ast.Call object representing the function call.
// It returns a pointer to a base.Child object and an error, if any.
func (t *Compiler) compileCall(entityName string, call *ast.Call) (*base.Child, error) {
	// Create a new base.Child to store the compiled information for the call.
	child := &base.Child{}

	// Create a slice to store the call arguments.
	var arguments []*base.Argument

	// Create a map to store the types of the rule arguments, only if reference validation is enabled.
	var types map[string]string

	// If reference validation is enabled, try to get the rule argument types from the schema for the specific call.
	// If the call's rule does not exist in the schema, return an error.
	if t.withReferenceValidation {
		var exist bool
		types, exist = t.schema.GetReferences().GetRuleArgumentTypesIfRuleExist(call.Name.Literal)
		if !exist {
			return nil, compileError(call.Name.PositionInfo, base.ErrorCode_ERROR_CODE_INVALID_RULE_REFERENCE.String())
		}

		if len(types) != len(call.Arguments) {
			return nil, compileError(call.Name.PositionInfo, base.ErrorCode_ERROR_CODE_MISSING_ARGUMENT.String())
		}
	}

	if len(call.Arguments) == 0 {
		return nil, compileError(call.Name.PositionInfo, base.ErrorCode_ERROR_CODE_MISSING_ARGUMENT.String())
	}

	// Loop through each argument in the call.
	for _, argument := range call.Arguments {

		// Check if the argument has no identifiers, which is not allowed.
		// Return an error if this is the case.
		if len(argument.Idents) == 0 {
			return nil, compileError(token.PositionInfo{
				LinePosition:   1,
				ColumnPosition: 1,
			}, base.ErrorCode_ERROR_CODE_SCHEMA_COMPILE.String())
		}

		// If the argument has only one identifier, it is a computed attribute.
		if len(argument.Idents) == 1 {

			// If reference validation is enabled, check if the attribute reference exists and its type matches the rule's argument type.
			if t.withReferenceValidation {
				atyp, exist := t.schema.GetReferences().GetAttributeReferenceTypeIfExist(utils.Key(entityName, argument.Idents[0].Literal))
				if !exist {
					return nil, compileError(call.Name.PositionInfo, base.ErrorCode_ERROR_CODE_INVALID_RULE_REFERENCE.String())
				}

				// Get the type of the rule argument from the types map and compare it with the attribute reference type.
				typeInfo, exist := types[argument.Idents[0].Literal]
				if !exist {
					return nil, compileError(argument.Idents[0].PositionInfo, base.ErrorCode_ERROR_CODE_INVALID_ARGUMENT.String())
				}

				if typeInfo != atyp.String() {
					return nil, compileError(argument.Idents[0].PositionInfo, base.ErrorCode_ERROR_CODE_INVALID_ARGUMENT.String())
				}
			}

			// Append the computed attribute to the arguments slice.
			arguments = append(arguments, &base.Argument{
				Type: &base.Argument_ComputedAttribute{
					ComputedAttribute: &base.ComputedAttribute{
						Name: argument.Idents[0].Literal,
					},
				},
			})
			continue
		}

		// If the argument has more than two identifiers, it indicates an unsupported relation walk.
		// Return an error in this case.
		return nil, compileError(argument.Idents[1].PositionInfo, base.ErrorCode_ERROR_CODE_NOT_SUPPORTED_WALK.String())
	}

	// Set the child's type to be a leaf with the compiled call information.
	child.Type = &base.Child_Leaf{Leaf: &base.Leaf{
		Type: &base.Leaf_Call{Call: &base.Call{
			RuleName:  call.Name.Literal,
			Arguments: arguments,
		}},
	}}

	// Return the compiled child and nil error to indicate success.
	return child, nil
}

// compileComputedUserSetIdentifier takes a string that represents a user set relation
// and compiles it into a base.Leaf object containing that relation. It returns the resulting Leaf and no error.
func (t *Compiler) compileComputedUserSetIdentifier(r string) (l *base.Leaf, err error) {
	// Initialize a new base.Leaf
	leaf := &base.Leaf{}

	// Initialize a new base.ComputedUserSet with the provided relation
	computedUserSet := &base.ComputedUserSet{
		Relation: r,
	}

	// Set the Type of the Leaf to the newly created ComputedUserSet
	leaf.Type = &base.Leaf_ComputedUserSet{ComputedUserSet: computedUserSet}

	// Return the Leaf and no error
	return leaf, nil
}

// compileComputedAttributeIdentifier compiles a string that represents a computed attribute
// into a base.Leaf object containing that attribute. It returns the resulting Leaf and no error.
func (t *Compiler) compileComputedAttributeIdentifier(r string) (l *base.Leaf, err error) {
	// Initialize a new base.Leaf
	leaf := &base.Leaf{}

	// Initialize a new base.ComputedAttribute with the provided name
	computedAttribute := &base.ComputedAttribute{
		Name: r,
	}

	// Set the Type of the Leaf to the newly created ComputedAttribute
	leaf.Type = &base.Leaf_ComputedAttribute{ComputedAttribute: computedAttribute}

	// Return the Leaf and no error
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

// validateReference checks if the provided identifier refers to a valid relation in the schema.
func (t *Compiler) validateTupleToUserSetReference(entityName string, identifier *ast.Identifier) error {
	// Stack to hold the types to be checked.
	typeCheckStack := make([]ast.RelationTypeStatement, 0)

	// Get initial relation types for the given entity.
	initialRelationTypes, doesExist := t.schema.GetReferences().GetRelationReferenceTypesIfExist(utils.Key(entityName, identifier.Idents[0].Literal))
	if !doesExist {
		// If initial relation does not exist, return an error.
		return compileError(identifier.Idents[0].PositionInfo, base.ErrorCode_ERROR_CODE_UNDEFINED_RELATION_REFERENCE.String())
	}

	// Add the initial relation types to the stack.
	typeCheckStack = append(typeCheckStack, initialRelationTypes...)

	// While there are types to be checked in the stack...
	for len(typeCheckStack) > 0 {
		// Pop the last type from the stack.
		stackSize := len(typeCheckStack) - 1
		currentType := typeCheckStack[stackSize]
		typeCheckStack = typeCheckStack[:stackSize]

		if currentType.Relation.Literal == "" {
			typ, exist := t.schema.GetReferences().GetReferenceType(utils.Key(currentType.Type.Literal, identifier.Idents[1].Literal))
			// If the relation type does not exist, check if it is a valid relational reference.
			if !exist || typ == ast.ATTRIBUTE {
				// If not, return an error.
				return compileError(identifier.Idents[1].PositionInfo, base.ErrorCode_ERROR_CODE_UNDEFINED_RELATION_REFERENCE.String())
			}
		} else {
			// If the relation type does exist, get the corresponding relation types.
			relationTypes, doesExist := t.schema.GetReferences().GetRelationReferenceTypesIfExist(utils.Key(currentType.Type.Literal, currentType.Relation.Literal))

			if !doesExist {
				// If these types do not exist, return an error.
				return compileError(identifier.Idents[0].PositionInfo, base.ErrorCode_ERROR_CODE_UNDEFINED_RELATION_REFERENCE.String())
			}

			// Add the newly found relation types to the stack.
			typeCheckStack = append(typeCheckStack, relationTypes...)
		}
	}

	// If the function didn't return until now, the reference is valid.
	return nil
}

// compileError creates an error with the given message and position information.
func compileError(info token.PositionInfo, message string) error {
	msg := fmt.Sprintf("%v:%v: %s", info.LinePosition, info.ColumnPosition, strings.ToLower(strings.Replace(strings.Replace(message, "ERROR_CODE_", "", -1), "_", " ", -1)))
	return errors.New(msg)
}

// getArgumentTypeIfExist takes a token and checks its literal value against
// the known attribute types ("string", "boolean", "integer", "float").
// If the literal value matches one of these types, it returns the corresponding base.AttributeType and no error.
// If the literal value does not match any of the known types, it returns an ATTRIBUTE_TYPE_UNSPECIFIED
// and an error indicating an invalid argument type.
func getArgumentTypeIfExist(tkn ast.AttributeTypeStatement) (base.AttributeType, error) {
	var attrType base.AttributeType

	switch tkn.Type.Literal {
	case "string":
		attrType = base.AttributeType_ATTRIBUTE_TYPE_STRING
	case "boolean":
		attrType = base.AttributeType_ATTRIBUTE_TYPE_BOOLEAN
	case "integer":
		attrType = base.AttributeType_ATTRIBUTE_TYPE_INTEGER
	case "double":
		attrType = base.AttributeType_ATTRIBUTE_TYPE_DOUBLE
	default:
		return base.AttributeType_ATTRIBUTE_TYPE_UNSPECIFIED, compileError(tkn.Type.PositionInfo, base.ErrorCode_ERROR_CODE_INVALID_ARGUMENT.String())
	}

	if tkn.IsArray {
		return attrType + 1, nil
	}

	return attrType, nil
}
