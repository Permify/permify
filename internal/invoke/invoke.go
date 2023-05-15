package invoke

import (
	"context"

	"go.opentelemetry.io/otel"

	"github.com/Permify/permify/internal/storage"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/token"
)

var tracer = otel.Tracer("invoke")

// Invoker is an interface that groups multiple permission-related interfaces.
// It is used to define a common contract for invoking various permission operations.
type Invoker interface {
	Check
	Expand
	LookupEntity
	LookupSubject
}

// Check is an interface that defines a method for checking permissions.
// It requires an implementation of InvokeCheck that takes a context and a PermissionCheckRequest,
// and returns a PermissionCheckResponse and an error if any.
type Check interface {
	Check(ctx context.Context, request *base.PermissionCheckRequest) (response *base.PermissionCheckResponse, err error)
}

// Expand is an interface that defines a method for expanding permissions.
// It requires an implementation of InvokeExpand that takes a context and a PermissionExpandRequest,
// and returns a PermissionExpandResponse and an error if any.
type Expand interface {
	Expand(ctx context.Context, request *base.PermissionExpandRequest) (response *base.PermissionExpandResponse, err error)
}

// LookupEntity is an interface that defines a method for looking up entities with permissions.
// It requires an implementation of InvokeLookupEntity that takes a context and a PermissionLookupEntityRequest,
// and returns a PermissionLookupEntityResponse and an error if any.
type LookupEntity interface {
	LookupEntity(ctx context.Context, request *base.PermissionLookupEntityRequest) (response *base.PermissionLookupEntityResponse, err error)
	LookupEntityStream(ctx context.Context, request *base.PermissionLookupEntityRequest, server base.Permission_LookupEntityStreamServer) (err error)
}

// LookupSubject -
type LookupSubject interface {
	LookupSubject(ctx context.Context, request *base.PermissionLookupSubjectRequest) (response *base.PermissionLookupSubjectResponse, err error)
	LookupSubjectStream(ctx context.Context, request *base.PermissionLookupSubjectRequest, server base.Permission_LookupSubjectStreamServer) (err error)
}

// DirectInvoker is a struct that implements the Invoker interface.
// It holds references to various engines needed for permission-related operations.
type DirectInvoker struct {
	// schemaReader is responsible for reading schema information
	schemaReader storage.SchemaReader
	// relationshipReader is responsible for reading relationship information
	relationshipReader storage.RelationshipReader
	// Check engine for permission checks
	cc Check
	// Expand engine for expanding permissions
	ec Expand
	// LookupEntity engine for looking up entities with permissions
	le LookupEntity
	// LookupSubject
	ls LookupSubject
}

// NewDirectInvoker is a constructor for DirectInvoker.
// It takes pointers to CheckEngine, ExpandEngine, LookupSchemaEngine, and LookupEntityEngine as arguments
// and returns an Invoker instance.
func NewDirectInvoker(
	schemaReader storage.SchemaReader,
	relationshipReader storage.RelationshipReader,
	cc Check,
	ec Expand,
	le LookupEntity,
	ls LookupSubject,
) *DirectInvoker {
	return &DirectInvoker{
		schemaReader:       schemaReader,
		relationshipReader: relationshipReader,
		cc:                 cc,
		ec:                 ec,
		le:                 le,
		ls:                 ls,
	}
}

