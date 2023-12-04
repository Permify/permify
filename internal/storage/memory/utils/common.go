package utils

import (
	"errors"
	"log/slog"
	"sync"

	base "github.com/Permify/permify/pkg/pb/base/v1"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var (
	RelationTuplesID AutoIncForRelationTuples
	AttributesID     AutoIncForRelationTuples
)

type AutoIncForRelationTuples struct {
	sync.Mutex
	id uint64
}

// ID -
func (a *AutoIncForRelationTuples) ID() (id uint64) {
	a.Lock()
	defer a.Unlock()
	if a.id == 0 {
		a.id++
	}
	id = a.id
	a.id++
	return
}

type AutoIncForAttributes struct {
	sync.Mutex
	id uint64
}

// ID -
func (a *AutoIncForAttributes) ID() (id uint64) {
	a.Lock()
	defer a.Unlock()
	if a.id == 0 {
		a.id++
	}
	id = a.id
	a.id++
	return
}

// GetRelationTuplesIndexNameAndArgsByFilters - Get index name and arguments by filters
func GetRelationTuplesIndexNameAndArgsByFilters(tenantID string, filter *base.TupleFilter) (string, []any) {
	if filter.GetEntity().GetType() != "" && filter.GetRelation() != "" {
		return "entity-type-and-relation-index", []any{tenantID, filter.GetEntity().GetType(), filter.GetRelation()}
	}
	if filter.GetEntity().GetType() != "" {
		return "entity-type-index", []any{tenantID, filter.GetEntity().GetType()}
	}
	return "id", nil
}

func GetAttributesIndexNameAndArgsByFilters(tenantID string, filter *base.AttributeFilter) (string, []any) {
	if filter.GetEntity().GetType() != "" {
		return "entity-type-index", []any{tenantID, filter.GetEntity().GetType()}
	}
	return "id", nil
}

// Union - Get union of two slices
func Union(a, b []string) []string {
	elements := make(map[string]bool)
	for _, item := range a {
		elements[item] = true
	}

	for _, item := range b {
		elements[item] = true
	}

	var union []string
	for key := range elements {
		union = append(union, key)
	}

	return union
}

func HandleError(span trace.Span, err error, errorCode base.ErrorCode) error {
	// Record the error on the span
	span.RecordError(err)

	// Set the status of the span
	span.SetStatus(codes.Error, err.Error())

	// Log the error
	slog.Error("Error encountered", slog.Any("error", err), slog.Any("errorCode", errorCode))

	// Return a new standardized error with the provided error code
	return errors.New(errorCode.String())
}
