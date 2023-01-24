package memory

import (
	"context"
	"errors"
	"sort"
	"time"

	"github.com/hashicorp/go-memdb"

	"github.com/Permify/permify/internal/repositories"
	"github.com/Permify/permify/internal/repositories/memory/snapshot"
	"github.com/Permify/permify/internal/repositories/memory/utils"
	"github.com/Permify/permify/pkg/database"
	db "github.com/Permify/permify/pkg/database/memory"
	"github.com/Permify/permify/pkg/helper"
	"github.com/Permify/permify/pkg/logger"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/token"
)

// RelationshipReader - Structure for Relationship Reader
type RelationshipReader struct {
	database *db.Memory
	// logger
	logger logger.Interface
}

// NewRelationshipReader - Creates a new RelationshipReader
func NewRelationshipReader(database *db.Memory, logger logger.Interface) *RelationshipReader {
	return &RelationshipReader{
		database: database,
		logger:   logger,
	}
}

// QueryRelationships - Reads relation tuples from the repository.
func (r *RelationshipReader) QueryRelationships(ctx context.Context, tenantID uint64, filter *base.TupleFilter, _ string) (it *database.TupleIterator, err error) {
	txn := r.database.DB.Txn(false)
	defer txn.Abort()

	collection := database.NewTupleCollection()

	index, args := utils.GetIndexNameAndArgsByFilters(tenantID, filter)
	var result memdb.ResultIterator

	result, err = txn.Get(RelationTuplesTable, index, args...)
	if err != nil {
		return nil, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
	}

	fit := memdb.NewFilterIterator(result, utils.FilterQuery(filter))
	for obj := fit.Next(); obj != nil; obj = fit.Next() {
		t, ok := obj.(repositories.RelationTuple)
		if !ok {
			return nil, errors.New(base.ErrorCode_ERROR_CODE_TYPE_CONVERSATION.String())
		}
		collection.Add(t.ToTuple())
	}

	return collection.CreateTupleIterator(), nil
}

// ReadRelationships - Gets all relationships for a given filter
func (r *RelationshipReader) ReadRelationships(ctx context.Context, tenantID uint64, filter *base.TupleFilter, _ string, pagination database.Pagination) (collection *database.TupleCollection, ct database.EncodedContinuousToken, err error) {
	txn := r.database.DB.Txn(false)
	defer txn.Abort()

	var lowerBound uint64
	if pagination.Token() != "" {
		var t database.ContinuousToken
		t, err = utils.EncodedContinuousToken{Value: pagination.Token()}.Decode()
		if err != nil {
			return nil, nil, err
		}
		lowerBound = t.(utils.ContinuousToken).Value
	}

	index, args := utils.GetIndexNameAndArgsByFilters(tenantID, filter)
	var result memdb.ResultIterator

	result, err = txn.LowerBound(RelationTuplesTable, index, args...)
	if err != nil {
		return nil, nil, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
	}

	tup := make([]repositories.RelationTuple, 0, 10)
	fit := memdb.NewFilterIterator(result, utils.FilterQuery(filter))
	for obj := fit.Next(); obj != nil; obj = fit.Next() {
		t, ok := obj.(repositories.RelationTuple)
		if !ok {
			return nil, nil, errors.New(base.ErrorCode_ERROR_CODE_TYPE_CONVERSATION.String())
		}
		tup = append(tup, t)
	}

	sort.Slice(tup, func(i, j int) bool {
		return tup[i].ID < tup[j].ID
	})

	tuples := make([]*base.Tuple, 0, pagination.PageSize()+1)

	for _, t := range tup {
		if t.ID >= lowerBound {
			tuples = append(tuples, t.ToTuple())
			if len(tuples) > int(pagination.PageSize()) {
				return database.NewTupleCollection(tuples[:pagination.PageSize()]...), utils.NewContinuousToken(t.ID).Encode(), nil
			}
		}
	}

	return database.NewTupleCollection(tuples...), utils.NewNoopContinuousToken().Encode(), nil
}

// GetUniqueEntityIDsByEntityType - Gets all entity IDs for a given entity type (unique)
func (r *RelationshipReader) GetUniqueEntityIDsByEntityType(ctx context.Context, tenantID uint64, typ, _ string) (array []string, err error) {
	txn := r.database.DB.Txn(false)
	defer txn.Abort()

	var it memdb.ResultIterator
	it, err = txn.Get(RelationTuplesTable, "entity-type-index", tenantID, typ)
	if err != nil {
		return nil, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
	}

	var result []string
	for obj := it.Next(); obj != nil; obj = it.Next() {
		t, ok := obj.(repositories.RelationTuple)
		if !ok {
			return nil, errors.New(base.ErrorCode_ERROR_CODE_TYPE_CONVERSATION.String())
		}
		result = append(result, t.EntityID)
	}

	return helper.RemoveDuplicate(result), nil
}

// HeadSnapshot - Reads the latest version of the snapshot from the repository.
func (r *RelationshipReader) HeadSnapshot(ctx context.Context, _ uint64) (token.SnapToken, error) {
	return snapshot.NewToken(time.Now()), nil
}
