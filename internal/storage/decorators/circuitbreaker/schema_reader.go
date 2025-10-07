package circuitBreaker

import (
	"context"

	"github.com/sony/gobreaker"

	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/pkg/database"
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

// ReadSchema returns the schema definition for a specific tenant and version as a structured object.
func (r *SchemaReader) ReadSchema(ctx context.Context, tenantID, version string) (*base.SchemaDefinition, error) {
	response, err := r.cb.Execute(func() (interface{}, error) {
		return r.delegate.ReadSchema(ctx, tenantID, version)
	})
	if err != nil {
		return nil, err
	}
	return response.(*base.SchemaDefinition), nil
}

// ReadSchemaString returns the schema definition for a specific tenant and version as a string.
func (r *SchemaReader) ReadSchemaString(ctx context.Context, tenantID, version string) (definitions []string, err error) {
	response, err := r.cb.Execute(func() (interface{}, error) {
		return r.delegate.ReadSchemaString(ctx, tenantID, version)
	})
	if err != nil {
		return nil, err
	}
	return response.([]string), nil
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

// ListSchemas - List all Schemas
func (r *SchemaReader) ListSchemas(ctx context.Context, tenantID string, pagination database.Pagination) (schemas []*base.SchemaList, ct database.EncodedContinuousToken, err error) {
	type circuitBreakerResponse struct {
		Schemas []*base.SchemaList
		Ct      database.EncodedContinuousToken
	}

	response, err := r.cb.Execute(func() (interface{}, error) {
		var err error
		var resp circuitBreakerResponse
		resp.Schemas, resp.Ct, err = r.delegate.ListSchemas(ctx, tenantID, pagination)
		return resp, err
	})
	if err != nil {
		return nil, nil, err
	}

	resp := response.(circuitBreakerResponse)
	return resp.Schemas, resp.Ct, nil
}
