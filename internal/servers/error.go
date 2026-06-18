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

	// Check specific error codes that need precise gRPC status mappings before
	// falling through to the range-based checks below.
	switch base.ErrorCode(code) {
	case base.ErrorCode_ERROR_CODE_CANCELLED:
		// A cancelled operation is not an internal server error — map it to the
		// standard gRPC Canceled code so interceptors and clients treat it
		// correctly and don't log it as an unexpected ERROR.
		return codes.Canceled
	case base.ErrorCode_ERROR_CODE_NOT_IMPLEMENTED:
		return codes.Unimplemented
	case base.ErrorCode_ERROR_CODE_SERIALIZATION:
		// Serialization failures (e.g. optimistic-lock conflicts) are transient
		// and should be signalled as Aborted so callers can safely retry.
		return codes.Aborted
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
