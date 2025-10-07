package invoke

import (
	"errors"
	"sync/atomic"

	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// checkDepth validates that the request has sufficient depth for permission checks
func checkDepth(request *base.PermissionCheckRequest) error {
	if atomic.LoadInt32(&request.GetMetadata().Depth) <= 0 { // Check depth is positive
		return errors.New(base.ErrorCode_ERROR_CODE_DEPTH_NOT_ENOUGH.String())
	}
	return nil
}
