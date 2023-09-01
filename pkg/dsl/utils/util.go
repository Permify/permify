package utils

import (
	"errors"
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"

	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// Key -
func Key(v1, v2 string) string {
	var sb strings.Builder
	sb.WriteString(v1)
	sb.WriteString("#")
	sb.WriteString(v2)
	return sb.String()
}

func ArgumentsAsCelEnv(arguments map[string]base.AttributeType) (*cel.Env, error) {
	opts := make([]cel.EnvOption, 0, len(arguments))
	for name, typ := range arguments {
		typ, err := GetCelType(typ)
		if err != nil {
			return nil, err
		}

		opts = append(opts, cel.Variable(name, typ))
	}
	return cel.NewEnv(opts...)
}

func GetCelType(attributeType base.AttributeType) (*types.Type, error) {
	switch attributeType {
	case base.AttributeType_ATTRIBUTE_TYPE_STRING:
		return types.StringType, nil
	case base.AttributeType_ATTRIBUTE_TYPE_STRING_ARRAY:
		return cel.ListType(cel.StringType), nil
	case base.AttributeType_ATTRIBUTE_TYPE_BOOLEAN:
		return types.BoolType, nil
	case base.AttributeType_ATTRIBUTE_TYPE_BOOLEAN_ARRAY:
		return cel.ListType(types.BoolType), nil
	case base.AttributeType_ATTRIBUTE_TYPE_INTEGER:
		return types.IntType, nil
	case base.AttributeType_ATTRIBUTE_TYPE_INTEGER_ARRAY:
		return cel.ListType(types.IntType), nil
	case base.AttributeType_ATTRIBUTE_TYPE_DOUBLE:
		return types.DoubleType, nil
	case base.AttributeType_ATTRIBUTE_TYPE_DOUBLE_ARRAY:
		return cel.ListType(types.DoubleType), nil
	default:
		return nil, errors.New("")
	}
}

func ConvertProtoAnyToInterface(a *anypb.Any) interface{} {
	switch a.GetTypeUrl() {
	case "type.googleapis.com/base.v1.String":
		stringValue := &base.String{}
		if err := anypb.UnmarshalTo(a, stringValue, proto.UnmarshalOptions{}); err != nil {
			return ""
		}
		return stringValue.GetValue()
	case "type.googleapis.com/base.v1.Boolean":
		boolValue := &base.Boolean{}
		if err := anypb.UnmarshalTo(a, boolValue, proto.UnmarshalOptions{}); err != nil {
			return false
		}
		return boolValue.GetValue()
	case "type.googleapis.com/base.v1.Integer":
		integerValue := &base.Integer{}
		if err := anypb.UnmarshalTo(a, integerValue, proto.UnmarshalOptions{}); err != nil {
			return 0
		}
		return integerValue.GetValue()
	case "type.googleapis.com/base.v1.Double":
		doubleValue := &base.Double{}
		if err := anypb.UnmarshalTo(a, doubleValue, proto.UnmarshalOptions{}); err != nil {
			return 0.0
		}
		return doubleValue.GetValue()
	case "type.googleapis.com/base.v1.StringArray":
		stringArrayValue := &base.StringArray{}
		if err := anypb.UnmarshalTo(a, stringArrayValue, proto.UnmarshalOptions{}); err != nil {
			return []string{}
		}
		return stringArrayValue.GetValues()
	case "type.googleapis.com/base.v1.BooleanArray":
		booleanArrayValue := &base.BooleanArray{}
		if err := anypb.UnmarshalTo(a, booleanArrayValue, proto.UnmarshalOptions{}); err != nil {
			return []bool{}
		}
		return booleanArrayValue.GetValues()
	case "type.googleapis.com/base.v1.IntegerArray":
		integerArrayValue := &base.IntegerArray{}
		if err := anypb.UnmarshalTo(a, integerArrayValue, proto.UnmarshalOptions{}); err != nil {
			return []int32{}
		}
		return integerArrayValue.GetValues()
	case "type.googleapis.com/base.v1.DoubleArray":
		doubleArrayValue := &base.DoubleArray{}
		if err := anypb.UnmarshalTo(a, doubleArrayValue, proto.UnmarshalOptions{}); err != nil {
			return []float64{}
		}
		return doubleArrayValue.GetValues()
	default:
		return ""
	}
}
