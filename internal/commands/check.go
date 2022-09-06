package commands

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	internalErrors "github.com/Permify/permify/internal/internal-errors"
	"github.com/Permify/permify/internal/repositories"
	"github.com/Permify/permify/internal/repositories/entities"
	"github.com/Permify/permify/pkg/dsl/schema"
	"github.com/Permify/permify/pkg/logger"
	"github.com/Permify/permify/pkg/tuple"
)

// CheckDecision -
type CheckDecision struct {
	Prefix string `json:"prefix"`
	Can    bool   `json:"can"`
	Err    error  `json:"-"`
}

// sendCheckDecision -
func sendCheckDecision(can bool, prefix string, err error) CheckDecision {
	return CheckDecision{
		Prefix: prefix,
		Can:    can,
		Err:    err,
	}
}

// CheckFunction -
type CheckFunction func(ctx context.Context, decisionChan chan<- CheckDecision)

// CheckCombiner .
type CheckCombiner func(ctx context.Context, functions []CheckFunction) CheckDecision

// CheckCommand -
type CheckCommand struct {
	relationTupleRepository repositories.IRelationTupleRepository
	logger                  logger.Interface
}

// NewCheckCommand -
func NewCheckCommand(rr repositories.IRelationTupleRepository, l logger.Interface) *CheckCommand {
	return &CheckCommand{
		relationTupleRepository: rr,
		logger:                  l,
	}
}

// CheckQuery -
type CheckQuery struct {
	Entity  tuple.Entity
	Subject tuple.Subject
	depth   int32
	visits  sync.Map
}

func (r *CheckQuery) SetVisit(key string, decision CheckDecision) {
	r.visits.Store(key, decision)
}

func (r *CheckQuery) LoadVisits() map[string]interface{} {
	m := map[string]interface{}{}
	r.visits.Range(func(key, value interface{}) bool {
		m[fmt.Sprint(key)] = value
		return true
	})
	return m
}

// SetDepth -
func (r *CheckQuery) SetDepth(i int32) {
	atomic.StoreInt32(&r.depth, i)
}

// decrease -
func (r *CheckQuery) decrease() int32 {
	return atomic.AddInt32(&r.depth, -1)
}

// loadDepth -
func (r *CheckQuery) loadDepth() int32 {
	return atomic.LoadInt32(&r.depth)
}

// isDepthFinish -
func (r *CheckQuery) isDepthFinish() bool {
	return r.loadDepth() <= 0
}

// CheckResponse -
type CheckResponse struct {
	Can            bool
	Visits         map[string]interface{}
	RemainingDepth int32
}

// Execute -
func (command *CheckCommand) Execute(ctx context.Context, q *CheckQuery, child schema.Child) (response CheckResponse, err error) {
	response.Can = false
	response.Can, err = command.c(ctx, q, child)
	response.Visits = q.LoadVisits()
	response.RemainingDepth = q.loadDepth()
	return
}

// c -
func (command *CheckCommand) c(ctx context.Context, q *CheckQuery, child schema.Child) (bool, error) {
	var fn CheckFunction
	switch child.GetKind() {
	case schema.RewriteKind.String():
		fn = command.checkRewrite(ctx, q, child.(schema.Rewrite))
	case schema.LeafKind.String():
		fn = command.checkLeaf(ctx, q, child.(schema.Leaf))
	}

	if fn == nil {
		return false, internalErrors.UndefinedChildKindError
	}

	result := checkUnion(ctx, []CheckFunction{fn})
	return result.Can, result.Err
}

// checkRewrite -
func (command *CheckCommand) checkRewrite(ctx context.Context, q *CheckQuery, child schema.Rewrite) CheckFunction {
	switch child.GetType() {
	case schema.Union.String():
		return command.set(ctx, q, child.Children, checkUnion)
	case schema.Intersection.String():
		return command.set(ctx, q, child.Children, checkIntersection)
	default:
		return checkFail(internalErrors.UndefinedChildTypeError)
	}
}

// checkLeaf -
func (command *CheckCommand) checkLeaf(ctx context.Context, q *CheckQuery, child schema.Leaf) CheckFunction {
	switch child.GetType() {
	case schema.TupleToUserSetType.String():
		return command.check(ctx, q.Entity, tuple.Relation(child.Value), q, child.Exclusion)
	case schema.ComputedUserSetType.String():
		return command.check(ctx, q.Entity, tuple.Relation(child.Value), q, child.Exclusion)
	default:
		return checkFail(internalErrors.UndefinedChildTypeError)
	}
}

// set -
func (command *CheckCommand) set(ctx context.Context, q *CheckQuery, children []schema.Child, combiner CheckCombiner) CheckFunction {
	var functions []CheckFunction
	for _, child := range children {
		switch child.GetKind() {
		case schema.RewriteKind.String():
			functions = append(functions, command.checkRewrite(ctx, q, child.(schema.Rewrite)))
		case schema.LeafKind.String():
			functions = append(functions, command.checkLeaf(ctx, q, child.(schema.Leaf)))
		default:
			return checkFail(internalErrors.UndefinedChildKindError)
		}
	}

	return func(ctx context.Context, resultChan chan<- CheckDecision) {
		resultChan <- combiner(ctx, functions)
	}
}

