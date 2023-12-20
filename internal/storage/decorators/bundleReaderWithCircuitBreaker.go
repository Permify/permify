package decorators

import (
	"context"
	"errors"

	"github.com/afex/hystrix-go/hystrix"

	"github.com/Permify/permify/internal/storage"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// BundleReaderWithCircuitBreaker - Add circuit breaker behaviour to bundle reader
type BundleReaderWithCircuitBreaker struct {
	delegate storage.BundleReader
	timeout  int
}

// NewBundleReaderWithCircuitBreaker - Add circuit breaker behaviour to new bundle reader
func NewBundleReaderWithCircuitBreaker(delegate storage.BundleReader, timeout int) *BundleReaderWithCircuitBreaker {
	return &BundleReaderWithCircuitBreaker{delegate: delegate, timeout: timeout}
}

// Read - Reads bundles from the repository
func (r *BundleReaderWithCircuitBreaker) Read(ctx context.Context, tenantID, name string) (bundle *base.DataBundle, err error) {
	type circuitBreakerResponse struct {
		Bundle *base.DataBundle
		Error  error
	}

	output := make(chan circuitBreakerResponse, 1)
	hystrix.ConfigureCommand("bundleReader.read", hystrix.CommandConfig{Timeout: r.timeout})
	bErrors := hystrix.Go("bundleReader.read", func() error {
		bundle, err := r.delegate.Read(ctx, tenantID, name)
		output <- circuitBreakerResponse{Bundle: bundle, Error: err}
		return nil
	}, func(err error) error {
		return nil
	})

	select {
	case out := <-output:
		return out.Bundle, out.Error
	case <-bErrors:
		return nil, errors.New(base.ErrorCode_ERROR_CODE_CIRCUIT_BREAKER.String())
	}
}
