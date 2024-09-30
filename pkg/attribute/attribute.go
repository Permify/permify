package attribute

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	base "github.com/Permify/permify/pkg/pb/base/v1"
)

const (
	ENTITY    = "%s:%s"
	ATTRIBUTE = "$%s"
	VALUE     = "%s:%s"
)

// Attribute function takes a string representation of an attribute and converts it back into the Attribute object.
func Attribute(attribute string) (*base.Attribute, error) {
	// Splitting the attribute string by "@" delimiter
	s := strings.Split(strings.TrimSpace(attribute), "|")
	if len(s) != 2 || s[0] == "" || s[1] == "" {
		// The attribute string should have exactly two parts
		return nil, ErrInvalidAttribute
	}

	// Splitting the entity part of the string by "#" delimiter
	e := strings.Split(s[0], "$")
	if len(e) != 2 || e[0] == "" || e[1] == "" {
		// The entity string should have exactly two parts
		return nil, ErrInvalidEntity
	}

	// Splitting the entity type and id by ":" delimiter
	et := strings.Split(e[0], ":")
	if len(et) != 2 || et[0] == "" || et[1] == "" {
		// The entity type and id should have exactly two parts
		return nil, ErrInvalidAttribute
	}

	// Splitting the attribute value part of the string by ":" delimiter
	v := strings.Split(s[1], ":")
	if len(v) != 2 || v[0] == "" || v[1] == "" {
		// The attribute value string should have exactly two parts
		return nil, ErrInvalidAttribute
	}

	// Declare a proto message to hold the attribute value
	var wrapped proto.Message

	// Parse the attribute value based on its type
	switch v[0] {
	case "boolean":
		boolVal, err := strconv.ParseBool(v[1])
		if err != nil {
			return nil, fmt.Errorf("failed to parse boolean: %w", err)
		}
		wrapped = &base.BooleanValue{Data: boolVal}
	case "boolean[]":
		val := strings.Split(v[1], ",")
		ba := make([]bool, len(val))
		for i, value := range val {
			boolVal, err := strconv.ParseBool(value)
			if err != nil {
				return nil, fmt.Errorf("failed to parse boolean: %w", err)
			}
			ba[i] = boolVal
		}
		wrapped = &base.BooleanArrayValue{Data: ba}
	case "string":
		wrapped = &base.StringValue{Data: v[1]}
	case "string[]":
		sa := strings.Split(v[1], ",")
		wrapped = &base.StringArrayValue{Data: sa}
	case "double":
		doubleVal, err := strconv.ParseFloat(v[1], 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse float: %w", err)
		}
		wrapped = &base.DoubleValue{Data: doubleVal}
	case "double[]":
		val := strings.Split(v[1], ",")
		da := make([]float64, len(val))
		for i, value := range val {
			doubleVal, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse float: %v", err)
			}
			da[i] = doubleVal
		}
		wrapped = &base.DoubleArrayValue{Data: da}
	case "integer":
		intVal, err := strconv.ParseInt(v[1], 10, 32)
		if err != nil {
			return nil, fmt.Errorf("failed to parse integer: %v", err)
		}
		wrapped = &base.IntegerValue{Data: int32(intVal)}
	case "integer[]":
		val := strings.Split(v[1], ",")
		ia := make([]int32, len(val))
		for i, value := range val {
			intVal, err := strconv.ParseInt(value, 10, 32)
			if err != nil {
				return nil, fmt.Errorf("failed to parse integer: %v", err)
			}
			ia[i] = int32(intVal)
		}
		wrapped = &base.IntegerArrayValue{Data: ia}
	default:
		return nil, ErrInvalidValue
	}

	// Convert the wrapped attribute value into Any proto message
	value, err := anypb.New(wrapped)
	if err != nil {
		return nil, err
	}

	// Return the attribute object
	return &base.Attribute{
		Entity: &base.Entity{
			Type: et[0],
			Id:   et[1],
		},
		Attribute: e[1],
		Value:     value,
	}, nil
}