// check -
func (command *CheckCommand) check(ctx context.Context, entity tuple.Entity, relation tuple.Relation, q *CheckQuery, exclusion bool) CheckFunction {
	return func(ctx context.Context, decisionChan chan<- CheckDecision) {
		var err error

		q.decrease()

		if q.isDepthFinish() {
			decisionChan <- sendCheckDecision(false, "", internalErrors.DepthError)
			return
		}

		var iterator tuple.ISubjectIterator
		iterator, err = command.getSubjects(ctx, entity, relation)
		if err != nil {
			checkFail(err)
			return
		}

		var checkFunctions []CheckFunction

		for iterator.HasNext() {
			subject := iterator.GetNext()
			if subject.Equals(q.Subject) {
				var dec CheckDecision
				if exclusion {
					dec = sendCheckDecision(false, "not", err)
				} else {
					dec = sendCheckDecision(true, "", err)
				}
				q.SetVisit(tuple.EntityAndRelation{Entity: entity, Relation: relation}.String(), dec)
				decisionChan <- dec
				return
			} else {
				if !subject.IsUser() {
					checkFunctions = append(checkFunctions, command.check(ctx, tuple.Entity{ID: subject.ID, Type: subject.Type}, subject.Relation, q, exclusion))
				}
			}
		}

		if len(checkFunctions) > 0 {
			decisionChan <- checkUnion(ctx, checkFunctions)
			return
		}

		var dec CheckDecision
		if exclusion {
			dec = sendCheckDecision(true, "not", err)
		} else {
			dec = sendCheckDecision(false, "", err)
		}
		q.SetVisit(tuple.EntityAndRelation{Entity: entity, Relation: relation}.String(), dec)
		decisionChan <- dec
		return
	}
}

// union -
func checkUnion(ctx context.Context, functions []CheckFunction) CheckDecision {
	if len(functions) == 0 {
		return sendCheckDecision(false, "", nil)
	}

	decisionChan := make(chan CheckDecision, len(functions))
	childCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	for _, fn := range functions {
		go fn(childCtx, decisionChan)
	}

	for i := 0; i < len(functions); i++ {
		select {
		case result := <-decisionChan:
			if result.Err == nil && result.Can {
				return sendCheckDecision(true, result.Prefix, nil)
			}
			if result.Err != nil {
				return sendCheckDecision(false, result.Prefix, result.Err)
			}
		case <-ctx.Done():
			return sendCheckDecision(false, "", internalErrors.CanceledError)
		}
	}

	return sendCheckDecision(false, "", nil)
}

// intersection -
func checkIntersection(ctx context.Context, functions []CheckFunction) CheckDecision {
	if len(functions) == 0 {
		return sendCheckDecision(false, "", nil)
	}

	decisionChan := make(chan CheckDecision, len(functions))
	childCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	for _, fn := range functions {
		go fn(childCtx, decisionChan)
	}

	for i := 0; i < len(functions); i++ {
		select {
		case result := <-decisionChan:
			if result.Err == nil && !result.Can {
				return sendCheckDecision(false, result.Prefix, nil)
			}
			if result.Err != nil {
				return sendCheckDecision(false, result.Prefix, result.Err)
			}
		case <-ctx.Done():
			return sendCheckDecision(false, "", internalErrors.CanceledError)
		}
	}

	return sendCheckDecision(true, "", nil)
}

// checkFail -
func checkFail(err error) CheckFunction {
	return func(ctx context.Context, decisionChan chan<- CheckDecision) {
		decisionChan <- sendCheckDecision(false, "", err)
	}
}

// getSubjects -
func (command *CheckCommand) getSubjects(ctx context.Context, entity tuple.Entity, relation tuple.Relation) (iterator tuple.ISubjectIterator, err error) {
	r := relation.Split()

	var tuples []entities.RelationTuple
	tuples, err = command.relationTupleRepository.QueryTuples(ctx, entity.Type, entity.ID, r[0].String())
	if err != nil {
		return nil, err
	}

	var subjects []*tuple.Subject
	for _, tup := range tuples {
		ct := tup.ToTuple()
		if !ct.Subject.IsUser() {
			subject := ct.Subject
			if tup.UsersetRelation == tuple.ELLIPSIS {
				subject.Relation = r[1]
			} else {
				subject.Relation = ct.Subject.Relation
			}
			subjects = append(subjects, &subject)
		} else {
			subjects = append(subjects, &tuple.Subject{
				Type: tuple.USER,
				ID:   tup.UsersetObjectID,
			})
		}
	}

	return tuple.NewSubjectIterator(subjects), err
}
