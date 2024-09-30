package parser

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Permify/permify/pkg/dsl/ast"
	"github.com/Permify/permify/pkg/dsl/lexer"
	"github.com/Permify/permify/pkg/dsl/token"
	"github.com/Permify/permify/pkg/dsl/utils"
)

const (
	// iota is a special identifier that is automatically set to 0 in this case, and increments by 1 for each subsequent constant declaration. By assigning the value to the blank identifier _, it is effectively ignored.
	_ int = iota

	// LOWEST precedence level for lowest precedence
	LOWEST
	// AND_OR_NOT precedence level for logical operators (AND, OR)
	AND_OR_NOT
)

var precedences = map[token.Type]int{ // a map that assigns precedence levels to different token types
	token.AND: AND_OR_NOT,
	token.OR:  AND_OR_NOT,
	token.NOT: AND_OR_NOT,
}

// Parser is a struct that contains information and functions related to parsing
type Parser struct {
	// a pointer to a Lexer object that will provide tokens for parsing
	l *lexer.Lexer
	// the current token being processed
	currentToken token.Token
	// the token before currentToken
	previousToken token.Token
	// the next token after currentToken
	peekToken token.Token
	// a slice of error messages that are generated during parsing
	errors []string
	// a map that associates prefix parsing functions with token types
	prefixParseFns map[token.Type]prefixParseFn
	// a map that associates infix parsing functions with token types
	infixParseFunc map[token.Type]infixParseFn
	// references to entities, rules, relations, attributes, and permissions
	references *ast.References
}

type (
	// a function that parses prefix expressions and returns an ast.Expression and error
	prefixParseFn func() (ast.Expression, error)

	// a function that parses infix expressions and returns an ast.Expression and error
	infixParseFn func(ast.Expression) (ast.Expression, error)
)

// NewParser creates a new Parser object with the given input string
func NewParser(str string) (p *Parser) {
	// initialize a new Parser object with the given input string and default values for other fields
	p = &Parser{
		l:          lexer.NewLexer(str), // create a new Lexer object with the input string
		errors:     []string{},          // initialize an empty slice of error messages
		references: ast.NewReferences(), // initialize an empty map for relational references
	}

	// register prefix parsing functions for token types IDENT and NOT
	p.prefixParseFns = make(map[token.Type]prefixParseFn)  // initialize an empty map for prefix parsing functions
	p.registerPrefix(token.IDENT, p.parseIdentifierOrCall) // associate the parseIdentifier function with the IDENT token type

	// register infix parsing functions for token types AND, OR, NOT
	p.infixParseFunc = make(map[token.Type]infixParseFn) // initialize an empty map for infix parsing functions
	p.registerInfix(token.AND, p.parseInfixExpression)   // associate the parseInfixExpression function with the AND token type
	p.registerInfix(token.OR, p.parseInfixExpression)    // associate the parseInfixExpression function with the OR token type
	p.registerInfix(token.NOT, p.parseInfixExpression)   // associate the parseInfixExpression function with the OR token type

	return p // return the newly created Parser object and no error
}

// next retrieves the next non-ignored token from the Parser's lexer and updates the Parser's currentToken and peekToken fields
func (p *Parser) next() {
	for {
		// retrieve the next token from the lexer
		peek := p.l.NextToken()
		// if the token is not an ignored token (e.g. whitespace or comments), update the currentToken and peekToken fields and exit the loop
		if !token.IsIgnores(peek.Type) {
			// set the previousToken before changing currentToken
			p.previousToken = p.currentToken
			// set the currentToken field to the previous peekToken value
			p.currentToken = p.peekToken
			// set the peekToken field to the new peek value
			p.peekToken = peek
			// exit the loop
			break
		}
	}
}

// nextWithIgnores advances the parser's token stream by one position.
// It updates the currentToken and peekToken of the Parser.
func (p *Parser) nextWithIgnores() {
	// Get the next token in the lexers token stream and store it in the variable peek.
	peek := p.l.NextToken()

	// Update the currentToken with the value of peekToken.
	p.currentToken = p.peekToken

	// Update the peekToken with the value of peek (the new next token in the lexers stream).
	p.peekToken = peek
}

