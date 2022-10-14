package decorators

import (
	"context"
	"errors"

	"github.com/afex/hystrix-go/hystrix"

	"github.com/Permify/permify/internal/repositories"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// EntityConfigWithCircuitBreaker -
type EntityConfigWithCircuitBreaker struct {
	repository repositories.IEntityConfigRepository
}

// NewEntityConfigWithCircuitBreaker -.
func NewEntityConfigWithCircuitBreaker(entityConfigRepository repositories.IEntityConfigRepository) *EntityConfigWithCircuitBreaker {
	return &EntityConfigWithCircuitBreaker{repository: entityConfigRepository}
}

// Migrate -
func (r *EntityConfigWithCircuitBreaker) Migrate() (err error) {
	return nil
}

// All -
func (r *EntityConfigWithCircuitBreaker) All(ctx context.Context, version string) (configs []repositories.EntityConfig, err error) {
	type circuitBreakerResponse struct {
		Configs []repositories.EntityConfig
		Error   error
	}

	output := make(chan circuitBreakerResponse, 1)

	hystrix.ConfigureCommand("entityConfigRepository.all", hystrix.CommandConfig{Timeout: 1000})
	bErrors := hystrix.Go("entityConfigRepository.all", func() error {
		conf, cErr := r.repository.All(ctx, version)
		output <- circuitBreakerResponse{Configs: conf, Error: cErr}
		return nil
	}, func(err error) error {
		return nil
	})

	select {
	case out := <-output:
		return out.Configs, out.Error
	case <-bErrors:
		return configs, errors.New(base.ErrorCode_circuit_breaker_error.String())
	}
}

// Read -
func (r *EntityConfigWithCircuitBreaker) Read(ctx context.Context, name string, version string) (config repositories.EntityConfig, err error) {
	type circuitBreakerResponse struct {
		Config repositories.EntityConfig
		Error  error
	}

	output := make(chan circuitBreakerResponse, 1)

	hystrix.ConfigureCommand("entityConfigRepository.read", hystrix.CommandConfig{Timeout: 1000})
	bErrors := hystrix.Go("entityConfigRepository.read", func() error {
		conf, cErr := r.repository.Read(ctx, name, version)
		output <- circuitBreakerResponse{Config: conf, Error: cErr}
		return nil
	}, func(err error) error {
		return nil
	})

	select {
	case out := <-output:
		return out.Config, out.Error
	case <-bErrors:
		return config, errors.New(base.ErrorCode_circuit_breaker_error.String())
	}
}

// Write -
func (r *EntityConfigWithCircuitBreaker) Write(ctx context.Context, configs []repositories.EntityConfig, version string) (err error) {
	outputErr := make(chan error, 1)
	hystrix.ConfigureCommand("entityConfigRepository.write", hystrix.CommandConfig{Timeout: 1000})
	bErrors := hystrix.Go("entityConfigRepository.write", func() error {
		err = r.repository.Write(ctx, configs, version)
		outputErr <- err
		return nil
	}, func(err error) error {
		return nil
	})

	select {
	case err = <-outputErr:
		return err
	case <-bErrors:
		return errors.New(base.ErrorCode_circuit_breaker_error.String())
	}
}
