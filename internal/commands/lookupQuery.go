package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/doug-martin/goqu/v9"

	internalErrors "github.com/Permify/permify/internal/errors"
	"github.com/Permify/permify/internal/repositories"
	"github.com/Permify/permify/internal/repositories/entities"
	"github.com/Permify/permify/pkg/dsl/schema"
	"github.com/Permify/permify/pkg/errors"
	"github.com/Permify/permify/pkg/helper"
	"github.com/Permify/permify/pkg/logger"
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
	Operation schema.OPType  `json:"operation"`
	Children  []IBuilderNode `json:"children"`
	Err       error          `json:"-"`
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

// BuildFunction -
type BuildFunction func(ctx context.Context, resultChan chan<- IBuilderNode)

// ResolverFunction -
type ResolverFunction func() ([]string, errors.Error)

// BuildCombiner .
type BuildCombiner func(ctx context.Context, functions []BuildFunction) IBuilderNode

// NewLookupQueryCommand -
func NewLookupQueryCommand(rr repositories.IRelationTupleRepository, l logger.Interface) *LookupQueryCommand {
	return &LookupQueryCommand{
		logger:                  l,
		relationTupleRepository: rr,
	}
}

// LookupQueryQuery -
type LookupQueryQuery struct {
	EntityType string
	Action     string
	Subject    tuple.Subject
	schema     schema.Schema
}

// SetSchema -
func (l *LookupQueryQuery) SetSchema(sch schema.Schema) {
	l.schema = sch
}

// LookupQueryResponse -
type LookupQueryResponse struct {
	Node  IBuilderNode
	Query string
	Args  []interface{}
}

// NewLookupQueryResponse -
func NewLookupQueryResponse(Node IBuilderNode, table string) (LookupQueryResponse, errors.Error) {
	query, args, err := rootNodeToSql(Node, table)
	return LookupQueryResponse{
		Node:  Node,
		Query: query,
		Args:  args,
	}, err
}

// Execute -
func (command *LookupQueryCommand) Execute(ctx context.Context, q *LookupQueryQuery, child schema.Child) (response LookupQueryResponse, err errors.Error) {
	return command.l(ctx, q, child)
}

func (command *LookupQueryCommand) l(ctx context.Context, q *LookupQueryQuery, child schema.Child) (response LookupQueryResponse, err errors.Error) {
	en, _ := q.schema.GetEntityByName(q.EntityType)

	var fn BuildFunction
	switch child.GetKind() {
	case schema.RewriteKind.String():
		fn = command.buildRewrite(ctx, q, child.(schema.Rewrite))
	case schema.LeafKind.String():
		fn = command.buildLeaf(ctx, q, child.(schema.Leaf))
	}

	if fn == nil {
		return LookupQueryResponse{}, internalErrors.UndefinedChildKindError
	}

	result := buildUnion(ctx, []BuildFunction{fn})
	return NewLookupQueryResponse(result, en.GetTable())
}

// set -
func (command *LookupQueryCommand) set(ctx context.Context, q *LookupQueryQuery, children []schema.Child, combiner BuildCombiner) BuildFunction {
	var functions []BuildFunction
	for _, child := range children {
		switch child.GetKind() {
		case schema.RewriteKind.String():
			functions = append(functions, command.buildRewrite(ctx, q, child.(schema.Rewrite)))
		case schema.LeafKind.String():
			functions = append(functions, command.buildLeaf(ctx, q, child.(schema.Leaf)))
		default:
			return buildFail(internalErrors.UndefinedChildKindError)
		}
	}

	return func(ctx context.Context, resultChan chan<- IBuilderNode) {
		resultChan <- combiner(ctx, functions)
	}
}

// buildRewrite -
func (command *LookupQueryCommand) buildRewrite(ctx context.Context, q *LookupQueryQuery, child schema.Rewrite) BuildFunction {
	switch child.GetType() {
	case schema.Union.String():
		return command.set(ctx, q, child.Children, buildUnion)
	case schema.Intersection.String():
		return command.set(ctx, q, child.Children, buildIntersection)
	default:
		return buildFail(internalErrors.UndefinedChildTypeError)
	}
}

