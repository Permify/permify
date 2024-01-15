package servers

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// GetStatus - Get status code and message from error
func GetStatus(err error) codes.Code {
	s, ok := status.FromError(err)
	if ok {
		// This was a custom error, so return its code directly
		return s.Code()
	}

	// If this wasn't a custom error, continue with your existing logic...
	code, ok := base.ErrorCode_value[err.Error()]
	if !ok {
		return codes.Internal
	}
	switch {
	case code > 999 && code < 1999:
		return codes.Unauthenticated
	case code > 1999 && code < 2999:
		return codes.InvalidArgument
	case code > 3999 && code < 4999:
		return codes.NotFound
	case code > 4999 && code < 5999:
		return codes.Internal
	default:
		return codes.Internal
	}
}
