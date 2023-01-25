package services

import (
	"context"

	"github.com/Permify/permify/internal/repositories"
	"github.com/Permify/permify/pkg/database"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// TenancyService -
type TenancyService struct {
	tr repositories.TenantReader
	tw repositories.TenantWriter
}

// NewTenancyService -
func NewTenancyService(tw repositories.TenantWriter, tr repositories.TenantReader) *TenancyService {
	return &TenancyService{
		tr: tr,
		tw: tw,
	}
}

// CreateTenant -
func (s *TenancyService) CreateTenant(ctx context.Context, id, name string) (tenant *base.Tenant, err error) {
	return s.tw.CreateTenant(ctx, id, name)
}

// DeleteTenant -
func (s *TenancyService) DeleteTenant(ctx context.Context, tenantID string) (tenant *base.Tenant, err error) {
	return s.tw.DeleteTenant(ctx, tenantID)
}

// ListTenants -
func (s *TenancyService) ListTenants(ctx context.Context, size uint32, ct string) (tenants []*base.Tenant, continuousToken database.EncodedContinuousToken, err error) {
	return s.tr.ListTenants(ctx, database.NewPagination(database.Size(size), database.Token(ct)))
}
