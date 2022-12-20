package commands

import (
	"context"

	"golang.org/x/sync/errgroup"

	"github.com/Permify/permify/internal/repositories"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/token"
)

// LookupEntityCommand -
type LookupEntityCommand struct {
	// commands
	checkCommand ICheckCommand
	// repositories
	schemaReader       repositories.SchemaReader
	relationshipReader repositories.RelationshipReader
}

// NewLookupEntityCommand -
func NewLookupEntityCommand(ck ICheckCommand, sr repositories.SchemaReader, rr repositories.RelationshipReader) *LookupEntityCommand {
	return &LookupEntityCommand{
		checkCommand:       ck,
		schemaReader:       sr,
		relationshipReader: rr,
	}
}

// Execute -
func (command *LookupEntityCommand) Execute(ctx context.Context, request *base.PermissionLookupEntityRequest) (response *base.PermissionLookupEntityResponse, err error) {
	ctx, span := tracer.Start(ctx, "permissions.LookUpSchema.execute")
	defer span.End()

	if request.GetMetadata().GetSnapToken() == "" {
		var st token.SnapToken
		st, err = command.relationshipReader.HeadSnapshot(ctx)
		if err != nil {
			return response, err
		}
		request.Metadata.SnapToken = st.Encode().String()
	}

	if request.GetMetadata().GetSchemaVersion() == "" {
		request.Metadata.SchemaVersion, err = command.schemaReader.HeadVersion(ctx)
		if err != nil {
			return response, err
		}
	}

	resultsChan := make(chan string, 100)
	errChan := make(chan error)

	go command.parallelChecker(ctx, request, resultsChan, errChan)

	entityIDs := make([]string, 0, len(resultsChan))
	for entityID := range resultsChan {
		entityIDs = append(entityIDs, entityID)
	}

	return &base.PermissionLookupEntityResponse{
		EntityIds: entityIDs,
	}, nil
}

// Stream -
func (command *LookupEntityCommand) Stream(ctx context.Context, request *base.PermissionLookupEntityRequest, server base.Permission_LookupEntityStreamServer) (err error) {
	ctx, span := tracer.Start(ctx, "permissions.LookupEntity.stream")
	defer span.End()

	if request.GetMetadata().GetSnapToken() == "" {
		var st token.SnapToken
		st, err = command.relationshipReader.HeadSnapshot(ctx)
		if err != nil {
			return err
		}
		request.Metadata.SnapToken = st.Encode().String()
	}

	if request.GetMetadata().GetSchemaVersion() == "" {
		request.Metadata.SchemaVersion, err = command.schemaReader.HeadVersion(ctx)
		if err != nil {
			return err
		}
	}

	resultChan := make(chan string, 100)
	errChan := make(chan error)

	go command.parallelChecker(ctx, request, resultChan, errChan)

	for {
		select {
		case id, ok := <-resultChan:
			if !ok {
				return nil
			}
			if err := server.Send(&base.PermissionLookupEntityStreamResponse{
				EntityId: id,
			}); err != nil {
				return err
			}
		case err, ok := <-errChan:
			if ok {
				return err
			}
		}
	}
}

// parallelChecker -
func (command *LookupEntityCommand) parallelChecker(ctx context.Context, request *base.PermissionLookupEntityRequest, resultChan chan<- string, errChan chan<- error) {
	ids, err := command.relationshipReader.GetUniqueEntityIDsByEntityType(ctx, request.GetEntityType(), request.GetMetadata().GetSnapToken())
	if err != nil {
		errChan <- err
	}

	g := new(errgroup.Group)
	g.SetLimit(100)

	for _, id := range ids {
		id := id
		g.Go(func() error {
			return command.internalCheck(ctx, &base.Entity{
				Type: request.GetEntityType(),
				Id:   id,
			}, request, resultChan)
		})
	}

	err = g.Wait()
	if err != nil {
		errChan <- err
	}

	close(resultChan)
}

// internalCheck -
func (command *LookupEntityCommand) internalCheck(ctx context.Context, en *base.Entity, request *base.PermissionLookupEntityRequest, resultChan chan<- string) error {
	result, err := command.checkCommand.Execute(ctx, &base.PermissionCheckRequest{
		Metadata: &base.PermissionCheckRequestMetadata{
			SnapToken:     request.GetMetadata().GetSnapToken(),
			SchemaVersion: request.GetMetadata().GetSchemaVersion(),
			Depth:         request.GetMetadata().GetDepth(),
			Exclusion:     false,
		},
		Entity:     en,
		Permission: request.GetPermission(),
		Subject:    request.GetSubject(),
	})
	if err != nil {
		return err
	}
	if result.Can == base.PermissionCheckResponse_RESULT_ALLOWED {
		resultChan <- en.GetId()
	}
	return nil
}
