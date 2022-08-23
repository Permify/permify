package decorators

import (
	"context"

	"github.com/afex/hystrix-go/hystrix"

	"github.com/Permify/permify/internal/entities"
	"github.com/Permify/permify/internal/repositories"
)

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
func (r *EntityConfigWithCircuitBreaker) All(ctx context.Context, version string) (configs entities.EntityConfigs, err error) {
	output := make(chan entities.EntityConfigs, 1)
	outputErr := make(chan error, 1)
	hystrix.ConfigureCommand("entityConfigRepository.all", hystrix.CommandConfig{Timeout: 1000})
	errors := hystrix.Go("entityConfigRepository.all", func() error {
		configs, err = r.repository.All(ctx, version)
		outputErr <- err
		output <- configs
		return nil
	}, nil)

	select {
	case out := <-output:
		return out, nil
	case err = <-outputErr:
		return configs, err
	case err = <-errors:
		return configs, err
	}
}

// Read -
func (r *EntityConfigWithCircuitBreaker) Read(ctx context.Context, name string, version string) (config entities.EntityConfig, err error) {
	output := make(chan entities.EntityConfig, 1)
	outputErr := make(chan error, 1)
	hystrix.ConfigureCommand("entityConfigRepository.read", hystrix.CommandConfig{Timeout: 1000})
	errors := hystrix.Go("entityConfigRepository.read", func() error {
		config, err = r.repository.Read(ctx, name, version)
		outputErr <- err
		output <- config
		return nil
	}, nil)

	select {
	case out := <-output:
		return out, nil
	case err = <-outputErr:
		return config, err
	case err = <-errors:
		return config, err
	}
}

// Write -
func (r *EntityConfigWithCircuitBreaker) Write(ctx context.Context, configs entities.EntityConfigs, version string) (err error) {
	outputErr := make(chan error, 1)
	hystrix.ConfigureCommand("entityConfigRepository.write", hystrix.CommandConfig{Timeout: 1000})
	errors := hystrix.Go("entityConfigRepository.write", func() error {
		err = r.repository.Write(ctx, configs, version)
		outputErr <- err
		return nil
	}, nil)

	select {
	case err = <-outputErr:
		return err
	case err = <-errors:
		return err
	}
}

// Clear -
func (r *EntityConfigWithCircuitBreaker) Clear(ctx context.Context, version string) (err error) {
	outputErr := make(chan error, 1)
	hystrix.ConfigureCommand("entityConfigRepository.clear", hystrix.CommandConfig{Timeout: 1000})
	errors := hystrix.Go("entityConfigRepository.clear", func() error {
		err = r.repository.Clear(ctx, version)
		outputErr <- err
		return nil
	}, nil)

	select {
	case err = <-outputErr:
		return err
	case err = <-errors:
		return err
	}
}
