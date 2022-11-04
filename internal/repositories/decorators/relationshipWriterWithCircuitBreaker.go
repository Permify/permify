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

// RelationshipWriterWithCircuitBreaker -
type RelationshipWriterWithCircuitBreaker struct {
	delegate repositories.RelationshipWriter
}

// NewRelationshipWriterWithCircuitBreaker -.
func NewRelationshipWriterWithCircuitBreaker(delegate repositories.RelationshipWriter) *RelationshipWriterWithCircuitBreaker {
	return &RelationshipWriterWithCircuitBreaker{delegate: delegate}
}

// WriteRelationships -
func (r *RelationshipWriterWithCircuitBreaker) WriteRelationships(ctx context.Context, collection database.ITupleCollection) (token.SnapToken, error) {
	type circuitBreakerResponse struct {
		Token token.SnapToken
		Error error
	}

	output := make(chan circuitBreakerResponse, 1)

	hystrix.ConfigureCommand("relationshipWriter.writeRelationships", hystrix.CommandConfig{Timeout: 1000})
	bErrors := hystrix.Go("relationshipWriter.writeRelationships", func() error {
		t, err := r.delegate.WriteRelationships(ctx, collection)
		output <- circuitBreakerResponse{Token: t, Error: err}
		return nil
	}, func(err error) error {
		return nil
	})

	select {
	case out := <-output:
		return out.Token, out.Error
	case <-bErrors:
		return token.SnapToken{}, errors.New(base.ErrorCode_ERROR_CODE_CIRCUIT_BREAKER.String())
	}
}

// DeleteRelationships -
func (r *RelationshipWriterWithCircuitBreaker) DeleteRelationships(ctx context.Context, filter *base.TupleFilter) (token.SnapToken, error) {
	type circuitBreakerResponse struct {
		Token token.SnapToken
		Error error
	}

	output := make(chan circuitBreakerResponse, 1)

	hystrix.ConfigureCommand("relationshipWriter.deleteRelationships", hystrix.CommandConfig{Timeout: 1000})
	bErrors := hystrix.Go("relationshipWriter.deleteRelationships", func() error {
		t, err := r.delegate.DeleteRelationships(ctx, filter)
		output <- circuitBreakerResponse{Token: t, Error: err}
		return nil
	}, func(err error) error {
		return nil
	})

	select {
	case out := <-output:
		return out.Token, out.Error
	case <-bErrors:
		return token.SnapToken{}, errors.New(base.ErrorCode_ERROR_CODE_CIRCUIT_BREAKER.String())
	}
}
