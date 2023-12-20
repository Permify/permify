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
	contextIds := make(map[string]struct{}) // Make an empty set to avoid duplicates

	// Querying relationships related to the entity type in the request context
	tit, err := storageContext.NewContextualTuples(request.GetContext().GetTuples()...).QueryRelationships(&base.TupleFilter{
		Entity: &base.EntityFilter{
			Type: request.GetEntityType(),
		},
	})
	if err != nil {
		return err
	}

	// Iterating through the results and adding the entity IDs to the contextIds map
	for tit.HasNext() {
		tuple := tit.GetNext()
		contextIds[tuple.GetEntity().GetId()] = struct{}{}
	}

	// Querying attributes related to the entity type in the request context
	ait, err := storageContext.NewContextualAttributes(request.GetContext().GetAttributes()...).QueryAttributes(&base.AttributeFilter{
		Entity: &base.EntityFilter{
			Type: request.GetEntityType(),
		},
	})
	if err != nil {
		return err
	}

	// Iterating through the results and adding the entity IDs to the contextIds map
	for ait.HasNext() {
		attribute := ait.GetNext()
		contextIds[attribute.GetEntity().GetId()] = struct{}{}
	}

	// Publishing the context IDs with the publisher
	for id := range contextIds {
		publisher.Publish(&base.Entity{
			Type: request.GetEntityType(),
			Id:   id,
		}, &base.PermissionCheckRequestMetadata{
			SnapToken:     request.GetMetadata().GetSnapToken(),
			SchemaVersion: request.GetMetadata().GetSchemaVersion(),
			Depth:         request.GetMetadata().GetDepth(),
		}, request.GetContext(), base.CheckResult_CHECK_RESULT_UNSPECIFIED)
	}

	// Creating a pagination object
	pagination := database.NewPagination(database.Size(100), database.Token(""))

	// This loop continues until all unique entities have been queried and published
	for {

		// Querying unique entities from the database
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

		// Publishing the unique entities
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

		// Checking if all entities have been retrieved. If the continuation token is empty, the loop breaks
		if continuationToken.String() == "" {
			break
		}

		// Updating the continuation token in the pagination object for the next loop iteration
		pagination = database.NewPagination(database.Size(100), database.Token(continuationToken.String()))
	}

	// Return successful completion of the function
	return nil
}