// currentTokenIs checks if the Parser's currentToken is any of the given token types
func (p *Parser) currentTokenIs(tokens ...token.Type) bool {
	// iterate through the given token types and check if any of them match the currentToken's type
	for _, t := range tokens {
		if p.currentToken.Type == t {
			// if a match is found, return true
			return true
		}
	}
	// if no match is found, return false
	return false
}

// previousTokenIs checks if the Parser's previousToken type is any of the given types
func (p *Parser) previousTokenIs(tokens ...token.Type) bool {
	for _, t := range tokens {
		if p.previousToken.Type == t {
			// if a match is found, return true
			return true
		}
	}
	// if no match is found, return false
	return false
}

// peekTokenIs checks if the Parser's peekToken is any of the given token types
func (p *Parser) peekTokenIs(tokens ...token.Type) bool {
	// iterate through the given token types and check if any of them match the peekToken's type
	for _, t := range tokens {
		if p.peekToken.Type == t {
			// if a match is found, return true
			return true
		}
	}
	// if no match is found, return false
	return false
}

// Error returns an error if there are any errors in the Parser's errors slice
func (p *Parser) Error() error {
	// if there are no errors, return nil
	if len(p.errors) == 0 {
		return nil
	}
	// if there are errors, return the first error message in the errors slice as an error type
	return errors.New(p.errors[0])
}

// Parse reads and parses the input string and returns an AST representation of the schema, along with any errors encountered during parsing
func (p *Parser) Parse() (*ast.Schema, error) {
	// create a new Schema object to store the parsed statements
	schema := ast.NewSchema()
	schema.Statements = []ast.Statement{}

	// loop through the input string until the end is reached
	for !p.currentTokenIs(token.EOF) {
		// parse the next statement in the input string
		stmt, err := p.parseStatement()
		if err != nil {
			// if there was an error parsing the statement, return the error message
			return nil, p.Error()
		}
		if stmt != nil {
			// add the parsed statement to the schema's Statements field if it is not nil
			schema.Statements = append(schema.Statements, stmt)
		}

		// move to the next token in the input string
		p.next()
	}

	schema.SetReferences(p.references)

	// return the parsed schema object and nil to indicate that there were no errors
	return schema, nil
}

func (p *Parser) ParsePartial(entityName string) (ast.Statement, error) {
	for !p.currentTokenIs(token.EOF) {
		// parse the next statement in the input string
		stmt, err := p.parsePartialStatement(entityName)
		if err != nil {
			return nil, p.Error()
		}
		if stmt != nil {
			return stmt, nil
		}
		p.next()
	}
	return nil, errors.New("no valid statement found")
}

func (p *Parser) parsePartialStatement(entityName string) (ast.Statement, error) {
	switch p.currentToken.Type {
	case token.ATTRIBUTE:
		return p.parseAttributeStatement(entityName)
	case token.RELATION:
		return p.parseRelationStatement(entityName)
	case token.PERMISSION:
		return p.parsePermissionStatement(entityName)
	default:
		return nil, nil
	}
}

// parseStatement method parses the current statement based on its defined token types
func (p *Parser) parseStatement() (ast.Statement, error) {
	// switch on the currentToken's type to determine which type of statement to parse
	switch p.currentToken.Type {
	case token.ENTITY:
		// if the currentToken is ENTITY, parse an EntityStatement
		return p.parseEntityStatement()
	case token.RULE:
		// if the currentToken is RULE, parse a RuleStatement
		return p.parseRuleStatement()
	default:
		return nil, nil
	}
}