// Check is a method that implements the Check interface.
// It calls the Run method of the CheckEngine with the provided context and PermissionCheckRequest,
// and returns a PermissionCheckResponse and an error if any.
func (invoker *DirectInvoker) Check(ctx context.Context, request *base.PermissionCheckRequest) (response *base.PermissionCheckResponse, err error) {
	// Start a new tracing span to measure the performance of the Check function.
	ctx, span := tracer.Start(ctx, "permissions.check")
	defer span.End()

	// Validate the depth of the request.
	err = checkDepth(request)
	if err != nil {
		return &base.PermissionCheckResponse{
			Can: base.PermissionCheckResponse_RESULT_DENIED,
			Metadata: &base.PermissionCheckResponseMetadata{
				CheckCount: 0,
			},
		}, err
	}

	// Set the SnapToken if it's not provided in the request.
	if request.GetMetadata().GetSnapToken() == "" {
		var st token.SnapToken
		st, err = invoker.relationshipReader.HeadSnapshot(ctx, request.GetTenantId())
		if err != nil {
			return &base.PermissionCheckResponse{
				Can: base.PermissionCheckResponse_RESULT_DENIED,
				Metadata: &base.PermissionCheckResponseMetadata{
					CheckCount: 0,
				},
			}, err
		}
		request.Metadata.SnapToken = st.Encode().String()
	}

	// Set the SchemaVersion if it's not provided in the request.
	if request.GetMetadata().GetSchemaVersion() == "" {
		request.Metadata.SchemaVersion, err = invoker.schemaReader.HeadVersion(ctx, request.GetTenantId())
		if err != nil {
			return &base.PermissionCheckResponse{
				Can: base.PermissionCheckResponse_RESULT_DENIED,
				Metadata: &base.PermissionCheckResponseMetadata{
					CheckCount: 0,
				},
			}, err
		}
	}

	// Decrease the depth of the request metadata.
	request.Metadata = decreaseDepth(request.GetMetadata())

	// Perform the actual permission check using the provided request.
	response, err = invoker.cc.Check(ctx, request)
	if err != nil {
		return &base.PermissionCheckResponse{
			Can: base.PermissionCheckResponse_RESULT_DENIED,
			Metadata: &base.PermissionCheckResponseMetadata{
				CheckCount: 0,
			},
		}, err
	}

	// Increase the check count in the response metadata.
	response.Metadata = increaseCheckCount(response.Metadata)
	return
}

// Expand is a method that implements the Expand interface.
// It calls the Run method of the ExpandEngine with the provided context and PermissionExpandRequest,
// and returns a PermissionExpandResponse and an error if any.
func (invoker *DirectInvoker) Expand(ctx context.Context, request *base.PermissionExpandRequest) (response *base.PermissionExpandResponse, err error) {
	ctx, span := tracer.Start(ctx, "permissions.expand")
	defer span.End()

	if request.GetMetadata().GetSnapToken() == "" {
		var st token.SnapToken
		st, err = invoker.relationshipReader.HeadSnapshot(ctx, request.GetTenantId())
		if err != nil {
			return response, err
		}
		request.Metadata.SnapToken = st.Encode().String()
	}

	if request.GetMetadata().GetSchemaVersion() == "" {
		request.Metadata.SchemaVersion, err = invoker.schemaReader.HeadVersion(ctx, request.GetTenantId())
		if err != nil {
			return response, err
		}
	}

	return invoker.ec.Expand(ctx, request)
}

// LookupEntity is a method that implements the LookupEntity interface.
// It calls the Run method of the LookupEntityEngine with the provided context and PermissionLookupEntityRequest,
// and returns a PermissionLookupEntityResponse and an error if any.
func (invoker *DirectInvoker) LookupEntity(ctx context.Context, request *base.PermissionLookupEntityRequest) (response *base.PermissionLookupEntityResponse, err error) {
	ctx, span := tracer.Start(ctx, "permissions.lookup-entity")
	defer span.End()

	// Set SnapToken if not provided
	if request.GetMetadata().GetSnapToken() == "" { // Check if the request has a SnapToken.
		var st token.SnapToken
		st, err = invoker.relationshipReader.HeadSnapshot(ctx, request.GetTenantId()) // Retrieve the head snapshot from the relationship reader.
		if err != nil {
			return nil, err
		}
		request.Metadata.SnapToken = st.Encode().String() // Set the SnapToken in the request metadata.
	}

	// Set SchemaVersion if not provided
	if request.GetMetadata().GetSchemaVersion() == "" { // Check if the request has a SchemaVersion.
		request.Metadata.SchemaVersion, err = invoker.schemaReader.HeadVersion(ctx, request.GetTenantId()) // Retrieve the head schema version from the schema reader.
		if err != nil {
			return nil, err
		}
	}

	return invoker.le.LookupEntity(ctx, request)
}

