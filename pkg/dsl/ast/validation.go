package ast

import (
	"errors"

	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/tuple"
)

// Validate - validate schema
func (sch *Schema) Validate() error {
	if !sch.IsEntityReferenceExist(tuple.USER) {
		return errors.New(base.ErrorCode_ERROR_CODE_SCHEMA_MUST_HAVE_USER_ENTITY_DEFINITION.String())
	}
	for _, st := range sch.relationReferences {
		entityReferenceCount := 0
		for _, s := range st {
			if sch.validateRelationTypeStatement(s) != nil {
				return errors.New(base.ErrorCode_ERROR_CODE_RELATION_REFERENCE_NOT_FOUND_IN_ENTITY_REFERENCES.String())
			}
			if IsDirectEntityReference(s) {
				entityReferenceCount++
			}
			if entityReferenceCount > 1 {
				return errors.New(base.ErrorCode_ERROR_CODE_RELATION_REFERENCE_MUST_HAVE_ONE_ENTITY_REFERENCE.String())
			}
		}
	}
	return nil
}

// validateRelationTypeStatement - validate relation type statement
func (sch *Schema) validateRelationTypeStatement(ref RelationTypeStatement) error {
	if !sch.IsEntityReferenceExist(ref.Type.Literal) {
		return errors.New(base.ErrorCode_ERROR_CODE_RELATION_REFERENCE_NOT_FOUND_IN_ENTITY_REFERENCES.String())
	}
	if !IsDirectEntityReference(ref) {
		if !sch.IsRelationReferenceExist(ref.Type.Literal + "#" + ref.Relation.Literal) {
			return errors.New(base.ErrorCode_ERROR_CODE_RELATION_REFERENCE_NOT_FOUND_IN_ENTITY_REFERENCES.String())
		}
	}
	return nil
}