// parseEntityStatement method parses an ENTITY statement and returns an EntityStatement AST node
func (p *Parser) parseEntityStatement() (*ast.EntityStatement, error) {
	// create a new EntityStatement object and set its Entity field to the currentToken
	stmt := &ast.EntityStatement{Entity: p.currentToken}
	// expect the next token to be an identifier token, and set the EntityStatement's Name field to the identifier's value
	if !p.expectAndNext(token.IDENT) {
		return nil, p.Error()
	}
	stmt.Name = p.currentToken

	// add the entity reference to the Parser's entityReferences map
	err := p.references.AddEntityReference(stmt.Name.Literal)
	if err != nil {
		p.duplicationError(stmt.Name.Literal) // Generate an error message indicating a duplication error
		return nil, p.Error()
	}

	// expect the next token to be a left brace token, indicating the start of the entity's body
	if !p.expectAndNext(token.LCB) {
		return nil, p.Error()
	}

	// loop through the entity's body until a right brace token is encountered
	for !p.currentTokenIs(token.RCB) {
		// if the currentToken is EOF, raise an error and return nil for both the statement and error values
		if p.currentTokenIs(token.EOF) {
			p.currentError(token.RCB)
			return nil, p.Error()
		}
		// based on the currentToken's type, parse a RelationStatement or PermissionStatement and add it to the EntityStatement's corresponding field
		switch p.currentToken.Type {
		case token.RELATION:
			relation, err := p.parseRelationStatement(stmt.Name.Literal)
			if err != nil {
				return nil, p.Error()
			}
			stmt.RelationStatements = append(stmt.RelationStatements, relation)
		case token.ATTRIBUTE:
			attribute, err := p.parseAttributeStatement(stmt.Name.Literal)
			if err != nil {
				return nil, p.Error()
			}
			stmt.AttributeStatements = append(stmt.AttributeStatements, attribute)
		case token.PERMISSION:
			action, err := p.parsePermissionStatement(stmt.Name.Literal)
			if err != nil {
				return nil, p.Error()
			}
			stmt.PermissionStatements = append(stmt.PermissionStatements, action)
		default:
			// if the currentToken is not recognized, check if it is a newline, left brace, or right brace token, and skip it if it is
			if !p.currentTokenIs(token.NEWLINE) && !p.currentTokenIs(token.LCB) && !p.currentTokenIs(token.RCB) {
				// if the currentToken is not recognized and not a newline, left brace, or right brace token, raise an error and return nil for both the statement and error values
				p.currentError(token.RELATION, token.PERMISSION, token.ATTRIBUTE)
				return nil, p.Error()
			}
		}
		// move to the next token in the input string
		p.next()
	}

	// return the parsed EntityStatement and nil for the error value
	return stmt, nil
}

