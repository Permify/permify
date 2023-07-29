package engines

import (
	"context"

	"github.com/Permify/permify/internal/storage"
	storageContext "github.com/Permify/permify/internal/storage/context"
	"github.com/Permify/permify/pkg/database"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

type MassSubjectFilter struct {
	dataReader storage.DataReader
}

func NewMassSubjectFilter(dataReader storage.DataReader) *MassSubjectFilter {
	return &MassSubjectFilter{
		dataReader: dataReader,
	}
}

func (filter *MassSubjectFilter) SubjectFilter(
	ctx context.Context,
	request *base.PermissionLookupSubjectRequest,
	publisher *BulkSubjectPublisher,
) (err error) {
	// Initialize the pagination with an initial continuation token.
	pagination := database.NewPagination(database.Size(100), database.Token(""))

	var contextIds []string

	tit, err := storageContext.NewContextualTuples(request.GetContext().GetTuples()...).QueryRelationships(&base.TupleFilter{
		Subject: &base.SubjectFilter{
			Type:     request.GetSubjectReference().GetType(),
			Relation: request.GetSubjectReference().GetRelation(),
		},
	})
	if err != nil {
		return err
	}

	for tit.HasNext() {
		tuple := tit.GetNext()
		contextIds = append(contextIds, tuple.GetSubject().GetId())
	}

	for _, id := range contextIds {
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

	for {
		// Query unique entities.
		ids, continuationToken, err := filter.dataReader.QueryUniqueSubjectReferences(
			ctx,
			request.GetTenantId(),
			request.GetSubjectReference(),
			request.GetMetadata().GetSnapToken(),
			pagination,
		)
		if err != nil {
			return err
		}

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

		// If the continuation token is empty, we've retrieved all entities.
		if continuationToken.String() == "" {
			break
		}

		// Update the continuation token in the pagination for the next loop iteration.
		pagination = database.NewPagination(database.Size(100), database.Token(continuationToken.String()))
	}

	// Return all IDs retrieved.
	return nil
}
