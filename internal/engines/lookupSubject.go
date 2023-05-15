package engines

import (
	"context"
	"fmt"
	"sync"

	base "github.com/Permify/permify/pkg/pb/base/v1"
)

type LookupSubjectEngine struct {
	// linkedEntityEngine is responsible for retrieving linked entities
	subjectFilterEngine *SubjectFilterEngine
	// concurrencyLimit is the maximum number of concurrent permission checks allowed
	concurrencyLimit int
}

func NewLookupSubjectEngine(filter *SubjectFilterEngine, opts ...LookupSubjectOption) *LookupSubjectEngine {
	engine := &LookupSubjectEngine{
		subjectFilterEngine: filter,
		concurrencyLimit:    _defaultConcurrencyLimit,
	}

	// options
	for _, opt := range opts {
		opt(engine)
	}

	return engine
}

func (engine *LookupSubjectEngine) LookupSubject(ctx context.Context, request *base.PermissionLookupSubjectRequest) (response *base.PermissionLookupSubjectResponse, err error) {
	var mu sync.Mutex
	var subjects []*base.Subject

	callback := func(subject *base.Subject) {
		mu.Lock()
		defer mu.Unlock()
		subjects = append(subjects, subject)
	}

	stream := NewStream(ctx, callback)
	stream.ConsumeData()

	err = engine.subjectFilterEngine.SubjectFilter(ctx, &base.PermissionSubjectFilterRequest{
		TenantId: request.GetTenantId(),
		Metadata: &base.PermissionSubjectFilterRequestMetadata{
			SnapToken:     request.GetMetadata().GetSnapToken(),
			SchemaVersion: request.GetMetadata().GetSchemaVersion(),
		},
		SubjectReference: request.GetSubjectReference(),
		Permission:       request.GetPermission(),
		Entity:           request.GetEntity(),
	}, stream)
	if err != nil {
		return nil, err
	}

	if err := stream.Wait(); err != nil {
		fmt.Println("Received error:", err)
	}

	return nil, err
}

func (engine *LookupSubjectEngine) LookupSubjectStream(ctx context.Context, request *base.PermissionLookupSubjectRequest, server base.Permission_LookupSubjectStreamServer) (err error) {
	callback := func(subject *base.Subject) {
		err := server.Send(&base.PermissionLookupSubjectStreamResponse{
			SubjectId: subject.GetId(),
		})
		if err != nil {
			return
		}
	}

	stream := NewStream(ctx, callback)
	stream.ConsumeData()

	err = engine.subjectFilterEngine.SubjectFilter(ctx, &base.PermissionSubjectFilterRequest{
		TenantId: request.GetTenantId(),
		Metadata: &base.PermissionSubjectFilterRequestMetadata{
			SnapToken:     request.GetMetadata().GetSnapToken(),
			SchemaVersion: request.GetMetadata().GetSchemaVersion(),
		},
		SubjectReference: request.GetSubjectReference(),
		Permission:       request.GetPermission(),
		Entity:           request.GetEntity(),
	}, stream)
	if err != nil {
		return err
	}

	if err := stream.Wait(); err != nil {
		fmt.Println("Received error:", err)
	}

	return nil
}