// parseRuleStatement is responsible for parsing a rule statement in the form:
//
//	rule name(typ1 string, typ2 boolean) {
//	    EXPRESSION
//	}
//
// This method assumes the current token points to the 'rule' token when it is called.
func (p *Parser) parseRuleStatement() (*ast.RuleStatement, error) {
	// Create a new RuleStatement
	stmt := &ast.RuleStatement{Rule: p.currentToken}

	// Expect the next token to be an identifier (the name of the rule).
	// If it's not an identifier, return an error.
	if !p.expectAndNext(token.IDENT) {
		return nil, p.Error()
	}
	stmt.Name = p.currentToken

	// Expect the next token to be a left parenthesis '(' starting the argument list.
	if !p.expectAndNext(token.LP) {
		return nil, p.Error()
	}

	arguments := map[token.Token]ast.AttributeTypeStatement{}
	args := map[string]string{}

	// Loop over the tokens until a right parenthesis ')' is encountered.
	// In each iteration, two tokens are processed: an identifier (arg name) and its type.
	for !p.peekTokenIs(token.RP) {
		// Expect the first token to be the parameter's identifier.
		if !p.expectAndNext(token.IDENT) {
			return nil, p.Error()
		}
		argument := p.currentToken
		arg := p.currentToken.Literal

		// Expect the second token to be the parameter's type.
		if !p.expectAndNext(token.IDENT) {
			return nil, p.Error()
		}

		if p.peekTokenIs(token.LSB) { // Check if the next token is '['
			arguments[argument] = ast.AttributeTypeStatement{
				Type:    p.currentToken,
				IsArray: true, // Marking the type as an array
			}
			args[arg] = p.currentToken.Literal + "[]" // Store the argument type as string with "[]" suffix
			p.next()                                  // Move to the '[' token
			if !p.expectAndNext(token.RSB) {          // Expect and move to the ']' token
				return nil, p.Error()
			}
		} else {
			arguments[argument] = ast.AttributeTypeStatement{
				Type:    p.currentToken,
				IsArray: false, // Marking the type as not an array
			}
			args[arg] = p.currentToken.Literal // Store the regular argument type
		}

		// If the next token is a comma, there are more parameters to parse.
		// Continue to the next iteration.
		if p.peekTokenIs(token.COMMA) {
			p.next()
			continue
		} else if !p.peekTokenIs(token.RP) {
			// If the next token is not a comma, it must be a closing parenthesis.
			// If it's not, return an error.
			p.peekError(token.RP)
			return nil, p.Error()
		}
	}

	// Save parsed arguments to the statement
	stmt.Arguments = arguments

	// Consume the right parenthesis.
	p.next()

	// Expect the next token to be a left curly bracket '{' starting the body.
	if !p.expectAndNext(token.LCB) {
		return nil, p.Error()
	}

	p.next()

	// Collect tokens for the body until a closing curly bracket '}' is encountered.
	var bodyTokens []token.Token
	for !p.peekTokenIs(token.RCB) {
		// If there's no closing bracket, return an error.
		if p.peekTokenIs(token.EOF) {
			p.peekError(token.RCB)
			return nil, p.Error()
		}

		bodyTokens = append(bodyTokens, p.currentToken)
		p.nextWithIgnores()
	}

	// Combine all the body tokens into a single string
	var bodyStr strings.Builder
	for _, t := range bodyTokens {
		bodyStr.WriteString(t.Literal)
	}
	stmt.Expression = bodyStr.String()

	// Expect and consume the closing curly bracket '}'.
	if !p.expectAndNext(token.RCB) {
		return nil, p.Error()
	}

	// Register the parsed rule in the parser's references.
	err := p.references.AddRuleReference(stmt.Name.Literal, args)
	if err != nil {
		// If there's an error (e.g., a duplicate rule), return an error.
		p.duplicationError(stmt.Name.Literal)
		return nil, p.Error()
	}

	// Return the successfully parsed RuleStatement.
	return stmt, nil
}

// parseRelationStatement method parses a RELATION statement and returns a RelationStatement AST node
func (p *Parser) parseAttributeStatement(entityName string) (*ast.AttributeStatement, error) {
	// create a new RelationStatement object and set its Relation field to the currentToken
	stmt := &ast.AttributeStatement{Attribute: p.currentToken}

	// expect the next token to be an identifier token, and set the RelationStatement's Name field to the identifier's value
	if !p.expectAndNext(token.IDENT) {
		return nil, p.Error()
	}
	stmt.Name = p.currentToken

	if !p.expectAndNext(token.IDENT) {
		return nil, p.Error()
	}

	atstmt := ast.AttributeTypeStatement{Type: p.currentToken}
	atstmt.IsArray = false

	if p.peekTokenIs(token.LSB) {
		p.next()
		if !p.expectAndNext(token.RSB) {
			return nil, p.Error()
		}
		atstmt.IsArray = true
	}

	stmt.AttributeType = atstmt

	key := utils.Key(entityName, stmt.Name.Literal)
	// add the relation reference to the Parser's relationReferences and relationalReferences maps
	err := p.references.AddAttributeReferences(key, atstmt)
	if err != nil {
		p.duplicationError(key) // Generate an error message indicating a duplication error
		return nil, p.Error()
	}

	// return the parsed RelationStatement and nil for the error value
	return stmt, nil
}

