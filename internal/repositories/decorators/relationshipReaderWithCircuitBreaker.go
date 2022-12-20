package decorators

import (
	"context"
	"errors"

	"github.com/afex/hystrix-go/hystrix"

	"github.com/Permify/permify/internal/repositories"
	"github.com/Permify/permify/pkg/database"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/token"
)

// RelationshipReaderWithCircuitBreaker - Add circuit breaker behaviour to relationship reader
type RelationshipReaderWithCircuitBreaker struct {
	delegate repositories.RelationshipReader
}

// NewRelationshipReaderWithCircuitBreaker - Add circuit breaker behaviour to new relationship reader
func NewRelationshipReaderWithCircuitBreaker(delegate repositories.RelationshipReader) *RelationshipReaderWithCircuitBreaker {
	return &RelationshipReaderWithCircuitBreaker{delegate: delegate}
}

// QueryRelationships - Reads relation tuples from the repository
func (r *RelationshipReaderWithCircuitBreaker) QueryRelationships(ctx context.Context, filter *base.TupleFilter, token string) (database.ITupleCollection, error) {
	type circuitBreakerResponse struct {
		Collection database.ITupleCollection
		Error      error
	}

	output := make(chan circuitBreakerResponse, 1)
	hystrix.ConfigureCommand("relationshipReader.queryRelationships", hystrix.CommandConfig{Timeout: 1000})
	bErrors := hystrix.Go("relationshipReader.queryRelationships", func() error {
		tup, err := r.delegate.QueryRelationships(ctx, filter, token)
		output <- circuitBreakerResponse{Collection: tup, Error: err}
		return nil
	}, func(err error) error {
		return nil
	})

	select {
	case out := <-output:
		return out.Collection, out.Error
	case <-bErrors:
		return nil, errors.New(base.ErrorCode_ERROR_CODE_CIRCUIT_BREAKER.String())
	}
}

// GetUniqueEntityIDsByEntityType - Reads relation tuples from the repository
func (r *RelationshipReaderWithCircuitBreaker) GetUniqueEntityIDsByEntityType(ctx context.Context, typ, token string) (array []string, err error) {
	type circuitBreakerResponse struct {
		IDs   []string
		Error error
	}

	output := make(chan circuitBreakerResponse, 1)
	hystrix.ConfigureCommand("relationshipReader.queryRelationships", hystrix.CommandConfig{Timeout: 1000})
	bErrors := hystrix.Go("relationshipReader.queryRelationships", func() error {
		ids, err := r.delegate.GetUniqueEntityIDsByEntityType(ctx, typ, token)
		output <- circuitBreakerResponse{IDs: ids, Error: err}
		return nil
	}, func(err error) error {
		return nil
	})

	select {
	case out := <-output:
		return out.IDs, out.Error
	case <-bErrors:
		return nil, errors.New(base.ErrorCode_ERROR_CODE_CIRCUIT_BREAKER.String())
	}
}

// HeadSnapshot - Reads the latest version of the snapshot from the repository.
func (r *RelationshipReaderWithCircuitBreaker) HeadSnapshot(ctx context.Context) (token.SnapToken, error) {
	type circuitBreakerResponse struct {
		Token token.SnapToken
		Error error
	}

	output := make(chan circuitBreakerResponse, 1)
	hystrix.ConfigureCommand("relationshipReader.headSnapshot", hystrix.CommandConfig{Timeout: 1000})
	bErrors := hystrix.Go("relationshipReader.headSnapshot", func() error {
		tok, err := r.delegate.HeadSnapshot(ctx)
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
