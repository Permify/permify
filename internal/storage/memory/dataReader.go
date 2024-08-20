package memory

import (
	"context"
	"errors"
	"sort"
	"strconv"
	"time"

	"github.com/hashicorp/go-memdb"

	"github.com/Permify/permify/internal/storage/memory/constants"

	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/internal/storage/memory/snapshot"
	"github.com/Permify/permify/internal/storage/memory/utils"
	"github.com/Permify/permify/pkg/database"
	db "github.com/Permify/permify/pkg/database/memory"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/token"
)

// DataReader -
type DataReader struct {
	database *db.Memory
}

// NewDataReader - Creates a new DataReader
func NewDataReader(database *db.Memory) *DataReader {
	return &DataReader{
		database: database,
	}
}

// QueryRelationships queries the database for relationships based on the provided filter.
func (r *DataReader) QueryRelationships(_ context.Context, tenantID string, filter *base.TupleFilter, _ string, pagination database.CursorPagination) (it *database.TupleIterator, err error) {
	txn := r.database.DB.Txn(false)
	defer txn.Abort()

	var lowerBound string

	if pagination.Cursor() != "" {
		var t database.ContinuousToken
		t, err = utils.EncodedContinuousToken{Value: pagination.Cursor()}.Decode()
		if err != nil {
			return nil, err
		}
		lowerBound = t.(utils.ContinuousToken).Value
	}

	// Get the index and arguments based on the filter.
	index, args := utils.GetRelationTuplesIndexNameAndArgsByFilters(tenantID, filter)

	// Get the result iterator based on the index and arguments.
	var result memdb.ResultIterator
	result, err = txn.LowerBound(constants.RelationTuplesTable, index, args...)
	if err != nil {
		return nil, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
	}

	// Filter the result iterator and add the tuples to the collection.
	tup := make([]storage.RelationTuple, 0, 10)
	fit := memdb.NewFilterIterator(result, utils.FilterRelationTuplesQuery(tenantID, filter))
	for obj := fit.Next(); obj != nil; obj = fit.Next() {
		t, ok := obj.(storage.RelationTuple)
		if !ok {
			return nil, errors.New(base.ErrorCode_ERROR_CODE_TYPE_CONVERSATION.String())
		}
		tup = append(tup, t)
	}

	// Sort tuples based on the provided order field
	sort.SliceStable(tup, func(i, j int) bool {
		switch pagination.Sort() {
		case "entity_id":
			return tup[i].EntityID < tup[j].EntityID
		case "subject_id":
			return tup[i].SubjectID < tup[j].SubjectID
		default:
			return false // No sorting if order field is invalid
		}
	})

	var tuples []*base.Tuple
	for _, t := range tup {
		switch pagination.Sort() {
		case "entity_id":
			if t.EntityID >= lowerBound {
				tuples = append(tuples, t.ToTuple())
			}
		case "subject_id":
			if t.SubjectID >= lowerBound {
				tuples = append(tuples, t.ToTuple())
			}
		default:
			tuples = append(tuples, t.ToTuple())
		}
	}

	return database.NewTupleCollection(tuples...).CreateTupleIterator(), nil
}