// ToString function takes an Attribute object and converts it into a string.
func ToString(attribute *base.Attribute) string {
	// Get the entity from the attribute
	entity := attribute.GetEntity()

	// Convert the entity to string
	strEntity := EntityToString(entity)

	// Create the string representation of the attribute
	result := fmt.Sprintf("%s$%s|%s:%s", strEntity, attribute.GetAttribute(), TypeUrlToString(attribute.GetValue().GetTypeUrl()), AnyToString(attribute.GetValue()))

	return result
}

// EntityAndAttributeToString converts an entity and attribute to a single string.
func EntityAndAttributeToString(entity *base.Entity, attr string) string {
	// Convert the entity to string format
	strEntity := EntityToString(entity)

	// Combine the entity string with the attribute using a dollar sign as the separator
	result := fmt.Sprintf("%s$%s", strEntity, attr)

	// Return the combined string
	return result
}

// EntityToString function takes an Entity object and converts it into a string.
func EntityToString(entity *base.Entity) string {
	return fmt.Sprintf(ENTITY, entity.GetType(), entity.GetId())
}

// EntityAndCallOrAttributeToString -
func EntityAndCallOrAttributeToString(entity *base.Entity, attributeOrCall string, arguments ...*base.Argument) string {
	return EntityToString(entity) + fmt.Sprintf(ATTRIBUTE, CallOrAttributeToString(attributeOrCall, arguments...))
}

// CallOrAttributeToString -
func CallOrAttributeToString(attributeOrCall string, arguments ...*base.Argument) string {
	if len(arguments) > 0 {
		var args []string
		for _, arg := range arguments {
			args = append(args, arg.GetComputedAttribute().GetName())
		}
		return fmt.Sprintf("%s(%s)", attributeOrCall, strings.Join(args, ","))
	}
	return attributeOrCall
}

func TypeUrlToString(url string) string {
	switch url {
	case "type.googleapis.com/base.v1.StringValue":
		return "string"
	case "type.googleapis.com/base.v1.BooleanValue":
		return "boolean"
	case "type.googleapis.com/base.v1.IntegerValue":
		return "integer"
	case "type.googleapis.com/base.v1.DoubleValue":
		return "double"
	case "type.googleapis.com/base.v1.StringArrayValue":
		return "string[]"
	case "type.googleapis.com/base.v1.BooleanArrayValue":
		return "boolean[]"
	case "type.googleapis.com/base.v1.IntegerArrayValue":
		return "integer[]"
	case "type.googleapis.com/base.v1.DoubleArrayValue":
		return "double[]"
	default:
		return ""
	}
}

// AnyToString function takes an Any proto message and converts it into a string.
func AnyToString(any *anypb.Any) string {
	var str string

	// Convert the Any proto message into string based on its TypeUrl
	switch any.TypeUrl {
	case "type.googleapis.com/base.v1.BooleanValue":
		boolVal := &base.BooleanValue{}
		if err := any.UnmarshalTo(boolVal); err != nil {
			return "undefined"
		}
		str = strconv.FormatBool(boolVal.Data)
	case "type.googleapis.com/base.v1.BooleanArrayValue":
		boolVal := &base.BooleanArrayValue{}
		if err := any.UnmarshalTo(boolVal); err != nil {
			return "undefined"
		}
		var strs []string
		for _, b := range boolVal.GetData() {
			strs = append(strs, strconv.FormatBool(b))
		}
		str = strings.Join(strs, ",")
	case "type.googleapis.com/base.v1.StringValue":
		stringVal := &base.StringValue{}
		if err := any.UnmarshalTo(stringVal); err != nil {
			return "undefined"
		}
		str = stringVal.Data
	case "type.googleapis.com/base.v1.StringArrayValue":
		stringVal := &base.StringArrayValue{}
		if err := any.UnmarshalTo(stringVal); err != nil {
			return "undefined"
		}
		str = strings.Join(stringVal.GetData(), ",")
	case "type.googleapis.com/base.v1.DoubleValue":
		doubleVal := &base.DoubleValue{}
		if err := any.UnmarshalTo(doubleVal); err != nil {
			return "undefined"
		}
		str = strconv.FormatFloat(doubleVal.Data, 'f', -1, 64)
	case "type.googleapis.com/base.v1.DoubleArrayValue":
		doubleVal := &base.DoubleArrayValue{}
		if err := any.UnmarshalTo(doubleVal); err != nil {
			return "undefined"
		}
		var strs []string
		for _, v := range doubleVal.GetData() {
			strs = append(strs, strconv.FormatFloat(v, 'f', -1, 64))
		}
		str = strings.Join(strs, ",")
	case "type.googleapis.com/base.v1.IntegerValue":
		intVal := &base.IntegerValue{}
		if err := any.UnmarshalTo(intVal); err != nil {
			return "undefined"
		}
		str = strconv.Itoa(int(intVal.Data))
	case "type.googleapis.com/base.v1.IntegerArrayValue":
		intVal := &base.IntegerArrayValue{}
		if err := any.UnmarshalTo(intVal); err != nil {
			return "undefined"
		}
		var strs []string
		for _, v := range intVal.GetData() {
			strs = append(strs, strconv.Itoa(int(v)))
		}
		str = strings.Join(strs, ",")
	default:
		return "undefined"
	}

	return str
}

