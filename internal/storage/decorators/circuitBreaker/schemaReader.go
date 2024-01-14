package circuitBreaker

import (
	"context"

	"github.com/sony/gobreaker"

	"github.com/Permify/permify/internal/storage"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// SchemaReader - Add circuit breaker behaviour to schema reader
type SchemaReader struct {
	delegate storage.SchemaReader
	cb       *gobreaker.CircuitBreaker
}

// NewSchemaReader - Add circuit breaker behaviour to new schema reader
func NewSchemaReader(delegate storage.SchemaReader, cb *gobreaker.CircuitBreaker) *SchemaReader {
	return &SchemaReader{delegate: delegate, cb: cb}
}

// ReadSchema - Read schema from repository
func (r *SchemaReader) ReadSchema(ctx context.Context, tenantID, version string) (*base.SchemaDefinition, error) {
	response, err := r.cb.Execute(func() (interface{}, error) {
		return r.delegate.ReadSchema(ctx, tenantID, version)
	})
	if err != nil {
		return nil, err
	}
	return response.(*base.SchemaDefinition), nil
}

// ReadEntityDefinition - Read entity definition from repository
func (r *SchemaReader) ReadEntityDefinition(ctx context.Context, tenantID, entityName, version string) (*base.EntityDefinition, string, error) {
	type circuitBreakerResponse struct {
		Definition *base.EntityDefinition
		Version    string
	}

	response, err := r.cb.Execute(func() (interface{}, error) {
		var err error
		var resp circuitBreakerResponse
		resp.Definition, resp.Version, err = r.delegate.ReadEntityDefinition(ctx, tenantID, entityName, version)
		return resp, err
	})
	if err != nil {
		return nil, "", err
	}

	resp := response.(circuitBreakerResponse)
	return resp.Definition, resp.Version, nil
}

// ReadRuleDefinition - Read rule definition from repository
func (r *SchemaReader) ReadRuleDefinition(ctx context.Context, tenantID, ruleName, version string) (*base.RuleDefinition, string, error) {
	type circuitBreakerResponse struct {
		Definition *base.RuleDefinition
		Version    string
	}

	response, err := r.cb.Execute(func() (interface{}, error) {
		var err error
		var resp circuitBreakerResponse
		resp.Definition, resp.Version, err = r.delegate.ReadRuleDefinition(ctx, tenantID, ruleName, version)
		return resp, err
	})
	if err != nil {
		return nil, "", err
	}

	resp := response.(circuitBreakerResponse)
	return resp.Definition, resp.Version, nil
}

// HeadVersion - Finds the latest version of the schema.
func (r *SchemaReader) HeadVersion(ctx context.Context, tenantID string) (version string, err error) {
	response, err := r.cb.Execute(func() (interface{}, error) {
		return r.delegate.HeadVersion(ctx, tenantID)
	})
	if err != nil {
		return "", err
	}
	return response.(string), nil
}
