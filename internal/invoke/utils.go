package invoke

import (
	"errors"
	"sync/atomic"

	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// checkDepth - a helper function that returns an error if the depth in a PermissionCheckRequest is <= zero.
func checkDepth(request *base.PermissionCheckRequest) error {
	if atomic.LoadInt32(&request.GetMetadata().Depth) <= 0 {
		return errors.New(base.ErrorCode_ERROR_CODE_DEPTH_NOT_ENOUGH.String())
	}
	return nil
}