// parseRelationStatement method parses a RELATION statement and returns a RelationStatement AST node
func (p *Parser) parseRelationStatement(entityName string) (*ast.RelationStatement, error) {
	// create a new RelationStatement object and set its Relation field to the currentToken
	stmt := &ast.RelationStatement{Relation: p.currentToken}

	// expect the next token to be an identifier token, and set the RelationStatement's Name field to the identifier's value
	if !p.expectAndNext(token.IDENT) {
		return nil, p.Error()
	}
	stmt.Name = p.currentToken
	relationName := stmt.Name.Literal

	// expect the next token to be a SIGN token, indicating the start of the relation type(s)
	if !p.expect(token.SIGN) {
		return nil, p.Error()
	}

	// loop through the relation types until no more SIGN tokens are encountered
	for p.peekTokenIs(token.SIGN) {
		// parse a RelationTypeStatement and append it to the RelationStatement's RelationTypes field
		relationStatement, err := p.parseRelationTypeStatement()
		if err != nil {
			return nil, p.Error()
		}
		stmt.RelationTypes = append(stmt.RelationTypes, *relationStatement)
	}

	key := utils.Key(entityName, relationName)

	// add the relation reference to the Parser's relationReferences and relationalReferences maps
	err := p.references.AddRelationReferences(key, stmt.RelationTypes)
	if err != nil {
		p.duplicationError(key) // Generate an error message indicating a duplication error
		return nil, p.Error()
	}

	// return the parsed RelationStatement and nil for the error value
	return stmt, nil
}

// parseRelationTypeStatement method parses a single relation type within a RELATION statement and returns a RelationTypeStatement AST node
func (p *Parser) parseRelationTypeStatement() (*ast.RelationTypeStatement, error) {
	// expect the currentToken to be a SIGN token, indicating the start of the relation type
	if !p.expectAndNext(token.SIGN) {
		return nil, p.Error()
	}
	// create a new RelationTypeStatement object and set its Sign field to the SIGN token
	stmt := &ast.RelationTypeStatement{Sign: p.currentToken}

	// expect the next token to be an identifier token, and set the RelationTypeStatement's Type field to the identifier's value
	if !p.expectAndNext(token.IDENT) {
		return nil, p.Error()
	}
	stmt.Type = p.currentToken

	// if the next token is a HASH token, indicating that a specific relation within the relation type is being referenced, parse it and set the RelationTypeStatement's Relation field to the identifier's value
	if p.peekTokenIs(token.HASH) {
		p.next()
		if !p.expectAndNext(token.IDENT) {
			return nil, p.Error()
		}
		stmt.Relation = p.currentToken
	}

	// return the parsed RelationTypeStatement and nil for the error value
	return stmt, nil
}

// parsePermissionStatement method parses an PERMISSION statement and returns an PermissionStatement AST node
func (p *Parser) parsePermissionStatement(entityName string) (ast.Statement, error) {
	// create a new PermissionStatement object and set its Permission field to the currentToken
	stmt := &ast.PermissionStatement{Permission: p.currentToken}

	// expect the next token to be an identifier token, and set the PermissionStatement's Name field to the identifier's value
	if !p.expectAndNext(token.IDENT) {
		return nil, p.Error()
	}
	stmt.Name = p.currentToken

	key := utils.Key(entityName, stmt.Name.Literal)
	// add the action reference to the Parser's actionReferences and relationalReferences maps
	err := p.references.AddPermissionReference(key)
	if err != nil {
		p.duplicationError(key) // Generate an error message indicating a duplication error
		return nil, p.Error()
	}

	// expect the next token to be an ASSIGN token, indicating the start of the expression to be assigned to the action
	if !p.expectAndNext(token.ASSIGN) {
		return nil, p.Error()
	}

	p.next()

	// parse the expression statement and set it as the PermissionStatement's ExpressionStatement field
	ex, err := p.parseExpressionStatement()
	if err != nil {
		return nil, p.Error()
	}
	stmt.ExpressionStatement = ex

	// return the parsed PermissionStatement and nil for the error value
	return stmt, nil
}

