package commands

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/doug-martin/goqu/v9"

	"github.com/Permify/permify/internal/repositories"
	"github.com/Permify/permify/pkg/dsl/schema"
	"github.com/Permify/permify/pkg/helper"
	"github.com/Permify/permify/pkg/logger"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/tuple"
)

type BuilderNodeKind string

const (
	QUERY BuilderNodeKind = "query"
	LOGIC BuilderNodeKind = "logic"
)

// IBuilderNode -
type IBuilderNode interface {
	GetKind() BuilderNodeKind
	Error() error
}

// LogicNode -
type LogicNode struct {
	Operation base.Rewrite_Operation
	Children  []IBuilderNode
	Err       error
}

// GetKind -
func (LogicNode) GetKind() BuilderNodeKind {
	return LOGIC
}

// Error -
func (e LogicNode) Error() error {
	return e.Err
}

// QueryNode -
type QueryNode struct {
	idResolver ResolverFunction
	Key        string          `json:"condition"`
	Join       map[string]join `json:"join"`
	Args       []string        `json:"vars"`
	Exclusion  bool            `json:"exclusion"`
	Err        error           `json:"-"`
}

type join struct {
	key   string
	value string
}

// GetKind -
func (QueryNode) GetKind() BuilderNodeKind {
	return QUERY
}

// Error -
func (e QueryNode) Error() error {
	return e.Err
}

// LookupQueryCommand -
type LookupQueryCommand struct {
	relationTupleRepository repositories.IRelationTupleRepository
	logger                  logger.Interface
}

// NewLookupQueryCommand -
func NewLookupQueryCommand(rr repositories.IRelationTupleRepository, l logger.Interface) *LookupQueryCommand {
	return &LookupQueryCommand{
		logger:                  l,
		relationTupleRepository: rr,
	}
}

// GetRelationTupleRepository -
func (command *LookupQueryCommand) GetRelationTupleRepository() repositories.IRelationTupleRepository {
	return command.relationTupleRepository
}

// BuildFunction -
type BuildFunction func(ctx context.Context, resultChan chan<- IBuilderNode)

// ResolverFunction -
type ResolverFunction func() ([]string, error)

// BuildCombiner .
type BuildCombiner func(ctx context.Context, functions []BuildFunction) IBuilderNode

// LookupQueryQuery -
type LookupQueryQuery struct {
	EntityType string
	Action     string
	Subject    *base.Subject
	schema     *base.Schema
}

// SetSchema -
func (l *LookupQueryQuery) SetSchema(sch *base.Schema) {
	l.schema = sch
}

// LookupQueryResponse -
type LookupQueryResponse struct {
	Node  IBuilderNode
	Query string
	Args  []string
}

// NewLookupQueryResponse -
func NewLookupQueryResponse(node IBuilderNode, table string) (LookupQueryResponse, error) {
	query, args, err := rootNodeToSql(node, table)
	strArgs := make([]string, len(args))
	for i, v := range args {
		strArgs[i] = v.(string)
	}
	return LookupQueryResponse{
		Node:  node,
		Query: query,
		Args:  strArgs,
	}, err
}

// Execute -
func (command *LookupQueryCommand) Execute(ctx context.Context, q *LookupQueryQuery, child *base.Child) (response LookupQueryResponse, err error) {
	return command.l(ctx, q, child)
}

func (command *LookupQueryCommand) l(ctx context.Context, q *LookupQueryQuery, child *base.Child) (response LookupQueryResponse, err error) {
	en, _ := schema.GetEntityByName(q.schema, q.EntityType)

	var fn BuildFunction
	switch op := child.GetType().(type) {
	case *base.Child_Rewrite:
		fn = command.buildRewrite(ctx, q, op.Rewrite)
	case *base.Child_Leaf:
		fn = command.buildLeaf(ctx, q, op.Leaf)
	}

	if fn == nil {
		return LookupQueryResponse{}, errors.New(base.ErrorCode_undefined_child_kind.String())
	}

	result := buildUnion(ctx, []BuildFunction{fn})
	return NewLookupQueryResponse(result, schema.GetTable(en))
}

