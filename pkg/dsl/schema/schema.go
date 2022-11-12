package schema

import (
	"errors"
	"fmt"

	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// NewSchema -
func NewSchema(entities ...*base.EntityDefinition) (schema *base.IndexedSchema) {
	schema = &base.IndexedSchema{
		EntityDefinitions:   map[string]*base.EntityDefinition{},
		RelationDefinitions: map[string]*base.RelationDefinition{},
		ActionDefinitions:   map[string]*base.ActionDefinition{},
	}
	for _, entity := range entities {
		if entity.Relations == nil {
			entity.Relations = map[string]*base.RelationDefinition{}
		}
		if entity.Actions == nil {
			entity.Actions = map[string]*base.ActionDefinition{}
		}
		schema.EntityDefinitions[entity.Name] = entity
		for _, relation := range entity.GetRelations() {
			schema.RelationDefinitions[fmt.Sprintf("%v#%v", entity.GetName(), relation.GetName())] = relation
		}
		for _, action := range entity.GetActions() {
			schema.ActionDefinitions[fmt.Sprintf("%v#%v", entity.GetName(), action.GetName())] = action
		}
	}
	return
}

// GetEntityByName -
func GetEntityByName(schema *base.IndexedSchema, name string) (entityDefinition *base.EntityDefinition, err error) {
	if en, ok := schema.GetEntityDefinitions()[name]; ok {
		return en, nil
	}
	return nil, errors.New(base.ErrorCode_ERROR_CODE_ENTITY_DEFINITION_NOT_FOUND.String())
}

// GetRelationWithKey Key -> entity_name#relation_name
func GetRelationWithKey(schema *base.IndexedSchema, key string) (relationDefinition *base.RelationDefinition, err error) {
	if en, ok := schema.GetRelationDefinitions()[key]; ok {
		return en, nil
	}
	return nil, errors.New(base.ErrorCode_ERROR_CODE_RELATION_DEFINITION_NOT_FOUND.String())
}

// GetActionWithKey Key -> entity_name#action_name
func GetActionWithKey(schema *base.IndexedSchema, key string) (actionDefinition *base.ActionDefinition, err error) {
	if en, ok := schema.GetActionDefinitions()[key]; ok {
		return en, nil
	}
	return nil, errors.New(base.ErrorCode_ERROR_CODE_ACTION_DEFINITION_NOT_FOUND.String())
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

// GetEntityReference -
func GetEntityReference(definition *base.RelationDefinition) string {
	return definition.GetEntityReference().GetName()
}

// GetTable -
func GetTable(definition *base.EntityDefinition) string {
	if table, ok := definition.GetOption()["table"]; ok {
		return table
	}
	return definition.GetName()
}

// GetIdentifier -
func GetIdentifier(definition *base.EntityDefinition) string {
	if identifier, ok := definition.GetOption()["identifier"]; ok {
		return identifier
	}
	return "id"
}

// GetColumn -
func GetColumn(definition *base.RelationDefinition) (string, bool) {
	if column, ok := definition.GetOption()["column"]; ok {
		return column, true
	}
	return "", false
}
