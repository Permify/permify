package utils

import (
	"errors"
	"strings"

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
	case base.AttributeType_ATTRIBUTE_TYPE_BOOLEAN:
		return types.BoolType, nil
	case base.AttributeType_ATTRIBUTE_TYPE_INTEGER:
		return types.IntType, nil
	case base.AttributeType_ATTRIBUTE_TYPE_DOUBLE:
		return types.DoubleType, nil
	default:
		return nil, errors.New("")
	}
}
