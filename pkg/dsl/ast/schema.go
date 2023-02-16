package ast

// Schema - it contains all statements
type Schema struct {
	Statements []Statement

	// entity references
	entityReferences map[string]struct{}

	// relational references
	actionReferences   map[string]struct{}
	relationReferences map[string][]RelationTypeStatement

	// all relational references
	relationalReferences map[string]RelationalReferenceType
}

// SetEntityReferences - it contains entity references
func (sch *Schema) SetEntityReferences(r map[string]struct{}) {
	if sch.entityReferences == nil {
		sch.entityReferences = map[string]struct{}{}
	}
	sch.entityReferences = r
}

// SetActionReferences - it contains action references
func (sch *Schema) SetActionReferences(r map[string]struct{}) {
	if sch.actionReferences == nil {
		sch.actionReferences = map[string]struct{}{}
	}
	sch.actionReferences = r
}

// SetRelationReferences - it contains relation references
func (sch *Schema) SetRelationReferences(r map[string][]RelationTypeStatement) {
	if sch.relationReferences == nil {
		sch.relationReferences = map[string][]RelationTypeStatement{}
	}
	sch.relationReferences = r
}

// SetRelationalReferences it contains action and relation references
func (sch *Schema) SetRelationalReferences(r map[string]RelationalReferenceType) {
	if sch.relationalReferences == nil {
		sch.relationalReferences = map[string]RelationalReferenceType{}
	}
	sch.relationalReferences = r
}

// GetRelationalReferenceTypeIfExist - it returns the relational reference type
func (sch *Schema) GetRelationalReferenceTypeIfExist(r string) (RelationalReferenceType, bool) {
	if _, ok := sch.relationalReferences[r]; ok {
		return sch.relationalReferences[r], true
	}
	return RELATION, false
}

// IsEntityReferenceExist - it checks if the entity reference exists
func (sch *Schema) IsEntityReferenceExist(name string) bool {
	if _, ok := sch.entityReferences[name]; ok {
		return ok
	}
	return false
}

// IsRelationReferenceExist - it checks if the relation reference exists
func (sch *Schema) IsRelationReferenceExist(name string) bool {
	if _, ok := sch.relationReferences[name]; ok {
		return true
	}
	return false
}

// IsRelationalReferenceExist - it checks if the relational reference exists
func (sch *Schema) IsRelationalReferenceExist(name string) bool {
	if _, ok := sch.relationalReferences[name]; ok {
		return true
	}
	return false
}

// GetRelationReferenceIfExist - it returns the relation reference
func (sch *Schema) GetRelationReferenceIfExist(name string) ([]RelationTypeStatement, bool) {
	if _, ok := sch.relationReferences[name]; ok {
		return sch.relationReferences[name], true
	}
	return nil, false
}
