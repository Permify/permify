package decorators

import (
	"context"
	"errors"

	"github.com/afex/hystrix-go/hystrix"

	"github.com/Permify/permify/internal/repositories"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// SchemaWriterWithCircuitBreaker - Add circuit breaker behaviour to schema writer
type SchemaWriterWithCircuitBreaker struct {
	delegate repositories.SchemaWriter
}

// NewSchemaWriterWithCircuitBreaker - Add circuit breaker behaviour to new schema writer
func NewSchemaWriterWithCircuitBreaker(delegate repositories.SchemaWriter) *SchemaWriterWithCircuitBreaker {
	return &SchemaWriterWithCircuitBreaker{delegate: delegate}
}

// WriteSchema - Write schema to repository
func (r *SchemaWriterWithCircuitBreaker) WriteSchema(ctx context.Context, definitions []repositories.SchemaDefinition) (string, error) {
	type circuitBreakerResponse struct {
		Version string
		Error   error
	}

	output := make(chan circuitBreakerResponse, 1)

	hystrix.ConfigureCommand("schemaWriter.writeSchema", hystrix.CommandConfig{Timeout: 1000})
	bErrors := hystrix.Go("schemaWriter.writeSchema", func() error {
		version, err := r.delegate.WriteSchema(ctx, definitions)
		output <- circuitBreakerResponse{Version: version, Error: err}
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
