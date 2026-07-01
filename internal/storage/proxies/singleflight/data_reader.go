package singleflight

import (
	"context"
	"fmt"
	"strings"

	"resenje.org/singleflight"

	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/pkg/database"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/token"
)

// DataReader - Add singleflight behaviour to data reader
type DataReader struct {
	delegate                storage.DataReader
	headSnapshotGroup       singleflight.Group[string, token.SnapToken]
	queryRelationshipsGroup singleflight.Group[string, []*base.Tuple]
	querySingleAttrGroup    singleflight.Group[string, *base.Attribute]
	queryAttributesGroup    singleflight.Group[string, []*base.Attribute]
}

// NewDataReader - Add singleflight behaviour to new data reader
func NewDataReader(delegate storage.DataReader) *DataReader {
	return &DataReader{delegate: delegate}
}

// QueryRelationships - Reads relation tuples from the repository with singleflight deduplication.
func (r *DataReader) QueryRelationships(ctx context.Context, tenantID string, filter *base.TupleFilter, token string, pagination database.CursorPagination) (*database.TupleIterator, error) {
	key := queryRelationshipsKey(tenantID, filter, token, pagination)
	tuples, _, err := r.queryRelationshipsGroup.Do(ctx, key, func(ctx context.Context) ([]*base.Tuple, error) {
		it, err := r.delegate.QueryRelationships(ctx, tenantID, filter, token, pagination)
		if err != nil {
			return nil, err
		}
		return drainTupleIterator(it), nil
	})
	if err != nil {
		return nil, err
	}
	return database.NewTupleIterator(tuples...), nil
}

// ReadRelationships - Reads relation tuples from the repository with different options.
func (r *DataReader) ReadRelationships(ctx context.Context, tenantID string, filter *base.TupleFilter, token string, pagination database.Pagination) (collection *database.TupleCollection, ct database.EncodedContinuousToken, err error) {
	return r.delegate.ReadRelationships(ctx, tenantID, filter, token, pagination)
}

// QuerySingleAttribute - Reads a single attribute from the repository with singleflight deduplication.
func (r *DataReader) QuerySingleAttribute(ctx context.Context, tenantID string, filter *base.AttributeFilter, token string) (*base.Attribute, error) {
	key := querySingleAttributeKey(tenantID, filter, token)
	attr, _, err := r.querySingleAttrGroup.Do(ctx, key, func(ctx context.Context) (*base.Attribute, error) {
		return r.delegate.QuerySingleAttribute(ctx, tenantID, filter, token)
	})
	return attr, err
}

// QueryAttributes - Reads multiple attributes from the repository with singleflight deduplication.
func (r *DataReader) QueryAttributes(ctx context.Context, tenantID string, filter *base.AttributeFilter, token string, pagination database.CursorPagination) (*database.AttributeIterator, error) {
	key := queryAttributesKey(tenantID, filter, token, pagination)
	attrs, _, err := r.queryAttributesGroup.Do(ctx, key, func(ctx context.Context) ([]*base.Attribute, error) {
		it, err := r.delegate.QueryAttributes(ctx, tenantID, filter, token, pagination)
		if err != nil {
			return nil, err
		}
		return drainAttributeIterator(it), nil
	})
	if err != nil {
		return nil, err
	}
	return database.NewAttributeIterator(attrs...), nil
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
	rev, _, err := r.headSnapshotGroup.Do(ctx, tenantID, func(ctx context.Context) (token.SnapToken, error) {
		return r.delegate.HeadSnapshot(ctx, tenantID)
	})
	return rev, err
}

// --- key builders ---

func queryRelationshipsKey(tenantID string, filter *base.TupleFilter, token string, pagination database.CursorPagination) string {
	var b strings.Builder
	fmt.Fprintf(&b, "qr\x00%q\x00%q\x00%#v\x00%q\x00%q\x00%#v\x00%q\x00%q\x00%q\x00%q\x00%d",
		tenantID,
		filter.GetEntity().GetType(),
		filter.GetEntity().GetIds(),
		filter.GetRelation(),
		filter.GetSubject().GetType(),
		filter.GetSubject().GetIds(),
		filter.GetSubject().GetRelation(),
		token,
		pagination.Cursor(),
		pagination.Sort(),
		pagination.Limit(),
	)
	return b.String()
}

func querySingleAttributeKey(tenantID string, filter *base.AttributeFilter, token string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "qsa\x00%q\x00%q\x00%#v\x00%#v\x00%q",
		tenantID,
		filter.GetEntity().GetType(),
		filter.GetEntity().GetIds(),
		filter.GetAttributes(),
		token,
	)
	return b.String()
}

func queryAttributesKey(tenantID string, filter *base.AttributeFilter, token string, pagination database.CursorPagination) string {
	var b strings.Builder
	fmt.Fprintf(&b, "qa\x00%q\x00%q\x00%#v\x00%#v\x00%q\x00%q\x00%q\x00%d",
		tenantID,
		filter.GetEntity().GetType(),
		filter.GetEntity().GetIds(),
		filter.GetAttributes(),
		token,
		pagination.Cursor(),
		pagination.Sort(),
		pagination.Limit(),
	)
	return b.String()
}

// --- iterator helpers ---

func drainTupleIterator(it *database.TupleIterator) []*base.Tuple {
	var tuples []*base.Tuple
	for it.HasNext() {
		tuples = append(tuples, it.GetNext())
	}
	return tuples
}

func drainAttributeIterator(it *database.AttributeIterator) []*base.Attribute {
	var attrs []*base.Attribute
	for it.HasNext() {
		attrs = append(attrs, it.GetNext())
	}
	return attrs
}
