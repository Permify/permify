package ast

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Permify/permify/pkg/dsl/token"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// Validate - validates the schema to ensure that it meets certain requirements.
func (sch *Schema) Validate() error {
	if len(sch.GetReferences().entityReferences) == 0 {
		return validationError(token.PositionInfo{
			LinePosition:   1,
			ColumnPosition: 1,
		}, base.ErrorCode_ERROR_CODE_NO_ENTITY_REFERENCES_FOUND_IN_SCHEMA.String())
	}

	// Loop through all relation references in the schema.
	for _, st := range sch.GetReferences().relationReferences {
		// Loop through all relation type statements in the relation reference.
		for _, s := range st {
			// Check that the relation type statement is valid.
			if sch.validateRelationTypeStatement(s) != nil {
				return validationError(s.Type.PositionInfo, base.ErrorCode_ERROR_CODE_RELATION_REFERENCE_NOT_FOUND_IN_ENTITY_REFERENCES.String())
			}
		}
	}
	return nil
}

// validateRelationTypeStatement - validates a single relation type statement to ensure that it meets certain requirements.
func (sch *Schema) validateRelationTypeStatement(ref RelationTypeStatement) error {
	// Check that the entity reference in the relation type statement is valid.
	if !sch.GetReferences().IsEntityReferenceExist(ref.Type.Literal) {
		return validationError(ref.Type.PositionInfo, base.ErrorCode_ERROR_CODE_RELATION_REFERENCE_NOT_FOUND_IN_ENTITY_REFERENCES.String())
	}
	// If the relation type statement does not have a direct entity reference, check that the relation reference is valid.
	if !IsDirectEntityReference(ref) {
		if !sch.GetReferences().IsRelationReferenceExist(ref.Type.Literal + "#" + ref.Relation.Literal) {
			return validationError(ref.Type.PositionInfo, base.ErrorCode_ERROR_CODE_RELATION_REFERENCE_NOT_FOUND_IN_ENTITY_REFERENCES.String())
		}
	}
	return nil
}

// validationError - returns a formatted error message.
func validationError(info token.PositionInfo, message string) error {
	msg := fmt.Sprintf("%v:%v: %s", info.LinePosition, info.ColumnPosition, strings.ToLower(strings.Replace(strings.Replace(message, "ERROR_CODE_", "", -1), "_", " ", -1)))
	return errors.New(msg)
}
