package ast

import (
	"errors"

	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/tuple"
)

// Validate - validates the schema to ensure that it meets certain requirements.
func (sch *Schema) Validate() error {
	// Check that the schema has a definition for the USER entity.
	if !sch.IsEntityReferenceExist(tuple.USER) {
		return errors.New(base.ErrorCode_ERROR_CODE_SCHEMA_MUST_HAVE_USER_ENTITY_DEFINITION.String())
	}

	// Loop through all relation references in the schema.
	for _, st := range sch.relationReferences {
		entityReferenceCount := 0
		// Loop through all relation type statements in the relation reference.
		for _, s := range st {
			// Check that the relation type statement is valid.
			if sch.validateRelationTypeStatement(s) != nil {
				return errors.New(base.ErrorCode_ERROR_CODE_RELATION_REFERENCE_NOT_FOUND_IN_ENTITY_REFERENCES.String())
			}
			// Count the number of direct entity references in the relation type statement.
			if IsDirectEntityReference(s) {
				entityReferenceCount++
			}
			// Check that the relation type statement has only one direct entity reference.
			if entityReferenceCount > 1 {
				return errors.New(base.ErrorCode_ERROR_CODE_RELATION_REFERENCE_MUST_HAVE_ONE_ENTITY_REFERENCE.String())
			}
		}
	}
	return nil
}

// validateRelationTypeStatement - validates a single relation type statement to ensure that it meets certain requirements.
func (sch *Schema) validateRelationTypeStatement(ref RelationTypeStatement) error {
	// Check that the entity reference in the relation type statement is valid.
	if !sch.IsEntityReferenceExist(ref.Type.Literal) {
		return errors.New(base.ErrorCode_ERROR_CODE_RELATION_REFERENCE_NOT_FOUND_IN_ENTITY_REFERENCES.String())
	}
	// If the relation type statement does not have a direct entity reference, check that the relation reference is valid.
	if !IsDirectEntityReference(ref) {
		if !sch.IsRelationReferenceExist(ref.Type.Literal + "#" + ref.Relation.Literal) {
			return errors.New(base.ErrorCode_ERROR_CODE_RELATION_REFERENCE_NOT_FOUND_IN_ENTITY_REFERENCES.String())
		}
	}
	return nil
}
