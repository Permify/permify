package ast

// Schema represents the parsed schema, which contains all the statements
// and extracted entity and relational references used by the schema. It
// is used as an intermediate representation before generating the
// corresponding metadata.
type Schema struct {
	// The list of statements in the schema
	Statements []Statement
	// Map of entity references extracted from the schema
	entityReferences map[string]struct{}
	// Map of permission references extracted from the schema
	permissionReferences map[string]struct{}
	// Map of relation references extracted from the schema
	relationReferences map[string][]RelationTypeStatement
	// Map of all relational references extracted from the schema
	relationalReferences map[string]RelationalReferenceType
}

// SetEntityReferences sets the entity references in the schema
func (sch *Schema) SetEntityReferences(r map[string]struct{}) {
	if sch.entityReferences == nil {
		sch.entityReferences = map[string]struct{}{}
	}
	sch.entityReferences = r
}

// SetPermissionReferences sets the permission references in the schema
func (sch *Schema) SetPermissionReferences(r map[string]struct{}) {
	if sch.permissionReferences == nil {
		sch.permissionReferences = map[string]struct{}{}
	}
	sch.permissionReferences = r
}

// SetRelationReferences sets the relation references in the schema
func (sch *Schema) SetRelationReferences(r map[string][]RelationTypeStatement) {
	if sch.relationReferences == nil {
		sch.relationReferences = map[string][]RelationTypeStatement{}
	}
	sch.relationReferences = r
}

// SetRelationalReferences sets the relational references in the schema
func (sch *Schema) SetRelationalReferences(r map[string]RelationalReferenceType) {
	if sch.relationalReferences == nil {
		sch.relationalReferences = map[string]RelationalReferenceType{}
	}
	sch.relationalReferences = r
}

// GetRelationalReferenceTypeIfExist returns the relational reference type if it exists
func (sch *Schema) GetRelationalReferenceTypeIfExist(r string) (RelationalReferenceType, bool) {
	if _, ok := sch.relationalReferences[r]; ok {
		return sch.relationalReferences[r], true
	}
	return RELATION, false
}

// IsEntityReferenceExist checks if the entity reference exists in the schema
func (sch *Schema) IsEntityReferenceExist(name string) bool {
	if _, ok := sch.entityReferences[name]; ok {
		return ok
	}
	return false
}

// IsRelationReferenceExist checks if the relation reference exists in the schema
func (sch *Schema) IsRelationReferenceExist(name string) bool {
	if _, ok := sch.relationReferences[name]; ok {
		return true
	}
	return false
}

// IsRelationalReferenceExist checks if the relational reference exists in the schema
func (sch *Schema) IsRelationalReferenceExist(name string) bool {
	if _, ok := sch.relationalReferences[name]; ok {
		return true
	}
	return false
}

// GetRelationReferenceIfExist returns the relation reference if it exists in the schema
func (sch *Schema) GetRelationReferenceIfExist(name string) ([]RelationTypeStatement, bool) {
	if _, ok := sch.relationReferences[name]; ok {
		return sch.relationReferences[name], true
	}
	return nil, false
}
