package decorators

import (
	"context"
	"errors"

	"github.com/afex/hystrix-go/hystrix"

	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/pkg/database"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/token"
)

// DataWriterWithCircuitBreaker - Add circuit breaker behaviour to data writer
type DataWriterWithCircuitBreaker struct {
	delegate storage.DataWriter
}

// NewDataWriterWithCircuitBreaker - Add circuit breaker behaviour to new data writer
func NewDataWriterWithCircuitBreaker(delegate storage.DataWriter) *DataWriterWithCircuitBreaker {
	return &DataWriterWithCircuitBreaker{delegate: delegate}
}

// WriteRelationships - Write relation tuples from the repository
func (r *DataWriterWithCircuitBreaker) Write(ctx context.Context, tenantID string, tupleCollection *database.TupleCollection, attributeCollection *database.AttributeCollection) (token.EncodedSnapToken, error) {
	type circuitBreakerResponse struct {
		Token token.EncodedSnapToken
		Error error
	}

	output := make(chan circuitBreakerResponse, 1)

	hystrix.ConfigureCommand("dataWriter.write", hystrix.CommandConfig{Timeout: 1000})
	bErrors := hystrix.Go("dataWriter.write", func() error {
		t, err := r.delegate.Write(ctx, tenantID, tupleCollection, attributeCollection)
		output <- circuitBreakerResponse{Token: t, Error: err}
		return nil
	}, func(err error) error {
		return nil
	})

	select {
	case out := <-output:
		return out.Token, out.Error
	case <-bErrors:
		return nil, errors.New(base.ErrorCode_ERROR_CODE_CIRCUIT_BREAKER.String())
	}
}

// Delete - Delete relation tuples and attributes from the repository
func (r *DataWriterWithCircuitBreaker) Delete(ctx context.Context, tenantID string, tupleFilter *base.TupleFilter, attrFilter *base.AttributeFilter) (token.EncodedSnapToken, error) {
	type circuitBreakerResponse struct {
		Token token.EncodedSnapToken
		Error error
	}

	output := make(chan circuitBreakerResponse, 1)

	hystrix.ConfigureCommand("dataWriter.deleteRelationships", hystrix.CommandConfig{Timeout: 1000})
	bErrors := hystrix.Go("dataWriter.deleteRelationships", func() error {
		t, err := r.delegate.Delete(ctx, tenantID, tupleFilter, attrFilter)
		output <- circuitBreakerResponse{Token: t, Error: err}
		return nil
	}, func(err error) error {
		return nil
	})

	select {
	case out := <-output:
		return out.Token, out.Error
	case <-bErrors:
		return nil, errors.New(base.ErrorCode_ERROR_CODE_CIRCUIT_BREAKER.String())
	}
}
