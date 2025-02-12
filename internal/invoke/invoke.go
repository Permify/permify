package invoke

import (
	"context"
	"sync/atomic"

	"go.opentelemetry.io/otel/attribute"
	otelCodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"github.com/Permify/permify/internal"
	"github.com/Permify/permify/internal/storage"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/telemetry"
	"github.com/Permify/permify/pkg/token"
	"github.com/Permify/permify/pkg/tuple"
)

// Invoker is an interface that groups multiple permission-related interfaces.
// It is used to define a common contract for invoking various permission operations.
type Invoker interface {
	Check
	Expand
	Lookup
	SubjectPermission
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

type Lookup interface {
	LookupEntity(ctx context.Context, request *base.PermissionLookupEntityRequest) (response *base.PermissionLookupEntityResponse, err error)
	LookupEntityStream(ctx context.Context, request *base.PermissionLookupEntityRequest, server base.Permission_LookupEntityStreamServer) (err error)
	LookupSubject(ctx context.Context, request *base.PermissionLookupSubjectRequest) (response *base.PermissionLookupSubjectResponse, err error)
}

// SubjectPermission -
type SubjectPermission interface {
	SubjectPermission(ctx context.Context, request *base.PermissionSubjectPermissionRequest) (response *base.PermissionSubjectPermissionResponse, err error)
}

// DirectInvoker is a struct that implements the Invoker interface.
// It holds references to various engines needed for permission-related operations.
type DirectInvoker struct {
	// schemaReader is responsible for reading schema information
	schemaReader storage.SchemaReader
	// relationshipReader is responsible for reading relationship information
	dataReader storage.DataReader
	// Check engine for permission checks
	cc Check
	// Expand engine for expanding permissions
	ec Expand
	// LookupEntity engine for looking up entities with permissions
	lo Lookup
	// LookupSubject
	sp SubjectPermission

	checkHistogram             metric.Int64Histogram
	lookupEntityHistogram      metric.Int64Histogram
	lookupSubjectHistogram     metric.Int64Histogram
	subjectPermissionHistogram metric.Int64Histogram
}

// NewDirectInvoker is a constructor for DirectInvoker.
// It takes pointers to CheckEngine, ExpandEngine, LookupSchemaEngine, and LookupEntityEngine as arguments
// and returns an Invoker instance.
func NewDirectInvoker(
	schemaReader storage.SchemaReader,
	dataReader storage.DataReader,
	cc Check,
	ec Expand,
	lo Lookup,
	sp SubjectPermission,
) *DirectInvoker {
	return &DirectInvoker{
		schemaReader:   schemaReader,
		dataReader:     dataReader,
		cc:             cc,
		ec:             ec,
		lo:             lo,
		sp:             sp,
		checkHistogram: telemetry.NewHistogram(internal.Meter, "check", "amount", "Number of checks"),

		lookupEntityHistogram:      telemetry.NewHistogram(internal.Meter, "lookup_entity", "amount", "Number of lookup entity"),
		lookupSubjectHistogram:     telemetry.NewHistogram(internal.Meter, "lookup_subject", "amount", "Number of lookup subject"),
		subjectPermissionHistogram: telemetry.NewHistogram(internal.Meter, "subject_permission", "amount", "Number of subject permission"),
	}
}

// Check is a method that implements the Check interface.
// It calls the Run method of the CheckEngine with the provided context and PermissionCheckRequest,
// and returns a PermissionCheckResponse and an error if any.
func (invoker *DirectInvoker) Check(ctx context.Context, request *base.PermissionCheckRequest) (response *base.PermissionCheckResponse, err error) {
	ctx, span := internal.Tracer.Start(ctx, "check", trace.WithAttributes(
		attribute.KeyValue{Key: "tenant_id", Value: attribute.StringValue(request.GetTenantId())},
		attribute.KeyValue{Key: "entity", Value: attribute.StringValue(tuple.EntityToString(request.GetEntity()))},
		attribute.KeyValue{Key: "permission", Value: attribute.StringValue(request.GetPermission())},
		attribute.KeyValue{Key: "subject", Value: attribute.StringValue(tuple.SubjectToString(request.GetSubject()))},
	))
	defer span.End()
	invoker.checkHistogram.Record(ctx, 1,
		metric.WithAttributeSet(
			attribute.NewSet(
				attribute.KeyValue{Key: "subject_id", Value: attribute.StringValue(request.GetSubject().GetId())},
				attribute.KeyValue{Key: "subject_type", Value: attribute.StringValue(request.GetSubject().GetType())},
			)),
	)

	// Validate the depth of the request.
	err = checkDepth(request)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		span.SetAttributes(attribute.KeyValue{Key: "can", Value: attribute.StringValue(base.CheckResult_CHECK_RESULT_DENIED.String())})
		return &base.PermissionCheckResponse{
			Can: base.CheckResult_CHECK_RESULT_DENIED,
			Metadata: &base.PermissionCheckResponseMetadata{
				CheckCount: 0,
			},
		}, err
	}

	// Set the SnapToken if it's not provided in the request.
	if request.GetMetadata().GetSnapToken() == "" {
		var st token.SnapToken
		st, err = invoker.dataReader.HeadSnapshot(ctx, request.GetTenantId())
		if err != nil {
			span.RecordError(err)
			span.SetStatus(otelCodes.Error, err.Error())
			span.SetAttributes(attribute.KeyValue{Key: "can", Value: attribute.StringValue(base.CheckResult_CHECK_RESULT_DENIED.String())})
			return &base.PermissionCheckResponse{
				Can: base.CheckResult_CHECK_RESULT_DENIED,
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
			span.RecordError(err)
			span.SetStatus(otelCodes.Error, err.Error())
			span.SetAttributes(attribute.KeyValue{Key: "can", Value: attribute.StringValue(base.CheckResult_CHECK_RESULT_DENIED.String())})
			return &base.PermissionCheckResponse{
				Can: base.CheckResult_CHECK_RESULT_DENIED,
				Metadata: &base.PermissionCheckResponseMetadata{
					CheckCount: 0,
				},
			}, err
		}
	}

	atomic.AddInt32(&request.GetMetadata().Depth, -1)

	// Perform the actual permission check using the provided request.
	response, err = invoker.cc.Check(ctx, request)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		span.SetAttributes(attribute.KeyValue{Key: "can", Value: attribute.StringValue(base.CheckResult_CHECK_RESULT_DENIED.String())})
		return &base.PermissionCheckResponse{
			Can: base.CheckResult_CHECK_RESULT_DENIED,
			Metadata: &base.PermissionCheckResponseMetadata{
				CheckCount: 0,
			},
		}, err
	}

	// increaseCheckCount increments the CheckCount value in the response metadata by 1.
	atomic.AddInt32(&response.GetMetadata().CheckCount, +1)

	span.SetAttributes(attribute.KeyValue{Key: "can", Value: attribute.StringValue(response.GetCan().String())})
	return
}

