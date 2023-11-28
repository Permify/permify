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

// DataReaderWithCircuitBreaker - Add circuit breaker behaviour to data reader
type DataReaderWithCircuitBreaker struct {
	delegate storage.DataReader
	timeout  int
}

// NewDataReaderWithCircuitBreaker - Add circuit breaker behaviour to new data reader
func NewDataReaderWithCircuitBreaker(delegate storage.DataReader, timeout int) *DataReaderWithCircuitBreaker {
	return &DataReaderWithCircuitBreaker{delegate: delegate, timeout: timeout}
}

// QueryRelationships - Reads relation tuples from the repository
func (r *DataReaderWithCircuitBreaker) QueryRelationships(ctx context.Context, tenantID string, filter *base.TupleFilter, token string) (*database.TupleIterator, error) {
	type circuitBreakerResponse struct {
		Iterator *database.TupleIterator
		Error    error
	}

	output := make(chan circuitBreakerResponse, 1)
	hystrix.ConfigureCommand("dataReader.queryRelationships", hystrix.CommandConfig{Timeout: r.timeout})
	bErrors := hystrix.Go("dataReader.queryRelationships", func() error {
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

// ReadRelationships - Reads relation tuples from the repository with different options.
func (r *DataReaderWithCircuitBreaker) ReadRelationships(ctx context.Context, tenantID string, filter *base.TupleFilter, token string, pagination database.Pagination) (collection *database.TupleCollection, ct database.EncodedContinuousToken, err error) {
	type circuitBreakerResponse struct {
		Collection      *database.TupleCollection
		ContinuousToken database.EncodedContinuousToken
		Error           error
	}

	output := make(chan circuitBreakerResponse, 1)
	hystrix.ConfigureCommand("dataReader.readRelationships", hystrix.CommandConfig{Timeout: r.timeout})
	bErrors := hystrix.Go("dataReader.readRelationships", func() error {
		tup, ct, err := r.delegate.ReadRelationships(ctx, tenantID, filter, token, pagination)
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

// QuerySingleAttribute - Reads a single attribute from the repository.
func (r *DataReaderWithCircuitBreaker) QuerySingleAttribute(ctx context.Context, tenantID string, filter *base.AttributeFilter, token string) (*base.Attribute, error) {
	type circuitBreakerResponse struct {
		Attribute *base.Attribute
		Error     error
	}

	output := make(chan circuitBreakerResponse, 1)
	hystrix.ConfigureCommand("dataReader.querySingleAttribute", hystrix.CommandConfig{Timeout: r.timeout})
	bErrors := hystrix.Go("dataReader.querySingleAttribute", func() error {
		attr, err := r.delegate.QuerySingleAttribute(ctx, tenantID, filter, token)
		output <- circuitBreakerResponse{Attribute: attr, Error: err}
		return nil
	}, func(err error) error {
		return nil
	})

	select {
	case out := <-output:
		return out.Attribute, out.Error
	case <-bErrors:
		return nil, errors.New(base.ErrorCode_ERROR_CODE_CIRCUIT_BREAKER.String())
	}
}

// QueryAttributes - Reads multiple attributes from the repository.
func (r *DataReaderWithCircuitBreaker) QueryAttributes(ctx context.Context, tenantID string, filter *base.AttributeFilter, token string) (*database.AttributeIterator, error) {
	type circuitBreakerResponse struct {
		Iterator *database.AttributeIterator
		Error    error
	}

	output := make(chan circuitBreakerResponse, 1)
	hystrix.ConfigureCommand("dataReader.queryAttributes", hystrix.CommandConfig{Timeout: r.timeout})
	bErrors := hystrix.Go("dataReader.queryAttributes", func() error {
		attr, err := r.delegate.QueryAttributes(ctx, tenantID, filter, token)
		output <- circuitBreakerResponse{Iterator: attr, Error: err}
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

// ReadAttributes - Reads multiple attributes from the repository with different options.
func (r *DataReaderWithCircuitBreaker) ReadAttributes(ctx context.Context, tenantID string, filter *base.AttributeFilter, token string, pagination database.Pagination) (collection *database.AttributeCollection, ct database.EncodedContinuousToken, err error) {
	type circuitBreakerResponse struct {
		Collection      *database.AttributeCollection
		ContinuousToken database.EncodedContinuousToken
		Error           error
	}

	output := make(chan circuitBreakerResponse, 1)
	hystrix.ConfigureCommand("dataReader.readAttributes", hystrix.CommandConfig{Timeout: r.timeout})
	bErrors := hystrix.Go("dataReader.readAttributes", func() error {
		attr, ct, err := r.delegate.ReadAttributes(ctx, tenantID, filter, token, pagination)
		output <- circuitBreakerResponse{Collection: attr, ContinuousToken: ct, Error: err}
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

// QueryUniqueEntities - Reads unique entities from the repository with different options.
func (r *DataReaderWithCircuitBreaker) QueryUniqueEntities(ctx context.Context, tenantID, name, token string, pagination database.Pagination) (ids []string, ct database.EncodedContinuousToken, err error) {
	type circuitBreakerResponse struct {
		IDs             []string
		ContinuousToken database.EncodedContinuousToken
		Error           error
	}

	output := make(chan circuitBreakerResponse, 1)
	hystrix.ConfigureCommand("dataReader.queryUniqueEntities", hystrix.CommandConfig{Timeout: r.timeout})
	bErrors := hystrix.Go("dataReader.queryUniqueEntities", func() error {
		ids, ct, err := r.delegate.QueryUniqueEntities(ctx, tenantID, name, token, pagination)
		output <- circuitBreakerResponse{IDs: ids, ContinuousToken: ct, Error: err}
		return nil
	}, func(err error) error {
		return nil
	})

	select {
	case out := <-output:
		return out.IDs, out.ContinuousToken, out.Error
	case <-bErrors:
		return nil, nil, errors.New(base.ErrorCode_ERROR_CODE_CIRCUIT_BREAKER.String())
	}
}

// QueryUniqueSubjectReferences - Reads unique subject references from the repository with different options.
func (r *DataReaderWithCircuitBreaker) QueryUniqueSubjectReferences(ctx context.Context, tenantID string, subjectReference *base.RelationReference, token string, pagination database.Pagination) (ids []string, ct database.EncodedContinuousToken, err error) {
	type circuitBreakerResponse struct {
		IDs             []string
		ContinuousToken database.EncodedContinuousToken
		Error           error
	}

	output := make(chan circuitBreakerResponse, 1)
	hystrix.ConfigureCommand("dataReader.queryUniqueSubjectReferences", hystrix.CommandConfig{Timeout: r.timeout})
	bErrors := hystrix.Go("dataReader.queryUniqueSubjectReferences", func() error {
		ids, ct, err := r.delegate.QueryUniqueSubjectReferences(ctx, tenantID, subjectReference, token, pagination)
		output <- circuitBreakerResponse{IDs: ids, ContinuousToken: ct, Error: err}
		return nil
	}, func(err error) error {
		return nil
	})

	select {
	case out := <-output:
		return out.IDs, out.ContinuousToken, out.Error
	case <-bErrors:
		return nil, nil, errors.New(base.ErrorCode_ERROR_CODE_CIRCUIT_BREAKER.String())
	}
}

// HeadSnapshot - Reads the latest version of the snapshot from the repository.
func (r *DataReaderWithCircuitBreaker) HeadSnapshot(ctx context.Context, tenantID string) (token.SnapToken, error) {
	type circuitBreakerResponse struct {
		Token token.SnapToken
		Error error
	}

	output := make(chan circuitBreakerResponse, 1)
	hystrix.ConfigureCommand("relationshipReader.headSnapshot", hystrix.CommandConfig{Timeout: r.timeout})
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
