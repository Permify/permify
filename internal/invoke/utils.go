package invoke

import (
	"errors"

	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// checkDepth validates that the request has sufficient depth for permission checks
func checkDepth(request *BatchCheckRequest) error {
	if request.Metadata.GetDepth() < 0 {
		return errors.New(base.ErrorCode_ERROR_CODE_DEPTH_NOT_ENOUGH.String())
	}
	return nil
}
