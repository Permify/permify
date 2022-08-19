package decorators

import (
	"github.com/afex/hystrix-go/hystrix"
	"golang.org/x/net/context"

	"github.com/Permify/permify/internal/entities"
	"github.com/Permify/permify/internal/repositories"
	"github.com/Permify/permify/internal/repositories/filters"
)

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
	hystrix.ConfigureCommand("relationTupleRepository.queryTuples", hystrix.CommandConfig{Timeout: 1000})
	errors := hystrix.Go("entityConfigRepository.queryTuples", func() error {
		tuples, _ = r.repository.QueryTuples(ctx, entity, objectID, relation)
		output <- tuples
		return nil
	}, nil)

	select {
	case out := <-output:
		return out, nil
	case err = <-errors:
		return tuples, err
	}
}

// QueryTuples -
func (r *RelationTupleWithCircuitBreaker) Read(ctx context.Context, filter filters.RelationTupleFilter) (tuples entities.RelationTuples, err error) {
	output := make(chan entities.RelationTuples, 1)
	hystrix.ConfigureCommand("relationTupleRepository.read", hystrix.CommandConfig{Timeout: 1000})
	errors := hystrix.Go("entityConfigRepository.read", func() error {
		tuples, _ = r.repository.Read(ctx, filter)
		output <- tuples
		return nil
	}, nil)

	select {
	case out := <-output:
		return out, nil
	case err = <-errors:
		return tuples, err
	}
}

// QueryTuples -
func (r *RelationTupleWithCircuitBreaker) Write(ctx context.Context, tuples entities.RelationTuples) (err error) {
	output := make(chan bool, 1)
	hystrix.ConfigureCommand("relationTupleRepository.write", hystrix.CommandConfig{Timeout: 1000})
	errors := hystrix.Go("entityConfigRepository.write", func() error {
		_ = r.repository.Write(ctx, tuples)
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

// Delete -
func (r *RelationTupleWithCircuitBreaker) Delete(ctx context.Context, tuples entities.RelationTuples) (err error) {
	output := make(chan bool, 1)
	hystrix.ConfigureCommand("relationTupleRepository.delete", hystrix.CommandConfig{Timeout: 1000})
	errors := hystrix.Go("entityConfigRepository.delete", func() error {
		_ = r.repository.Delete(ctx, tuples)
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
