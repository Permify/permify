package engines

import (
	"errors"
	"sync"

	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/tuple"
)

const (
	_defaultConcurrencyLimit = 100
)

// CheckOption - a functional option type for configuring the CheckEngine.
type CheckOption func(engine *CheckEngine)

// CheckConcurrencyLimit - a functional option that sets the concurrency limit for the CheckEngine.
func CheckConcurrencyLimit(limit int) CheckOption {
	return func(c *CheckEngine) {
		c.concurrencyLimit = limit
	}
}

// LookupEntityOption - a functional option type for configuring the LookupEntityEngine.
type LookupEntityOption func(engine *LookupEntityEngine)

// LookupEntityConcurrencyLimit - a functional option that sets the concurrency limit for the LookupEntityEngine.
func LookupEntityConcurrencyLimit(limit int) LookupEntityOption {
	return func(c *LookupEntityEngine) {
		c.concurrencyLimit = limit
	}
}

// LookupSubjectOption - a functional option type for configuring the LookupSubjectEngine.
type LookupSubjectOption func(engine *LookupSubjectEngine)

// LookupSubjectConcurrencyLimit - a functional option that sets the concurrency limit for the LookupSubjectEngine.
func LookupSubjectConcurrencyLimit(limit int) LookupSubjectOption {
	return func(c *LookupSubjectEngine) {
		c.concurrencyLimit = limit
	}
}

// SchemaBaseSubjectFilterOption - a functional option type for configuring the LookupSubjectEngine.
type SchemaBaseSubjectFilterOption func(engine *SchemaBasedSubjectFilter)

// SchemaBaseSubjectFilterConcurrencyLimit - a functional option that sets the concurrency limit for the LookupSubjectEngine.
func SchemaBaseSubjectFilterConcurrencyLimit(limit int) SchemaBaseSubjectFilterOption {
	return func(c *SchemaBasedSubjectFilter) {
		c.concurrencyLimit = limit
	}
}

// SubjectPermissionOption - a functional option type for configuring the SubjectPermissionEngine.
type SubjectPermissionOption func(engine *SubjectPermissionEngine)

// SubjectPermissionConcurrencyLimit - a functional option that sets the concurrency limit for the SubjectPermissionEngine.
func SubjectPermissionConcurrencyLimit(limit int) SubjectPermissionOption {
	return func(c *SubjectPermissionEngine) {
		c.concurrencyLimit = limit
	}
}

// joinResponseMetas - a helper function that merges multiple PermissionCheckResponseMetadata structs into one.
func joinResponseMetas(meta ...*base.PermissionCheckResponseMetadata) *base.PermissionCheckResponseMetadata {
	response := &base.PermissionCheckResponseMetadata{}
	for _, m := range meta {
		response.CheckCount += m.CheckCount
	}
	return response
}

// SubjectPermissionResponse - a struct that holds a SubjectPermissionResponse and an error for a single subject permission check result.
type SubjectPermissionResponse struct {
	permission string
	result     base.CheckResult
	err        error
}

// CheckResponse - a struct that holds a PermissionCheckResponse and an error for a single check function.
type CheckResponse struct {
	resp *base.PermissionCheckResponse
	err  error
}

// ERMap - a thread-safe map of ENR records.
type ERMap struct {
	value sync.Map
}

func (s *ERMap) Add(onr *base.EntityAndRelation) bool {
	key := tuple.EntityAndRelationToString(onr)
	_, existed := s.value.LoadOrStore(key, struct{}{})
	return !existed
}

// SubjectFilterResponse -
type SubjectFilterResponse struct {
	resp []string
	err  error
}

// getDuplicates is a function that accepts a slice of strings and returns a slice of duplicated strings in the input slice
func getDuplicates(s []string) []string {
	// "seen" will keep track of all strings we have encountered in the slice
	seen := make(map[string]bool)

	// "duplicates" will keep track of all strings that are duplicated in the slice
	duplicates := make(map[string]bool)

	// Iterate over every string in the input slice
	for _, str := range s {
		// If we have seen the string before, then it is a duplicate.
		// So, we add it to our duplicates map.
		if _, value := seen[str]; value {
			duplicates[str] = true
		} else {
			// If we haven't seen the string before, add it to the seen map.
			seen[str] = true
		}
	}

	// "duplicatesSlice" will eventually hold our results: all strings that are duplicated in the input slice
	duplicatesSlice := make([]string, 0, len(duplicates))

	// Now, duplicates map contains all the duplicated strings.
	// We iterate over it and add each string to our result slice.
	for str := range duplicates {
		duplicatesSlice = append(duplicatesSlice, str)
	}

	// Return the slice that contains all the duplicated strings
	return duplicatesSlice
}

