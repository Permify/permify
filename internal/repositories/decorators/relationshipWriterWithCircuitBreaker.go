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

// RelationshipWriterWithCircuitBreaker - Add circuit breaker behaviour to relationship writer
type RelationshipWriterWithCircuitBreaker struct {
	delegate repositories.RelationshipWriter
}

// NewRelationshipWriterWithCircuitBreaker - Add circuit breaker behaviour to new relationship writer
func NewRelationshipWriterWithCircuitBreaker(delegate repositories.RelationshipWriter) *RelationshipWriterWithCircuitBreaker {
	return &RelationshipWriterWithCircuitBreaker{delegate: delegate}
}

// WriteRelationships - Write relation tuples from the repository
func (r *RelationshipWriterWithCircuitBreaker) WriteRelationships(ctx context.Context, tenantID string, collection *database.TupleCollection) (token.EncodedSnapToken, error) {
	type circuitBreakerResponse struct {
		Token token.EncodedSnapToken
		Error error
	}

	output := make(chan circuitBreakerResponse, 1)

	hystrix.ConfigureCommand("relationshipWriter.writeRelationships", hystrix.CommandConfig{Timeout: 1000})
	bErrors := hystrix.Go("relationshipWriter.writeRelationships", func() error {
		t, err := r.delegate.WriteRelationships(ctx, tenantID, collection)
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

// DeleteRelationships - Delete relation tuples from the repository
func (r *RelationshipWriterWithCircuitBreaker) DeleteRelationships(ctx context.Context, tenantID string, filter *base.TupleFilter) (token.EncodedSnapToken, error) {
	type circuitBreakerResponse struct {
		Token token.EncodedSnapToken
		Error error
	}

	output := make(chan circuitBreakerResponse, 1)

	hystrix.ConfigureCommand("relationshipWriter.deleteRelationships", hystrix.CommandConfig{Timeout: 1000})
	bErrors := hystrix.Go("relationshipWriter.deleteRelationships", func() error {
		t, err := r.delegate.DeleteRelationships(ctx, tenantID, filter)
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