// LookupEntityStream is a method that implements the LookupEntityStream interface.
// It calls the Stream method of the LookupEntityEngine with the provided context, PermissionLookupEntityRequest, and Permission_LookupEntityStreamServer,
// and returns an error if any.
func (invoker *DirectInvoker) LookupEntityStream(ctx context.Context, request *base.PermissionLookupEntityRequest, server base.Permission_LookupEntityStreamServer) (err error) {
	ctx, span := tracer.Start(ctx, "permissions.lookup-entity-stream")
	defer span.End()

	// Set SnapToken if not provided
	if request.GetMetadata().GetSnapToken() == "" { // Check if the request has a SnapToken.
		var st token.SnapToken
		st, err = invoker.relationshipReader.HeadSnapshot(ctx, request.GetTenantId()) // Retrieve the head snapshot from the relationship reader.
		if err != nil {
			return err
		}
		request.Metadata.SnapToken = st.Encode().String() // Set the SnapToken in the request metadata.
	}

	// Set SchemaVersion if not provided
	if request.GetMetadata().GetSchemaVersion() == "" { // Check if the request has a SchemaVersion.
		request.Metadata.SchemaVersion, err = invoker.schemaReader.HeadVersion(ctx, request.GetTenantId()) // Retrieve the head schema version from the schema reader.
		if err != nil {
			return err
		}
	}

	return invoker.le.LookupEntityStream(ctx, request, server)
}

func (invoker *DirectInvoker) LookupSubject(ctx context.Context, request *base.PermissionLookupSubjectRequest) (response *base.PermissionLookupSubjectResponse, err error) {
	ctx, span := tracer.Start(ctx, "permissions.lookup-subject")
	defer span.End()

	// Set SnapToken if not provided
	if request.GetMetadata().GetSnapToken() == "" { // Check if the request has a SnapToken.
		var st token.SnapToken
		st, err = invoker.relationshipReader.HeadSnapshot(ctx, request.GetTenantId()) // Retrieve the head snapshot from the relationship reader.
		if err != nil {
			return nil, err
		}
		request.Metadata.SnapToken = st.Encode().String() // Set the SnapToken in the request metadata.
	}

	// Set SchemaVersion if not provided
	if request.GetMetadata().GetSchemaVersion() == "" { // Check if the request has a SchemaVersion.
		request.Metadata.SchemaVersion, err = invoker.schemaReader.HeadVersion(ctx, request.GetTenantId()) // Retrieve the head schema version from the schema reader.
		if err != nil {
			return nil, err
		}
	}

	return invoker.ls.LookupSubject(ctx, request)
}

func (invoker *DirectInvoker) LookupSubjectStream(ctx context.Context, request *base.PermissionLookupSubjectRequest, server base.Permission_LookupSubjectStreamServer) (err error) {
	ctx, span := tracer.Start(ctx, "permissions.lookup-subject-stream")
	defer span.End()

	// Set SnapToken if not provided
	if request.GetMetadata().GetSnapToken() == "" { // Check if the request has a SnapToken.
		var st token.SnapToken
		st, err = invoker.relationshipReader.HeadSnapshot(ctx, request.GetTenantId()) // Retrieve the head snapshot from the relationship reader.
		if err != nil {
			return err
		}
		request.Metadata.SnapToken = st.Encode().String() // Set the SnapToken in the request metadata.
	}

	// Set SchemaVersion if not provided
	if request.GetMetadata().GetSchemaVersion() == "" { // Check if the request has a SchemaVersion.
		request.Metadata.SchemaVersion, err = invoker.schemaReader.HeadVersion(ctx, request.GetTenantId()) // Retrieve the head schema version from the schema reader.
		if err != nil {
			return err
		}
	}

	return invoker.ls.LookupSubjectStream(ctx, request, server)
}
