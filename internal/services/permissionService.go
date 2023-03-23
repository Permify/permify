package services

import (
	"context"

	"github.com/Permify/permify/internal/engines"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// PermissionService -
type PermissionService struct {
	// engines
	cc *engines.CheckEngine
	ec *engines.ExpandEngine
	ls *engines.LookupSchemaEngine
	le *engines.LookupEntityEngine
}

// NewPermissionService -
func NewPermissionService(cc *engines.CheckEngine, ec *engines.ExpandEngine, ls *engines.LookupSchemaEngine, le *engines.LookupEntityEngine) *PermissionService {
	return &PermissionService{
		cc: cc,
		ec: ec,
		ls: ls,
		le: le,
	}
}

// CheckPermissions -
func (service *PermissionService) CheckPermissions(ctx context.Context, request *base.PermissionCheckRequest) (response *base.PermissionCheckResponse, err error) {
	return service.cc.Run(ctx, request)
}

// ExpandPermissions -
func (service *PermissionService) ExpandPermissions(ctx context.Context, request *base.PermissionExpandRequest) (response *base.PermissionExpandResponse, err error) {
	return service.ec.Run(ctx, request)
}

// LookupSchema -
func (service *PermissionService) LookupSchema(ctx context.Context, request *base.PermissionLookupSchemaRequest) (response *base.PermissionLookupSchemaResponse, err error) {
	return service.ls.Run(ctx, request)
}

// LookupEntity -
func (service *PermissionService) LookupEntity(ctx context.Context, request *base.PermissionLookupEntityRequest) (response *base.PermissionLookupEntityResponse, err error) {
	return service.le.Run(ctx, request)
}

// LookupEntityStream -
func (service *PermissionService) LookupEntityStream(ctx context.Context, request *base.PermissionLookupEntityRequest, server base.Permission_LookupEntityStreamServer) (err error) {
	return service.le.Stream(ctx, request, server)
}
