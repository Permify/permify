package multitenant

import (
	"context"

	"google.golang.org/grpc"

	"github.com/Permify/permify/internal/authn"
	basev1 "github.com/Permify/permify/pkg/pb/base/v1"
)

// UnaryServerInterceptor -
func UnaryServerInterceptor(t TenantAuthenticator) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		var reqTenantID string
		switch r := req.(type) {
		case *basev1.PermissionCheckRequest:
			reqTenantID = r.TenantId
		case *basev1.PermissionLookupEntityRequest:
			reqTenantID = r.TenantId
		case *basev1.PermissionLookupSchemaRequest:
			reqTenantID = r.TenantId
		case *basev1.PermissionExpandRequest:
			reqTenantID = r.TenantId
		case *basev1.RelationshipWriteRequest:
			reqTenantID = r.TenantId
		case *basev1.RelationshipDeleteRequest:
			reqTenantID = r.TenantId
		case *basev1.RelationshipReadRequest:
			reqTenantID = r.TenantId
		case *basev1.SchemaReadRequest:
			reqTenantID = r.TenantId
		case *basev1.SchemaWriteRequest:
			reqTenantID = r.TenantId
		case *basev1.TenantDeleteRequest:
			reqTenantID = r.Id
		default:
			return handler(ctx, req)
		}

		err := t.Authenticate(ctx, reqTenantID)
		if err != nil {
			return nil, authn.Unauthenticated
		}

		return handler(ctx, req)
	}
}

// StreamServerInterceptor -
func StreamServerInterceptor(t TenantAuthenticator) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		wrapper := &authnWrapper{ServerStream: stream, authenticator: t}
		return handler(srv, wrapper)
	}
}

// authnWrapper -
type authnWrapper struct {
	grpc.ServerStream
	authenticator TenantAuthenticator
}

// RecvMsg -
func (s *authnWrapper) RecvMsg(req interface{}) error {
	if err := s.ServerStream.RecvMsg(req); err != nil {
		return err
	}
	var reqTenantID string
	switch r := req.(type) {
	case *basev1.PermissionLookupEntityRequest:
		reqTenantID = r.TenantId
	default:
		return nil
	}
	err := s.authenticator.Authenticate(s.ServerStream.Context(), reqTenantID)
	if err != nil {
		return authn.Unauthenticated
	}
	return nil
}
