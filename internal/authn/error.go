package authn

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	base "github.com/Permify/permify/pkg/pb/base/v1"
)

var (
	Unauthenticated         = status.Error(codes.Code(base.ErrorCode_ERROR_CODE_UNAUTHENTICATED), "unauthenticated")
	MissingBearerTokenError = status.Error(codes.Code(base.ErrorCode_ERROR_CODE_MISSING_BEARER_TOKEN), "missing bearer token")
	// MissingTenantIDError    = status.Error(codes.Code(base.ErrorCode_ERROR_CODE_MISSING_TENANT_ID), "missing tenant id")
)
