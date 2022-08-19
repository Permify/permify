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
	hystrix.ConfigureCommand("entityConfigRepository.all", hystrix.CommandConfig{Timeout: 1000})
	errors := hystrix.Go("entityConfigRepository.all", func() error {
		configs, _ = r.repository.All(ctx, version)
		output <- configs
		return nil
	}, nil)

	select {
	case out := <-output:
		return out, nil
	case err = <-errors:
		return configs, err
	}
}

// Read -
func (r *EntityConfigWithCircuitBreaker) Read(ctx context.Context, name string, version string) (config entities.EntityConfig, err error) {
	output := make(chan entities.EntityConfig, 1)
	hystrix.ConfigureCommand("entityConfigRepository.read", hystrix.CommandConfig{Timeout: 1000})
	errors := hystrix.Go("entityConfigRepository.read", func() error {
		config, _ = r.repository.Read(ctx, name, version)
		output <- config
		return nil
	}, nil)

	select {
	case out := <-output:
		return out, nil
	case err = <-errors:
		return config, err
	}
}

// Write -
func (r *EntityConfigWithCircuitBreaker) Write(ctx context.Context, configs entities.EntityConfigs, version string) (err error) {
	output := make(chan bool, 1)
	hystrix.ConfigureCommand("entityConfigRepository.write", hystrix.CommandConfig{Timeout: 1000})
	errors := hystrix.Go("entityConfigRepository.write", func() error {
		_ = r.repository.Write(ctx, configs, version)
		output <- true
		return nil
	}, nil)

	select {
	case _ = <-output:
		return nil
	case err = <-errors:
		return err
	}
}

// Clear -
func (r *EntityConfigWithCircuitBreaker) Clear(ctx context.Context, version string) (err error) {
	output := make(chan bool, 1)
	hystrix.ConfigureCommand("entityConfigRepository.clear", hystrix.CommandConfig{Timeout: 1000})
	errors := hystrix.Go("entityConfigRepository.clear", func() error {
		_ = r.repository.Clear(ctx, version)
		output <- true
		return nil
	}, nil)

	select {
	case _ = <-output:
		return nil
	case err = <-errors:
		return err
	}
}
