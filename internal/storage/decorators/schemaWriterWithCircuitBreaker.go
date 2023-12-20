package decorators

import (
	"context"
	"errors"

	"github.com/afex/hystrix-go/hystrix"

	"github.com/Permify/permify/internal/storage"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// SchemaWriterWithCircuitBreaker - Add circuit breaker behaviour to schema writer
type SchemaWriterWithCircuitBreaker struct {
	delegate storage.SchemaWriter
	timeout  int
}

// NewSchemaWriterWithCircuitBreaker - Add circuit breaker behaviour to new schema writer
func NewSchemaWriterWithCircuitBreaker(delegate storage.SchemaWriter, timeout int) *SchemaWriterWithCircuitBreaker {
	return &SchemaWriterWithCircuitBreaker{delegate: delegate, timeout: timeout}
}

// WriteSchema - Write schema to repository
func (r *SchemaWriterWithCircuitBreaker) WriteSchema(ctx context.Context, definitions []storage.SchemaDefinition) error {
	type circuitBreakerResponse struct {
		Error error
	}

	output := make(chan circuitBreakerResponse, 1)

	hystrix.ConfigureCommand("schemaWriter.writeSchema", hystrix.CommandConfig{Timeout: r.timeout})
	bErrors := hystrix.Go("schemaWriter.writeSchema", func() error {
		err := r.delegate.WriteSchema(ctx, definitions)
		output <- circuitBreakerResponse{Error: err}
		return nil
	}, func(err error) error {
		return nil
	})

	select {
	case out := <-output:
		return out.Error
	case <-bErrors:
		return errors.New(base.ErrorCode_ERROR_CODE_CIRCUIT_BREAKER.String())
	}
}
