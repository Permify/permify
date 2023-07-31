package attribute

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	base "github.com/Permify/permify/pkg/pb/base/v1"
)

const (
	ENTITY    = "%s:%s"
	ATTRIBUTE = "#%s"
	VALUE     = "%s:%s"
)

// Attribute function takes a string representation of an attribute and converts it back into the Attribute object.
func Attribute(attribute string) (*base.Attribute, error) {
	// Splitting the attribute string by "@" delimiter
	s := strings.Split(strings.TrimSpace(attribute), "@")
	if len(s) != 2 {
		// The attribute string should have exactly two parts
		return nil, ErrInvalidAttribute
	}

	// Splitting the entity part of the string by "#" delimiter
	e := strings.Split(s[0], "#")
	if len(e) != 2 {
		// The entity string should have exactly two parts
		return nil, ErrInvalidEntity
	}

	// Splitting the entity type and id by ":" delimiter
	et := strings.Split(e[0], ":")
	if len(et) != 2 {
		// The entity type and id should have exactly two parts
		return nil, ErrInvalidAttribute
	}

	// Splitting the attribute value part of the string by ":" delimiter
	v := strings.Split(s[1], ":")
	if len(v) != 2 {
		// The attribute value string should have exactly two parts
		return nil, ErrInvalidAttribute
	}

	// Declare a proto message to hold the attribute value
	var wrapped proto.Message

	// Parse the attribute value based on its type
	switch v[0] {
	// In case of boolean
	case "boolean":
		boolVal, err := strconv.ParseBool(v[1])
		if err != nil {
			return nil, fmt.Errorf("failed to parse boolean: %v", err)
		}
		wrapped = wrapperspb.Bool(boolVal)
	// In case of string
	case "string":
		wrapped = wrapperspb.String(v[1])
	// In case of double
	case "double":
		doubleVal, err := strconv.ParseFloat(v[1], 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse float: %v", err)
		}
		wrapped = wrapperspb.Double(doubleVal)
	// In case of integer
	case "integer":
		intVal, err := strconv.Atoi(v[1])
		if err != nil {
			return nil, fmt.Errorf("failed to parse integer: %v", err)
		}
		wrapped = wrapperspb.Int32(int32(intVal))
	// In case the type is none of the above
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
		Type:      v[0],
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
	result := fmt.Sprintf("%s#%s@%s:%s", strEntity, attribute.GetAttribute(), attribute.GetType(), AnyToString(attribute.GetValue()))

	return result
}

// EntityToString function takes an Entity object and converts it into a string.
func EntityToString(entity *base.Entity) string {
	return fmt.Sprintf(ENTITY, entity.GetType(), entity.GetId())
}

// AnyToString function takes an Any proto message and converts it into a string.
func AnyToString(any *anypb.Any) string {
	var str string

	// Convert the Any proto message into string based on its TypeUrl
	switch any.TypeUrl {
	case "type.googleapis.com/google.protobuf.BoolValue":
		boolVal := &wrapperspb.BoolValue{}
		if err := any.UnmarshalTo(boolVal); err != nil {
			return "undefined"
		}
		str = strconv.FormatBool(boolVal.Value)
	case "type.googleapis.com/google.protobuf.StringValue":
		stringVal := &wrapperspb.StringValue{}
		if err := any.UnmarshalTo(stringVal); err != nil {
			return "undefined"
		}
		str = stringVal.Value
	case "type.googleapis.com/google.protobuf.DoubleValue":
		doubleVal := &wrapperspb.DoubleValue{}
		if err := any.UnmarshalTo(doubleVal); err != nil {
			return "undefined"
		}
		str = strconv.FormatFloat(doubleVal.Value, 'f', -1, 64)
	case "type.googleapis.com/google.protobuf.Int32Value":
		intVal := &wrapperspb.Int32Value{}
		if err := any.UnmarshalTo(intVal); err != nil {
			return "undefined"
		}
		str = strconv.Itoa(int(intVal.Value))
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
	case base.AttributeType_ATTRIBUTE_TYPE_DOUBLE:
		return "double"
	case base.AttributeType_ATTRIBUTE_TYPE_STRING:
		return "string"
	case base.AttributeType_ATTRIBUTE_TYPE_BOOLEAN:
		return "boolean"
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
		target = &wrapperspb.Int32Value{} // Expected integer type
	case base.AttributeType_ATTRIBUTE_TYPE_DOUBLE:
		target = &wrapperspb.DoubleValue{} // Expected double type
	case base.AttributeType_ATTRIBUTE_TYPE_STRING:
		target = &wrapperspb.StringValue{} // Expected string type
	case base.AttributeType_ATTRIBUTE_TYPE_BOOLEAN:
		target = &wrapperspb.BoolValue{} // Expected boolean type
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