// parseExpressionStatement method parses an expression statement and returns an ExpressionStatement AST node
func (p *Parser) parseExpressionStatement() (*ast.ExpressionStatement, error) {
	// create a new ExpressionStatement object
	stmt := &ast.ExpressionStatement{}
	var err error
	// parse the expression using the lowest precedence value as the initial precedence level
	stmt.Expression, err = p.parseExpression(LOWEST)
	if err != nil {
		return nil, p.Error()
	}

	// return the parsed ExpressionStatement and nil for the error value
	return stmt, nil
}

// expectAndNext method checks if the next token is of the expected type and advances the lexer to the next token if it is. It returns true if the next token is of the expected type, and false otherwise.
func (p *Parser) expectAndNext(t token.Type) bool {
	// if the next token is of the expected type, advance the lexer to the next token and return true
	if p.peekTokenIs(t) {
		p.next()
		return true
	}
	// otherwise, generate an error message indicating that the expected token type was not found and return false
	p.peekError(t)
	return false
}

// expect method checks if the next token is of the expected type, without advancing the lexer. It returns true if the next token is of the expected type, and false otherwise.
func (p *Parser) expect(t token.Type) bool {
	// if the next token is of the expected type, return true
	if p.peekTokenIs(t) {
		return true
	}
	// otherwise, generate an error message indicating that the expected token type was not found and return false
	p.peekError(t)
	return false
}

// parseExpression method parses an expression with a given precedence level and returns the parsed expression as an AST node. It takes an integer value indicating the precedence level.
func (p *Parser) parseExpression(precedence int) (ast.Expression, error) {
	var exp ast.Expression
	var err error

	if p.currentTokenIs(token.NEWLINE) && p.previousTokenIs(token.LP, token.AND, token.OR, token.NOT, token.ASSIGN) {
		// advance to the next token
		p.next()
	}

	if p.currentTokenIs(token.LP) {
		p.next() // Consume the left parenthesis.
		exp, err = p.parseExpression(LOWEST)
		if err != nil {
			return nil, err
		}

		if !p.expect(token.RP) {
			return nil, p.Error()
		}
		p.next() // Consume the right parenthesis.
	} else {
		// get the prefix parsing function for the current token type
		prefix := p.prefixParseFns[p.currentToken.Type]
		if prefix == nil {
			p.noPrefixParseFnError(p.currentToken.Type)
			return nil, p.Error()
		}

		// parse the prefix expression
		exp, err = prefix()
		if err != nil {
			return nil, p.Error()
		}
	}

	// continue parsing the expression while the next token has a higher precedence level than the current precedence level
	for !p.peekTokenIs(token.NEWLINE) && precedence < p.peekPrecedence() {
		// get the infix parsing function for the next token type
		infix := p.infixParseFunc[p.peekToken.Type]
		if infix == nil {
			return exp, nil
		}
		p.next()
		// parse the infix expression with the current expression as its left-hand side
		exp, err = infix(exp)
		if err != nil {
			return nil, p.Error()
		}
	}

	// return the parsed expression and nil for the error value
	return exp, nil
}

// parseInfixExpression parses an infix expression that has a left operand and an operator followed by
// a right operand, such as "a or b" or "x and y".
// It takes the left operand as an argument, constructs an InfixExpression with the current operator
// and left operand, and parses the right operand with a higher precedence to construct the final
// expression tree.
// It returns the resulting InfixExpression and any error encountered.
func (p *Parser) parseInfixExpression(left ast.Expression) (ast.Expression, error) {
	// Ensure the current token is a valid infix operator before proceeding.
	if !p.isInfixOperator(p.currentToken.Type) {
		p.currentError(token.AND, token.OR, token.NOT) // Replace with your actual valid infix token types
		return nil, p.Error()
	}

	// Create a new InfixExpression with the left operand and the current operator.
	expression := &ast.InfixExpression{
		Op:       p.currentToken,
		Left:     left,
		Operator: ast.Operator(p.currentToken.Literal),
	}

	// Get the precedence of the current operator and consume the operator token.
	precedence := p.currentPrecedence()
	p.next()

	// Parse the right operand with a higher precedence to construct the final expression tree.
	right, err := p.parseExpression(precedence)
	if err != nil {
		return nil, err
	}

	// Ensure the right operand is not nil.
	if right == nil {
		p.currentError(token.IDENT, token.LP) // Replace with your actual valid right operand token types
		return nil, p.Error()
	}

	// Set the right operand of the InfixExpression and return it.
	expression.Right = right
	return expression, nil
}

