package cmd

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	basev1 "github.com/Permify/permify/pkg/pb/base/v1"
)

type stubPermissionClient struct {
	checkFn func(context.Context, *basev1.PermissionCheckRequest, ...grpc.CallOption) (*basev1.PermissionCheckResponse, error)
}

func (s *stubPermissionClient) Check(ctx context.Context, in *basev1.PermissionCheckRequest, opts ...grpc.CallOption) (*basev1.PermissionCheckResponse, error) {
	if s.checkFn != nil {
		return s.checkFn(ctx, in, opts...)
	}
	return nil, status.Errorf(codes.Unimplemented, "Check")
}

func (s *stubPermissionClient) BulkCheck(context.Context, *basev1.PermissionBulkCheckRequest, ...grpc.CallOption) (*basev1.PermissionBulkCheckResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "BulkCheck")
}

func (s *stubPermissionClient) Expand(context.Context, *basev1.PermissionExpandRequest, ...grpc.CallOption) (*basev1.PermissionExpandResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "Expand")
}

func (s *stubPermissionClient) LookupEntity(context.Context, *basev1.PermissionLookupEntityRequest, ...grpc.CallOption) (*basev1.PermissionLookupEntityResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "LookupEntity")
}

func (s *stubPermissionClient) LookupEntityStream(context.Context, *basev1.PermissionLookupEntityRequest, ...grpc.CallOption) (grpc.ServerStreamingClient[basev1.PermissionLookupEntityStreamResponse], error) {
	return nil, status.Errorf(codes.Unimplemented, "LookupEntityStream")
}

func (s *stubPermissionClient) LookupSubject(context.Context, *basev1.PermissionLookupSubjectRequest, ...grpc.CallOption) (*basev1.PermissionLookupSubjectResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "LookupSubject")
}

func (s *stubPermissionClient) SubjectPermission(context.Context, *basev1.PermissionSubjectPermissionRequest, ...grpc.CallOption) (*basev1.PermissionSubjectPermissionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "SubjectPermission")
}

type stubSchemaClient struct {
	writeFn func(context.Context, *basev1.SchemaWriteRequest, ...grpc.CallOption) (*basev1.SchemaWriteResponse, error)
	readFn  func(context.Context, *basev1.SchemaReadRequest, ...grpc.CallOption) (*basev1.SchemaReadResponse, error)
}

func (s *stubSchemaClient) Write(ctx context.Context, in *basev1.SchemaWriteRequest, opts ...grpc.CallOption) (*basev1.SchemaWriteResponse, error) {
	if s.writeFn != nil {
		return s.writeFn(ctx, in, opts...)
	}
	return nil, status.Errorf(codes.Unimplemented, "Write")
}

func (s *stubSchemaClient) Read(ctx context.Context, in *basev1.SchemaReadRequest, opts ...grpc.CallOption) (*basev1.SchemaReadResponse, error) {
	if s.readFn != nil {
		return s.readFn(ctx, in, opts...)
	}
	return nil, status.Errorf(codes.Unimplemented, "Read")
}

func (s *stubSchemaClient) PartialWrite(context.Context, *basev1.SchemaPartialWriteRequest, ...grpc.CallOption) (*basev1.SchemaPartialWriteResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "PartialWrite")
}

func (s *stubSchemaClient) List(context.Context, *basev1.SchemaListRequest, ...grpc.CallOption) (*basev1.SchemaListResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "List")
}

type stubDataClient struct {
	writeFn   func(context.Context, *basev1.DataWriteRequest, ...grpc.CallOption) (*basev1.DataWriteResponse, error)
	readRelFn func(context.Context, *basev1.RelationshipReadRequest, ...grpc.CallOption) (*basev1.RelationshipReadResponse, error)
}

func (s *stubDataClient) Write(ctx context.Context, in *basev1.DataWriteRequest, opts ...grpc.CallOption) (*basev1.DataWriteResponse, error) {
	if s.writeFn != nil {
		return s.writeFn(ctx, in, opts...)
	}
	return nil, status.Errorf(codes.Unimplemented, "Write")
}

func (s *stubDataClient) ReadRelationships(ctx context.Context, in *basev1.RelationshipReadRequest, opts ...grpc.CallOption) (*basev1.RelationshipReadResponse, error) {
	if s.readRelFn != nil {
		return s.readRelFn(ctx, in, opts...)
	}
	return nil, status.Errorf(codes.Unimplemented, "ReadRelationships")
}

func (s *stubDataClient) WriteRelationships(context.Context, *basev1.RelationshipWriteRequest, ...grpc.CallOption) (*basev1.RelationshipWriteResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "WriteRelationships")
}

func (s *stubDataClient) ReadAttributes(context.Context, *basev1.AttributeReadRequest, ...grpc.CallOption) (*basev1.AttributeReadResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "ReadAttributes")
}

func (s *stubDataClient) Delete(context.Context, *basev1.DataDeleteRequest, ...grpc.CallOption) (*basev1.DataDeleteResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "Delete")
}

func (s *stubDataClient) DeleteRelationships(context.Context, *basev1.RelationshipDeleteRequest, ...grpc.CallOption) (*basev1.RelationshipDeleteResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "DeleteRelationships")
}

func (s *stubDataClient) RunBundle(context.Context, *basev1.BundleRunRequest, ...grpc.CallOption) (*basev1.BundleRunResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "RunBundle")
}

type stubTenancyClient struct {
	createFn func(context.Context, *basev1.TenantCreateRequest, ...grpc.CallOption) (*basev1.TenantCreateResponse, error)
	listFn   func(context.Context, *basev1.TenantListRequest, ...grpc.CallOption) (*basev1.TenantListResponse, error)
	deleteFn func(context.Context, *basev1.TenantDeleteRequest, ...grpc.CallOption) (*basev1.TenantDeleteResponse, error)
}

func (s *stubTenancyClient) Create(ctx context.Context, in *basev1.TenantCreateRequest, opts ...grpc.CallOption) (*basev1.TenantCreateResponse, error) {
	if s.createFn != nil {
		return s.createFn(ctx, in, opts...)
	}
	return nil, status.Errorf(codes.Unimplemented, "Create")
}

func (s *stubTenancyClient) List(ctx context.Context, in *basev1.TenantListRequest, opts ...grpc.CallOption) (*basev1.TenantListResponse, error) {
	if s.listFn != nil {
		return s.listFn(ctx, in, opts...)
	}
	return nil, status.Errorf(codes.Unimplemented, "List")
}

func (s *stubTenancyClient) Delete(ctx context.Context, in *basev1.TenantDeleteRequest, opts ...grpc.CallOption) (*basev1.TenantDeleteResponse, error) {
	if s.deleteFn != nil {
		return s.deleteFn(ctx, in, opts...)
	}
	return nil, status.Errorf(codes.Unimplemented, "Delete")
}

// Compile-time checks: stubs must satisfy the full gRPC client interfaces.
var (
	_ basev1.PermissionClient = (*stubPermissionClient)(nil)
	_ basev1.SchemaClient     = (*stubSchemaClient)(nil)
	_ basev1.DataClient       = (*stubDataClient)(nil)
	_ basev1.TenancyClient    = (*stubTenancyClient)(nil)
)
