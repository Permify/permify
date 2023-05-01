package invoke

import (
	"context"

	"github.com/Permify/permify/internal/engines"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// Invoker is an interface that groups multiple permission-related interfaces.
// It is used to define a common contract for invoking various permission operations.
type Invoker interface {
	Check
	Expand
	LookupEntity
	LookupEntityStream
	LookupSchema
}

// Check is an interface that defines a method for checking permissions.
// It requires an implementation of InvokeCheck that takes a context and a PermissionCheckRequest,
// and returns a PermissionCheckResponse and an error if any.
type Check interface {
	InvokeCheck(ctx context.Context, request *base.PermissionCheckRequest) (response *base.PermissionCheckResponse, err error)
}

// Expand is an interface that defines a method for expanding permissions.
// It requires an implementation of InvokeExpand that takes a context and a PermissionExpandRequest,
// and returns a PermissionExpandResponse and an error if any.
type Expand interface {
	InvokeExpand(ctx context.Context, request *base.PermissionExpandRequest) (response *base.PermissionExpandResponse, err error)
}

// LookupEntity is an interface that defines a method for looking up entities with permissions.
// It requires an implementation of InvokeLookupEntity that takes a context and a PermissionLookupEntityRequest,
// and returns a PermissionLookupEntityResponse and an error if any.
type LookupEntity interface {
	InvokeLookupEntity(ctx context.Context, request *base.PermissionLookupEntityRequest) (response *base.PermissionLookupEntityResponse, err error)
}

// LookupEntityStream is an interface that defines a method for streaming entities with permissions.
// It requires an implementation of InvokeLookupEntityStream that takes a context, a PermissionLookupEntityRequest,
// and a Permission_LookupEntityStreamServer, and returns an error if any.
type LookupEntityStream interface {
	InvokeLookupEntityStream(ctx context.Context, request *base.PermissionLookupEntityRequest, server base.Permission_LookupEntityStreamServer) (err error)
}

// LookupSchema is an interface that defines a method for looking up schemas with permissions.
// It requires an implementation of InvokeLookupSchema that takes a context and a PermissionLookupSchemaRequest,
// and returns a PermissionLookupSchemaResponse and an error if any.
type LookupSchema interface {
	InvokeLookupSchema(ctx context.Context, request *base.PermissionLookupSchemaRequest) (response *base.PermissionLookupSchemaResponse, err error)
}

// DirectInvoker is a struct that implements the Invoker interface.
// It holds references to various engines needed for permission-related operations.
type DirectInvoker struct {
	cc *engines.CheckEngine        // Check engine for permission checks
	ec *engines.ExpandEngine       // Expand engine for expanding permissions
	ls *engines.LookupSchemaEngine // LookupSchema engine for looking up schemas with permissions
	le *engines.LookupEntityEngine // LookupEntity engine for looking up entities with permissions
}

// NewDirectInvoker is a constructor for DirectInvoker.
// It takes pointers to CheckEngine, ExpandEngine, LookupSchemaEngine, and LookupEntityEngine as arguments
// and returns an Invoker instance.
func NewDirectInvoker(
	cc *engines.CheckEngine,
	ec *engines.ExpandEngine,
	ls *engines.LookupSchemaEngine,
	le *engines.LookupEntityEngine,
) Invoker {
	return &DirectInvoker{
		cc: cc,
		ec: ec,
		ls: ls,
		le: le,
	}
}

// InvokeCheck is a method that implements the Check interface.
// It calls the Run method of the CheckEngine with the provided context and PermissionCheckRequest,
// and returns a PermissionCheckResponse and an error if any.
func (invoker *DirectInvoker) InvokeCheck(ctx context.Context, request *base.PermissionCheckRequest) (response *base.PermissionCheckResponse, err error) {
	return invoker.cc.Run(ctx, request)
}

// InvokeExpand is a method that implements the Expand interface.
// It calls the Run method of the ExpandEngine with the provided context and PermissionExpandRequest,
// and returns a PermissionExpandResponse and an error if any.
func (invoker *DirectInvoker) InvokeExpand(ctx context.Context, request *base.PermissionExpandRequest) (response *base.PermissionExpandResponse, err error) {
	return invoker.ec.Run(ctx, request)
}

// InvokeLookupEntity is a method that implements the LookupEntity interface.
// It calls the Run method of the LookupEntityEngine with the provided context and PermissionLookupEntityRequest,
// and returns a PermissionLookupEntityResponse and an error if any.
func (invoker *DirectInvoker) InvokeLookupEntity(ctx context.Context, request *base.PermissionLookupEntityRequest) (response *base.PermissionLookupEntityResponse, err error) {
	return invoker.le.Run(ctx, request)
}

// InvokeLookupEntityStream is a method that implements the LookupEntityStream interface.
// It calls the Stream method of the LookupEntityEngine with the provided context, PermissionLookupEntityRequest, and Permission_LookupEntityStreamServer,
// and returns an error if any.
func (invoker *DirectInvoker) InvokeLookupEntityStream(ctx context.Context, request *base.PermissionLookupEntityRequest, server base.Permission_LookupEntityStreamServer) (err error) {
	return invoker.le.Stream(ctx, request, server)
}

// InvokeLookupSchema is a method that implements the LookupSchema interface.
// It calls the Run method of the LookupSchemaEngine with the provided context and PermissionLookupSchemaRequest,
// and returns a PermissionLookupSchemaResponse and an error if any.
func (invoker *DirectInvoker) InvokeLookupSchema(ctx context.Context, request *base.PermissionLookupSchemaRequest) (response *base.PermissionLookupSchemaResponse, err error) {
	return invoker.ls.Run(ctx, request)
}