// ReadRelationships reads relationships from the database taking into account the pagination.
func (r *DataReader) ReadRelationships(_ context.Context, tenantID string, filter *base.TupleFilter, _ string, pagination database.Pagination) (collection *database.TupleCollection, ct database.EncodedContinuousToken, err error) {
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

	index, args := utils.GetRelationTuplesIndexNameAndArgsByFilters(tenantID, filter)

	// Get the result iterator using lower bound.
	var result memdb.ResultIterator
	result, err = txn.LowerBound(constants.RelationTuplesTable, index, args...)
	if err != nil {
		return nil, database.NewNoopContinuousToken().Encode(), errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
	}

	// Filter the result iterator and add the tuples to the array.
	tup := make([]storage.RelationTuple, 0, 10)
	fit := memdb.NewFilterIterator(result, utils.FilterRelationTuplesQuery(tenantID, filter))
	for obj := fit.Next(); obj != nil; obj = fit.Next() {
		t, ok := obj.(storage.RelationTuple)
		if !ok {
			return nil, database.NewNoopContinuousToken().Encode(), errors.New(base.ErrorCode_ERROR_CODE_TYPE_CONVERSATION.String())
		}
		tup = append(tup, t)
	}

	// Sort the tuples and append them to the collection.
	sort.Slice(tup, func(i, j int) bool {
		return tup[i].ID < tup[j].ID
	})

	tuples := make([]*base.Tuple, 0, pagination.PageSize()+1)
	for _, t := range tup {
		if t.ID >= lowerBound {
			tuples = append(tuples, t.ToTuple())
			if pagination.PageSize() != 0 && len(tuples) > int(pagination.PageSize()) {
				return database.NewTupleCollection(tuples[:pagination.PageSize()]...), utils.NewContinuousToken(strconv.FormatUint(t.ID, 10)).Encode(), nil
			}
		}
	}

	return database.NewTupleCollection(tuples...), database.NewNoopContinuousToken().Encode(), nil
}

// QuerySingleAttribute queries the database for a single attribute based on the provided filter.
func (r *DataReader) QuerySingleAttribute(_ context.Context, tenantID string, filter *base.AttributeFilter, _ string) (attribute *base.Attribute, err error) {
	txn := r.database.DB.Txn(false)
	defer txn.Abort()

	// Get the index and arguments based on the filter.
	index, args := utils.GetAttributesIndexNameAndArgsByFilters(tenantID, filter)

	// Get the result iterator based on the index and arguments.
	var result memdb.ResultIterator
	result, err = txn.Get(constants.AttributesTable, index, args...)
	if err != nil {
		return nil, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
	}

	// Filter the result iterator and add the attributes to the collection.
	fit := memdb.NewFilterIterator(result, utils.FilterAttributesQuery(tenantID, filter))
	for obj := fit.Next(); obj != nil; {
		t, ok := obj.(storage.Attribute)
		if !ok {
			return nil, errors.New(base.ErrorCode_ERROR_CODE_TYPE_CONVERSATION.String())
		}
		return t.ToAttribute(), nil
	}

	return nil, nil
}

// QueryAttributes queries the database for attributes based on the provided filter.
func (r *DataReader) QueryAttributes(_ context.Context, tenantID string, filter *base.AttributeFilter, _ string) (iterator *database.AttributeIterator, err error) {
	txn := r.database.DB.Txn(false)
	defer txn.Abort()

	collection := database.NewAttributeCollection()

	// Get the index and arguments based on the filter.
	index, args := utils.GetAttributesIndexNameAndArgsByFilters(tenantID, filter)

	// Get the result iterator based on the index and arguments.
	var result memdb.ResultIterator
	result, err = txn.Get(constants.AttributesTable, index, args...)
	if err != nil {
		return nil, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
	}

	// Filter the result iterator and add the attributes to the collection.
	fit := memdb.NewFilterIterator(result, utils.FilterAttributesQuery(tenantID, filter))
	for obj := fit.Next(); obj != nil; obj = fit.Next() {
		t, ok := obj.(storage.Attribute)
		if !ok {
			return nil, errors.New(base.ErrorCode_ERROR_CODE_TYPE_CONVERSATION.String())
		}
		collection.Add(t.ToAttribute())
	}

	return collection.CreateAttributeIterator(), nil
}

