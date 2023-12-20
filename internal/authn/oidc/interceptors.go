package oidc

import (
	"context"

	"google.golang.org/grpc"
)

// UnaryServerInterceptor returns a gRPC unary server interceptor that
// performs authentication using the provided Authenticator.
func UnaryServerInterceptor(t Authenticator) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Authenticate the request.
		err := t.Authenticate(ctx)
		if err != nil {
			// If authentication fails, return the error.
			return nil, err
		}
		// If authentication succeeds, proceed with the request.
		return handler(ctx, req)
	}
}

// StreamServerInterceptor returns a gRPC stream server interceptor that
// wraps the incoming stream with an authenticator.
func StreamServerInterceptor(t Authenticator) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// Wrap the stream with the authenticator.
		wrapper := &authnWrapper{ServerStream: stream, authenticator: t}
		return handler(srv, wrapper)
	}
}

// authnWrapper wraps a grpc.ServerStream and intercepts its RecvMsg
// method to perform authentication on each message received.
type authnWrapper struct {
	grpc.ServerStream
	authenticator Authenticator
}

// RecvMsg intercepts the RecvMsg call of the wrapped grpc.ServerStream
// to perform authentication before processing the message.
func (s *authnWrapper) RecvMsg(req interface{}) error {
	// Receive the message from the original stream.
	if err := s.ServerStream.RecvMsg(req); err != nil {
		return err
	}
	// Authenticate the received message.
	err := s.authenticator.Authenticate(s.ServerStream.Context())
	if err != nil {
		// If authentication fails, return the error.
		return err
	}
	// If authentication succeeds, proceed with the message.
	return nil
}
