package commands

import (
	"errors"

	base "github.com/Permify/permify/pkg/pb/base/v1"
)

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

// decreaseDepth -
func decreaseDepth(md *base.PermissionCheckRequestMetadata) *base.PermissionCheckRequestMetadata {
	return &base.PermissionCheckRequestMetadata{
		SchemaVersion: md.GetSchemaVersion(),
		Exclusion:     md.GetExclusion(),
		SnapToken:     md.GetSnapToken(),
		Depth:         md.Depth - 1,
	}
}

// checkDepth -
func checkDepth(request *base.PermissionCheckRequest) error {
	if request.GetMetadata().Depth == 0 {
		return errors.New(base.ErrorCode_ERROR_CODE_DEPTH_NOT_ENOUGH.String())
	}
	return nil
}
