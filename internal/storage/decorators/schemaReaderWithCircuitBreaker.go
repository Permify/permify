package decorators

import (
	"context"
	"errors"

	"github.com/afex/hystrix-go/hystrix"

	"github.com/Permify/permify/internal/storage"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// SchemaReaderWithCircuitBreaker - Add circuit breaker behaviour to schema reader
type SchemaReaderWithCircuitBreaker struct {
	delegate storage.SchemaReader
}

// NewSchemaReaderWithCircuitBreaker - Add circuit breaker behaviour to new schema reader
func NewSchemaReaderWithCircuitBreaker(delegate storage.SchemaReader) *SchemaReaderWithCircuitBreaker {
	return &SchemaReaderWithCircuitBreaker{delegate: delegate}
}

// ReadSchema - Read schema from repository
func (r *SchemaReaderWithCircuitBreaker) ReadSchema(ctx context.Context, tenantID, version string) (*base.SchemaDefinition, error) {
	type circuitBreakerResponse struct {
		Schema *base.SchemaDefinition
		Error  error
	}

	output := make(chan circuitBreakerResponse, 1)

	hystrix.ConfigureCommand("schemaReader.readSchema", hystrix.CommandConfig{Timeout: 1000})
	bErrors := hystrix.Go("schemaReader.readSchema", func() error {
		sch, err := r.delegate.ReadSchema(ctx, tenantID, version)
		output <- circuitBreakerResponse{Schema: sch, Error: err}
		return nil
	}, func(err error) error {
		return nil
	})

	select {
	case out := <-output:
		return out.Schema, out.Error
	case <-bErrors:
		return nil, errors.New(base.ErrorCode_ERROR_CODE_CIRCUIT_BREAKER.String())
	}
}

// ReadEntityDefinition - Read entity definition from repository
func (r *SchemaReaderWithCircuitBreaker) ReadEntityDefinition(ctx context.Context, tenantID, entityName, version string) (*base.EntityDefinition, string, error) {
	type circuitBreakerResponse struct {
		Definition *base.EntityDefinition
		Version    string
		Error      error
	}

	output := make(chan circuitBreakerResponse, 1)

	hystrix.ConfigureCommand("schemaReader.readEntityDefinition", hystrix.CommandConfig{Timeout: 1000})
	bErrors := hystrix.Go("schemaReader.readEntityDefinition", func() error {
		conf, v, err := r.delegate.ReadEntityDefinition(ctx, tenantID, entityName, version)
		output <- circuitBreakerResponse{Definition: conf, Version: v, Error: err}
		return nil
	}, func(err error) error {
		return nil
	})

	select {
	case out := <-output:
		return out.Definition, out.Version, out.Error
	case <-bErrors:
		return nil, "", errors.New(base.ErrorCode_ERROR_CODE_CIRCUIT_BREAKER.String())
	}
}

// ReadRuleDefinition - Read rule definition from repository
func (r *SchemaReaderWithCircuitBreaker) ReadRuleDefinition(ctx context.Context, tenantID, ruleName, version string) (*base.RuleDefinition, string, error) {
	type circuitBreakerResponse struct {
		Definition *base.RuleDefinition
		Version    string
		Error      error
	}

	output := make(chan circuitBreakerResponse, 1)

	hystrix.ConfigureCommand("schemaReader.readRuleDefinition", hystrix.CommandConfig{Timeout: 1000})
	bErrors := hystrix.Go("schemaReader.readRuleDefinition", func() error {
		conf, v, err := r.delegate.ReadRuleDefinition(ctx, tenantID, ruleName, version)
		output <- circuitBreakerResponse{Definition: conf, Version: v, Error: err}
		return nil
	}, func(err error) error {
		return nil
	})

	select {
	case out := <-output:
		return out.Definition, out.Version, out.Error
	case <-bErrors:
		return nil, "", errors.New(base.ErrorCode_ERROR_CODE_CIRCUIT_BREAKER.String())
	}
}

// HeadVersion - Finds the latest version of the schema.
func (r *SchemaReaderWithCircuitBreaker) HeadVersion(ctx context.Context, tenantID string) (version string, err error) {
	type circuitBreakerResponse struct {
		Version string
		Error   error
	}

	output := make(chan circuitBreakerResponse, 1)

	hystrix.ConfigureCommand("schemaReader.headVersion", hystrix.CommandConfig{Timeout: 1000})
	bErrors := hystrix.Go("schemaReader.headVersion", func() error {
		v, err := r.delegate.HeadVersion(ctx, tenantID)
		output <- circuitBreakerResponse{Version: v, Error: err}
		return nil
	}, func(err error) error {
		return nil
	})

	select {
	case out := <-output:
		return out.Version, out.Error
	case <-bErrors:
		return "", errors.New(base.ErrorCode_ERROR_CODE_CIRCUIT_BREAKER.String())
	}
}
