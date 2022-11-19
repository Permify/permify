package memory

import (
	"context"
	"errors"

	"github.com/hashicorp/go-memdb"

	"github.com/Permify/permify/internal/repositories"
	"github.com/Permify/permify/internal/repositories/memory/snapshot"
	"github.com/Permify/permify/internal/repositories/memory/utils"
	"github.com/Permify/permify/pkg/database"
	db "github.com/Permify/permify/pkg/database/memory"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/token"
)

// RelationshipReader - Structure for Relationship Reader
type RelationshipReader struct {
	database *db.Memory
}

// NewRelationshipReader - Creates a new RelationshipReader
func NewRelationshipReader(database *db.Memory) *RelationshipReader {
	return &RelationshipReader{
		database: database,
	}
}

// QueryRelationships - Reads relation tuples from the repository.
func (r *RelationshipReader) QueryRelationships(ctx context.Context, filter *base.TupleFilter, _ string) (collection database.ITupleCollection, err error) {
	txn := r.database.DB.Txn(false)
	defer txn.Abort()

	collection = database.NewTupleCollection()

	index, args := utils.GetIndexNameAndArgsByFilters(filter)
	var it memdb.ResultIterator
	it, err = txn.Get(RelationTuplesTable, index, args...)
	if err != nil {
		return nil, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
	}

	fit := memdb.NewFilterIterator(it, utils.FilterQuery(filter))
	for obj := fit.Next(); obj != nil; obj = it.Next() {
		t := obj.(repositories.RelationTuple)
		collection.Add(t.ToTuple())
	}

	return collection, nil
}

// HeadSnapshot - Reads the latest version of the snapshot from the repository.
func (r *RelationshipReader) HeadSnapshot(ctx context.Context) (token.SnapToken, error) {
	return snapshot.Token{}, nil
}
