package schema

import (
	"errors"
	"strings"

	"github.com/Permify/permify/pkg/dsl/compiler"
	"github.com/Permify/permify/pkg/dsl/parser"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// NewSchemaFromStringDefinitions creates a new `SchemaDefinition` from a list of string definitions.
// The `validation` argument determines whether to validate the input definitions before creating the schema.
// If the validation is successful, it returns a pointer to the newly created `SchemaDefinition`.
// If there's an error during validation or creating the schema, it returns an error.
func NewSchemaFromStringDefinitions(validation bool, definitions ...string) (*base.SchemaDefinition, error) {
	// Create entity definitions from the input string definitions
	defs, err := NewEntityDefinitionsFromStringDefinitions(validation, definitions...)
	if err != nil {
		// If there's an error, return the error
		return nil, err
	}
	// Create a schema from the entity definitions
	return NewSchemaFromEntityDefinitions(defs...), nil
}

// NewSchemaFromEntityDefinitions creates a new `SchemaDefinition` from a list of `EntityDefinition`s.
// It initializes the `EntityDefinitions` property of the `SchemaDefinition` as an empty map,
// and then adds each `EntityDefinition` to the map using the entity name as the key.
// If an `EntityDefinition` doesn't have a `Relations` or `Actions` property, it initializes it as an empty map.
// Finally, it returns a pointer to the newly created `SchemaDefinition`.
func NewSchemaFromEntityDefinitions(entities ...*base.EntityDefinition) *base.SchemaDefinition {
	// Initialize the schema definition
	schema := &base.SchemaDefinition{
		EntityDefinitions: map[string]*base.EntityDefinition{},
	}
	// Loop through each entity definition
	for _, entity := range entities {
		// If the entity definition doesn't have a Relations property, initialize it as an empty map
		if entity.Relations == nil {
			entity.Relations = map[string]*base.RelationDefinition{}
		}
		// If the entity definition doesn't have an Actions property, initialize it as an empty map
		if entity.Actions == nil {
			entity.Actions = map[string]*base.ActionDefinition{}
		}
		// Add the entity definition to the schema definition's EntityDefinitions map
		schema.EntityDefinitions[entity.Name] = entity
	}
	// Return the schema definition
	return schema
}

// NewEntityDefinitionsFromStringDefinitions creates a list of `EntityDefinition`s from a list of string definitions.
// The `validation` argument determines whether to validate the input definitions before creating the entity definitions.
// It first uses the `parser` package to parse the string definitions into a `Schema` object.
// If there's an error during parsing, it returns an error.
// Then it uses the `compiler` package to compile the `Schema` into a list of `EntityDefinition`s.
// If the validation is successful, it returns the list of `EntityDefinition`s.
// If there's an error during validation or compiling, it returns an error.
func NewEntityDefinitionsFromStringDefinitions(validation bool, definitions ...string) ([]*base.EntityDefinition, error) {
	// Use the parser to parse the string definitions into a Schema object
	sch, err := parser.NewParser(strings.Join(definitions, "\n")).Parse()
	if err != nil {
		// If there's an error, return the error
		return nil, err
	}
	// Use the compiler to compile the Schema into a list of EntityDefinitions
	var s []*base.EntityDefinition
	s, err = compiler.NewCompiler(!validation, sch).Compile()
	if err != nil {
		// If there's an error, return the error
		return nil, err
	}
	// Return the list of EntityDefinitions
	return s, nil
}

// GetEntityByName retrieves an `EntityDefinition` from a `SchemaDefinition` by its name.
// It returns a pointer to the `EntityDefinition` if it is found in the `SchemaDefinition`.
// If the `EntityDefinition` is not found, it returns an error with error code `ERROR_CODE_ENTITY_DEFINITION_NOT_FOUND`.
func GetEntityByName(schema *base.SchemaDefinition, name string) (entityDefinition *base.EntityDefinition, err error) {
	// Look up the entity definition in the schema definition's EntityDefinitions map
	if en, ok := schema.GetEntityDefinitions()[name]; ok {
		// If the entity definition is found, return it and a nil error
		return en, nil
	}
	// If the entity definition is not found, return a nil entity definition and an error
	return nil, errors.New(base.ErrorCode_ERROR_CODE_ENTITY_DEFINITION_NOT_FOUND.String())
}

// GetTypeOfRelationalReferenceByNameInEntityDefinition retrieves the type of a relational reference in an `EntityDefinition` by its name.
// It returns the type of the relational reference if it is found in the `EntityDefinition`.
// If the relational reference is not found, it returns an error with error code `ERROR_CODE_RELATION_DEFINITION_NOT_FOUND`.
func GetTypeOfRelationalReferenceByNameInEntityDefinition(entityDefinition *base.EntityDefinition, name string) (relationalDefinitionType base.EntityDefinition_RelationalReference, err error) {
	// Look up the relational reference in the entity definition's References map
	if re, ok := entityDefinition.GetReferences()[name]; ok {
		// If the relational reference is found, return its type and a nil error
		return re, nil
	}
	// If the relational reference is not found, return the default type and an error
	return base.EntityDefinition_RELATIONAL_REFERENCE_UNSPECIFIED, errors.New(base.ErrorCode_ERROR_CODE_RELATION_DEFINITION_NOT_FOUND.String())
}

// GetActionByNameInEntityDefinition retrieves an `ActionDefinition` from an `EntityDefinition` by its name.
// It returns a pointer to the `ActionDefinition` if it is found in the `EntityDefinition`.
// If the `ActionDefinition` is not found, it returns an error with error code `ERROR_CODE_ACTION_DEFINITION_NOT_FOUND`.
func GetActionByNameInEntityDefinition(entityDefinition *base.EntityDefinition, name string) (actionDefinition *base.ActionDefinition, err error) {
	// Look up the action definition in the entity definition's Actions map
	if re, ok := entityDefinition.GetActions()[name]; ok {
		// If the action definition is found, return it and a nil error
		return re, nil
	}
	// If the action definition is not found, return a nil action definition and an error
	return nil, errors.New(base.ErrorCode_ERROR_CODE_ACTION_DEFINITION_NOT_FOUND.String())
}

// GetRelationByNameInEntityDefinition retrieves a `RelationDefinition` from an `EntityDefinition` by its name.
// It returns a pointer to the `RelationDefinition` if it is found in the `EntityDefinition`.
// If the `RelationDefinition` is not found, it returns an error with error code `ERROR_CODE_RELATION_DEFINITION_NOT_FOUND`.
func GetRelationByNameInEntityDefinition(entityDefinition *base.EntityDefinition, name string) (relationDefinition *base.RelationDefinition, err error) {
	// Look up the relation definition in the entity definition's Relations map
	if re, ok := entityDefinition.GetRelations()[name]; ok {
		// If the relation definition is found, return it and a nil error
		return re, nil
	}
	// If the relation definition is not found, return a nil relation definition and an error
	return nil, errors.New(base.ErrorCode_ERROR_CODE_RELATION_DEFINITION_NOT_FOUND.String())
}

// IsDirectlyRelated checks if a source `RelationReference` is directly related to a target `RelationDefinition`.
// It returns true if the source and target have the same type and relation, false otherwise.
func IsDirectlyRelated(target *base.RelationDefinition, source *base.RelationReference) bool {
	// Loop through all the relation references of the target
	for _, refs := range target.GetRelationReferences() {
		// Check if the source and reference have the same type
		if source.GetType() == refs.GetType() {
			// Check if the source and reference have the same relation
			if refs.GetRelation() == source.GetRelation() {
				// If both type and relation match, return true
				return true
			}
		}
	}
	// If no match is found, return false
	return false
}
