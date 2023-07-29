package engines

import (
	"context"

	"github.com/Permify/permify/internal/storage"
	storageContext "github.com/Permify/permify/internal/storage/context"
	"github.com/Permify/permify/pkg/database"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// MassEntityFilter is a struct that performs permission checks on a set of entities
type MassEntityFilter struct {
	// dataReader is responsible for reading relationship information
	dataReader storage.DataReader
}

// NewMassEntityFilter creates a new MassEntityFilter instance.
func NewMassEntityFilter(dataReader storage.DataReader) *MassEntityFilter {
	return &MassEntityFilter{
		dataReader: dataReader,
	}
}

// EntityFilter performs a permission check on a set of entities and returns a response
func (filter *MassEntityFilter) EntityFilter(
	ctx context.Context,
	request *base.PermissionLookupEntityRequest,
	publisher *BulkEntityPublisher,
) (err error) {
	// Initialize the pagination with an initial continuation token.
	pagination := database.NewPagination(database.Size(100), database.Token(""))

	var contextIds []string

	tit, err := storageContext.NewContextualTuples(request.GetContext().GetTuples()...).QueryRelationships(&base.TupleFilter{
		Entity: &base.EntityFilter{
			Type: request.GetEntityType(),
		},
	})
	if err != nil {
		return err
	}

	for tit.HasNext() {
		tuple := tit.GetNext()
		contextIds = append(contextIds, tuple.GetEntity().GetId())
	}

	ait, err := storageContext.NewContextualAttributes(request.GetContext().GetAttributes()...).QueryAttributes(&base.AttributeFilter{
		Entity: &base.EntityFilter{
			Type: request.GetEntityType(),
		},
	})
	if err != nil {
		return err
	}

	for ait.HasNext() {
		attribute := ait.GetNext()
		contextIds = append(contextIds, attribute.GetEntity().GetId())
	}

	for _, id := range contextIds {
		publisher.Publish(&base.Entity{
			Type: request.GetEntityType(),
			Id:   id,
		}, &base.PermissionCheckRequestMetadata{
			SnapToken:     request.GetMetadata().GetSnapToken(),
			SchemaVersion: request.GetMetadata().GetSchemaVersion(),
			Depth:         request.GetMetadata().GetDepth(),
		}, request.GetContext(), base.CheckResult_CHECK_RESULT_UNSPECIFIED)
	}

	for {
		// Query unique entities.
		ids, continuationToken, err := filter.dataReader.QueryUniqueEntities(
			ctx,
			request.GetTenantId(),
			request.GetEntityType(),
			request.GetMetadata().GetSnapToken(),
			pagination,
		)
		if err != nil {
			return err
		}

		for _, id := range ids {
			publisher.Publish(&base.Entity{
				Type: request.GetEntityType(),
				Id:   id,
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
