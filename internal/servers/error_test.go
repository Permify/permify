package servers

import (
	"errors"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	base "github.com/Permify/permify/pkg/pb/base/v1"
)

func TestGetStatus(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected codes.Code
	}{
		// Specific overrides that must not be swallowed by the 5xxx→Internal range.
		{
			name:     "ERROR_CODE_CANCELLED maps to codes.Canceled",
			err:      errors.New(base.ErrorCode_ERROR_CODE_CANCELLED.String()),
			expected: codes.Canceled,
		},
		{
			name:     "ERROR_CODE_NOT_IMPLEMENTED maps to codes.Unimplemented",
			err:      errors.New(base.ErrorCode_ERROR_CODE_NOT_IMPLEMENTED.String()),
			expected: codes.Unimplemented,
		},
		{
			name:     "ERROR_CODE_SERIALIZATION maps to codes.Aborted",
			err:      errors.New(base.ErrorCode_ERROR_CODE_SERIALIZATION.String()),
			expected: codes.Aborted,
		},
		// Range-based mappings.
		{
			name:     "authn error maps to codes.Unauthenticated",
			err:      errors.New(base.ErrorCode_ERROR_CODE_UNAUTHENTICATED.String()),
			expected: codes.Unauthenticated,
		},
		{
			name:     "validation error maps to codes.InvalidArgument",
			err:      errors.New(base.ErrorCode_ERROR_CODE_VALIDATION.String()),
			expected: codes.InvalidArgument,
		},
		{
			name:     "not-found error maps to codes.NotFound",
			err:      errors.New(base.ErrorCode_ERROR_CODE_NOT_FOUND.String()),
			expected: codes.NotFound,
		},
		{
			name:     "generic internal error maps to codes.Internal",
			err:      errors.New(base.ErrorCode_ERROR_CODE_INTERNAL.String()),
			expected: codes.Internal,
		},
		{
			name:     "unknown error string maps to codes.Internal",
			err:      errors.New("some unexpected error"),
			expected: codes.Internal,
		},
		// A pre-wrapped gRPC status error must be passed through unchanged.
		{
			name:     "existing gRPC status error is forwarded as-is",
			err:      status.Error(codes.PermissionDenied, "access denied"),
			expected: codes.PermissionDenied,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := GetStatus(tc.err)
			if got != tc.expected {
				t.Errorf("GetStatus(%q) = %v, want %v", tc.err, got, tc.expected)
			}
		})
	}
}
