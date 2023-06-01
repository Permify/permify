package memory

import (
	"context"
	"errors"
	"sort"
	"strconv"
	"time"

	"github.com/hashicorp/go-memdb"

	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/internal/storage/memory/snapshot"
	"github.com/Permify/permify/internal/storage/memory/utils"
	"github.com/Permify/permify/pkg/database"
	db "github.com/Permify/permify/pkg/database/memory"
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
func (r *RelationshipReader) QueryRelationships(ctx context.Context, tenantID string, filter *base.TupleFilter, _ string) (it *database.TupleIterator, err error) {
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
		t, ok := obj.(storage.RelationTuple)
		if !ok {
			return nil, errors.New(base.ErrorCode_ERROR_CODE_TYPE_CONVERSATION.String())
		}
		collection.Add(t.ToTuple())
	}

	return collection.CreateTupleIterator(), nil
}

// ReadRelationships - Gets all relationships for a given filter
func (r *RelationshipReader) ReadRelationships(ctx context.Context, tenantID string, filter *base.TupleFilter, _ string, pagination database.Pagination) (collection *database.TupleCollection, ct database.EncodedContinuousToken, err error) {
	txn := r.database.DB.Txn(false)
	defer txn.Abort()

	var lowerBound uint64
	if pagination.Token() != "" {
		var t database.ContinuousToken
		t, err = utils.EncodedContinuousToken{Value: pagination.Token()}.Decode()
		if err != nil {
			return nil, database.NewNoopContinuousToken().Encode(), err
		}
		lowerBound, err = strconv.ParseUint(t.(utils.ContinuousToken).Value, 10, 64)
		if err != nil {
			return nil, database.NewNoopContinuousToken().Encode(), errors.New(base.ErrorCode_ERROR_CODE_INVALID_CONTINUOUS_TOKEN.String())
		}
	}

	index, args := utils.GetIndexNameAndArgsByFilters(tenantID, filter)

	var result memdb.ResultIterator
	result, err = txn.LowerBound(RelationTuplesTable, index, args...)
	if err != nil {
		return nil, database.NewNoopContinuousToken().Encode(), errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
	}

	tup := make([]storage.RelationTuple, 0, 10)
	fit := memdb.NewFilterIterator(result, utils.FilterQuery(filter))
	for obj := fit.Next(); obj != nil; obj = fit.Next() {
		t, ok := obj.(storage.RelationTuple)
		if !ok {
			return nil, database.NewNoopContinuousToken().Encode(), errors.New(base.ErrorCode_ERROR_CODE_TYPE_CONVERSATION.String())
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
				return database.NewTupleCollection(tuples[:pagination.PageSize()]...), utils.NewContinuousToken(strconv.FormatUint(t.ID, 10)).Encode(), nil
			}
		}
	}

	return database.NewTupleCollection(tuples...), database.NewNoopContinuousToken().Encode(), nil
}

// HeadSnapshot - Reads the latest version of the snapshot from the repository.
func (r *RelationshipReader) HeadSnapshot(ctx context.Context, _ string) (token.SnapToken, error) {
	return snapshot.NewToken(time.Now()), nil
}