// parseIntegerLiteral parses an integer literal and returns the resulting IntegerLiteral expression.
func (p *Parser) isInfixOperator(tokenType token.Type) bool {
	return tokenType == token.AND || tokenType == token.OR || tokenType == token.NOT
}

// peekPrecedence returns the precedence of the next token in the input, if it is a known
// operator, or the lowest precedence otherwise.
func (p *Parser) peekPrecedence() int {
	if pr, ok := precedences[p.peekToken.Type]; ok {
		return pr
	}
	return LOWEST
}

// currentPrecedence returns the precedence of the current token in the input, if it is a known
// operator, or the lowest precedence otherwise.
func (p *Parser) currentPrecedence() int {
	if pr, ok := precedences[p.currentToken.Type]; ok {
		return pr
	}
	return LOWEST
}

func (p *Parser) parseIdentifierOrCall() (ast.Expression, error) {
	// Ensure the current token is a valid identifier before proceeding.
	if !p.currentTokenIs(token.IDENT) {
		return nil, fmt.Errorf("unexpected token type for identifier expression: %s", p.currentToken.Type)
	}

	if p.peekTokenIs(token.LP) {
		return p.parseCallExpression()
	}

	return p.parseIdentifierExpression()
}

// parseIdentifier parses an identifier expression that may consist of one or more dot-separated
// identifiers, such as "x", "foo.bar", or "a.b.c.d".
// It constructs a new Identifier expression with the first token as the prefix and subsequent
// tokens as identifiers, and returns the resulting expression and any error encountered.
func (p *Parser) parseIdentifierExpression() (ast.Expression, error) {
	// Ensure the current token is a valid identifier before proceeding.
	if !p.currentTokenIs(token.IDENT) {
		p.currentError(token.IDENT)
		return nil, p.Error()
	}

	// Create a new Identifier expression with the first token as the prefix.
	ident := &ast.Identifier{Idents: []token.Token{p.currentToken}}

	// If the next token is a dot, consume it and continue parsing the next identifier.
	for p.peekTokenIs(token.DOT) {
		p.next() // Consume the dot token

		// Check if the next token after the dot is a valid identifier
		if !p.expectAndNext(token.IDENT) {
			return nil, p.Error()
		}

		ident.Idents = append(ident.Idents, p.currentToken)
	}

	// Return the resulting Identifier expression.
	return ident, nil
}

// call_func(variable1, variable2)
func (p *Parser) parseCallExpression() (ast.Expression, error) {
	// Ensure the current token is a valid identifier before proceeding.
	if !p.currentTokenIs(token.IDENT) {
		p.currentError(token.IDENT)
		return nil, p.Error()
	}

	// Create a new Identifier expression with the first token as the prefix.
	call := &ast.Call{Name: p.currentToken}

	if !p.expectAndNext(token.LP) {
		return nil, p.Error()
	}

	// Check if there are no arguments
	if p.peekTokenIs(token.RP) {
		p.next() // consume the RP token
		return call, nil
	}

	p.next()

	// Parse the first argument
	ident, err := p.parseIdentifierExpression()
	if err != nil {
		return nil, err
	}

	i, ok := ident.(*ast.Identifier)
	if !ok {
		return nil, fmt.Errorf("expected identifier, got %T", ident)
	}
	call.Arguments = append(call.Arguments, *i)

	// Parse remaining arguments
	for p.peekTokenIs(token.COMMA) {
		p.next()

		if !p.expectAndNext(token.IDENT) {
			return nil, p.Error()
		}

		ident, err = p.parseIdentifierExpression()
		if err != nil {
			return nil, err
		}

		i, ok = ident.(*ast.Identifier)
		if !ok {
			return nil, fmt.Errorf("expected identifier, got %T", ident)
		}
		call.Arguments = append(call.Arguments, *i)
	}

	if !p.expectAndNext(token.RP) {
		return nil, p.Error()
	}

	// Return the resulting Identifier expression.
	return call, nil
}

