package decorators

import (
	"context"
	"errors"

	"github.com/afex/hystrix-go/hystrix"

	"github.com/Permify/permify/internal/repositories"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// SchemaReaderWithCircuitBreaker - Add circuit breaker behaviour to schema reader
type SchemaReaderWithCircuitBreaker struct {
	delegate repositories.SchemaReader
}

// NewSchemaReaderWithCircuitBreaker - Add circuit breaker behaviour to new schema reader
func NewSchemaReaderWithCircuitBreaker(delegate repositories.SchemaReader) *SchemaReaderWithCircuitBreaker {
	return &SchemaReaderWithCircuitBreaker{delegate: delegate}
}

// ReadSchema - Read schema from repository
func (r *SchemaReaderWithCircuitBreaker) ReadSchema(ctx context.Context, version string) (*base.IndexedSchema, error) {
	type circuitBreakerResponse struct {
		Schema *base.IndexedSchema
		Error  error
	}

	output := make(chan circuitBreakerResponse, 1)

	hystrix.ConfigureCommand("schemaReader.readSchema", hystrix.CommandConfig{Timeout: 1000})
	bErrors := hystrix.Go("schemaReader.readSchema", func() error {
		sch, err := r.delegate.ReadSchema(ctx, version)
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

// ReadSchemaDefinition - Read schema definition from repository
func (r *SchemaReaderWithCircuitBreaker) ReadSchemaDefinition(ctx context.Context, entityType string, version string) (*base.EntityDefinition, string, error) {
	type circuitBreakerResponse struct {
		Definition *base.EntityDefinition
		Version    string
		Error      error
	}

	output := make(chan circuitBreakerResponse, 1)

	hystrix.ConfigureCommand("schemaReader.readSchemaDefinition", hystrix.CommandConfig{Timeout: 1000})
	bErrors := hystrix.Go("schemaReader.readSchemaDefinition", func() error {
		conf, v, err := r.delegate.ReadSchemaDefinition(ctx, entityType, version)
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
func (r *SchemaReaderWithCircuitBreaker) HeadVersion(ctx context.Context) (version string, err error) {
	type circuitBreakerResponse struct {
		Version string
		Error   error
	}

	output := make(chan circuitBreakerResponse, 1)

	hystrix.ConfigureCommand("schemaReader.headVersion", hystrix.CommandConfig{Timeout: 1000})
	bErrors := hystrix.Go("schemaReader.headVersion", func() error {
		v, err := r.delegate.HeadVersion(ctx)
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
