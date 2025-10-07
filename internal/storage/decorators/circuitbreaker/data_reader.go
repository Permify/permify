package circuitBreaker

import (
	"context"

	"github.com/sony/gobreaker"

	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/pkg/database"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/token"
)

// DataReader - Add circuit breaker behaviour to data reader
type DataReader struct {
	delegate storage.DataReader
	cb       *gobreaker.CircuitBreaker
}

// NewDataReader - Add circuit breaker behaviour to new data reader
func NewDataReader(delegate storage.DataReader, cb *gobreaker.CircuitBreaker) *DataReader {
	return &DataReader{delegate: delegate, cb: cb}
}

// QueryRelationships - Reads relation tuples from the repository
func (r *DataReader) QueryRelationships(ctx context.Context, tenantID string, filter *base.TupleFilter, token string, pagination database.CursorPagination) (*database.TupleIterator, error) {
	response, err := r.cb.Execute(func() (interface{}, error) {
		return r.delegate.QueryRelationships(ctx, tenantID, filter, token, pagination)
	})
	if err != nil {
		return nil, err
	}
	return response.(*database.TupleIterator), nil
}

// ReadRelationships - Reads relation tuples from the repository with different options.
func (r *DataReader) ReadRelationships(ctx context.Context, tenantID string, filter *base.TupleFilter, token string, pagination database.Pagination) (collection *database.TupleCollection, ct database.EncodedContinuousToken, err error) {
	type circuitBreakerResponse struct {
		Collection      *database.TupleCollection
		ContinuousToken database.EncodedContinuousToken
	}

	response, err := r.cb.Execute(func() (interface{}, error) {
		var err error
		var resp circuitBreakerResponse
		resp.Collection, resp.ContinuousToken, err = r.delegate.ReadRelationships(ctx, tenantID, filter, token, pagination)
		return resp, err
	})
	if err != nil {
		return nil, nil, err
	}

	resp := response.(circuitBreakerResponse)
	return resp.Collection, resp.ContinuousToken, nil
}

// QuerySingleAttribute - Reads a single attribute from the repository.
func (r *DataReader) QuerySingleAttribute(ctx context.Context, tenantID string, filter *base.AttributeFilter, token string) (*base.Attribute, error) {
	response, err := r.cb.Execute(func() (interface{}, error) {
		return r.delegate.QuerySingleAttribute(ctx, tenantID, filter, token)
	})
	if err != nil {
		return nil, err
	}
	return response.(*base.Attribute), nil
}

// QueryAttributes - Reads multiple attributes from the repository.
func (r *DataReader) QueryAttributes(ctx context.Context, tenantID string, filter *base.AttributeFilter, token string, pagination database.CursorPagination) (*database.AttributeIterator, error) {
	response, err := r.cb.Execute(func() (interface{}, error) {
		return r.delegate.QueryAttributes(ctx, tenantID, filter, token, pagination)
	})
	if err != nil {
		return nil, err
	}
	return response.(*database.AttributeIterator), nil
}

// ReadAttributes - Reads multiple attributes from the repository with different options.
func (r *DataReader) ReadAttributes(ctx context.Context, tenantID string, filter *base.AttributeFilter, token string, pagination database.Pagination) (collection *database.AttributeCollection, ct database.EncodedContinuousToken, err error) {
	type circuitBreakerResponse struct {
		Collection      *database.AttributeCollection
		ContinuousToken database.EncodedContinuousToken
	}

	response, err := r.cb.Execute(func() (interface{}, error) {
		var err error
		var resp circuitBreakerResponse
		resp.Collection, resp.ContinuousToken, err = r.delegate.ReadAttributes(ctx, tenantID, filter, token, pagination)
		return resp, err
	})
	if err != nil {
		return nil, nil, err
	}

	resp := response.(circuitBreakerResponse)
	return resp.Collection, resp.ContinuousToken, nil
}

// QueryUniqueSubjectReferences - Reads unique subject references from the repository with different options.
func (r *DataReader) QueryUniqueSubjectReferences(ctx context.Context, tenantID string, subjectReference *base.RelationReference, excluded []string, token string, pagination database.Pagination) (ids []string, ct database.EncodedContinuousToken, err error) {
	type circuitBreakerResponse struct {
		IDs             []string
		ContinuousToken database.EncodedContinuousToken
	}

	response, err := r.cb.Execute(func() (interface{}, error) {
		var err error
		var resp circuitBreakerResponse
		resp.IDs, resp.ContinuousToken, err = r.delegate.QueryUniqueSubjectReferences(ctx, tenantID, subjectReference, excluded, token, pagination)
		return resp, err
	})
	if err != nil {
		return nil, nil, err
	}

	resp := response.(circuitBreakerResponse)
	return resp.IDs, resp.ContinuousToken, nil
}

// HeadSnapshot - Reads the latest version of the snapshot from the repository.
func (r *DataReader) HeadSnapshot(ctx context.Context, tenantID string) (token.SnapToken, error) {
	response, err := r.cb.Execute(func() (interface{}, error) {
		return r.delegate.HeadSnapshot(ctx, tenantID)
	})
	if err != nil {
		return nil, err
	}
	return response.(token.SnapToken), nil
}
