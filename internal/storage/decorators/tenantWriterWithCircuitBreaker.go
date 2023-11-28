package decorators

import (
	"context"
	"errors"

	"github.com/afex/hystrix-go/hystrix"

	"github.com/Permify/permify/internal/storage"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// TenantWriterWithCircuitBreaker - Add circuit breaker behaviour to tenant writer
type TenantWriterWithCircuitBreaker struct {
	delegate storage.TenantWriter
	timeout  int
}

// NewTenantWriterWithCircuitBreaker - Add circuit breaker behaviour to new bundle reader
func NewTenantWriterWithCircuitBreaker(delegate storage.TenantWriter, timeout int) *TenantWriterWithCircuitBreaker {
	return &TenantWriterWithCircuitBreaker{delegate: delegate, timeout: timeout}
}

// CreateTenant - Create tenant from the repository
func (r *TenantWriterWithCircuitBreaker) CreateTenant(ctx context.Context, id, name string) (result *base.Tenant, err error) {
	type circuitBreakerResponse struct {
		Tenant *base.Tenant
		Error  error
	}

	output := make(chan circuitBreakerResponse, 1)
	hystrix.ConfigureCommand("tenantWriter.createTenant", hystrix.CommandConfig{Timeout: r.timeout})
	bErrors := hystrix.Go("tenantWriter.createTenant", func() error {
		tenant, err := r.delegate.CreateTenant(ctx, id, name)
		output <- circuitBreakerResponse{Tenant: tenant, Error: err}
		return nil
	}, func(err error) error {
		return nil
	})

	select {
	case out := <-output:
		return out.Tenant, out.Error
	case <-bErrors:
		return nil, errors.New(base.ErrorCode_ERROR_CODE_CIRCUIT_BREAKER.String())
	}
}

// DeleteTenant - Delete tenant from the repository
func (r *TenantWriterWithCircuitBreaker) DeleteTenant(ctx context.Context, tenantID string) (result *base.Tenant, err error) {
	type circuitBreakerResponse struct {
		Tenant *base.Tenant
		Error  error
	}

	output := make(chan circuitBreakerResponse, 1)
	hystrix.ConfigureCommand("tenantWriter.deleteTenant", hystrix.CommandConfig{Timeout: r.timeout})
	bErrors := hystrix.Go("tenantWriter.deleteTenant", func() error {
		tenant, err := r.delegate.DeleteTenant(ctx, tenantID)
		output <- circuitBreakerResponse{Tenant: tenant, Error: err}
		return nil
	}, func(err error) error {
		return nil
	})

	select {
	case out := <-output:
		return out.Tenant, out.Error
	case <-bErrors:
		return nil, errors.New(base.ErrorCode_ERROR_CODE_CIRCUIT_BREAKER.String())
	}
}
