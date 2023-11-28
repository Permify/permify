package decorators

import (
	"context"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/pkg/errors"

	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/pkg/database"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// TenantReaderWithCircuitBreaker - Add circuit breaker behaviour to tenant reader
type TenantReaderWithCircuitBreaker struct {
	delegate storage.TenantReader
	timeout  int
}

// NewTenantReaderWithCircuitBreaker - Add circuit breaker behaviour to new tenant reader
func NewTenantReaderWithCircuitBreaker(delegate storage.TenantReader, timeout int) *TenantReaderWithCircuitBreaker {
	return &TenantReaderWithCircuitBreaker{delegate: delegate, timeout: timeout}
}

// ListTenants - List tenants from the repository
func (r *TenantReaderWithCircuitBreaker) ListTenants(ctx context.Context, pagination database.Pagination) (tenants []*base.Tenant, ct database.EncodedContinuousToken, err error) {
	type circuitBreakerResponse struct {
		Tenants []*base.Tenant
		Ct      database.EncodedContinuousToken
		Error   error
	}

	output := make(chan circuitBreakerResponse, 1)
	hystrix.ConfigureCommand("tenantReader.listTenants", hystrix.CommandConfig{Timeout: r.timeout})
	bErrors := hystrix.Go("tenantReader.listTenants", func() error {
		tenants, ct, err := r.delegate.ListTenants(ctx, pagination)
		output <- circuitBreakerResponse{Tenants: tenants, Ct: ct, Error: err}
		return nil
	}, func(err error) error {
		return nil
	})

	select {
	case out := <-output:
		return out.Tenants, out.Ct, out.Error
	case <-bErrors:
		return nil, nil, errors.New(base.ErrorCode_ERROR_CODE_CIRCUIT_BREAKER.String())
	}
}
