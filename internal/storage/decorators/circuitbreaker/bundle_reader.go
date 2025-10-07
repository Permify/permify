package circuitBreaker

import (
	"context"

	"github.com/sony/gobreaker"

	"github.com/Permify/permify/internal/storage"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// BundleReader - Add circuit breaker behaviour to bundle reader
type BundleReader struct {
	delegate storage.BundleReader
	cb       *gobreaker.CircuitBreaker
}

// NewBundleReader - Add circuit breaker behaviour to new bundle reader
func NewBundleReader(delegate storage.BundleReader, cb *gobreaker.CircuitBreaker) *BundleReader {
	return &BundleReader{delegate: delegate, cb: cb}
}

// Read - Reads bundles from the repository
func (r *BundleReader) Read(ctx context.Context, tenantID, name string) (bundle *base.DataBundle, err error) {
	response, err := r.cb.Execute(func() (interface{}, error) {
		return r.delegate.Read(ctx, tenantID, name)
	})
	if err != nil {
		return nil, err
	}
	return response.(*base.DataBundle), nil
}
