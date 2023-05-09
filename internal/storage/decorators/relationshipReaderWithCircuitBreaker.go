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

// RelationshipReaderWithCircuitBreaker - Add circuit breaker behaviour to relationship reader
type RelationshipReaderWithCircuitBreaker struct {
	delegate storage.RelationshipReader
}

// NewRelationshipReaderWithCircuitBreaker - Add circuit breaker behaviour to new relationship reader
func NewRelationshipReaderWithCircuitBreaker(delegate storage.RelationshipReader) *RelationshipReaderWithCircuitBreaker {
	return &RelationshipReaderWithCircuitBreaker{delegate: delegate}
}

// QueryRelationships - Reads relation tuples from the repository
func (r *RelationshipReaderWithCircuitBreaker) QueryRelationships(ctx context.Context, tenantID string, filter *base.TupleFilter, token string) (*database.TupleIterator, error) {
	type circuitBreakerResponse struct {
		Iterator *database.TupleIterator
		Error    error
	}

	output := make(chan circuitBreakerResponse, 1)
	hystrix.ConfigureCommand("relationshipReader.queryRelationships", hystrix.CommandConfig{Timeout: 1000})
	bErrors := hystrix.Go("relationshipReader.queryRelationships", func() error {
		tup, err := r.delegate.QueryRelationships(ctx, tenantID, filter, token)
		output <- circuitBreakerResponse{Iterator: tup, Error: err}
		return nil
	}, func(err error) error {
		return nil
	})

	select {
	case out := <-output:
		return out.Iterator, out.Error
	case <-bErrors:
		return nil, errors.New(base.ErrorCode_ERROR_CODE_CIRCUIT_BREAKER.String())
	}
}

// ReadRelationships reads relation tuples from the repository with different options.
func (r *RelationshipReaderWithCircuitBreaker) ReadRelationships(ctx context.Context, tenantID string, filter *base.TupleFilter, snap string, pagination database.Pagination) (collection *database.TupleCollection, ct database.EncodedContinuousToken, err error) {
	type circuitBreakerResponse struct {
		Collection      *database.TupleCollection
		ContinuousToken database.EncodedContinuousToken
		Error           error
	}

	output := make(chan circuitBreakerResponse, 1)
	hystrix.ConfigureCommand("relationshipReader.readRelationships", hystrix.CommandConfig{Timeout: 1000})
	bErrors := hystrix.Go("relationshipReader.readRelationships", func() error {
		tup, ct, err := r.delegate.ReadRelationships(ctx, tenantID, filter, snap, pagination)
		output <- circuitBreakerResponse{Collection: tup, ContinuousToken: ct, Error: err}
		return nil
	}, func(err error) error {
		return nil
	})

	select {
	case out := <-output:
		return out.Collection, out.ContinuousToken, out.Error
	case <-bErrors:
		return nil, nil, errors.New(base.ErrorCode_ERROR_CODE_CIRCUIT_BREAKER.String())
	}
}

// HeadSnapshot - Reads the latest version of the snapshot from the repository.
func (r *RelationshipReaderWithCircuitBreaker) HeadSnapshot(ctx context.Context, tenantID string) (token.SnapToken, error) {
	type circuitBreakerResponse struct {
		Token token.SnapToken
		Error error
	}

	output := make(chan circuitBreakerResponse, 1)
	hystrix.ConfigureCommand("relationshipReader.headSnapshot", hystrix.CommandConfig{Timeout: 1000})
	bErrors := hystrix.Go("relationshipReader.headSnapshot", func() error {
		tok, err := r.delegate.HeadSnapshot(ctx, tenantID)
		output <- circuitBreakerResponse{Token: tok, Error: err}
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
