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
	en, ru, err := NewEntityAndRuleDefinitionsFromStringDefinitions(validation, definitions...)
	if err != nil {
		// If there's an error, return the error
		return nil, err
	}
	// Create a schema from the entity definitions
	return NewSchemaFromEntityAndRuleDefinitions(en, ru), nil
}

// NewSchemaFromEntityAndRuleDefinitions creates a new base.SchemaDefinition from entity and rule definitions.
// It takes two slices of pointers to base.EntityDefinition and base.RuleDefinition, representing the entities and rules to include in the schema.
// It returns a pointer to the created base.SchemaDefinition.
func NewSchemaFromEntityAndRuleDefinitions(entities []*base.EntityDefinition, rules []*base.RuleDefinition) *base.SchemaDefinition {
	// Create a new base.SchemaDefinition with empty maps for EntityDefinitions, RuleDefinitions, and References.
	schema := &base.SchemaDefinition{
		EntityDefinitions: map[string]*base.EntityDefinition{},
		RuleDefinitions:   map[string]*base.RuleDefinition{},
		References:        map[string]base.SchemaDefinition_Reference{},
	}

	// Process each entity in the entities slice.
	for _, entity := range entities {
		// If the entity's Relations map is nil, initialize it as an empty map.
		if entity.Relations == nil {
			entity.Relations = map[string]*base.RelationDefinition{}
		}

		// If the entity's Permissions map is nil, initialize it as an empty map.
		if entity.Permissions == nil {
			entity.Permissions = map[string]*base.PermissionDefinition{}
		}

		// Add the entity to the EntityDefinitions map of the schema, using its name as the key.
		schema.EntityDefinitions[entity.Name] = entity

		// Set a reference for the entity in the References map of the schema, indicating it's an entity reference.
		schema.References[entity.Name] = base.SchemaDefinition_REFERENCE_ENTITY
	}

	// Process each rule in the rules slice.
	for _, rule := range rules {
		// If the rule's Arguments map is nil, initialize it as an empty map.
		if rule.Arguments == nil {
			rule.Arguments = map[string]base.AttributeType{}
		}

		// Add the rule to the RuleDefinitions map of the schema, using its name as the key.
		schema.RuleDefinitions[rule.Name] = rule

		// Set a reference for the rule in the References map of the schema, indicating it's a rule reference.
		schema.References[rule.Name] = base.SchemaDefinition_REFERENCE_RULE
	}

	// Return the created schema.
	return schema
}

func NewEntityAndRuleDefinitionsFromStringDefinitions(validation bool, definitions ...string) ([]*base.EntityDefinition, []*base.RuleDefinition, error) {
	// Use the parser to parse the string definitions into a Schema object
	sch, err := parser.NewParser(strings.Join(definitions, "\n")).Parse()
	if err != nil {
		// If there's an error, return the error
		return nil, nil, err
	}
	// Return the list of EntityDefinitions
	return compiler.NewCompiler(validation, sch).Compile()
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

// GetRuleByName retrieves the rule definition from the given schema by its name.
// It takes a pointer to a base.SchemaDefinition and the name of the rule to look for.
// If the rule with the specified name is found in the schema's RuleDefinitions map, it returns the rule definition and a nil error.
// If the rule is not found, it returns nil for the rule definition and an error indicating that the entity definition was not found.
func GetRuleByName(schema *base.SchemaDefinition, name string) (ruleDefinition *base.RuleDefinition, err error) {
	// Check if the rule with the specified name exists in the schema's RuleDefinitions map.
	if en, ok := schema.GetRuleDefinitions()[name]; ok {
		// If the rule definition is found, return it and a nil error.
		return en, nil
	}

	// If the rule is not found, return nil for the rule definition and an error indicating that the entity definition was not found.
	return nil, errors.New(base.ErrorCode_ERROR_CODE_RULE_DEFINITION_NOT_FOUND.String())
}

// GetTypeOfReferenceByNameInEntityDefinition retrieves the type of reference in an `EntityDefinition` by its name.
// It returns the type of the relational reference if it is found in the `EntityDefinition`.
// If the relational reference is not found, it returns an error with error code `ERROR_CODE_RELATION_DEFINITION_NOT_FOUND`.
func GetTypeOfReferenceByNameInEntityDefinition(entityDefinition *base.EntityDefinition, name string) (relationalDefinitionType base.EntityDefinition_Reference, err error) {
	// Look up the relational reference in the entity definition's References map
	if re, ok := entityDefinition.GetReferences()[name]; ok {
		// If the relational reference is found, return its type and a nil error
		return re, nil
	}
	// If the relational reference is not found, return the default type and an error
	return base.EntityDefinition_REFERENCE_UNSPECIFIED, errors.New(base.ErrorCode_ERROR_CODE_RELATION_DEFINITION_NOT_FOUND.String())
}

// GetPermissionByNameInEntityDefinition retrieves an `PermissionDefinition` from an `EntityDefinition` by its name.
// It returns a pointer to the `PermissionDefinition` if it is found in the `EntityDefinition`.
// If the `PermissionDefinition` is not found, it returns an error with error code `ERROR_CODE_ACTION_DEFINITION_NOT_FOUND`.
func GetPermissionByNameInEntityDefinition(entityDefinition *base.EntityDefinition, name string) (permissionDefinition *base.PermissionDefinition, err error) {
	// Look up the permission definition in the entity definition's Permissions map
	if re, ok := entityDefinition.GetPermissions()[name]; ok {
		// If the permission definition is found, return it and a nil error
		return re, nil
	}
	// If the permission definition is not found, return a nil permission definition and an error
	return nil, errors.New(base.ErrorCode_ERROR_CODE_PERMISSION_DEFINITION_NOT_FOUND.String())
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

// GetAttributeByNameInEntityDefinition retrieves an `AttributeDefinition` from an `EntityDefinition` by its name.
// It returns a pointer to the `AttributeDefinition` if it is found in the `EntityDefinition`.
// If the `AttributeDefinition` is not found, it returns an error with error code `ERROR_CODE_ATTRIBUTE_DEFINITION_NOT_FOUND`.
func GetAttributeByNameInEntityDefinition(entityDefinition *base.EntityDefinition, name string) (attributeDefinition *base.AttributeDefinition, err error) {
	// Look up the attribute definition in the entity definition's Attributes map
	if ad, ok := entityDefinition.GetAttributes()[name]; ok {
		// If the attribute definition is found, return it and a nil error
		return ad, nil
	}
	// If the attribute definition is not found, return a nil attribute definition and an error
	return nil, errors.New(base.ErrorCode_ERROR_CODE_ATTRIBUTE_DEFINITION_NOT_FOUND.String())
}

// IsDirectlyRelated checks if a source `RelationReference` is directly related to a target `RelationDefinition`.
// It returns true if the source and target have the same type and relation, false otherwise.
func IsDirectlyRelated(target *base.RelationDefinition, source *base.Entrance) bool {
	// Loop through all the relation references of the target
	for _, refs := range target.GetRelationReferences() {
		// Check if the source and reference have the same type
		if source.GetType() == refs.GetType() {
			// Check if the source and reference have the same relation
			if refs.GetRelation() == source.GetValue() {
				// If both type and relation match, return true
				return true
			}
		}
	}
	// If no match is found, return false
	return false
}
