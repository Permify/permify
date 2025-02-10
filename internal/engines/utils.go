package engines

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"

	"google.golang.org/protobuf/types/known/anypb"

	"github.com/Permify/permify/internal/schema"
	"github.com/Permify/permify/pkg/attribute"
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

// SubjectFilterOption - a functional option type for configuring the LookupSubjectEngine.
type SubjectFilterOption func(engine *SubjectFilter)

// SubjectFilterConcurrencyLimit - a functional option that sets the concurrency limit for the LookupSubjectEngine.
func SubjectFilterConcurrencyLimit(limit int) SubjectFilterOption {
	return func(c *SubjectFilter) {
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

// VisitsMap - a thread-safe map of ENR records.
type VisitsMap struct {
	er        sync.Map
	published sync.Map
}

func (s *VisitsMap) AddER(entity *base.Entity, relation string) bool {
	key := tuple.EntityAndRelationToString(entity, relation)
	_, existed := s.er.LoadOrStore(key, struct{}{})
	return !existed
}

func (s *VisitsMap) AddEA(entityType, attribute string) bool {
	key := fmt.Sprintf("%s$%s", entityType, attribute)
	_, existed := s.er.LoadOrStore(key, struct{}{})
	return !existed
}

func (s *VisitsMap) AddPublished(entity *base.Entity) bool {
	key := tuple.EntityToString(entity)
	_, existed := s.published.LoadOrStore(key, struct{}{})
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
		value, err := anypb.New(&base.StringValue{Data: ""})
		if err != nil {
			return nil, err
		}
		return value, nil

	case base.AttributeType_ATTRIBUTE_TYPE_STRING_ARRAY:
		// Create an empty protobuf StringArray message
		value, err := anypb.New(&base.StringArrayValue{Data: []string{}})
		if err != nil {
			return nil, err
		}
		return value, nil

	case base.AttributeType_ATTRIBUTE_TYPE_INTEGER:
		// Create an empty protobuf Integer message
		value, err := anypb.New(&base.IntegerValue{Data: 0})
		if err != nil {
			return nil, err
		}
		return value, nil

	case base.AttributeType_ATTRIBUTE_TYPE_INTEGER_ARRAY:
		// Create an empty protobuf IntegerArray message
		value, err := anypb.New(&base.IntegerArrayValue{Data: []int32{}})
		if err != nil {
			return nil, err
		}
		return value, nil

	case base.AttributeType_ATTRIBUTE_TYPE_DOUBLE:
		// Create an empty protobuf Double message
		value, err := anypb.New(&base.DoubleValue{Data: 0.0})
		if err != nil {
			return nil, err
		}
		return value, nil

	case base.AttributeType_ATTRIBUTE_TYPE_DOUBLE_ARRAY:
		// Create an empty protobuf DoubleArray message
		value, err := anypb.New(&base.DoubleArrayValue{Data: []float64{}})
		if err != nil {
			return nil, err
		}
		return value, nil

	case base.AttributeType_ATTRIBUTE_TYPE_BOOLEAN:
		// Create an empty protobuf Boolean message with a default value of false
		value, err := anypb.New(&base.BooleanValue{Data: false})
		if err != nil {
			return nil, err
		}
		return value, nil

	case base.AttributeType_ATTRIBUTE_TYPE_BOOLEAN_ARRAY:
		// Create an empty protobuf BooleanArray message
		value, err := anypb.New(&base.BooleanArrayValue{Data: []bool{}})
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
		anyValue, err = anypb.New(&base.BooleanValue{Data: v})
	case []bool:
		anyValue, err = anypb.New(&base.BooleanArrayValue{Data: v})
	case int:
		anyValue, err = anypb.New(&base.IntegerValue{Data: int32(v)})
	case []int32:
		anyValue, err = anypb.New(&base.IntegerArrayValue{Data: v})
	case float64:
		anyValue, err = anypb.New(&base.DoubleValue{Data: v})
	case []float64:
		anyValue, err = anypb.New(&base.DoubleArrayValue{Data: v})
	case string:
		anyValue, err = anypb.New(&base.StringValue{Data: v})
	case []string:
		anyValue, err = anypb.New(&base.StringArrayValue{Data: v})
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

// GenerateKey function takes a PermissionCheckRequest and generates a unique key
// Key format: check|{tenant_id}|{schema_version}|{snap_token}|{context}|{entity:id#permission(optional_arguments)@subject:id#optional_relation}
func GenerateKey(key *base.PermissionCheckRequest, isRelational bool) string {
	// Initialize the parts slice with the string "check"
	parts := []string{"check"}

	// If tenantId is not empty, append it to parts
	if tenantId := key.GetTenantId(); tenantId != "" {
		parts = append(parts, tenantId)
	}

	// If Metadata exists, extract schema version and snap token and append them to parts if they are not empty
	if meta := key.GetMetadata(); meta != nil {
		if version := meta.GetSchemaVersion(); version != "" {
			parts = append(parts, version)
		}
		if token := meta.GetSnapToken(); token != "" {
			parts = append(parts, token)
		}
	}

	// If Context exists, convert it to string and append it to parts
	if ctx := key.GetContext(); ctx != nil {
		parts = append(parts, ContextToString(ctx))
	}

	if isRelational {
		// Convert entity and relation to string with any optional arguments and append to parts
		entityRelationString := tuple.EntityAndRelationToString(key.GetEntity(), key.GetPermission())

		subjectString := tuple.SubjectToString(key.GetSubject())

		if entityRelationString != "" {
			parts = append(parts, fmt.Sprintf("%s@%s", entityRelationString, subjectString))
		}
	} else {
		parts = append(parts, attribute.EntityAndCallOrAttributeToString(
			key.GetEntity(),
			key.GetPermission(),
			key.GetArguments()...,
		))
	}

	// Join all parts with "|" delimiter to generate the final key
	return strings.Join(parts, "|")
}

// ContextToString function takes a Context object and converts it into a string
func ContextToString(context *base.Context) string {
	// Initialize an empty slice to store parts of the context
	var parts []string

	// For each Tuple in the Context, convert it to a string and append to parts
	for _, tup := range context.GetTuples() {
		parts = append(parts, tuple.ToString(tup)) // replace with your function
	}

	// For each Attribute in the Context, convert it to a string and append to parts
	for _, attr := range context.GetAttributes() {
		parts = append(parts, attribute.ToString(attr)) // replace with your function
	}

	// If Data exists in the Context, convert it to JSON string and append to parts
	if data := context.GetData(); data != nil {
		parts = append(parts, mapToString(data.AsMap()))
	}

	// Join all parts with "," delimiter to generate the final context string
	return strings.Join(parts, ",")
}

// mapToString function takes a map[string]interface{} and converts it into a string
func mapToString(m map[string]interface{}) string {
	var keys []string
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var parts []string
	for _, key := range keys {
		value := m[key]
		parts = append(parts, fmt.Sprintf("%s:%v", key, value))
	}
	return strings.Join(parts, ",")
}

// IsRelational determines if a given permission corresponds to a relational attribute
// in the provided entity definition.
func IsRelational(en *base.EntityDefinition, permission string) bool {
	// Default to non-relational
	isRelational := false

	// Attempt to get the type of reference for the given permission in the entity definition
	tor, err := schema.GetTypeOfReferenceByNameInEntityDefinition(en, permission)
	if err == nil && tor != base.EntityDefinition_REFERENCE_ATTRIBUTE {
		// If the type of reference is anything other than REFERENCE_ATTRIBUTE,
		// treat it as a relational attribute
		isRelational = true
	}

	return isRelational
}
