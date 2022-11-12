package commands

import (
	"context"
	"errors"
	"sync/atomic"

	"github.com/Permify/permify/internal/keys"
	"github.com/Permify/permify/internal/repositories"
	"github.com/Permify/permify/pkg/database"
	"github.com/Permify/permify/pkg/logger"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/tuple"
)

// CheckDecision -
type CheckDecision struct {
	Response *base.PermissionCheckResponse
	Err      error `json:"-"`
}

// newCheckDecision -
func newCheckDecision(decision *base.PermissionCheckResponse, err error) CheckDecision {
	return CheckDecision{
		Response: decision,
		Err:      err,
	}
}

// CheckFunction -
type CheckFunction func(ctx context.Context, decisionChan chan<- CheckDecision)

// CheckCombiner .
type CheckCombiner func(ctx context.Context, functions []CheckFunction) CheckDecision

// CheckCommand -
type CheckCommand struct {
	commandKeyManager  keys.CommandKeyManager
	relationshipReader repositories.RelationshipReader
	logger             logger.Interface
}

// NewCheckCommand -
func NewCheckCommand(km keys.CommandKeyManager, rr repositories.RelationshipReader, l logger.Interface) *CheckCommand {
	return &CheckCommand{
		commandKeyManager:  km,
		relationshipReader: rr,
		logger:             l,
	}
}

// RelationshipReader -
func (command *CheckCommand) RelationshipReader() repositories.RelationshipReader {
	return command.relationshipReader
}

// decrease -
func decrease(request *base.PermissionCheckRequest) int32 {
	return atomic.AddInt32(&request.Depth.Value, -1)
}

// loadDepth -
func loadDepth(request *base.PermissionCheckRequest) int32 {
	return atomic.LoadInt32(&request.Depth.Value)
}

// isDepthFinish -
func isDepthFinish(request *base.PermissionCheckRequest) bool {
	return loadDepth(request) <= 0
}

// Execute -
func (command *CheckCommand) Execute(ctx context.Context, request *base.PermissionCheckRequest, child *base.Child) (response *base.PermissionCheckResponse, err error) {
	response, err = command.c(ctx, request, child)
	response.RemainingDepth = request.GetDepth().Value
	return
}

// c -
func (command *CheckCommand) c(ctx context.Context, request *base.PermissionCheckRequest, child *base.Child) (*base.PermissionCheckResponse, error) {
	var fn CheckFunction
	switch op := child.GetType().(type) {
	case *base.Child_Rewrite:
		fn = command.checkRewrite(ctx, request, op.Rewrite)
	case *base.Child_Leaf:
		fn = command.checkLeaf(ctx, request, op.Leaf)
	}

	if fn == nil {
		return &base.PermissionCheckResponse{
			Can: false,
		}, errors.New(base.ErrorCode_ERROR_CODE_UNDEFINED_CHILD_KIND.String())
	}

	result := checkUnion(ctx, []CheckFunction{fn})
	return result.Response, result.Err
}

// checkRewrite -
func (command *CheckCommand) checkRewrite(ctx context.Context, q *base.PermissionCheckRequest, rewrite *base.Rewrite) CheckFunction {
	switch rewrite.GetRewriteOperation() {
	case *base.Rewrite_OPERATION_UNION.Enum():
		return command.set(ctx, q, rewrite.GetChildren(), checkUnion)
	case *base.Rewrite_OPERATION_INTERSECTION.Enum():
		return command.set(ctx, q, rewrite.GetChildren(), checkIntersection)
	default:
		return checkFail(errors.New(base.ErrorCode_ERROR_CODE_UNDEFINED_CHILD_TYPE.String()))
	}
}

// checkLeaf -
func (command *CheckCommand) checkLeaf(ctx context.Context, request *base.PermissionCheckRequest, leaf *base.Leaf) CheckFunction {
	switch op := leaf.GetType().(type) {
	case *base.Leaf_TupleToUserSet:
		return command.check(ctx, &base.PermissionCheckRequest{
			Entity:        request.GetEntity(),
			Action:        op.TupleToUserSet.GetRelation(),
			Subject:       request.GetSubject(),
			SnapToken:     request.GetSnapToken(),
			SchemaVersion: request.GetSchemaVersion(),
			Depth:         request.GetDepth(),
		}, leaf.GetExclusion())
	case *base.Leaf_ComputedUserSet:
		return command.check(ctx, &base.PermissionCheckRequest{
			Entity:        request.GetEntity(),
			Action:        op.ComputedUserSet.GetRelation(),
			Subject:       request.GetSubject(),
			SnapToken:     request.GetSnapToken(),
			SchemaVersion: request.GetSchemaVersion(),
			Depth:         request.GetDepth(),
		}, leaf.GetExclusion())
	default:
		return checkFail(errors.New(base.ErrorCode_ERROR_CODE_UNDEFINED_CHILD_TYPE.String()))
	}
}

// set -
func (command *CheckCommand) set(ctx context.Context, q *base.PermissionCheckRequest, children []*base.Child, combiner CheckCombiner) CheckFunction {
	var functions []CheckFunction
	for _, child := range children {
		switch child.GetType().(type) {
		case *base.Child_Rewrite:
			functions = append(functions, command.checkRewrite(ctx, q, child.GetRewrite()))
		case *base.Child_Leaf:
			functions = append(functions, command.checkLeaf(ctx, q, child.GetLeaf()))
		default:
			return checkFail(errors.New(base.ErrorCode_ERROR_CODE_UNDEFINED_CHILD_TYPE.String()))
		}
	}

	return func(ctx context.Context, resultChan chan<- CheckDecision) {
		resultChan <- combiner(ctx, functions)
	}
}

