package circuitBreaker

import (
	"context"

	"github.com/sony/gobreaker"

	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/pkg/database"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// TenantReader - Add circuit breaker behaviour to tenant reader
type TenantReader struct {
	delegate storage.TenantReader
	cb       *gobreaker.CircuitBreaker
}

// NewTenantReader - Add circuit breaker behaviour to new tenant reader
func NewTenantReader(delegate storage.TenantReader, cb *gobreaker.CircuitBreaker) *TenantReader {
	return &TenantReader{delegate: delegate, cb: cb}
}

// ListTenants - List tenants from the repository
func (r *TenantReader) ListTenants(ctx context.Context, pagination database.Pagination) (tenants []*base.Tenant, ct database.EncodedContinuousToken, err error) {
	type circuitBreakerResponse struct {
		Tenants []*base.Tenant
		Ct      database.EncodedContinuousToken
	}

	response, err := r.cb.Execute(func() (interface{}, error) {
		var err error
		var resp circuitBreakerResponse
		resp.Tenants, resp.Ct, err = r.delegate.ListTenants(ctx, pagination)
		return resp, err
	})
	if err != nil {
		return nil, nil, err
	}

	resp := response.(circuitBreakerResponse)
	return resp.Tenants, resp.Ct, nil
}
