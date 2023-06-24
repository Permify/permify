package engines

import (
	"context"
	"errors"
	`golang.org/x/exp/slices`
	"sync"

	"github.com/Permify/permify/internal/storage"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

type SubjectPermissionEngine struct {
	// checkEngine is responsible for performing permission checks
	checkEngine *CheckEngine
	// schemaReader is responsible for reading schema information
	schemaReader storage.SchemaReader
	// concurrencyLimit is the maximum number of concurrent permission checks allowed
	concurrencyLimit int
}

func NewSubjectPermission(check *CheckEngine, sr storage.SchemaReader, opts ...SubjectPermissionOption) *SubjectPermissionEngine {
	// Initialize a CheckEngine with default concurrency limit and provided parameters
	engine := &SubjectPermissionEngine{
		checkEngine:      check,
		schemaReader:     sr,
		concurrencyLimit: _defaultConcurrencyLimit,
	}

	// Apply provided options to configure the CheckEngine
	for _, opt := range opts {
		opt(engine)
	}

	return engine
}

// SubjectPermission is a method on the SubjectPermissionEngine struct.
// It checks permissions for a given subject based on the supplied request and context.
func (engine *SubjectPermissionEngine) SubjectPermission(ctx context.Context, request *base.PermissionSubjectPermissionRequest) (*base.PermissionSubjectPermissionResponse, error) {
	// emptyResp is a default, empty response that we will return in case of an error or when the context is cancelled.
	emptyResp := &base.PermissionSubjectPermissionResponse{
		Results: map[string]base.CheckResult{},
	}

	// The schema definition for the entity is read from the engine's schemaReader.
	// The tenant ID, entity type, and schema version are all taken from the request.
	en, _, err := engine.schemaReader.ReadSchemaDefinition(ctx, request.GetTenantId(), request.GetEntity().GetType(), request.GetMetadata().GetSchemaVersion())
	if err != nil {
		// If there's an error reading the schema definition, we wrap it and return.
		return emptyResp, err
	}

	// Initialize a response object with an empty map for the Results.
	res := &base.PermissionSubjectPermissionResponse{
		Results: map[string]base.CheckResult{},
	}

	typs := []base.EntityDefinition_RelationalReference{
		base.EntityDefinition_RELATIONAL_REFERENCE_PERMISSION,
	}

	if !request.GetMetadata().GetOnlyPermission() {
		typs = append(typs, base.EntityDefinition_RELATIONAL_REFERENCE_RELATION)
	}

	var refs []string

	for ref, typ := range en.GetReferences() {
		if slices.Contains(typs, typ) {
			refs = append(refs, ref)
		}
	}

	// Create a buffered channel for SubjectPermissionResponses.
	// The buffer size is equal to the number of references in the entity.
	resultChannel := make(chan SubjectPermissionResponse, len(refs))

	// The WaitGroup and Mutex are used for synchronization.
	var wg sync.WaitGroup
	var mutex sync.Mutex

	// Loop over each reference in the entity.
	for _, p := range refs {
		// For each reference, we add a count to the WaitGroup and start a new goroutine.
		wg.Add(1)
		go func(permission string) {
			// When the goroutine completes, it calls Done on the WaitGroup.
			defer wg.Done()

			// The checkEngine's Check method is called with a new PermissionCheckRequest.
			// The request is created using the data from the original request, and the permission from the current iteration.
			cr, err := engine.checkEngine.Check(ctx, &base.PermissionCheckRequest{
				TenantId: request.GetTenantId(),
				Metadata: &base.PermissionCheckRequestMetadata{
					SchemaVersion: request.GetMetadata().GetSchemaVersion(),
					SnapToken:     request.GetMetadata().GetSnapToken(),
					Depth:         request.GetMetadata().GetDepth(),
				},
				Entity:           request.GetEntity(),
				Permission:       permission,
				Subject:          request.GetSubject(),
				ContextualTuples: request.GetContextualTuples(),
			})
			// If there's an error, it is sent over the resultChannel along with the permission and a "denied" result.
			if err != nil {
				resultChannel <- SubjectPermissionResponse{permission: permission, result: base.CheckResult_RESULT_DENIED, err: err}
				return
			}

			// If there's no error, the result of the check (along with the permission and a nil error) is sent over the resultChannel.
			resultChannel <- SubjectPermissionResponse{permission: permission, result: cr.Can, err: nil}
		}(p)
	}

	// Once the function returns, we wait for all goroutines to finish, then close the resultChannel.
	defer func() {
		wg.Wait()
		close(resultChannel)
	}()

	// We read the responses from the resultChannel.
	// We expect as many responses as there are references in the entity.
	for i := 0; i < len(refs); i++ {
		select {
		// If we receive a response from the resultChannel, we check for errors.
		case response := <-resultChannel:
			// If there's an error, we return an empty
			// response and the error.
			if response.err != nil {
				return emptyResp, response.err
			}
			// If there's no error, we add the result to our response's Results map.
			// We use a mutex to safely update the map since multiple goroutines may be writing to it concurrently.
			mutex.Lock()
			res.Results[response.permission] = response.result
			mutex.Unlock()

		// If the context is done (i.e., canceled or deadline exceeded), we return an empty response and an error.
		case <-ctx.Done():
			return emptyResp, errors.New(base.ErrorCode_ERROR_CODE_CANCELLED.String())
		}
	}

	// Once all results are processed, we return the response and a nil error.
	return res, nil
}