// Expand is a method that implements the Expand interface.
// It calls the Run method of the ExpandEngine with the provided context and PermissionExpandRequest,
// and returns a PermissionExpandResponse and an error if any.
func (invoker *DirectInvoker) Expand(ctx context.Context, request *base.PermissionExpandRequest) (response *base.PermissionExpandResponse, err error) {
	ctx, span := internal.Tracer.Start(ctx, "expand", trace.WithAttributes(
		attribute.KeyValue{Key: "tenant_id", Value: attribute.StringValue(request.GetTenantId())},
		attribute.KeyValue{Key: "entity", Value: attribute.StringValue(tuple.EntityToString(request.GetEntity()))},
		attribute.KeyValue{Key: "permission", Value: attribute.StringValue(request.GetPermission())},
	))
	defer span.End()

	if request.GetMetadata().GetSnapToken() == "" {
		var st token.SnapToken
		st, err = invoker.dataReader.HeadSnapshot(ctx, request.GetTenantId())
		if err != nil {
			span.RecordError(err)
			span.SetStatus(otelCodes.Error, err.Error())
			return response, err
		}
		request.Metadata.SnapToken = st.Encode().String()
	}

	if request.GetMetadata().GetSchemaVersion() == "" {
		request.Metadata.SchemaVersion, err = invoker.schemaReader.HeadVersion(ctx, request.GetTenantId())
		if err != nil {
			span.RecordError(err)
			span.SetStatus(otelCodes.Error, err.Error())
			return response, err
		}
	}

	return invoker.ec.Expand(ctx, request)
}