// set -
func (command *LookupQueryCommand) set(ctx context.Context, q *LookupQueryQuery, children []*base.Child, combiner BuildCombiner) BuildFunction {
	var functions []BuildFunction
	for _, child := range children {
		switch child.GetType().(type) {
		case *base.Child_Rewrite:
			functions = append(functions, command.buildRewrite(ctx, q, child.GetRewrite()))
		case *base.Child_Leaf:
			functions = append(functions, command.buildLeaf(ctx, q, child.GetLeaf()))
		default:
			return buildFail(errors.New(base.ErrorCode_undefined_child_kind.String()))
		}
	}

	return func(ctx context.Context, resultChan chan<- IBuilderNode) {
		resultChan <- combiner(ctx, functions)
	}
}

// buildRewrite -
func (command *LookupQueryCommand) buildRewrite(ctx context.Context, q *LookupQueryQuery, rewrite *base.Rewrite) BuildFunction {
	switch rewrite.GetRewriteOperation() {
	case *base.Rewrite_UNION.Enum():
		return command.set(ctx, q, rewrite.GetChildren(), buildUnion)
	case *base.Rewrite_INTERSECTION.Enum():
		return command.set(ctx, q, rewrite.GetChildren(), buildIntersection)
	default:
		return buildFail(errors.New(base.ErrorCode_undefined_child_type.String()))
	}
}

// buildLeaf -
func (command *LookupQueryCommand) buildLeaf(ctx context.Context, q *LookupQueryQuery, leaf *base.Leaf) BuildFunction {
	switch op := leaf.GetType().(type) {
	case *base.Leaf_TupleToUserSet:
		return command.build(ctx, leaf.GetExclusion(), command.buildTupleToUserSetQuery(ctx, q.Subject, q.EntityType, op.TupleToUserSet.GetRelation(), q.schema))
	case *base.Leaf_ComputedUserSet:
		return command.build(ctx, leaf.GetExclusion(), command.buildComputedUserSetQuery(ctx, q.Subject, q.EntityType, op.ComputedUserSet.GetRelation(), q.schema))
	default:
		return buildFail(errors.New(base.ErrorCode_undefined_child_type.String()))
	}
}

// build -
func (command *LookupQueryCommand) build(ctx context.Context, exclusion bool, q QueryNode) BuildFunction {
	return func(ctx context.Context, builderChan chan<- IBuilderNode) {
		qu := q
		qu.Exclusion = exclusion
		if q.idResolver != nil {
			ids, err := q.idResolver()
			if err != nil {
				buildFail(err)
				return
			}
			qu.Args = append(qu.Args, ids...)
		}
		builderChan <- qu
		return
	}
}

// buildTupleToUserSetQuery -
func (command *LookupQueryCommand) buildTupleToUserSetQuery(ctx context.Context, subject *base.Subject, entityType string, relation string, sch *base.Schema) QueryNode {
	var qu QueryNode
	var err error
	r := tuple.SplitRelation(relation)
	var entity *base.EntityDefinition
	entity, err = schema.GetEntityByName(sch, entityType)
	if err != nil {
		return QueryNode{
			Err: err,
		}
	}
	var rel *base.RelationDefinition
	rel, err = schema.GetRelationByNameInEntityDefinition(entity, r[0])
	if err != nil {
		return QueryNode{
			Err: err,
		}
	}
	column, columnExist := schema.GetColumn(rel)
	var parentEntity *base.EntityDefinition
	parentEntity, err = schema.GetEntityByName(sch, schema.GetEntityReference(rel))
	if err != nil {
		return QueryNode{
			Err: err,
		}
	}
	var parentRel *base.RelationDefinition
	parentRel, err = schema.GetRelationByNameInEntityDefinition(parentEntity, r[1])
	if err != nil {
		return QueryNode{
			Err: err,
		}
	}
	parentColumn, parentColumnExist := schema.GetColumn(parentRel)
	if !columnExist {
		qu.Key = fmt.Sprintf("%s.%s", schema.GetTable(entity), schema.GetIdentifier(entity))
		qu.idResolver = func() ([]string, error) {
			parentIDs, err := command.getEntityIDs(ctx, schema.GetEntityReference(rel), r[1], subject.GetType(), []string{subject.GetId()}, subject.GetRelation())
			if err != nil {
				return nil, err
			}
			if len(parentIDs) > 0 {
				return command.getEntityIDs(ctx, entity.Name, rel.Name, schema.GetEntityReference(rel), parentIDs, tuple.ELLIPSIS)
			}
			return []string{}, nil
		}
	} else {
		if !parentColumnExist {
			qu.Key = fmt.Sprintf("%s.%s", schema.GetTable(entity), column)
			qu.idResolver = func() ([]string, error) {
				return command.getEntityIDs(ctx, schema.GetEntityReference(rel), r[1], subject.Type, []string{subject.GetId()}, subject.Relation)
			}
		} else {
			qu.Key = fmt.Sprintf("%s.%s", schema.GetTable(parentEntity), parentColumn)
			j := join{
				key:   fmt.Sprintf("%s.%s", schema.GetTable(entity), column),
				value: fmt.Sprintf("%s.%s", schema.GetTable(parentEntity), schema.GetIdentifier(parentEntity)),
			}
			qu.Join = map[string]join{
				schema.GetTable(parentEntity): j,
			}
			qu.idResolver = func() ([]string, error) {
				return command.getUserIDs(ctx, subject)
			}
		}
	}
	return qu
}

