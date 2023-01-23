package commands

import (
	"errors"

	"go.opentelemetry.io/otel"

	base "github.com/Permify/permify/pkg/pb/base/v1"
)

var tracer = otel.Tracer("commands")

const (
	_defaultConcurrencyLimit = 100
)

// CheckOption - Option type
type CheckOption func(*CheckCommand)

// ConcurrencyLimit - Defines concurrency limit
func ConcurrencyLimit(limit int) CheckOption {
	return func(c *CheckCommand) {
		c.concurrencyLimit = limit
	}
}

// joinResponseMetas -
func joinResponseMetas(meta ...*base.PermissionCheckResponseMetadata) *base.PermissionCheckResponseMetadata {
	response := &base.PermissionCheckResponseMetadata{}
	for _, m := range meta {
		response.CheckCount += m.CheckCount
	}
	return response
}

// increaseCheckCount -
func increaseCheckCount(metadata *base.PermissionCheckResponseMetadata) *base.PermissionCheckResponseMetadata {
	return &base.PermissionCheckResponseMetadata{
		CheckCount: metadata.CheckCount + 1,
	}
}

// CheckResponse -
type CheckResponse struct {
	resp *base.PermissionCheckResponse
	err  error
}

// checkDepth -
func checkDepth(request *base.PermissionCheckRequest) error {
	if request.GetMetadata().Depth == 0 {
		return errors.New(base.ErrorCode_ERROR_CODE_DEPTH_NOT_ENOUGH.String())
	}
	return nil
}
