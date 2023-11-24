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
func ValidateTupleFilter(tupleFilter *base.TupleFilter) (err error) {
	if IsTupleFilterEmpty(tupleFilter) {
		return errors.New(base.ErrorCode_ERROR_CODE_VALIDATION.String())
	}
	return nil
}

// ValidateFilters checks if both provided filters, tupleFilter and attributeFilter, are empty.
// It returns an error if both filters are empty, as at least one filter should contain criteria for the operation to be valid.
func ValidateFilters(tupleFilter *base.TupleFilter, attributeFilter *base.AttributeFilter) (err error) {
	// Check if both tupleFilter and attributeFilter are empty using the respective functions.
	// If both are empty, then there is nothing to validate, and an error is returned.
	if IsTupleFilterEmpty(tupleFilter) && IsAttributeFilterEmpty(attributeFilter) {
		// The error returned indicates a validation error due to both filters being empty.
		return errors.New(base.ErrorCode_ERROR_CODE_VALIDATION.String())
	}

	// If at least one of the filters is not empty, then the validation is successful, and no error is returned.
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

// IsTupleFilterEmpty checks whether any of the fields in a TupleFilter are filled.
// It assumes that a filter is "empty" if all its fields are unset or have zero values.
func IsTupleFilterEmpty(filter *base.TupleFilter) bool {
	// If the entity type is set, the filter is not empty.
	if filter.GetEntity().GetType() != "" {
		return false // Entity type is present, therefore filter is not empty.
	}

	// If the entity IDs slice is not empty, the filter is not empty.
	if len(filter.GetEntity().GetIds()) > 0 {
		return false // Entity IDs are present, therefore filter is not empty.
	}

	// If the relation is set, the filter is not empty.
	if filter.GetRelation() != "" {
		return false // Relation is present, therefore filter is not empty.
	}

	// If the subject type is set, the filter is not empty.
	if filter.GetSubject().GetType() != "" {
		return false // Subject type is present, therefore filter is not empty.
	}

	// If the subject IDs slice is not empty, the filter is not empty.
	if len(filter.GetSubject().GetIds()) > 0 {
		return false // Subject IDs are present, therefore filter is not empty.
	}

	// If the subject relation is set, the filter is not empty.
	if filter.GetSubject().GetRelation() != "" {
		return false // Subject relation is present, therefore filter is not empty.
	}

	// If none of the above conditions are met, then all fields are unset or have zero values,
	// which means the filter is empty.
	return true
}

// IsAttributeFilterEmpty checks if the provided AttributeFilter object is empty.
// An AttributeFilter is considered empty if none of its fields have been set.
func IsAttributeFilterEmpty(filter *base.AttributeFilter) bool {
	// Check if the Entity type field of the filter is set. If it is, the filter is not empty.
	if filter.GetEntity().GetType() != "" {
		return false // Entity type is specified, hence the filter is not empty.
	}

	// Check if the Entity IDs field of the filter has any IDs. If it does, the filter is not empty.
	if len(filter.GetEntity().GetIds()) > 0 {
		return false // Entity IDs are specified, hence the filter is not empty.
	}

	// Check if the Attributes field of the filter has any attributes set. If it does, the filter is not empty.
	if len(filter.GetAttributes()) > 0 {
		return false // Attributes are specified, hence the filter is not empty.
	}

	// If none of the above fields are set, then the filter is considered empty.
	return true
}

// ValidateBundleOperation checks for duplicate keys in various operations of a bundle.
func ValidateBundleOperation(op *base.Operation) error {
	// Initialize a map to store keys for write operations on relationships.
	relationshipsWrite := map[string]struct{}{}

	// Iterate over all keys obtained from the GetRelationshipsWrite method.
	for _, rw := range op.GetRelationshipsWrite() {
		// Check if the key already exists in the relationshipsWrite map.
		if _, exists := relationshipsWrite[rw]; exists {
			// If the key exists, return an error indicating a duplicate key.
			return errors.New(base.ErrorCode_ERROR_CODE_ALREADY_EXIST.String())
		}

		// Add the key to the relationshipsWrite map.
		relationshipsWrite[rw] = struct{}{}
	}

	// Initialize a map to store keys for delete operations on relationships.
	relationshipsDelete := map[string]struct{}{}

	// Iterate over all keys obtained from the GetRelationshipsDelete method.
	for _, rd := range op.GetRelationshipsDelete() {
		// Check if the key already exists in the relationshipsDelete map.
		if _, exists := relationshipsDelete[rd]; exists {
			// If the key exists, return an error indicating a duplicate key.
			return errors.New(base.ErrorCode_ERROR_CODE_ALREADY_EXIST.String())
		}

		// Add the key to the relationshipsDelete map.
		relationshipsDelete[rd] = struct{}{}
	}

	// Initialize a map to store keys for write operations on attributes.
	attributesWrite := map[string]struct{}{}

	// Iterate over all keys obtained from the GetAttributesWrite method.
	for _, aw := range op.GetAttributesWrite() {
		// Check if the key already exists in the attributesWrite map.
		if _, exists := attributesWrite[aw]; exists {
			// If the key exists, return an error indicating a duplicate key.
			return errors.New(base.ErrorCode_ERROR_CODE_ALREADY_EXIST.String())
		}

		// Add the key to the attributesWrite map.
		attributesWrite[aw] = struct{}{}
	}

	// Initialize a map to store keys for delete operations on attributes.
	attributesDelete := map[string]struct{}{}

	// Iterate over all keys obtained from the GetAttributesWrite method.
	// Note: Should this be GetAttributesDelete instead of GetAttributesWrite?
	for _, ad := range op.GetAttributesDelete() {
		// Check if the key already exists in the attributesDelete map.
		if _, exists := attributesDelete[ad]; exists {
			// If the key exists, return an error indicating a duplicate key.
			return errors.New(base.ErrorCode_ERROR_CODE_ALREADY_EXIST.String())
		}

		// Add the key to the attributesDelete map.
		attributesDelete[ad] = struct{}{}
	}

	// Return nil if no duplicates were found in any of the operations.
	return nil
}

// ValidateBundleArguments checks if all strings in slice 'a' are keys in map 'b'.
func ValidateBundleArguments(a []string, b map[string]string) error {
	// Iterate through each string in slice 'a'.
	for _, str := range a {
		// Check if the current string is a key in map 'b'.
		if _, exists := b[str]; !exists {
			// If the string is not a key in the map, return an error.
			return errors.New(base.ErrorCode_ERROR_CODE_MISSING_ARGUMENT.String())
		}
	}
	// If all strings in 'a' are keys in 'b', return nil indicating success.
	return nil
}
