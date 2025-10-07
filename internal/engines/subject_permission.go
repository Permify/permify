package engines

import (
	"context"
	"errors"
	"slices"
	"sync"

	"github.com/Permify/permify/internal/invoke"
	"github.com/Permify/permify/internal/storage"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

type SubjectPermissionEngine struct {
	// checkEngine is responsible for performing permission checks
	checker invoke.Check
	// schemaReader is responsible for reading schema information
	schemaReader storage.SchemaReader
	// concurrencyLimit is the maximum number of concurrent permission checks allowed
	concurrencyLimit int
}

func NewSubjectPermission(checker invoke.Check, sr storage.SchemaReader, opts ...SubjectPermissionOption) *SubjectPermissionEngine {
	// Initialize a CheckEngine with default concurrency limit and provided parameters
	engine := &SubjectPermissionEngine{
		checker:          checker,
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
	en, _, err := engine.schemaReader.ReadEntityDefinition(ctx, request.GetTenantId(), request.GetEntity().GetType(), request.GetMetadata().GetSchemaVersion())
	if err != nil {
		// If there's an error reading the schema definition, we wrap it and return.
		return emptyResp, err
	}

	// Initialize a response object with an empty map for the Results.
	res := &base.PermissionSubjectPermissionResponse{
		Results: map[string]base.CheckResult{},
	}

	// Initialize a reference types with permission
	rtyps := []base.EntityDefinition_Reference{
		base.EntityDefinition_REFERENCE_PERMISSION,
	}

	// If the request is not for only permissions, we add relation reference to the list of relational reference types.
	if !request.GetMetadata().GetOnlyPermission() {
		rtyps = append(rtyps, base.EntityDefinition_REFERENCE_RELATION)
	}

	// If allowed reference types contains reference. append to refs list
	var refs []string
	for ref, typ := range en.GetReferences() {
		if slices.Contains(rtyps, typ) {
			refs = append(refs, ref)
		}
	}

	if len(refs) == 0 {
		return emptyResp, nil
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
			cr, err := engine.checker.Check(ctx, &base.PermissionCheckRequest{
				TenantId: request.GetTenantId(),
				Metadata: &base.PermissionCheckRequestMetadata{
					SchemaVersion: request.GetMetadata().GetSchemaVersion(),
					SnapToken:     request.GetMetadata().GetSnapToken(),
					Depth:         request.GetMetadata().GetDepth(),
				},
				Entity:     request.GetEntity(),
				Permission: permission,
				Subject:    request.GetSubject(),
				Context:    request.GetContext(),
			})
			// If there's an error, it is sent over the resultChannel along with the permission and a "denied" result.
			if err != nil {
				resultChannel <- SubjectPermissionResponse{permission: permission, result: base.CheckResult_CHECK_RESULT_DENIED, err: err}
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
