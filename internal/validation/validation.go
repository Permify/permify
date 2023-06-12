package validation

import (
	"errors"
	"fmt"

	"github.com/Permify/permify/internal/schema"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/tuple"
)

// ValidateTuple checks if the provided tuple conforms to the entity definition
// and relation schema provided. It returns an error if the tuple is invalid.
func ValidateTuple(definition *base.EntityDefinition, tup *base.Tuple) (err error) {
	// Check if the subject of the tuple is a user
	if tuple.IsSubjectUser(tup.GetSubject()) {
		// If the subject is a user, the relation must be empty
		if tup.GetSubject().GetRelation() != "" {
			return errors.New(base.ErrorCode_ERROR_CODE_SUBJECT_RELATION_MUST_BE_EMPTY.String())
		}
	}

	// Check if the entity and the subject of the tuple are the same
	if tuple.IsEntityAndSubjectEquals(tup) {
		return errors.New(base.ErrorCode_ERROR_CODE_ENTITY_AND_SUBJECT_CANNOT_BE_EQUAL.String())
	}

	// Initialize variables for the relation definition and valid types
	var rel *base.RelationDefinition
	var vt []string

	// Get the relation definition for the tuple's relation within the entity definition
	rel, err = schema.GetRelationByNameInEntityDefinition(definition, tup.GetRelation())
	if err != nil {
		return err
	}

	// Iterate over relation references and build the list of valid types
	for _, t := range rel.GetRelationReferences() {
		if t.GetRelation() != "" {
			vt = append(vt, fmt.Sprintf("%s#%s", t.GetType(), t.GetRelation()))
		} else {
			vt = append(vt, t.GetType())
		}
	}

	// Validate if the subject type is among the valid types
	err = tuple.ValidateSubjectType(tup.GetSubject(), vt)
	if err != nil {
		return err
	}

	// If no errors were encountered, return nil
	return nil
}

// ValidateFilter checks if the provided filter conforms to the entity definition
func ValidateFilter(filter *base.TupleFilter) (err error) {
	if filter.GetEntity().GetType() == "" {
		return errors.New(base.ErrorCode_ERROR_CODE_ENTITY_TYPE_REQUIRED.String())
	}
	return nil
}
