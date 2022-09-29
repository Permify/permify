package decorators

import (
	"context"

	"github.com/afex/hystrix-go/hystrix"

	"github.com/Permify/permify/internal/repositories"
	"github.com/Permify/permify/internal/repositories/entities"
	"github.com/Permify/permify/pkg/errors"
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
func (r *EntityConfigWithCircuitBreaker) Migrate() (err errors.Error) {
	return nil
}

// All -
func (r *EntityConfigWithCircuitBreaker) All(ctx context.Context, version string) (configs entities.EntityConfigs, err errors.Error) {
	type circuitBreakerResponse struct {
		Configs entities.EntityConfigs
		Error   errors.Error
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
		return configs, errors.CircuitBreakerError
	}
}

// Read -
func (r *EntityConfigWithCircuitBreaker) Read(ctx context.Context, name string, version string) (config entities.EntityConfig, err errors.Error) {
	type circuitBreakerResponse struct {
		Config entities.EntityConfig
		Error  errors.Error
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
		return config, errors.CircuitBreakerError
	}
}

// Write -
func (r *EntityConfigWithCircuitBreaker) Write(ctx context.Context, configs entities.EntityConfigs, version string) (err errors.Error) {
	outputErr := make(chan errors.Error, 1)
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
		return errors.CircuitBreakerError
	}
}