// TypeToString function takes an AttributeType enum and converts it into a string.
func TypeToString(attributeType base.AttributeType) string {
	switch attributeType {
	case base.AttributeType_ATTRIBUTE_TYPE_INTEGER:
		return "integer"
	case base.AttributeType_ATTRIBUTE_TYPE_INTEGER_ARRAY:
		return "integer[]"
	case base.AttributeType_ATTRIBUTE_TYPE_DOUBLE:
		return "double"
	case base.AttributeType_ATTRIBUTE_TYPE_DOUBLE_ARRAY:
		return "double[]"
	case base.AttributeType_ATTRIBUTE_TYPE_STRING:
		return "string"
	case base.AttributeType_ATTRIBUTE_TYPE_STRING_ARRAY:
		return "string[]"
	case base.AttributeType_ATTRIBUTE_TYPE_BOOLEAN:
		return "boolean"
	case base.AttributeType_ATTRIBUTE_TYPE_BOOLEAN_ARRAY:
		return "boolean[]"
	default:
		return "undefined"
	}
}

// ValidateValue checks the validity of the 'any' parameter which is a protobuf 'Any' type,
// based on the attribute type provided.
//
// 'any' is a protobuf 'Any' type which should contain a value of a specific type.
// 'attributeType' is an enum indicating the expected type of the value within 'any'.
// The function returns an error if the value within 'any' is not of the expected type, or if unmarshalling fails.
//
// The function returns nil if the value is valid (i.e., it is of the expected type and can be successfully unmarshalled).
func ValidateValue(any *anypb.Any, attributeType base.AttributeType) error {
	// Declare a variable 'target' of type proto.Message to hold the unmarshalled value.
	var target proto.Message

	// Depending on the expected attribute type, assign 'target' a new instance of the corresponding specific type.
	switch attributeType {
	case base.AttributeType_ATTRIBUTE_TYPE_INTEGER:
		target = &base.IntegerValue{}
	case base.AttributeType_ATTRIBUTE_TYPE_INTEGER_ARRAY:
		target = &base.IntegerArrayValue{}
	case base.AttributeType_ATTRIBUTE_TYPE_DOUBLE:
		target = &base.DoubleValue{}
	case base.AttributeType_ATTRIBUTE_TYPE_DOUBLE_ARRAY:
		target = &base.DoubleArrayValue{}
	case base.AttributeType_ATTRIBUTE_TYPE_STRING:
		target = &base.StringValue{}
	case base.AttributeType_ATTRIBUTE_TYPE_STRING_ARRAY:
		target = &base.StringArrayValue{}
	case base.AttributeType_ATTRIBUTE_TYPE_BOOLEAN:
		target = &base.BooleanValue{}
	case base.AttributeType_ATTRIBUTE_TYPE_BOOLEAN_ARRAY:
		target = &base.BooleanArrayValue{}
	default:
		// If attributeType doesn't match any of the known types, return an error indicating invalid argument.
		return errors.New(base.ErrorCode_ERROR_CODE_INVALID_ARGUMENT.String())
	}

	// Attempt to unmarshal the value in 'any' into 'target'.
	// If this fails, return the error from UnmarshalTo.
	if err := any.UnmarshalTo(target); err != nil {
		return err
	}

	// If the value was successfully unmarshalled and is of the expected type, return nil to indicate success.
	return nil
}
