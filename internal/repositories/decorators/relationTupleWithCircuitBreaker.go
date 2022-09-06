package decorators

import (
	"github.com/afex/hystrix-go/hystrix"
	"golang.org/x/net/context"

	"github.com/Permify/permify/internal/repositories"
	"github.com/Permify/permify/internal/repositories/entities"
	"github.com/Permify/permify/internal/repositories/filters"
)

// RelationTupleWithCircuitBreaker -
type RelationTupleWithCircuitBreaker struct {
	repository repositories.IRelationTupleRepository
}

// NewRelationTupleWithCircuitBreaker -.
func NewRelationTupleWithCircuitBreaker(relationTupleRepository repositories.IRelationTupleRepository) *RelationTupleWithCircuitBreaker {
	return &RelationTupleWithCircuitBreaker{repository: relationTupleRepository}
}

// Migrate -
func (r *RelationTupleWithCircuitBreaker) Migrate() (err error) {
	return nil
}

// QueryTuples -
func (r *RelationTupleWithCircuitBreaker) QueryTuples(ctx context.Context, entity string, objectID string, relation string) (tuples entities.RelationTuples, err error) {
	output := make(chan entities.RelationTuples, 1)
	outputErr := make(chan error, 1)
	hystrix.ConfigureCommand("relationTupleRepository.queryTuples", hystrix.CommandConfig{Timeout: 1000})
	errors := hystrix.Go("entityConfigRepository.queryTuples", func() error {
		tuples, err = r.repository.QueryTuples(ctx, entity, objectID, relation)
		outputErr <- err
		output <- tuples
		return nil
	}, nil)

	select {
	case out := <-output:
		return out, nil
	case err = <-outputErr:
		return tuples, err
	case err = <-errors:
		return tuples, err
	}
}

// Read -
func (r *RelationTupleWithCircuitBreaker) Read(ctx context.Context, filter filters.RelationTupleFilter) (tuples entities.RelationTuples, err error) {
	output := make(chan entities.RelationTuples, 1)
	outputErr := make(chan error, 1)
	hystrix.ConfigureCommand("relationTupleRepository.read", hystrix.CommandConfig{Timeout: 1000})
	errors := hystrix.Go("entityConfigRepository.read", func() error {
		tuples, err = r.repository.Read(ctx, filter)
		outputErr <- err
		output <- tuples
		return nil
	}, nil)

	select {
	case out := <-output:
		return out, err
	case err = <-outputErr:
		return tuples, err
	case err = <-errors:
		return tuples, err
	}
}

// Write -
func (r *RelationTupleWithCircuitBreaker) Write(ctx context.Context, tuples entities.RelationTuples) (err error) {
	outputErr := make(chan error, 1)
	hystrix.ConfigureCommand("relationTupleRepository.write", hystrix.CommandConfig{Timeout: 1000})
	errors := hystrix.Go("entityConfigRepository.write", func() error {
		err = r.repository.Write(ctx, tuples)
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

// Delete -
func (r *RelationTupleWithCircuitBreaker) Delete(ctx context.Context, tuples entities.RelationTuples) (err error) {
	outputErr := make(chan error, 1)
	hystrix.ConfigureCommand("relationTupleRepository.delete", hystrix.CommandConfig{Timeout: 1000})
	errors := hystrix.Go("entityConfigRepository.delete", func() error {
		err = r.repository.Delete(ctx, tuples)
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
