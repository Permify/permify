package schema

import (
	"errors"
	"strings"

	"github.com/Permify/permify/pkg/dsl/compiler"
	"github.com/Permify/permify/pkg/dsl/parser"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// NewSchemaFromStringDefinitions -
func NewSchemaFromStringDefinitions(validation bool, definitions ...string) (*base.SchemaDefinition, error) {
	defs, err := NewEntityDefinitionsFromStringDefinitions(validation, definitions...)
	if err != nil {
		return nil, err
	}
	return NewSchemaFromEntityDefinitions(defs...), nil
}

// NewSchemaFromEntityDefinitions -
func NewSchemaFromEntityDefinitions(entities ...*base.EntityDefinition) *base.SchemaDefinition {
	schema := &base.SchemaDefinition{
		EntityDefinitions: map[string]*base.EntityDefinition{},
	}
	for _, entity := range entities {
		if entity.Relations == nil {
			entity.Relations = map[string]*base.RelationDefinition{}
		}
		if entity.Actions == nil {
			entity.Actions = map[string]*base.ActionDefinition{}
		}
		schema.EntityDefinitions[entity.Name] = entity
	}
	return schema
}

// NewEntityDefinitionsFromStringDefinitions -
func NewEntityDefinitionsFromStringDefinitions(validation bool, definitions ...string) ([]*base.EntityDefinition, error) {
	sch, err := parser.NewParser(strings.Join(definitions, "\n")).Parse()
	if err != nil {
		return nil, err
	}
	var s []*base.EntityDefinition
	s, err = compiler.NewCompiler(!validation, sch).Compile()
	if err != nil {
		return nil, err
	}
	return s, nil
}

// GetEntityByName -
func GetEntityByName(schema *base.SchemaDefinition, name string) (entityDefinition *base.EntityDefinition, err error) {
	if en, ok := schema.GetEntityDefinitions()[name]; ok {
		return en, nil
	}
	return nil, errors.New(base.ErrorCode_ERROR_CODE_ENTITY_DEFINITION_NOT_FOUND.String())
}

// GetTypeOfRelationalReferenceByNameInEntityDefinition -
func GetTypeOfRelationalReferenceByNameInEntityDefinition(entityDefinition *base.EntityDefinition, name string) (relationalDefinitionType base.EntityDefinition_RelationalReference, err error) {
	if re, ok := entityDefinition.GetReferences()[name]; ok {
		return re, nil
	}
	return base.EntityDefinition_RELATIONAL_REFERENCE_UNSPECIFIED, errors.New(base.ErrorCode_ERROR_CODE_RELATION_DEFINITION_NOT_FOUND.String())
}

// GetActionByNameInEntityDefinition -
func GetActionByNameInEntityDefinition(entityDefinition *base.EntityDefinition, name string) (actionDefinition *base.ActionDefinition, err error) {
	if re, ok := entityDefinition.GetActions()[name]; ok {
		return re, nil
	}
	return nil, errors.New(base.ErrorCode_ERROR_CODE_ACTION_DEFINITION_NOT_FOUND.String())
}

// GetRelationByNameInEntityDefinition -
func GetRelationByNameInEntityDefinition(entityDefinition *base.EntityDefinition, name string) (relationDefinition *base.RelationDefinition, err error) {
	if re, ok := entityDefinition.GetRelations()[name]; ok {
		return re, nil
	}
	return nil, errors.New(base.ErrorCode_ERROR_CODE_RELATION_DEFINITION_NOT_FOUND.String())
}

// IsDirectlyRelated -
func IsDirectlyRelated(targetRelation *base.RelationDefinition, source *base.RelationReference) bool {
	for _, refs := range targetRelation.GetRelationReferences() {
		if source.GetType() == refs.GetType() {
			if refs.GetRelation() == source.GetRelation() {
				return true
			}
		}
	}
	return false
}
