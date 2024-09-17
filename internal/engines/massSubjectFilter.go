package engines

import (
	"context"

	"github.com/Permify/permify/internal/storage"
	storageContext "github.com/Permify/permify/internal/storage/context"
	"github.com/Permify/permify/pkg/database"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// MassSubjectFilter is a struct that represents a filter for mass subjects.
type MassSubjectFilter struct {
	dataReader storage.DataReader // An interface for reading data.
}

// NewMassSubjectFilter returns a new MassSubjectFilter.
func NewMassSubjectFilter(dataReader storage.DataReader) *MassSubjectFilter {
	return &MassSubjectFilter{
		dataReader: dataReader, // Set the data reader.
	}
}

// SubjectFilter is a method on MassSubjectFilter that filters subjects.
func (filter *MassSubjectFilter) SubjectFilter(
	ctx context.Context, // The context in which the method is executed.
	request *base.PermissionLookupSubjectRequest, // The request containing the subject to look up.
	publisher *BulkSubjectPublisher, // The publisher to publish the filtered subjects to.
) (err error) {
	// Make an empty set to avoid duplicates.
	contextIds := make(map[string]struct{})

	// Query relationships based on the given filter criteria.
	tit, err := storageContext.NewContextualTuples(request.GetContext().GetTuples()...).QueryRelationships(&base.TupleFilter{
		Subject: &base.SubjectFilter{
			Type:     request.GetSubjectReference().GetType(),
			Relation: request.GetSubjectReference().GetRelation(),
		},
	}, database.NewCursorPagination(database.Cursor(request.GetContinuousToken()), database.Sort("subject_id")))
	// Return any error encountered during the query.
	if err != nil {
		return err
	}

	// Add each unique subject ID to the set.
	for tit.HasNext() {
		tuple := tit.GetNext()
		contextIds[tuple.GetSubject().GetId()] = struct{}{}
	}

	// Publish each subject in the set.
	for id := range contextIds {
		publisher.Publish(&base.Subject{
			Type:     request.GetSubjectReference().GetType(),
			Id:       id,
			Relation: request.GetSubjectReference().GetRelation(),
		}, &base.PermissionCheckRequestMetadata{
			SnapToken:     request.GetMetadata().GetSnapToken(),
			SchemaVersion: request.GetMetadata().GetSchemaVersion(),
			Depth:         request.GetMetadata().GetDepth(),
		}, request.GetContext(), base.CheckResult_CHECK_RESULT_UNSPECIFIED)
	}

	// Prepare the initial pagination object.
	pagination := database.NewPagination(database.Token(request.GetContinuousToken()))

	ids, _, err := filter.dataReader.QueryUniqueSubjectReferences(
		ctx,
		request.GetTenantId(),
		request.GetSubjectReference(),
		request.GetMetadata().GetSnapToken(),
		pagination,
	)
	// Return any error encountered during the query.
	if err != nil {
		return err
	}

	// Publish each subject retrieved.
	for _, id := range ids {
		publisher.Publish(&base.Subject{
			Type:     request.GetSubjectReference().GetType(),
			Id:       id,
			Relation: request.GetSubjectReference().GetRelation(),
		}, &base.PermissionCheckRequestMetadata{
			SnapToken:     request.GetMetadata().GetSnapToken(),
			SchemaVersion: request.GetMetadata().GetSchemaVersion(),
			Depth:         request.GetMetadata().GetDepth(),
		}, request.GetContext(), base.CheckResult_CHECK_RESULT_UNSPECIFIED)
	}

	// Return all IDs retrieved.
	return nil
}
