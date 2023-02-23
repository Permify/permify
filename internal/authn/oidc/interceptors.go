package oidc

import (
	"context"

	"google.golang.org/grpc"
)

// UnaryServerInterceptor -
func UnaryServerInterceptor(t OidcAuthenticator) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		err := t.Authenticate(ctx)
		if err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}
}

// StreamServerInterceptor -
func StreamServerInterceptor(t OidcAuthenticator) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		wrapper := &authnWrapper{ServerStream: stream, authenticator: t}
		return handler(srv, wrapper)
	}
}

// authnWrapper -
type authnWrapper struct {
	grpc.ServerStream
	authenticator OidcAuthenticator
}

// RecvMsg -
func (s *authnWrapper) RecvMsg(req interface{}) error {
	if err := s.ServerStream.RecvMsg(req); err != nil {
		return err
	}
	err := s.authenticator.Authenticate(s.ServerStream.Context())
	if err != nil {
		return err
	}
	return nil
}
