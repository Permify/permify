package invoke

import (
	"errors"

	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// checkDepth - a helper function that returns an error if the depth in a PermissionCheckRequest is zero.
func checkDepth(request *base.PermissionCheckRequest) error {
	if request.GetMetadata().Depth == 0 {
		return errors.New(base.ErrorCode_ERROR_CODE_DEPTH_NOT_ENOUGH.String())
	}
	return nil
}

// increaseCheckCount - a helper function that increments the check count in a PermissionCheckResponseMetadata struct.
func increaseCheckCount(metadata *base.PermissionCheckResponseMetadata) *base.PermissionCheckResponseMetadata {
	return &base.PermissionCheckResponseMetadata{
		CheckCount: metadata.CheckCount + 1,
	}
}

// increaseCheckCount - a helper function that increments the check count in a PermissionCheckResponseMetadata struct.
func decreaseDepth(metadata *base.PermissionCheckRequestMetadata) *base.PermissionCheckRequestMetadata {
	return &base.PermissionCheckRequestMetadata{
		SchemaVersion: metadata.GetSchemaVersion(),
		SnapToken:     metadata.GetSnapToken(),
		Depth:         metadata.Depth - 1,
	}
}
