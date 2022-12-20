package services

import (
	"context"
	
	"github.com/Permify/permify/internal/commands"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// PermissionService -
type PermissionService struct {
	// commands
	cc commands.ICheckCommand
	ec commands.IExpandCommand
	ls commands.ILookupSchemaCommand
	le commands.ILookupEntityCommand
}

// NewPermissionService -
func NewPermissionService(cc commands.ICheckCommand, ec commands.IExpandCommand, ls commands.ILookupSchemaCommand, le commands.ILookupEntityCommand) *PermissionService {
	return &PermissionService{
		cc: cc,
		ec: ec,
		ls: ls,
		le: le,
	}
}

// CheckPermissions -
func (service *PermissionService) CheckPermissions(ctx context.Context, request *base.PermissionCheckRequest) (response *base.PermissionCheckResponse, err error) {
	return service.cc.Execute(ctx, request)
}

// ExpandPermissions -
func (service *PermissionService) ExpandPermissions(ctx context.Context, request *base.PermissionExpandRequest) (response *base.PermissionExpandResponse, err error) {
	return service.ec.Execute(ctx, request)
}

// LookupSchema -
func (service *PermissionService) LookupSchema(ctx context.Context, request *base.PermissionLookupSchemaRequest) (response *base.PermissionLookupSchemaResponse, err error) {
	return service.ls.Execute(ctx, request)
}

// LookupEntity -
func (service *PermissionService) LookupEntity(ctx context.Context, request *base.PermissionLookupEntityRequest) (response *base.PermissionLookupEntityResponse, err error) {
	return service.le.Execute(ctx, request)
}

// LookupEntityStream -
func (service *PermissionService) LookupEntityStream(ctx context.Context, request *base.PermissionLookupEntityRequest, server base.Permission_LookupEntityStreamServer) (err error) {
	return service.le.Stream(ctx, request, server)
}
