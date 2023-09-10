package engines

import (
	"errors"
	"sync"

	"google.golang.org/protobuf/types/known/anypb"

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

type LookupOption func(engine *LookupEngine)

func LookupConcurrencyLimit(limit int) LookupOption {
	return func(c *LookupEngine) {
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
	case base.AttributeType_ATTRIBUTE_TYPE_STRING_ARRAY:
		// In the case of a string type, an empty string "" is considered the empty value.
		return []string{}
	case base.AttributeType_ATTRIBUTE_TYPE_INTEGER:
		// In the case of an integer type, zero (0) is considered the empty value.
		return 0
	case base.AttributeType_ATTRIBUTE_TYPE_INTEGER_ARRAY:
		// In the case of an integer type, zero (0) is considered the empty value.
		return []int32{}
	case base.AttributeType_ATTRIBUTE_TYPE_DOUBLE:
		// In the case of a double (or floating point) type, zero (0.0) is considered the empty value.
		return 0.0
	case base.AttributeType_ATTRIBUTE_TYPE_DOUBLE_ARRAY:
		// In the case of a double (or floating point) type, zero (0.0) is considered the empty value.
		return []float64{}
	case base.AttributeType_ATTRIBUTE_TYPE_BOOLEAN:
		// In the case of a boolean type, false is considered the empty value.
		return false
	case base.AttributeType_ATTRIBUTE_TYPE_BOOLEAN_ARRAY:
		// In the case of a boolean type, false is considered the empty value.
		return []bool{}
	default:
		// For any other types that are not explicitly handled, the function returns nil.
		// This may need to be adjusted if there are other types that need specific empty values.
		return nil
	}
}

// getEmptyProtoValueForType returns an empty protobuf value of the specified type.
// It takes a base.AttributeType as input and generates an empty protobuf Any message
// containing a default value for that type.
func getEmptyProtoValueForType(typ base.AttributeType) (*anypb.Any, error) {
	switch typ {
	case base.AttributeType_ATTRIBUTE_TYPE_STRING:
		// Create an empty protobuf String message
		value, err := anypb.New(&base.String{Value: ""})
		if err != nil {
			return nil, err
		}
		return value, nil

	case base.AttributeType_ATTRIBUTE_TYPE_STRING_ARRAY:
		// Create an empty protobuf StringArray message
		value, err := anypb.New(&base.StringArray{Values: []string{}})
		if err != nil {
			return nil, err
		}
		return value, nil

	case base.AttributeType_ATTRIBUTE_TYPE_INTEGER:
		// Create an empty protobuf Integer message
		value, err := anypb.New(&base.Integer{Value: 0})
		if err != nil {
			return nil, err
		}
		return value, nil

	case base.AttributeType_ATTRIBUTE_TYPE_INTEGER_ARRAY:
		// Create an empty protobuf IntegerArray message
		value, err := anypb.New(&base.IntegerArray{Values: []int32{}})
		if err != nil {
			return nil, err
		}
		return value, nil

	case base.AttributeType_ATTRIBUTE_TYPE_DOUBLE:
		// Create an empty protobuf Double message
		value, err := anypb.New(&base.Double{Value: 0.0})
		if err != nil {
			return nil, err
		}
		return value, nil

	case base.AttributeType_ATTRIBUTE_TYPE_DOUBLE_ARRAY:
		// Create an empty protobuf DoubleArray message
		value, err := anypb.New(&base.DoubleArray{Values: []float64{}})
		if err != nil {
			return nil, err
		}
		return value, nil

	case base.AttributeType_ATTRIBUTE_TYPE_BOOLEAN:
		// Create an empty protobuf Boolean message with a default value of false
		value, err := anypb.New(&base.Boolean{Value: false})
		if err != nil {
			return nil, err
		}
		return value, nil

	case base.AttributeType_ATTRIBUTE_TYPE_BOOLEAN_ARRAY:
		// Create an empty protobuf BooleanArray message
		value, err := anypb.New(&base.BooleanArray{Values: []bool{}})
		if err != nil {
			return nil, err
		}
		return value, nil

	default:
		// Handle the case where the provided attribute type is unknown
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
		anyValue, err = anypb.New(&base.Boolean{Value: v})
	case []bool:
		anyValue, err = anypb.New(&base.BooleanArray{Values: v})
	case int:
		anyValue, err = anypb.New(&base.Integer{Value: int32(v)})
	case []int32:
		anyValue, err = anypb.New(&base.IntegerArray{Values: v})
	case float64:
		anyValue, err = anypb.New(&base.Double{Value: v})
	case []float64:
		anyValue, err = anypb.New(&base.DoubleArray{Values: v})
	case string:
		anyValue, err = anypb.New(&base.String{Value: v})
	case []string:
		anyValue, err = anypb.New(&base.StringArray{Values: v})
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
