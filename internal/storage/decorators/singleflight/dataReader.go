package singleflight

import (
	"context"

	"resenje.org/singleflight"

	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/pkg/database"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/token"
)

// DataReader - Add singleflight behaviour to data reader
type DataReader struct {
	delegate storage.DataReader
	group    singleflight.Group[string, token.SnapToken]
}

// NewDataReader - Add singleflight behaviour to new data reader
func NewDataReader(delegate storage.DataReader) *DataReader {
	return &DataReader{delegate: delegate}
}

// QueryRelationships - Reads relation tuples from the repository
func (r *DataReader) QueryRelationships(ctx context.Context, tenantID string, filter *base.TupleFilter, token string, pagination database.CursorPagination) (*database.TupleIterator, error) {
	return r.delegate.QueryRelationships(ctx, tenantID, filter, token, pagination)
}

// ReadRelationships - Reads relation tuples from the repository with different options.
func (r *DataReader) ReadRelationships(ctx context.Context, tenantID string, filter *base.TupleFilter, token string, pagination database.Pagination) (collection *database.TupleCollection, ct database.EncodedContinuousToken, err error) {
	return r.delegate.ReadRelationships(ctx, tenantID, filter, token, pagination)
}

// QuerySingleAttribute - Reads a single attribute from the repository.
func (r *DataReader) QuerySingleAttribute(ctx context.Context, tenantID string, filter *base.AttributeFilter, token string) (*base.Attribute, error) {
	return r.delegate.QuerySingleAttribute(ctx, tenantID, filter, token)
}

// QueryAttributes - Reads multiple attributes from the repository.
func (r *DataReader) QueryAttributes(ctx context.Context, tenantID string, filter *base.AttributeFilter, token string, pagination database.CursorPagination) (*database.AttributeIterator, error) {
	return r.delegate.QueryAttributes(ctx, tenantID, filter, token, pagination)
}

// ReadAttributes - Reads multiple attributes from the repository with different options.
func (r *DataReader) ReadAttributes(ctx context.Context, tenantID string, filter *base.AttributeFilter, token string, pagination database.Pagination) (collection *database.AttributeCollection, ct database.EncodedContinuousToken, err error) {
	return r.delegate.ReadAttributes(ctx, tenantID, filter, token, pagination)
}

// QueryUniqueSubjectReferences - Reads unique subject references from the repository with different options.
func (r *DataReader) QueryUniqueSubjectReferences(ctx context.Context, tenantID string, subjectReference *base.RelationReference, excluded []string, token string, pagination database.Pagination) (ids []string, ct database.EncodedContinuousToken, err error) {
	return r.delegate.QueryUniqueSubjectReferences(ctx, tenantID, subjectReference, excluded, token, pagination)
}

// HeadSnapshot - Reads the latest version of the snapshot from the repository.
func (r *DataReader) HeadSnapshot(ctx context.Context, tenantID string) (token.SnapToken, error) {
	rev, _, err := r.group.Do(ctx, "", func(ctx context.Context) (token.SnapToken, error) {
		return r.delegate.HeadSnapshot(ctx, tenantID)
	})
	return rev, err
}
