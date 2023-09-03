package validation

import (
	"errors"
	"fmt"

	"github.com/Permify/permify/internal/schema"
	"github.com/Permify/permify/pkg/attribute"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/tuple"
)

// ValidateTuple checks if the provided tuple conforms to the entity definition
// and relation schema provided. It returns an error if the tuple is invalid.
func ValidateTuple(definition *base.EntityDefinition, tup *base.Tuple) (err error) {
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

// ValidateTupleFilter checks if the provided filter conforms to the entity definition
func ValidateTupleFilter(filter *base.TupleFilter) (err error) {
	if filter.GetEntity().GetType() == "" {
		return errors.New(base.ErrorCode_ERROR_CODE_ENTITY_TYPE_REQUIRED.String())
	}
	return nil
}

// ValidateAttribute checks whether a given attribute request (reqAttribute) aligns with
// the attribute definition in a given entity definition. It verifies if the attribute exists
// in the entity definition and if the attribute type matches the type specified in the request.
// If the attribute value is not of the expected type, an error is returned.
//
// 'definition' is the entity definition containing the definitions of valid attributes.
// 'reqAttribute' is the attribute that needs to be validated against the entity definition.
//
// Returns nil if the attribute is valid, otherwise returns an error.
func ValidateAttribute(definition *base.EntityDefinition, reqAttribute *base.Attribute) (err error) {
	// Initialize a pointer to an AttributeDefinition.
	var attr *base.AttributeDefinition

	// Fetch the attribute from the entity definition by the attribute name from reqAttribute.
	// If it does not exist, an error is returned.
	attr, err = schema.GetAttributeByNameInEntityDefinition(definition, reqAttribute.GetAttribute())
	if err != nil {
		return err
	}

	// Check whether the type of the attribute in the request matches the type defined in the entity.
	// If the types do not match, an error indicating a type mismatch is returned.
	if attribute.TypeToString(attr.GetType()) != attribute.TypeUrlToString(reqAttribute.GetValue().GetTypeUrl()) {
		return errors.New(base.ErrorCode_ERROR_CODE_ATTRIBUTE_TYPE_MISMATCH.String())
	}

	// Validate the value of the attribute in the request against the type defined in the entity.
	// If the value is not valid, an error is returned.
	err = attribute.ValidateValue(reqAttribute.GetValue(), attr.GetType())
	if err != nil {
		return err
	}

	// If all checks pass without returning, the attribute is considered valid and the function returns nil.
	return nil
}
