package invoke

import (
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// BatchCheckResponse holds per-entity check results.
// Each entity ID maps to its individual CheckResult.
type BatchCheckResponse struct {
	Results  map[string]base.CheckResult // entityID → result
	Metadata *base.PermissionCheckResponseMetadata
}

// NewBatchCheckResponse creates a response from a set of entity IDs with the same result.
func NewBatchCheckResponse(result base.CheckResult, entityIDs ...string) *BatchCheckResponse {
	results := make(map[string]base.CheckResult, len(entityIDs))
	for _, id := range entityIDs {
		results[id] = result
	}
	return &BatchCheckResponse{
		Results:  results,
		Metadata: &base.PermissionCheckResponseMetadata{},
	}
}

// IsAllowed returns true if ANY entity in the batch was allowed (union semantics).
func (r *BatchCheckResponse) IsAllowed() bool {
	for _, result := range r.Results {
		if result == base.CheckResult_CHECK_RESULT_ALLOWED {
			return true
		}
	}
	return false
}

// UnionResult returns a single CheckResult with union semantics.
func (r *BatchCheckResponse) UnionResult() base.CheckResult {
	if r.IsAllowed() {
		return base.CheckResult_CHECK_RESULT_ALLOWED
	}
	return base.CheckResult_CHECK_RESULT_DENIED
}

// ToPermissionCheckResponse converts to a single-entity proto response (union semantics).
// Used at the public API boundary.
func (r *BatchCheckResponse) ToPermissionCheckResponse() *base.PermissionCheckResponse {
	return &base.PermissionCheckResponse{
		Can:      r.UnionResult(),
		Metadata: r.Metadata,
	}
}

// Merge combines results from another response into this one.
func (r *BatchCheckResponse) Merge(other *BatchCheckResponse) {
	for id, result := range other.Results {
		r.Results[id] = result
	}
	if other.Metadata != nil {
		r.Metadata.CheckCount += other.Metadata.CheckCount
	}
}

// BatchCheckRequest is the internal batch-capable check request.
// Unlike PermissionCheckRequest which has a single Entity, this holds
// multiple entity IDs of the same type, enabling IN (...) SQL queries.
type BatchCheckRequest struct {
	TenantID   string
	EntityType string
	EntityIDs  []string // batch: multiple entity IDs of the same type
	Permission string
	Subject    *base.Subject
	Metadata   *base.PermissionCheckRequestMetadata
	Context    *base.Context
	Arguments  []*base.Argument
}

// NewBatchCheckRequest creates a BatchCheckRequest from a single-entity PermissionCheckRequest.
func NewBatchCheckRequest(req *base.PermissionCheckRequest) *BatchCheckRequest {
	return &BatchCheckRequest{
		TenantID:   req.GetTenantId(),
		EntityType: req.GetEntity().GetType(),
		EntityIDs:  []string{req.GetEntity().GetId()},
		Permission: req.GetPermission(),
		Subject:    req.GetSubject(),
		Metadata:   req.GetMetadata(),
		Context:    req.GetContext(),
		Arguments:  req.GetArguments(),
	}
}

// ToPermissionCheckRequest converts back to a single-entity request for the given entity ID.
// Used for cache key generation, single-entity invoke, etc.
func (r *BatchCheckRequest) ToPermissionCheckRequest(entityID string) *base.PermissionCheckRequest {
	return &base.PermissionCheckRequest{
		TenantId: r.TenantID,
		Entity: &base.Entity{
			Type: r.EntityType,
			Id:   entityID,
		},
		Permission: r.Permission,
		Subject:    r.Subject,
		Metadata:   r.Metadata,
		Context:    r.Context,
		Arguments:  r.Arguments,
	}
}

// Clone creates a shallow copy of the request with a new EntityIDs slice.
func (r *BatchCheckRequest) Clone() *BatchCheckRequest {
	ids := make([]string, len(r.EntityIDs))
	copy(ids, r.EntityIDs)
	return &BatchCheckRequest{
		TenantID:   r.TenantID,
		EntityType: r.EntityType,
		EntityIDs:  ids,
		Permission: r.Permission,
		Subject:    r.Subject,
		Metadata:   r.Metadata,
		Context:    r.Context,
		Arguments:  r.Arguments,
	}
}

// CloneWithDepth creates a clone with a modified depth value.
func (r *BatchCheckRequest) CloneWithDepth(depth int32) *BatchCheckRequest {
	c := r.Clone()
	c.Metadata = &base.PermissionCheckRequestMetadata{
		SchemaVersion: r.Metadata.GetSchemaVersion(),
		SnapToken:     r.Metadata.GetSnapToken(),
		Depth:         depth,
	}
	return c
}
