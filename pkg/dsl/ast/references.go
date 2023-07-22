package ast

import (
	"fmt"
)

type (
	// ReferenceType defines the type of reference.
	ReferenceType string
)

const (
	UNSPECIFIED ReferenceType = "unspecified"
	PERMISSION  ReferenceType = "permission"
	RELATION    ReferenceType = "relation"
	ATTRIBUTE   ReferenceType = "attribute"
	ENTITY      ReferenceType = "entity"
	RULE        ReferenceType = "rule"
)

// References - Map of all relational references extracted from the schema
type References struct {
	// Map of entity references extracted from the schema
	entityReferences map[string]struct{}
	// Map of rule references extracted from the schema
	ruleReferences map[string]struct{}

	// Map of permission references extracted from the schema
	// -> ["entity_name#permission_name"] = {}
	permissionReferences map[string]struct{}
	// Map of attribute references extracted from the schema
	// -> ["entity_name#attribute_name"] = "type"
	attributeReferences map[string]AttributeTypeStatement
	// Map of relation references extracted from the schema
	// -> ["entity_name#relation_name"] = []{"type", "type#relation"}
	relationReferences map[string][]RelationTypeStatement

	// Map of all relational references extracted from the schema
	// its include relation, attribute and permission references
	references map[string]ReferenceType
}

// NewReferences creates a new instance of References
func NewReferences() *References {
	return &References{
		entityReferences:     map[string]struct{}{},
		ruleReferences:       map[string]struct{}{},
		permissionReferences: map[string]struct{}{},
		attributeReferences:  map[string]AttributeTypeStatement{},
		relationReferences:   map[string][]RelationTypeStatement{},
		references:           map[string]ReferenceType{},
	}
}

func (refs *References) SetEntityReference(name string) error {
	if len(name) == 0 {
		return fmt.Errorf("name cannot be empty")
	}
	if refs.IsReferenceExist(name) {
		return fmt.Errorf("reference %s already exists", name)
	}
	refs.entityReferences[name] = struct{}{}
	refs.references[name] = ENTITY
	return nil
}

func (refs *References) SetRuleReference(name string) error {
	if len(name) == 0 {
		return fmt.Errorf("name cannot be empty")
	}
	if refs.IsReferenceExist(name) {
		return fmt.Errorf("reference %s already exists", name)
	}
	refs.ruleReferences[name] = struct{}{}
	refs.references[name] = RULE
	return nil
}

func (refs *References) SetRelationReferences(key string, types []RelationTypeStatement) error {
	if len(key) == 0 {
		return fmt.Errorf("key cannot be empty")
	}
	if refs.IsReferenceExist(key) {
		return fmt.Errorf("reference %s already exists", key)
	}
	refs.relationReferences[key] = types
	refs.references[key] = RELATION
	return nil
}

func (refs *References) SetPermissionReference(key string) error {
	if len(key) == 0 {
		return fmt.Errorf("key cannot be empty")
	}
	if refs.IsReferenceExist(key) {
		return fmt.Errorf("reference %s already exists", key)
	}
	refs.permissionReferences[key] = struct{}{}
	refs.references[key] = PERMISSION
	return nil
}

func (refs *References) SetAttributeReferences(key string, typ AttributeTypeStatement) error {
	if len(key) == 0 {
		return fmt.Errorf("key cannot be empty")
	}
	if refs.IsReferenceExist(key) {
		return fmt.Errorf("reference %s already exists", key)
	}
	refs.attributeReferences[key] = typ
	refs.references[key] = ATTRIBUTE
	return nil
}

func (refs *References) GetReferenceType(key string) (ReferenceType, bool) {
	if _, ok := refs.references[key]; ok {
		return refs.references[key], true
	}
	return UNSPECIFIED, false
}

func (refs *References) IsEntityReferenceExist(name string) bool {
	if _, ok := refs.entityReferences[name]; ok {
		return ok
	}
	return false
}

func (refs *References) IsRelationReferenceExist(name string) bool {
	if _, ok := refs.relationReferences[name]; ok {
		return true
	}
	return false
}

func (refs *References) IsAttributeReferenceExist(name string) bool {
	if _, ok := refs.attributeReferences[name]; ok {
		return true
	}
	return false
}

func (refs *References) IsRuleReferenceExist(name string) bool {
	if _, ok := refs.ruleReferences[name]; ok {
		return true
	}
	return false
}

func (refs *References) IsReferenceExist(name string) bool {
	if _, ok := refs.references[name]; ok {
		return true
	}
	return false
}

func (refs *References) GetAttributeReferenceTypeIfExist(name string) (AttributeTypeStatement, bool) {
	if _, ok := refs.attributeReferences[name]; ok {
		return refs.attributeReferences[name], true
	}
	return AttributeTypeStatement{}, false
}

func (refs *References) GetRelationReferenceTypesIfExist(name string) ([]RelationTypeStatement, bool) {
	if _, ok := refs.relationReferences[name]; ok {
		return refs.relationReferences[name], true
	}
	return nil, false
}
