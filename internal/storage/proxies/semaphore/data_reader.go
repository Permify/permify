package semaphore

import (
	"context"

	"github.com/Permify/permify/internal/invoke"
	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/pkg/database"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/token"
)

// DataReader wraps a storage.DataReader with per-request semaphore protection.
// Each DB query acquires a slot from the request-scoped semaphore (stored in context),
// preventing connection pool exhaustion from concurrent goroutines.
type DataReader struct {
	delegate storage.DataReader
}

// NewDataReader creates a semaphore-protected DataReader.
func NewDataReader(delegate storage.DataReader) *DataReader {
	return &DataReader{delegate: delegate}
}

func (r *DataReader) QueryRelationships(ctx context.Context, tenantID string, filter *base.TupleFilter, snap string, pagination database.CursorPagination) (*database.TupleIterator, error) {
	sem := invoke.ConcurrencySemaphoreFromContext(ctx)
	if err := sem.Acquire(ctx, 1); err != nil {
		return nil, err
	}
	defer sem.Release(1)
	return r.delegate.QueryRelationships(ctx, tenantID, filter, snap, pagination)
}

func (r *DataReader) QueryRelationshipsWithSubjectFilter(ctx context.Context, tenantID string, filter *base.TupleFilter, subject *base.Subject, snap string, pagination database.CursorPagination) (*database.TupleIterator, error) {
	sem := invoke.ConcurrencySemaphoreFromContext(ctx)
	if err := sem.Acquire(ctx, 1); err != nil {
		return nil, err
	}
	defer sem.Release(1)
	return r.delegate.QueryRelationshipsWithSubjectFilter(ctx, tenantID, filter, subject, snap, pagination)
}

func (r *DataReader) ReadRelationships(ctx context.Context, tenantID string, filter *base.TupleFilter, snap string, pagination database.Pagination) (*database.TupleCollection, database.EncodedContinuousToken, error) {
	sem := invoke.ConcurrencySemaphoreFromContext(ctx)
	if err := sem.Acquire(ctx, 1); err != nil {
		return nil, nil, err
	}
	defer sem.Release(1)
	return r.delegate.ReadRelationships(ctx, tenantID, filter, snap, pagination)
}

func (r *DataReader) QuerySingleAttribute(ctx context.Context, tenantID string, filter *base.AttributeFilter, snap string) (*base.Attribute, error) {
	sem := invoke.ConcurrencySemaphoreFromContext(ctx)
	if err := sem.Acquire(ctx, 1); err != nil {
		return nil, err
	}
	defer sem.Release(1)
	return r.delegate.QuerySingleAttribute(ctx, tenantID, filter, snap)
}

func (r *DataReader) QueryAttributes(ctx context.Context, tenantID string, filter *base.AttributeFilter, snap string, pagination database.CursorPagination) (*database.AttributeIterator, error) {
	sem := invoke.ConcurrencySemaphoreFromContext(ctx)
	if err := sem.Acquire(ctx, 1); err != nil {
		return nil, err
	}
	defer sem.Release(1)
	return r.delegate.QueryAttributes(ctx, tenantID, filter, snap, pagination)
}

func (r *DataReader) ReadAttributes(ctx context.Context, tenantID string, filter *base.AttributeFilter, snap string, pagination database.Pagination) (*database.AttributeCollection, database.EncodedContinuousToken, error) {
	sem := invoke.ConcurrencySemaphoreFromContext(ctx)
	if err := sem.Acquire(ctx, 1); err != nil {
		return nil, nil, err
	}
	defer sem.Release(1)
	return r.delegate.ReadAttributes(ctx, tenantID, filter, snap, pagination)
}

func (r *DataReader) QueryUniqueSubjectReferences(ctx context.Context, tenantID string, subjectReference *base.RelationReference, excluded []string, snap string, pagination database.Pagination) ([]string, database.EncodedContinuousToken, error) {
	sem := invoke.ConcurrencySemaphoreFromContext(ctx)
	if err := sem.Acquire(ctx, 1); err != nil {
		return nil, nil, err
	}
	defer sem.Release(1)
	return r.delegate.QueryUniqueSubjectReferences(ctx, tenantID, subjectReference, excluded, snap, pagination)
}

func (r *DataReader) HeadSnapshot(ctx context.Context, tenantID string) (token.SnapToken, error) {
	sem := invoke.ConcurrencySemaphoreFromContext(ctx)
	if err := sem.Acquire(ctx, 1); err != nil {
		return nil, err
	}
	defer sem.Release(1)
	return r.delegate.HeadSnapshot(ctx, tenantID)
}