// check -
func (command *CheckCommand) check(ctx context.Context, request *base.PermissionCheckRequest, exclusion bool) CheckFunction {
	return func(ctx context.Context, decisionChan chan<- CheckDecision) {
		var err error

		res, found := command.commandKeyManager.GetCheckKey(request)
		if found {
			var can bool
			if exclusion {
				can = !res.GetCan()
			} else {
				can = res.GetCan()
			}
			decisionChan <- CheckDecision{
				Response: &base.PermissionCheckResponse{
					Can: can,
				},
				Err: nil,
			}
			return
		}

		decrease(request)

		if isDepthFinish(request) {
			decisionChan <- newCheckDecision(&base.PermissionCheckResponse{
				Can: false,
			}, errors.New(base.ErrorCode_ERROR_CODE_DEPTH_NOT_ENOUGH.String()))
		}

		var iterator database.ISubjectIterator
		iterator, err = getSubjects(ctx, command, &base.EntityAndRelation{
			Entity: &base.Entity{
				Type: request.GetEntity().GetType(),
				Id:   request.GetEntity().GetId(),
			},
			Relation: request.GetAction(),
		}, request.GetSnapToken())
		if err != nil {
			checkFail(err)
			return
		}

		var checkFunctions []CheckFunction

		for iterator.HasNext() {
			subject := iterator.GetNext()
			if tuple.AreSubjectsEqual(subject, request.GetSubject()) {
				var dec CheckDecision
				if exclusion {
					dec = newCheckDecision(&base.PermissionCheckResponse{
						Can: false,
					}, nil)
				} else {
					dec = newCheckDecision(&base.PermissionCheckResponse{
						Can: true,
					}, nil)
				}
				command.commandKeyManager.SetCheckKey(request, &base.PermissionCheckResponse{
					Can: true,
				})
				decisionChan <- dec
				return
			} else {
				if !tuple.IsSubjectUser(subject) {
					checkFunctions = append(checkFunctions, command.check(ctx, &base.PermissionCheckRequest{
						Entity: &base.Entity{
							Type: subject.GetType(),
							Id:   subject.GetId(),
						},
						Action:        subject.GetRelation(),
						Subject:       request.GetSubject(),
						SnapToken:     request.GetSnapToken(),
						SchemaVersion: request.GetSchemaVersion(),
						Depth:         request.GetDepth(),
					}, exclusion))
				}
			}
		}

		if len(checkFunctions) > 0 {
			decisionChan <- checkUnion(ctx, checkFunctions)
			return
		}

		var dec CheckDecision
		if exclusion {
			dec = newCheckDecision(&base.PermissionCheckResponse{
				Can: true,
			}, nil)
		} else {
			dec = newCheckDecision(&base.PermissionCheckResponse{
				Can: false,
			}, nil)
		}
		command.commandKeyManager.SetCheckKey(request, &base.PermissionCheckResponse{
			Can: false,
		})
		decisionChan <- dec
		return
	}
}

// union -
func checkUnion(ctx context.Context, functions []CheckFunction) CheckDecision {
	if len(functions) == 0 {
		return newCheckDecision(&base.PermissionCheckResponse{
			Can: false,
		}, nil)
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
			if result.Err == nil && result.Response.GetCan() {
				return newCheckDecision(&base.PermissionCheckResponse{
					Can: true,
				}, nil)
			}
			if result.Err != nil {
				return newCheckDecision(&base.PermissionCheckResponse{
					Can: false,
				}, result.Err)
			}
		case <-ctx.Done():
			return newCheckDecision(&base.PermissionCheckResponse{
				Can: false,
			}, errors.New(base.ErrorCode_ERROR_CODE_CANCELLED.String()))
		}
	}

	return newCheckDecision(&base.PermissionCheckResponse{
		Can: false,
	}, nil)
}

// intersection -
func checkIntersection(ctx context.Context, functions []CheckFunction) CheckDecision {
	if len(functions) == 0 {
		return newCheckDecision(&base.PermissionCheckResponse{
			Can: false,
		}, nil)
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
			if result.Err == nil && !result.Response.GetCan() {
				return newCheckDecision(&base.PermissionCheckResponse{
					Can: false,
				}, nil)
			}
			if result.Err != nil {
				return newCheckDecision(&base.PermissionCheckResponse{
					Can: false,
				}, result.Err)
			}
		case <-ctx.Done():
			return newCheckDecision(&base.PermissionCheckResponse{
				Can: false,
			}, errors.New(base.ErrorCode_ERROR_CODE_CANCELLED.String()))
		}
	}

	return newCheckDecision(&base.PermissionCheckResponse{
		Can: true,
	}, nil)
}

// checkFail -
func checkFail(err error) CheckFunction {
	return func(ctx context.Context, decisionChan chan<- CheckDecision) {
		decisionChan <- newCheckDecision(&base.PermissionCheckResponse{
			Can: false,
		}, err)
	}
}