// ReadAttributes reads attributes from the database taking into account the pagination.
func (r *DataReader) ReadAttributes(_ context.Context, tenantID string, filter *base.AttributeFilter, _ string, pagination database.Pagination) (collection *database.AttributeCollection, ct database.EncodedContinuousToken, err error) {
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

	// Get the index and arguments based on the filter.
	index, args := utils.GetAttributesIndexNameAndArgsByFilters(tenantID, filter)

	// Get the result iterator using lower bound.
	var result memdb.ResultIterator
	result, err = txn.LowerBound(constants.AttributesTable, index, args...)
	if err != nil {
		return nil, database.NewNoopContinuousToken().Encode(), errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
	}

	// Filter the result iterator and add the attributes to the array.
	attr := make([]storage.Attribute, 0, 10)
	fit := memdb.NewFilterIterator(result, utils.FilterAttributesQuery(tenantID, filter))
	for obj := fit.Next(); obj != nil; obj = fit.Next() {
		a, ok := obj.(storage.Attribute)
		if !ok {
			return nil, database.NewNoopContinuousToken().Encode(), errors.New(base.ErrorCode_ERROR_CODE_TYPE_CONVERSATION.String())
		}
		attr = append(attr, a)
	}

	// Sort the attributes and append them to the collection.
	sort.Slice(attr, func(i, j int) bool {
		return attr[i].ID < attr[j].ID
	})

	attributes := make([]*base.Attribute, 0, pagination.PageSize()+1)
	for _, t := range attr {
		if t.ID >= lowerBound {
			attributes = append(attributes, t.ToAttribute())
			if pagination.PageSize() != 0 && len(attributes) > int(pagination.PageSize()) {
				return database.NewAttributeCollection(attributes[:pagination.PageSize()]...), utils.NewContinuousToken(strconv.FormatUint(t.ID, 10)).Encode(), nil
			}
		}
	}

	return database.NewAttributeCollection(attributes...), database.NewNoopContinuousToken().Encode(), nil
}

// QueryUniqueEntities is a function that searches for unique entities in a given database.
func (r *DataReader) QueryUniqueEntities(_ context.Context, tenantID, name, _ string, pagination database.Pagination) (ids []string, ct database.EncodedContinuousToken, err error) {
	// Starts a new read-only transaction
	txn := r.database.DB.Txn(false)
	defer txn.Abort()

	var lowerBound string
	if pagination.Token() != "" {
		var t database.ContinuousToken
		t, err := utils.EncodedContinuousToken{Value: pagination.Token()}.Decode()
		if err != nil {
			return nil, database.NewNoopContinuousToken().Encode(), err
		}
		lowerBound = t.(utils.ContinuousToken).Value
	}

	var tupleIds []string

	// Query the database for entities matching the given tenant ID and name
	var entityResult memdb.ResultIterator
	entityResult, err = txn.LowerBound(constants.RelationTuplesTable, "entity-type-index", tenantID, name)
	if err != nil {
		// Returns an error if execution fails
		return nil, database.NewNoopContinuousToken().Encode(), errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
	}

	// Iterates over the resulting entities and append their IDs to the tupleIds slice
	for obj := entityResult.Next(); obj != nil; obj = entityResult.Next() {
		t, ok := obj.(storage.RelationTuple)
		if !ok {
			// Returns an error if type conversion fails
			return nil, database.NewNoopContinuousToken().Encode(), errors.New(base.ErrorCode_ERROR_CODE_TYPE_CONVERSATION.String())
		}
		tupleIds = append(tupleIds, t.EntityID)
	}

	var attributeIds []string

	// Query the database for attributes matching the given tenant ID and name
	var attributeResult memdb.ResultIterator
	attributeResult, err = txn.LowerBound(constants.AttributesTable, "entity-type-index", tenantID, name)
	if err != nil {
		// Returns an error if execution fails
		return nil, database.NewNoopContinuousToken().Encode(), errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
	}

	// Iterates over the resulting attributes and append their IDs to the tupleIds slice
	for obj := attributeResult.Next(); obj != nil; obj = attributeResult.Next() {
		t, ok := obj.(storage.Attribute)
		if !ok {
			// Returns an error if type conversion fails
			return nil, database.NewNoopContinuousToken().Encode(), errors.New(base.ErrorCode_ERROR_CODE_TYPE_CONVERSATION.String())
		}
		attributeIds = append(attributeIds, t.EntityID)
	}

	all := append(tupleIds, attributeIds...)

	// Sort the combined slice
	sort.SliceStable(all, func(i, j int) bool {
		return all[i] < all[j]
	})

	mp := make(map[string]bool)
	var lastID string

	for _, b := range all {
		if _, exists := mp[b]; !exists && b >= lowerBound {
			ids = append(ids, b)
			mp[b] = true

			// Capture the last ID after adding pagesize + 1 elements
			if len(ids) == int(pagination.PageSize())+1 {
				lastID = b
			}

			// Stop appending if we've reached the page size
			if pagination.PageSize() != 0 && len(ids) > int(pagination.PageSize()) {
				return ids[:pagination.PageSize()], utils.NewContinuousToken(lastID).Encode(), nil
			}
		}
	}

	// If page size is not exceeded, return the entire list with a noop token
	return ids, database.NewNoopContinuousToken().Encode(), nil
}

