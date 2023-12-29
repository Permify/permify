package decorators

import (
	"context"
	"errors"

	"github.com/afex/hystrix-go/hystrix"

	"github.com/Permify/permify/internal/storage"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// BundleWriterWithCircuitBreaker - Add circuit breaker behaviour to bundle writer
type BundleWriterWithCircuitBreaker struct {
	delegate storage.BundleWriter
	timeout  int
}

// NewBundleWriterWithCircuitBreaker - Add circuit breaker behaviour to new bundle writer
func NewBundleWriterWithCircuitBreaker(delegate storage.BundleWriter, timeout int) *BundleWriterWithCircuitBreaker {
	return &BundleWriterWithCircuitBreaker{delegate: delegate, timeout: timeout}
}

// Write - Write bundles from the repository
func (r *BundleWriterWithCircuitBreaker) Write(ctx context.Context, bundles []storage.Bundle) (names []string, err error) {
	type circuitBreakerResponse struct {
		Names []string
		Error error
	}

	output := make(chan circuitBreakerResponse, 1)
	hystrix.ConfigureCommand("bundleWriter.write", hystrix.CommandConfig{Timeout: r.timeout})
	bErrors := hystrix.Go("bundleWriter.write", func() error {
		names, err := r.delegate.Write(ctx, bundles)
		output <- circuitBreakerResponse{Names: names, Error: err}
		return nil
	}, func(err error) error {
		return nil
	})

	select {
	case out := <-output:
		return out.Names, out.Error
	case <-bErrors:
		return nil, errors.New(base.ErrorCode_ERROR_CODE_CIRCUIT_BREAKER.String())
	}
}

// Delete - Delete bundles from the repository
func (r *BundleWriterWithCircuitBreaker) Delete(ctx context.Context, tenantID, name string) (err error) {
	type circuitBreakerResponse struct {
		Error error
	}

	output := make(chan circuitBreakerResponse, 1)
	hystrix.ConfigureCommand("bundleWriter.write", hystrix.CommandConfig{Timeout: r.timeout})
	bErrors := hystrix.Go("bundleWriter.write", func() error {
		err := r.delegate.Delete(ctx, tenantID, name)
		output <- circuitBreakerResponse{Error: err}
		return nil
	}, func(err error) error {
		return nil
	})

	select {
	case out := <-output:
		return out.Error
	case <-bErrors:
		return errors.New(base.ErrorCode_ERROR_CODE_CIRCUIT_BREAKER.String())
	}
}