// buildComputedUserSetQuery -
func (command *LookupQueryCommand) buildComputedUserSetQuery(ctx context.Context, subject *base.Subject, entityType string, relation string, sch *base.Schema) (qu QueryNode) {
	var err error
	var entity *base.EntityDefinition
	entity, err = schema.GetEntityByName(sch, entityType)
	if err != nil {
		return QueryNode{
			Err: err,
		}
	}
	var rel *base.RelationDefinition
	rel, err = schema.GetRelationByNameInEntityDefinition(entity, relation)
	if err != nil {
		return QueryNode{
			Err: err,
		}
	}
	column, columnExist := schema.GetColumn(rel)
	if !columnExist {
		qu.Key = fmt.Sprintf("%s.%s", schema.GetTable(entity), schema.GetIdentifier(entity))
		qu.idResolver = func() ([]string, error) {
			return command.getEntityIDs(ctx, entity.Name, rel.Name, subject.Type, []string{subject.GetId()}, subject.GetRelation())
		}
	} else {
		qu.Key = fmt.Sprintf("%s.%s", schema.GetTable(entity), column)
		qu.idResolver = func() ([]string, error) {
			return command.getUserIDs(ctx, subject)
		}
	}
	return qu
}

// buildSetOperation -
func buildSetOperation(
	ctx context.Context,
	functions []BuildFunction,
	op base.Rewrite_Operation,
) IBuilderNode {
	children := make([]IBuilderNode, 0, len(functions))

	c, cancel := context.WithCancel(ctx)
	defer cancel()

	result := make([]chan IBuilderNode, 0, len(functions))
	for _, fn := range functions {
		en := make(chan IBuilderNode)
		result = append(result, en)
		go fn(c, en)
	}

	for _, resultChan := range result {
		select {
		case res := <-resultChan:
			if res.Error() != nil {
				return LogicNode{
					Err: res.Error(),
				}
			}
			children = append(children, res)
		case <-ctx.Done():
			return nil
		}
	}

	return LogicNode{
		Operation: op,
		Children:  children,
	}
}

// buildUnion -
func buildUnion(ctx context.Context, functions []BuildFunction) IBuilderNode {
	return buildSetOperation(ctx, functions, base.Rewrite_UNION)
}

// buildIntersection -
func buildIntersection(ctx context.Context, functions []BuildFunction) IBuilderNode {
	return buildSetOperation(ctx, functions, base.Rewrite_INTERSECTION)
}

// buildFail -
func buildFail(err error) BuildFunction {
	return func(ctx context.Context, builderChan chan<- IBuilderNode) {
		builderChan <- LogicNode{
			Err: err,
		}
	}
}