// buildLeaf -
func (command *LookupQueryCommand) buildLeaf(ctx context.Context, q *LookupQueryQuery, child schema.Leaf) BuildFunction {
	switch child.GetType() {
	case schema.TupleToUserSetType.String():
		return command.build(ctx, child.Exclusion, command.buildTupleToUserSetQuery(ctx, q.Subject, q.EntityType, tuple.Relation(child.Value), q.schema))
	case schema.ComputedUserSetType.String():
		return command.build(ctx, child.Exclusion, command.buildComputedUserSetQuery(ctx, q.Subject, q.EntityType, tuple.Relation(child.Value), q.schema))
	default:
		return buildFail(internalErrors.UndefinedChildTypeError)
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
func (command *LookupQueryCommand) buildTupleToUserSetQuery(ctx context.Context, subject tuple.Subject, entityType string, relation tuple.Relation, sch schema.Schema) QueryNode {
	var qu QueryNode
	var err error
	r := relation.Split()
	var entity schema.Entity
	entity, err = sch.GetEntityByName(entityType)
	if err != nil {
		return QueryNode{
			Err: err,
		}
	}
	var rel schema.Relation
	rel, err = schema.Relations(entity.Relations).GetRelationByName(r[0].String())
	if err != nil {
		return QueryNode{
			Err: err,
		}
	}
	column, columnExist := rel.GetColumn()
	var parentEntity schema.Entity
	parentEntity, err = sch.GetEntityByName(rel.Type())
	if err != nil {
		return QueryNode{
			Err: err,
		}
	}
	var parentRel schema.Relation
	parentRel, err = schema.Relations(parentEntity.Relations).GetRelationByName(r[1].String())
	if err != nil {
		return QueryNode{
			Err: err,
		}
	}
	parentColumn, parentColumnExist := parentRel.GetColumn()
	if !columnExist {
		qu.Key = fmt.Sprintf("%s.%s", entity.GetTable(), entity.GetIdentifier())
		qu.idResolver = func() ([]string, errors.Error) {
			parentIDs, err := command.getEntityIDs(ctx, rel.Type(), tuple.Relation(r[1].String()), subject.Type, []string{subject.ID}, subject.Relation)
			if err != nil {
				return nil, err
			}
			if len(parentIDs) > 0 {
				return command.getEntityIDs(ctx, entity.Name, tuple.Relation(rel.Name), rel.Type(), parentIDs, tuple.ELLIPSIS)
			}
			return []string{}, nil
		}
	} else {
		if !parentColumnExist {
			qu.Key = fmt.Sprintf("%s.%s", entity.GetTable(), column)
			qu.idResolver = func() ([]string, errors.Error) {
				return command.getEntityIDs(ctx, rel.Type(), tuple.Relation(r[1].String()), subject.Type, []string{subject.ID}, subject.Relation)
			}
		} else {
			qu.Key = fmt.Sprintf("%s.%s", parentEntity.GetTable(), parentColumn)
			j := join{
				key:   fmt.Sprintf("%s.%s", entity.GetTable(), column),
				value: fmt.Sprintf("%s.%s", parentEntity.GetTable(), parentEntity.GetIdentifier()),
			}
			qu.Join = map[string]join{
				parentEntity.GetTable(): j,
			}
			qu.idResolver = func() ([]string, errors.Error) {
				return command.getUserIDs(ctx, subject)
			}
		}
	}
	return qu
}

// buildComputedUserSetQuery -
func (command *LookupQueryCommand) buildComputedUserSetQuery(ctx context.Context, subject tuple.Subject, entityType string, relation tuple.Relation, sch schema.Schema) (qu QueryNode) {
	var err error
	var entity schema.Entity
	entity, err = sch.GetEntityByName(entityType)
	if err != nil {
		return QueryNode{
			Err: err,
		}
	}
	var rel schema.Relation
	rel, err = schema.Relations(entity.Relations).GetRelationByName(relation.String())
	if err != nil {
		return QueryNode{
			Err: err,
		}
	}
	column, columnExist := rel.GetColumn()
	if !columnExist {
		qu.Key = fmt.Sprintf("%s.%s", entity.GetTable(), entity.GetIdentifier())
		qu.idResolver = func() ([]string, errors.Error) {
			return command.getEntityIDs(ctx, entity.Name, tuple.Relation(rel.Name), subject.Type, []string{subject.ID}, subject.Relation)
		}
	} else {
		qu.Key = fmt.Sprintf("%s.%s", entity.GetTable(), column)
		qu.idResolver = func() ([]string, errors.Error) {
			return command.getUserIDs(ctx, subject)
		}
	}
	return qu
}

// buildSetOperation -
func buildSetOperation(
	ctx context.Context,
	functions []BuildFunction,
	op schema.OPType,
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

	return &LogicNode{
		Operation: op,
		Children:  children,
	}
}

// buildUnion -
func buildUnion(ctx context.Context, functions []BuildFunction) IBuilderNode {
	return buildSetOperation(ctx, functions, schema.Union)
}

// buildIntersection -
func buildIntersection(ctx context.Context, functions []BuildFunction) IBuilderNode {
	return buildSetOperation(ctx, functions, schema.Intersection)
}

// buildFail -
func buildFail(err errors.Error) BuildFunction {
	return func(ctx context.Context, builderChan chan<- IBuilderNode) {
		builderChan <- LogicNode{
			Err: err,
		}
	}
}

// getUserIDs -
func (command *LookupQueryCommand) getUserIDs(ctx context.Context, s tuple.Subject) (r []string, err errors.Error) {
	if s.IsUser() {
		r = append(r, s.ID)
	} else {
		iterator, err := command.getSubjects(ctx, tuple.Entity{
			Type: s.Type,
			ID:   s.ID,
		}, s.Relation)
		if err != nil {
			return nil, err
		}
		for iterator.HasNext() {
			ids, err := command.getUserIDs(ctx, *iterator.GetNext())
			if err != nil {
				return nil, err
			}
			r = append(r, ids...)
		}
	}
	return helper.RemoveDuplicate(r), nil
}

// getEntityIDs -
func (command *LookupQueryCommand) getEntityIDs(ctx context.Context, entityType string, relation tuple.Relation, subjectType string, subjectIDs []string, subjectRelation tuple.Relation) (r []string, error2 errors.Error) {
	iterator, err := command.getEntities(ctx, entityType, relation, subjectType, subjectIDs, subjectRelation)
	if err != nil {
		return nil, err
	}
	for iterator.HasNext() {
		entity := *iterator.GetNext()
		r = append(r, entity.ID)
	}
	return helper.RemoveDuplicate(r), nil
}

// getSubjects -
func (command *LookupQueryCommand) getSubjects(ctx context.Context, entity tuple.Entity, relation tuple.Relation) (iterator tuple.ISubjectIterator, err errors.Error) {
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

// getEntities -
func (command *LookupQueryCommand) getEntities(ctx context.Context, entityType string, relation tuple.Relation, subjectType string, subjectIDs []string, subjectRelation tuple.Relation) (iterator tuple.IEntityIterator, err errors.Error) {
	var tuples []entities.RelationTuple
	tuples, err = command.relationTupleRepository.ReverseQueryTuples(ctx, entityType, relation.String(), subjectType, subjectIDs, subjectRelation.String())
	if err != nil {
		return nil, err
	}

	var ent []*tuple.Entity
	for _, tup := range tuples {
		ct := tup.ToTuple()
		ent = append(ent, &ct.Entity)
	}

	return tuple.NewEntityIterator(ent), err
}

// StatementBuilder -
type StatementBuilder struct {
	joins map[string]join
}

// rootNodeToSql -
func rootNodeToSql(node IBuilderNode, table string) (sql string, args []interface{}, e errors.Error) {
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
			return "", []interface{}{}, errors.NewError(errors.Service).SetMessage("sql convert error")
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
			ln := child.(*LogicNode)
			b, sub := builder.buildSql(ln.Children, exp)
			sub = append(sub, b.Expression())
			switch ln.Operation {
			case schema.Union:
				exp = goqu.Or(sub...)
			case schema.Intersection:
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
