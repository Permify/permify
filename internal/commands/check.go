package commands

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	internalErrors "github.com/Permify/permify/internal/errors"
	"github.com/Permify/permify/internal/repositories"
	"github.com/Permify/permify/pkg/errors"
	"github.com/Permify/permify/pkg/logger"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/tuple"
)

// CheckDecision -
type CheckDecision struct {
	Prefix string       `json:"prefix"`
	Can    bool         `json:"can"`
	Err    errors.Error `json:"-"`
}

// newCheckDecision -
func newCheckDecision(can bool, prefix string, err errors.Error) CheckDecision {
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

// GetRelationTupleRepository -
func (command *CheckCommand) GetRelationTupleRepository() repositories.IRelationTupleRepository {
	return command.relationTupleRepository
}

// CheckQuery -
type CheckQuery struct {
	Entity  *base.Entity
	Subject *base.Subject
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
func (command *CheckCommand) Execute(ctx context.Context, q *CheckQuery, child *base.Child) (response CheckResponse, err errors.Error) {
	response.Can = false
	response.Can, err = command.c(ctx, q, child)
	response.Visits = q.LoadVisits()
	response.RemainingDepth = q.loadDepth()
	return
}

// c -
func (command *CheckCommand) c(ctx context.Context, q *CheckQuery, child *base.Child) (bool, errors.Error) {
	var fn CheckFunction
	switch op := child.GetType().(type) {
	case *base.Child_Rewrite:
		fn = command.checkRewrite(ctx, q, op.Rewrite)
	case *base.Child_Leaf:
		fn = command.checkLeaf(ctx, q, op.Leaf)
	}

	if fn == nil {
		return false, internalErrors.UndefinedChildKindError
	}

	result := checkUnion(ctx, []CheckFunction{fn})
	return result.Can, result.Err
}

// checkRewrite -
func (command *CheckCommand) checkRewrite(ctx context.Context, q *CheckQuery, rewrite *base.Rewrite) CheckFunction {
	switch rewrite.GetRewriteOperation() {
	case *base.Rewrite_UNION.Enum():
		return command.set(ctx, q, rewrite.GetChildren(), checkUnion)
	case *base.Rewrite_INTERSECTION.Enum():
		return command.set(ctx, q, rewrite.GetChildren(), checkIntersection)
	default:
		return checkFail(internalErrors.UndefinedChildTypeError)
	}
}

// checkLeaf -
func (command *CheckCommand) checkLeaf(ctx context.Context, q *CheckQuery, leaf *base.Leaf) CheckFunction {
	switch op := leaf.GetType().(type) {
	case *base.Leaf_TupleToUserSet:
		return command.check(ctx, q.Entity, op.TupleToUserSet.GetRelation(), q, leaf.GetExclusion())
	case *base.Leaf_ComputedUserSet:
		return command.check(ctx, q.Entity, op.ComputedUserSet.GetRelation(), q, leaf.GetExclusion())
	default:
		return checkFail(internalErrors.UndefinedChildTypeError)
	}
}

// set -
func (command *CheckCommand) set(ctx context.Context, q *CheckQuery, children []*base.Child, combiner CheckCombiner) CheckFunction {
	var functions []CheckFunction
	for _, child := range children {
		switch child.GetType().(type) {
		case *base.Child_Rewrite:
			functions = append(functions, command.checkRewrite(ctx, q, child.GetRewrite()))
		case *base.Child_Leaf:
			functions = append(functions, command.checkLeaf(ctx, q, child.GetLeaf()))
		default:
			return checkFail(internalErrors.UndefinedChildKindError)
		}
	}

	return func(ctx context.Context, resultChan chan<- CheckDecision) {
		resultChan <- combiner(ctx, functions)
	}
}

// check -
func (command *CheckCommand) check(ctx context.Context, entity *base.Entity, relation string, q *CheckQuery, exclusion bool) CheckFunction {
	return func(ctx context.Context, decisionChan chan<- CheckDecision) {
		var err errors.Error

		q.decrease()

		if q.isDepthFinish() {
			decisionChan <- newCheckDecision(false, "", internalErrors.DepthError)
			return
		}

		var iterator tuple.ISubjectIterator
		iterator, err = getSubjects(ctx, command, entity, relation)
		if err != nil {
			checkFail(err)
			return
		}

		var checkFunctions []CheckFunction

		for iterator.HasNext() {
			subject := iterator.GetNext()
			if tuple.AreSubjectsEqual(subject, q.Subject) {
				var dec CheckDecision
				if exclusion {
					dec = newCheckDecision(false, "not", err)
				} else {
					dec = newCheckDecision(true, "", err)
				}
				q.SetVisit(tuple.EntityAndRelationToString(&base.EntityAndRelation{
					Entity:   entity,
					Relation: relation,
				}), dec)
				decisionChan <- dec
				return
			} else {
				if !tuple.IsSubjectUser(subject) {
					checkFunctions = append(checkFunctions, command.check(ctx, &base.Entity{Id: subject.GetId(), Type: subject.GetType()}, subject.GetRelation(), q, exclusion))
				}
			}
		}

		if len(checkFunctions) > 0 {
			decisionChan <- checkUnion(ctx, checkFunctions)
			return
		}

		var dec CheckDecision
		if exclusion {
			dec = newCheckDecision(true, "not", err)
		} else {
			dec = newCheckDecision(false, "", err)
		}

		q.SetVisit(tuple.EntityAndRelationToString(&base.EntityAndRelation{
			Entity:   entity,
			Relation: relation,
		}), dec)
		decisionChan <- dec
		return
	}
}

// union -
func checkUnion(ctx context.Context, functions []CheckFunction) CheckDecision {
	if len(functions) == 0 {
		return newCheckDecision(false, "", nil)
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
				return newCheckDecision(true, result.Prefix, nil)
			}
			if result.Err != nil {
				return newCheckDecision(false, result.Prefix, result.Err)
			}
		case <-ctx.Done():
			return newCheckDecision(false, "", internalErrors.CanceledError)
		}
	}

	return newCheckDecision(false, "", nil)
}

// intersection -
func checkIntersection(ctx context.Context, functions []CheckFunction) CheckDecision {
	if len(functions) == 0 {
		return newCheckDecision(false, "", nil)
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
				return newCheckDecision(false, result.Prefix, nil)
			}
			if result.Err != nil {
				return newCheckDecision(false, result.Prefix, result.Err)
			}
		case <-ctx.Done():
			return newCheckDecision(false, "", internalErrors.CanceledError)
		}
	}

	return newCheckDecision(true, "", nil)
}

// checkFail -
func checkFail(err errors.Error) CheckFunction {
	return func(ctx context.Context, decisionChan chan<- CheckDecision) {
		decisionChan <- newCheckDecision(false, "", err)
	}
}