// registerPrefix safely registers a parsing function for a prefix token type in the parser's prefixParseFns map.
// It takes a token type and a prefix parsing function as arguments, and stores the function in the map
// under the given token type key.
func (p *Parser) registerPrefix(tokenType token.Type, fn prefixParseFn) {
	if fn == nil {
		p.duplicationError(fmt.Sprintf("registerPrefix: nil function for token type %s", tokenType))
		return
	}

	if _, exists := p.prefixParseFns[tokenType]; exists {
		p.duplicationError(fmt.Sprintf("registerPrefix: token type %s already registered", tokenType))
		return
	}

	p.prefixParseFns[tokenType] = fn
}

// registerInfix safely registers a parsing function for an infix token type in the parser's infixParseFunc map.
// It takes a token type and an infix parsing function as arguments, and stores the function in the map
// under the given token type key.
func (p *Parser) registerInfix(tokenType token.Type, fn infixParseFn) {
	if fn == nil {
		p.duplicationError(fmt.Sprintf("registerInfix: nil function for token type %s", tokenType))
		return
	}

	if _, exists := p.infixParseFunc[tokenType]; exists {
		p.duplicationError(fmt.Sprintf("registerInfix: token type %s already registered", tokenType))
		return
	}

	p.infixParseFunc[tokenType] = fn
}

// duplicationError adds an error message to the parser's error list indicating that a duplication was found.
// It takes a key string as an argument that is used to identify the source of the duplication in the input.
func (p *Parser) duplicationError(key string) {
	msg := fmt.Sprintf("%v:%v:duplication found for %s", p.l.GetLinePosition(), p.l.GetColumnPosition(), key)
	p.errors = append(p.errors, msg)
}

// noPrefixParseFnError adds an error message to the parser's error list indicating that no prefix parsing
// function was found for a given token type.
// It takes a token type as an argument that indicates the type of the token for which a parsing function is missing.
func (p *Parser) noPrefixParseFnError(t token.Type) {
	msg := fmt.Sprintf("%v:%v:no prefix parse function for %s found", p.l.GetLinePosition(), p.l.GetColumnPosition(), t)
	p.errors = append(p.errors, msg)
}

// peekError adds an error message to the parser's error list indicating that the next token in the input
// did not match the expected type(s).
// It takes one or more token types as arguments that indicate the expected types.
func (p *Parser) peekError(t ...token.Type) {
	expected := strings.Join(tokenTypesToStrings(t), ", ")
	msg := fmt.Sprintf("%v:%v:expected next token to be %s, got %s instead", p.l.GetLinePosition(), p.l.GetColumnPosition(), expected, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

// currentError adds an error message to the parser's error list indicating that the current token in the input
// did not match the expected type(s).
// It takes one or more token types as arguments that indicate the expected types.
func (p *Parser) currentError(t ...token.Type) {
	expected := strings.Join(tokenTypesToStrings(t), ", ")
	msg := fmt.Sprintf("%v:%v:expected token to be %s, got %s instead", p.l.GetLinePosition(),
		p.l.GetColumnPosition(), expected, p.currentToken.Type)
	p.errors = append(p.errors, msg)
}

// tokenTypesToStrings converts a slice of token types to a slice of their string representations.
func tokenTypesToStrings(types []token.Type) []string {
	strs := make([]string, len(types))
	for i, t := range types {
		strs[i] = t.String()
	}
	return strs
}
