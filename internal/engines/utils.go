package engines

import (
	"errors"
	"sync"

	"go.opentelemetry.io/otel"

	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/tuple"
)

var tracer = otel.Tracer("engines")

const (
	_defaultConcurrencyLimit = 100
)

// CheckOption - a functional option type for configuring the CheckEngine.
type CheckOption func(engine *CheckEngine)

// CheckConcurrencyLimit - a functional option that sets the concurrency limit for the CheckEngine.
func CheckConcurrencyLimit(limit int) CheckOption {
	return func(c *CheckEngine) {
		c.concurrencyLimit = limit
	}
}

// LookupEntityOption - a functional option type for configuring the LookupEntityEngine.
type LookupEntityOption func(engine *LookupEntityEngine)

// LookupEntityConcurrencyLimit - a functional option that sets the concurrency limit for the LookupEntityEngine.
func LookupEntityConcurrencyLimit(limit int) LookupEntityOption {
	return func(c *LookupEntityEngine) {
		c.concurrencyLimit = limit
	}
}

// joinResponseMetas - a helper function that merges multiple PermissionCheckResponseMetadata structs into one.
func joinResponseMetas(meta ...*base.PermissionCheckResponseMetadata) *base.PermissionCheckResponseMetadata {
	response := &base.PermissionCheckResponseMetadata{}
	for _, m := range meta {
		response.CheckCount += m.CheckCount
	}
	return response
}

// increaseCheckCount - a helper function that increments the check count in a PermissionCheckResponseMetadata struct.
func increaseCheckCount(metadata *base.PermissionCheckResponseMetadata) *base.PermissionCheckResponseMetadata {
	return &base.PermissionCheckResponseMetadata{
		CheckCount: metadata.CheckCount + 1,
	}
}

// CheckResponse - a struct that holds a PermissionCheckResponse and an error for a single check function.
type CheckResponse struct {
	resp *base.PermissionCheckResponse
	err  error
}

// checkDepth - a helper function that returns an error if the depth in a PermissionCheckRequest is zero.
func checkDepth(request *base.PermissionCheckRequest) error {
	if request.GetMetadata().Depth == 0 {
		return errors.New(base.ErrorCode_ERROR_CODE_DEPTH_NOT_ENOUGH.String())
	}
	return nil
}

// ERMap - a thread-safe map of ENR records.
type ERMap struct {
	value sync.Map
}

func (s *ERMap) Add(onr *base.EntityAndRelation) bool {
	key := tuple.EntityAndRelationToString(onr)
	_, existed := s.value.LoadOrStore(key, struct{}{})
	return !existed
}
