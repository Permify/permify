package servers

import (
	"google.golang.org/grpc/codes"

	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// GetStatus - Get error status
func GetStatus(err error) codes.Code {
	code, ok := base.ErrorCode_value[err.Error()]
	if !ok {
		return codes.Internal
	}
	switch {
	case code > 999 && code < 1999:
		return codes.Unauthenticated
	case code > 1999 && code < 2999:
		return codes.InvalidArgument
	case code > 2999 && code < 3999:
		return codes.NotFound
	case code > 3999 && code < 4999:
		return codes.Internal
	default:
		return codes.Internal
	}
}