// getEmptyValueForType is a helper function that takes a string representation of a type
// and returns an "empty" value for that type.
// An empty value is a value that is generally considered a default or initial state for a variable of a given type.
// The purpose of this function is to be able to initialize a variable of a given type without knowing the type in advance.
// The function uses a switch statement to handle different possible type values and returns a corresponding empty value.
func getEmptyValueForType(typ base.AttributeType) interface{} {
	switch typ {
	case base.AttributeType_ATTRIBUTE_TYPE_STRING:
		// In the case of a string type, an empty string "" is considered the empty value.
		return ""
	case base.AttributeType_ATTRIBUTE_TYPE_INTEGER:
		// In the case of an integer type, zero (0) is considered the empty value.
		return 0
	case base.AttributeType_ATTRIBUTE_TYPE_DOUBLE:
		// In the case of a double (or floating point) type, zero (0.0) is considered the empty value.
		return 0.0
	case base.AttributeType_ATTRIBUTE_TYPE_BOOLEAN:
		// In the case of a boolean type, false is considered the empty value.
		return false
	default:
		// For any other types that are not explicitly handled, the function returns nil.
		// This may need to be adjusted if there are other types that need specific empty values.
		return nil
	}
}

// 'getEmptyProtoValueForType' is a function which creates an 'anypb.Any' value that
// corresponds to the base value of the provided attribute type.
func getEmptyProtoValueForType(typ base.AttributeType) (*anypb.Any, error) {
	// Based on the provided attribute type, create a new 'anypb.Any' value that corresponds
	// to the base value of that type.
	switch typ {

	// If the attribute type is a string, create an 'anypb.Any' value that corresponds to an empty string.
	case base.AttributeType_ATTRIBUTE_TYPE_STRING:
		value, err := anypb.New(wrapperspb.String(""))
		if err != nil {
			return nil, err
		}
		return value, nil

	// If the attribute type is an integer, create an 'anypb.Any' value that corresponds to 0.
	case base.AttributeType_ATTRIBUTE_TYPE_INTEGER:
		value, err := anypb.New(wrapperspb.Int64(0))
		if err != nil {
			return nil, err
		}
		return value, nil

	// If the attribute type is a double, create an 'anypb.Any' value that corresponds to 0.0.
	case base.AttributeType_ATTRIBUTE_TYPE_DOUBLE:
		value, err := anypb.New(wrapperspb.Double(0.0))
		if err != nil {
			return nil, err
		}
		return value, nil

	// If the attribute type is a boolean, create an 'anypb.Any' value that corresponds to 'false'.
	case base.AttributeType_ATTRIBUTE_TYPE_BOOLEAN:
		value, err := anypb.New(wrapperspb.Bool(false))
		if err != nil {
			return nil, err
		}
		return value, nil

	// If the attribute type is not recognized, return an error.
	default:
		return nil, errors.New("unknown type")
	}
}

// ConvertToAnyPB is a function to convert various basic Go types into *anypb.Any.
// It supports conversion from bool, int, float64, and string.
// It uses a type switch to detect the type of the input value.
// If the type is unsupported or unknown, it returns an error.
func ConvertToAnyPB(value interface{}) (*anypb.Any, error) {
	// anyValue will store the converted value, err will store any error occurred during conversion.
	var anyValue *anypb.Any
	var err error

	// Use a type switch to handle different types of value.
	switch v := value.(type) {
	case bool:
		// In case of a bool type, we convert it to a protobuf BoolValue.
		anyValue, err = anypb.New(wrapperspb.Bool(v))
	case int:
		// In case of an int type, we convert it to a protobuf Int64Value.
		// Note that this involves a type conversion from int to int64.
		anyValue, err = anypb.New(wrapperspb.Int64(int64(v)))
	case float64:
		// In case of a float64 type, we convert it to a protobuf DoubleValue.
		anyValue, err = anypb.New(wrapperspb.Double(v))
	case string:
		// In case of a string type, we convert it to a protobuf StringValue.
		anyValue, err = anypb.New(wrapperspb.String(v))
	default:
		// In case of an unsupported or unknown type, we return an error.
		return nil, errors.New("unknown type")
	}

	// If there was an error during the conversion, return the error.
	if err != nil {
		return nil, err
	}

	// If the conversion was successful, return the converted value.
	return anyValue, nil
}