// LookupEntity is a method that implements the LookupEntity interface.
// It calls the Run method of the LookupEntityEngine with the provided context and PermissionLookupEntityRequest,
// and returns a PermissionLookupEntityResponse and an error if any.
func (invoker *DirectInvoker) LookupEntity(ctx context.Context, request *base.PermissionLookupEntityRequest) (response *base.PermissionLookupEntityResponse, err error) {
	ctx, span := internal.Tracer.Start(ctx, "lookup-entity", trace.WithAttributes(
		attribute.KeyValue{Key: "tenant_id", Value: attribute.StringValue(request.GetTenantId())},
		attribute.KeyValue{Key: "entity_type", Value: attribute.StringValue(request.GetEntityType())},
		attribute.KeyValue{Key: "permission", Value: attribute.StringValue(request.GetPermission())},
		attribute.KeyValue{Key: "subject", Value: attribute.StringValue(tuple.SubjectToString(request.GetSubject()))},
	))
	defer span.End()

	// Set SnapToken if not provided
	if request.GetMetadata().GetSnapToken() == "" { // Check if the request has a SnapToken.
		var st token.SnapToken
		st, err = invoker.dataReader.HeadSnapshot(ctx, request.GetTenantId()) // Retrieve the head snapshot from the relationship reader.
		if err != nil {
			span.RecordError(err)
			span.SetStatus(otelCodes.Error, err.Error())
			return response, err
		}
		request.Metadata.SnapToken = st.Encode().String() // Set the SnapToken in the request metadata.
	}

	// Set SchemaVersion if not provided
	if request.GetMetadata().GetSchemaVersion() == "" { // Check if the request has a SchemaVersion.
		request.Metadata.SchemaVersion, err = invoker.schemaReader.HeadVersion(ctx, request.GetTenantId()) // Retrieve the head schema version from the schema reader.
		if err != nil {
			span.RecordError(err)
			span.SetStatus(otelCodes.Error, err.Error())
			return response, err
		}
	}

	resp, err := invoker.lo.LookupEntity(ctx, request)

	invoker.lookupEntityHistogram.Record(ctx, 1)

	return resp, err
}

// LookupEntityStream is a method that implements the LookupEntityStream interface.
// It calls the Stream method of the LookupEntityEngine with the provided context, PermissionLookupEntityRequest, and Permission_LookupEntityStreamServer,
// and returns an error if any.
func (invoker *DirectInvoker) LookupEntityStream(ctx context.Context, request *base.PermissionLookupEntityRequest, server base.Permission_LookupEntityStreamServer) (err error) {
	ctx, span := internal.Tracer.Start(ctx, "lookup-entity-stream", trace.WithAttributes(
		attribute.KeyValue{Key: "tenant_id", Value: attribute.StringValue(request.GetTenantId())},
		attribute.KeyValue{Key: "entity_type", Value: attribute.StringValue(request.GetEntityType())},
		attribute.KeyValue{Key: "permission", Value: attribute.StringValue(request.GetPermission())},
		attribute.KeyValue{Key: "subject", Value: attribute.StringValue(tuple.SubjectToString(request.GetSubject()))},
	))
	defer span.End()

	// Set SnapToken if not provided
	if request.GetMetadata().GetSnapToken() == "" { // Check if the request has a SnapToken.
		var st token.SnapToken
		st, err = invoker.dataReader.HeadSnapshot(ctx, request.GetTenantId()) // Retrieve the head snapshot from the relationship reader.
		if err != nil {
			span.RecordError(err)
			span.SetStatus(otelCodes.Error, err.Error())
			return err
		}
		request.Metadata.SnapToken = st.Encode().String() // Set the SnapToken in the request metadata.
	}

	// Set SchemaVersion if not provided
	if request.GetMetadata().GetSchemaVersion() == "" { // Check if the request has a SchemaVersion.
		request.Metadata.SchemaVersion, err = invoker.schemaReader.HeadVersion(ctx, request.GetTenantId()) // Retrieve the head schema version from the schema reader.
		if err != nil {
			span.RecordError(err)
			span.SetStatus(otelCodes.Error, err.Error())
			return err
		}
	}

	resp := invoker.lo.LookupEntityStream(ctx, request, server)

	invoker.lookupEntityHistogram.Record(ctx, 1)

	return resp
}