// QueryUniqueSubjectReferences is a function that searches for unique subject references in a given database.
func (r *DataReader) QueryUniqueSubjectReferences(_ context.Context, tenantID string, subjectReference *base.RelationReference, _ string, pagination database.Pagination) (ids []string, _ database.EncodedContinuousToken, err error) {
	txn := r.database.DB.Txn(false)
	defer txn.Abort()

	var lowerBound string
	if pagination.Token() != "" {
		var t database.ContinuousToken
		t, err := utils.EncodedContinuousToken{Value: pagination.Token()}.Decode()
		if err != nil {
			return nil, database.NewNoopContinuousToken().Encode(), err
		}
		lowerBound = t.(utils.ContinuousToken).Value
	}

	// Get the result iterator based on the index and arguments.
	result, err := txn.LowerBound(constants.RelationTuplesTable, "id")
	if err != nil {
		return nil, database.NewNoopContinuousToken().Encode(), errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
	}

	var subjectIDs []string

	// Filter the result iterator and add the tuples to the collection.
	fit := memdb.NewFilterIterator(result, utils.FilterRelationTuplesQuery(tenantID, &base.TupleFilter{
		Subject: &base.SubjectFilter{
			Type:     subjectReference.GetType(),
			Relation: subjectReference.GetRelation(),
		},
	}))
	for obj := fit.Next(); obj != nil; obj = fit.Next() {
		t, ok := obj.(storage.RelationTuple)
		if !ok {
			return nil, database.NewNoopContinuousToken().Encode(), errors.New(base.ErrorCode_ERROR_CODE_TYPE_CONVERSATION.String())
		}
		subjectIDs = append(subjectIDs, t.SubjectID)
	}

	// Sort the tuples and append them to the collection.
	sort.Slice(subjectIDs, func(i, j int) bool {
		return subjectIDs[i] < subjectIDs[j]
	})

	mp := make(map[string]bool)
	var lastID string

	for _, b := range subjectIDs {
		if _, exists := mp[b]; !exists && b >= lowerBound {
			ids = append(ids, b)
			mp[b] = true

			// Capture the last ID after adding pagesize + 1 elements
			if len(ids) == int(pagination.PageSize())+1 {
				lastID = b
			}

			// Stop appending if we've reached the page size
			if pagination.PageSize() != 0 && len(ids) > int(pagination.PageSize()) {
				return ids[:pagination.PageSize()], utils.NewContinuousToken(lastID).Encode(), nil
			}
		}
	}

	return ids, database.NewNoopContinuousToken().Encode(), nil
}

// HeadSnapshot - Reads the latest version of the snapshot from the repository.
func (r *DataReader) HeadSnapshot(_ context.Context, _ string) (token.SnapToken, error) {
	return snapshot.NewToken(time.Now()), nil
}