// getUserIDs -
func (command *LookupQueryCommand) getUserIDs(ctx context.Context, s *base.Subject) (r []string, err error) {
	if tuple.IsSubjectUser(s) {
		r = append(r, s.GetId())
	} else {
		var iterator tuple.ISubjectIterator
		iterator, err = getSubjects(ctx, command, &base.EntityAndRelation{
			Entity: &base.Entity{
				Type: s.GetType(),
				Id:   s.GetId(),
			},
			Relation: s.Relation,
		})
		if err != nil {
			return nil, err
		}
		for iterator.HasNext() {
			ids, err := command.getUserIDs(ctx, iterator.GetNext())
			if err != nil {
				return nil, err
			}
			r = append(r, ids...)
		}
	}
	return helper.RemoveDuplicate(r), nil
}

// getEntityIDs -
func (command *LookupQueryCommand) getEntityIDs(ctx context.Context, entityType string, relation string, subjectType string, subjectIDs []string, subjectRelation string) (r []string, err error) {
	var iterator tuple.IEntityIterator
	iterator, err = command.getEntities(ctx, entityType, relation, subjectType, subjectIDs, subjectRelation)
	if err != nil {
		return nil, err
	}
	for iterator.HasNext() {
		r = append(r, iterator.GetNext().GetId())
	}
	return helper.RemoveDuplicate(r), nil
}

// getEntities -
func (command *LookupQueryCommand) getEntities(ctx context.Context, entityType string, relation string, subjectType string, subjectIDs []string, subjectRelation string) (iterator tuple.IEntityIterator, err error) {
	var tupleIterator tuple.ITupleIterator
	tupleIterator, err = command.relationTupleRepository.ReverseQueryTuples(ctx, entityType, relation, subjectType, subjectIDs, subjectRelation)
	if err != nil {
		return nil, err
	}

	collection := tuple.NewEntityCollection()

	for tupleIterator.HasNext() {
		collection.Add(tupleIterator.GetNext().GetEntity())
	}

	return collection.CreateEntityIterator(), err
}

// StatementBuilder -
type StatementBuilder struct {
	joins map[string]join
}

// rootNodeToSql -
func rootNodeToSql(node IBuilderNode, table string) (sql string, args []interface{}, e error) {
	var err error
	st := &StatementBuilder{
		joins: map[string]join{},
	}
	expressions, _ := st.buildSql([]IBuilderNode{node}, goqu.Ex{})
	if len(st.joins) > 0 {
		ex := goqu.From(table).Where(expressions)
		for t, j := range st.joins {
			ex = ex.InnerJoin(goqu.T(t), goqu.On(goqu.Ex{j.key: goqu.I(j.value)}))
		}
		sql, args, err = ex.Prepared(true).ToSQL()
	} else {
		q := goqu.From(table).Where(expressions)
		sql, args, err = q.Prepared(true).ToSQL()
		if err != nil {
			return "", []interface{}{}, errors.New(base.ErrorCode_sql_builder_error.String())
		}
	}
	return strings.ReplaceAll(sql, "\"", ""), args, nil
}

// buildSql -
func (builder *StatementBuilder) buildSql(children []IBuilderNode, exp goqu.Expression) (m goqu.Expression, s []goqu.Expression) {
	var ex []goqu.Expression
	for _, child := range children {
		switch child.GetKind() {
		case LOGIC:
			ln := child.(LogicNode)
			b, sub := builder.buildSql(ln.Children, exp)
			sub = append(sub, b.Expression())
			switch ln.Operation {
			case base.Rewrite_UNION:
				exp = goqu.Or(sub...)
			case base.Rewrite_INTERSECTION:
				exp = goqu.And(sub...)
			}
		case QUERY:
			qn := child.(QueryNode)
			if qn.Exclusion {
				if qn.Args != nil {
					if len(qn.Args) == 1 {
						ex = append(ex, goqu.I(qn.Key).Neq(qn.Args[0]))
					} else {
						ex = append(ex, goqu.I(qn.Key).NotIn(qn.Args))
					}
				}
			} else {
				if qn.Args == nil {
					ex = append(ex, goqu.I(qn.Key).Is(nil))
				} else if len(qn.Args) == 1 {
					ex = append(ex, goqu.I(qn.Key).Eq(qn.Args[0]))
				} else {
					ex = append(ex, goqu.I(qn.Key).In(qn.Args))
				}
			}
			if len(qn.Join) > 0 {
				for k, j := range qn.Join {
					builder.joins[k] = j
				}
			}
		}
	}
	return exp, ex
}
