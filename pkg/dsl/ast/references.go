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
	ruleReferences map[string]map[string]string

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
		ruleReferences:       map[string]map[string]string{},
		permissionReferences: map[string]struct{}{},
		attributeReferences:  map[string]AttributeTypeStatement{},
		relationReferences:   map[string][]RelationTypeStatement{},
		references:           map[string]ReferenceType{},
	}
}

// AddEntityReference sets a reference for an entity.
func (refs *References) AddEntityReference(name string) error {
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if refs.IsReferenceExist(name) {
		return fmt.Errorf("reference %s already exists", name)
	}
	refs.entityReferences[name] = struct{}{}
	refs.references[name] = ENTITY
	return nil
}

// AddRuleReference sets a reference for a rule.
func (refs *References) AddRuleReference(name string, types map[string]string) error {
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if refs.IsReferenceExist(name) {
		return fmt.Errorf("reference %s already exists", name)
	}
	refs.ruleReferences[name] = types
	refs.references[name] = RULE
	return nil
}

// AddRelationReferences sets references for a relation with its types.
func (refs *References) AddRelationReferences(key string, types []RelationTypeStatement) error {
	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}
	if refs.IsReferenceExist(key) {
		return fmt.Errorf("reference %s already exists", key)
	}
	refs.relationReferences[key] = types
	refs.references[key] = RELATION
	return nil
}

// AddPermissionReference sets a reference for a permission.
func (refs *References) AddPermissionReference(key string) error {
	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}
	if refs.IsReferenceExist(key) {
		return fmt.Errorf("reference %s already exists", key)
	}
	refs.permissionReferences[key] = struct{}{}
	refs.references[key] = PERMISSION
	return nil
}

// AddAttributeReferences sets references for an attribute with its type.
func (refs *References) AddAttributeReferences(key string, typ AttributeTypeStatement) error {
	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}
	if refs.IsReferenceExist(key) {
		return fmt.Errorf("reference %s already exists", key)
	}
	refs.attributeReferences[key] = typ
	refs.references[key] = ATTRIBUTE
	return nil
}

// UpdateRelationReferences updates the relationship references for a given key.
// It replaces the existing relation types with the new ones provided.
// If the key is empty, it returns an error.
func (refs *References) UpdateRelationReferences(key string, types []RelationTypeStatement) error {
	if len(key) == 0 {
		return fmt.Errorf("key cannot be empty")
	}
	refs.relationReferences[key] = types // Update or set the relationship types for the key.
	refs.references[key] = RELATION      // Mark this key as a relation reference.
	return nil
}

// UpdatePermissionReference updates the permission references for a given key.
// This typically means marking a particular entity or action as permitted.
// If the key is empty, it returns an error.
func (refs *References) UpdatePermissionReference(key string) error {
	if len(key) == 0 {
		return fmt.Errorf("key cannot be empty")
	}
	refs.permissionReferences[key] = struct{}{} // Set the permission flag for the key.
	refs.references[key] = PERMISSION           // Mark this key as a permission reference.
	return nil
}

// UpdateAttributeReferences updates the attribute references for a given key.
// It associates a specific attribute type with the key.
// If the key is empty, it returns an error.
func (refs *References) UpdateAttributeReferences(key string, typ AttributeTypeStatement) error {
	if len(key) == 0 {
		return fmt.Errorf("key cannot be empty")
	}
	refs.attributeReferences[key] = typ // Set the attribute type for the key.
	refs.references[key] = ATTRIBUTE    // Mark this key as an attribute reference.
	return nil
}

// RemoveRelationReferences removes the relationship references associated with a given key.
// If the key is empty, it returns an error.
func (refs *References) RemoveRelationReferences(key string) error {
	if len(key) == 0 {
		return fmt.Errorf("key cannot be empty")
	}
	delete(refs.relationReferences, key) // Remove the relationship reference for the key.
	delete(refs.references, key)         // Remove the general reference.
	return nil
}

// RemovePermissionReference removes the permission references associated with a given key.
// If the key is empty, it returns an error.
func (refs *References) RemovePermissionReference(key string) error {
	if len(key) == 0 {
		return fmt.Errorf("key cannot be empty")
	}
	delete(refs.permissionReferences, key) // Remove the permission reference for the key.
	delete(refs.references, key)           // Remove the general reference.
	return nil
}

// RemoveAttributeReferences removes the attribute references associated with a given key.
// If the key is empty, it returns an error.
func (refs *References) RemoveAttributeReferences(key string) error {
	if len(key) == 0 {
		return fmt.Errorf("key cannot be empty")
	}
	delete(refs.attributeReferences, key) // Remove the attribute reference for the key.
	delete(refs.references, key)          // Remove the general reference.
	return nil
}

// GetReferenceType retrieves the type of the reference for a given key.
func (refs *References) GetReferenceType(key string) (ReferenceType, bool) {
	if _, ok := refs.references[key]; ok {
		return refs.references[key], true
	}
	return UNSPECIFIED, false
}

// IsEntityReferenceExist checks if an entity reference exists for the given name.
func (refs *References) IsEntityReferenceExist(name string) bool {
	if _, ok := refs.entityReferences[name]; ok {
		return ok
	}
	return false
}

// IsRelationReferenceExist checks if a relation reference exists for the given key.
func (refs *References) IsRelationReferenceExist(name string) bool {
	if _, ok := refs.relationReferences[name]; ok {
		return true
	}
	return false
}

// IsAttributeReferenceExist checks if an attribute reference exists for the given key.
func (refs *References) IsAttributeReferenceExist(name string) bool {
	if _, ok := refs.attributeReferences[name]; ok {
		return true
	}
	return false
}

// IsRuleReferenceExist checks if a rule reference exists for the given name.
func (refs *References) IsRuleReferenceExist(name string) bool {
	if _, ok := refs.ruleReferences[name]; ok {
		return true
	}
	return false
}

// IsReferenceExist checks if a reference exists for the given key.
func (refs *References) IsReferenceExist(name string) bool {
	if _, ok := refs.references[name]; ok {
		return true
	}
	return false
}

// GetAttributeReferenceTypeIfExist retrieves the attribute type for a given attribute reference key.
func (refs *References) GetAttributeReferenceTypeIfExist(name string) (AttributeTypeStatement, bool) {
	if _, ok := refs.attributeReferences[name]; ok {
		return refs.attributeReferences[name], true
	}
	return AttributeTypeStatement{}, false
}

// GetRelationReferenceTypesIfExist retrieves the relation types for a given relation reference key.
func (refs *References) GetRelationReferenceTypesIfExist(name string) ([]RelationTypeStatement, bool) {
	if _, ok := refs.relationReferences[name]; ok {
		return refs.relationReferences[name], true
	}
	return nil, false
}

// GetRuleArgumentTypesIfRuleExist retrieves the rule argument types for a given rule reference key.
func (refs *References) GetRuleArgumentTypesIfRuleExist(name string) (map[string]string, bool) {
	if _, ok := refs.ruleReferences[name]; ok {
		return refs.ruleReferences[name], true
	}
	return map[string]string{}, false
}