// LookupSubject is a method of the DirectInvoker structure. It handles the task of looking up subjects
// and returning the results in a response.
func (invoker *DirectInvoker) LookupSubject(ctx context.Context, request *base.PermissionLookupSubjectRequest) (response *base.PermissionLookupSubjectResponse, err error) {
	ctx, span := internal.Tracer.Start(ctx, "lookup-subject", trace.WithAttributes(
		attribute.KeyValue{Key: "tenant_id", Value: attribute.StringValue(request.GetTenantId())},
		attribute.KeyValue{Key: "entity", Value: attribute.StringValue(tuple.EntityToString(request.GetEntity()))},
		attribute.KeyValue{Key: "permission", Value: attribute.StringValue(request.GetPermission())},
		attribute.KeyValue{Key: "subject_reference", Value: attribute.StringValue(tuple.ReferenceToString(request.GetSubjectReference()))},
	))
	defer span.End()

	// Check if the request has a SnapToken. If not, a SnapToken is set.
	if request.GetMetadata().GetSnapToken() == "" {
		// Create an instance of SnapToken
		var st token.SnapToken
		// Retrieve the head snapshot from the relationship reader
		st, err = invoker.dataReader.HeadSnapshot(ctx, request.GetTenantId())
		// If there's an error retrieving the snapshot, return the response and the error
		if err != nil {
			span.RecordError(err)
			span.SetStatus(otelCodes.Error, err.Error())
			return response, err
		}
		// Set the SnapToken in the request metadata
		request.Metadata.SnapToken = st.Encode().String()
	}

	// Similar to SnapToken, check if the request has a SchemaVersion. If not, a SchemaVersion is set.
	if request.GetMetadata().GetSchemaVersion() == "" {
		// Retrieve the head schema version from the schema reader
		request.Metadata.SchemaVersion, err = invoker.schemaReader.HeadVersion(ctx, request.GetTenantId())
		// If there's an error retrieving the schema version, return the response and the error
		if err != nil {
			span.RecordError(err)
			span.SetStatus(otelCodes.Error, err.Error())
			return response, err
		}
	}

	resp, err := invoker.lo.LookupSubject(ctx, request)

	invoker.lookupSubjectHistogram.Record(ctx, 1)

	// Call the LookupSubject function of the ls field in the invoker, pass the context and request,
	// and return its response and error
	return resp, err
}

// SubjectPermission is a method of the DirectInvoker structure. It handles the task of subject's permissions
// and returning the results in a response.
func (invoker *DirectInvoker) SubjectPermission(ctx context.Context, request *base.PermissionSubjectPermissionRequest) (response *base.PermissionSubjectPermissionResponse, err error) {
	ctx, span := internal.Tracer.Start(ctx, "subject-permission", trace.WithAttributes(
		attribute.KeyValue{Key: "tenant_id", Value: attribute.StringValue(request.GetTenantId())},
		attribute.KeyValue{Key: "entity", Value: attribute.StringValue(tuple.EntityToString(request.GetEntity()))},
		attribute.KeyValue{Key: "subject", Value: attribute.StringValue(tuple.SubjectToString(request.GetSubject()))},
	))
	defer span.End()

	// Check if the request has a SnapToken. If not, a SnapToken is set.
	if request.GetMetadata().GetSnapToken() == "" {
		// Create an instance of SnapToken
		var st token.SnapToken
		// Retrieve the head snapshot from the relationship reader
		st, err = invoker.dataReader.HeadSnapshot(ctx, request.GetTenantId())
		// If there's an error retrieving the snapshot, return the response and the error
		if err != nil {
			span.RecordError(err)
			span.SetStatus(otelCodes.Error, err.Error())
			return response, err
		}
		// Set the SnapToken in the request metadata
		request.Metadata.SnapToken = st.Encode().String()
	}

	// Similar to SnapToken, check if the request has a SchemaVersion. If not, a SchemaVersion is set.
	if request.GetMetadata().GetSchemaVersion() == "" {
		// Retrieve the head schema version from the schema reader
		request.Metadata.SchemaVersion, err = invoker.schemaReader.HeadVersion(ctx, request.GetTenantId())
		// If there's an error retrieving the schema version, return the response and the error
		if err != nil {
			span.RecordError(err)
			span.SetStatus(otelCodes.Error, err.Error())
			return response, err
		}
	}
	resp, err := invoker.sp.SubjectPermission(ctx, request)

	invoker.subjectPermissionHistogram.Record(ctx, 1)

	// Call the SubjectPermission function of the ls field in the invoker, pass the context and request,
	// and return its response and error
	return resp, err
}
