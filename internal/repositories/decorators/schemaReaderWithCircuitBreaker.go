package decorators

import (
	"context"
	"errors"

	"github.com/afex/hystrix-go/hystrix"

	"github.com/Permify/permify/internal/repositories"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// SchemaReaderWithCircuitBreaker -
type SchemaReaderWithCircuitBreaker struct {
	delegate repositories.SchemaReader
}

// NewSchemaReaderWithCircuitBreaker -.
func NewSchemaReaderWithCircuitBreaker(delegate repositories.SchemaReader) *SchemaReaderWithCircuitBreaker {
	return &SchemaReaderWithCircuitBreaker{delegate: delegate}
}

// ReadSchema -
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

// ReadSchemaDefinition -
func (r *SchemaReaderWithCircuitBreaker) ReadSchemaDefinition(ctx context.Context, entityType string, version string) (*base.EntityDefinition, error) {
	type circuitBreakerResponse struct {
		Definition *base.EntityDefinition
		Error      error
	}

	output := make(chan circuitBreakerResponse, 1)

	hystrix.ConfigureCommand("schemaReader.readSchemaDefinition", hystrix.CommandConfig{Timeout: 1000})
	bErrors := hystrix.Go("schemaReader.readSchemaDefinition", func() error {
		conf, err := r.delegate.ReadSchemaDefinition(ctx, entityType, version)
		output <- circuitBreakerResponse{Definition: conf, Error: err}
		return nil
	}, func(err error) error {
		return nil
	})

	select {
	case out := <-output:
		return out.Definition, out.Error
	case <-bErrors:
		return nil, errors.New(base.ErrorCode_ERROR_CODE_CIRCUIT_BREAKER.String())
	}
}
