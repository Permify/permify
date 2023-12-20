package oidc

import (
	"context"
	"errors"
	"testing"

	"google.golang.org/grpc/metadata"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"
)

func TestAuthInterceptors(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "authentication interceptors suite")
}

var _ = Describe("Auth Interceptors", func() {
	fakeError := errors.New("fake authentication error")

	Describe("UnaryServerInterceptor", func() {
		var authenticator Authenticator
		var interceptor grpc.UnaryServerInterceptor
		var handlerCalled bool

		BeforeEach(func() {
			handlerCalled = false
		})

		Context("when authentication is successful", func() {
			BeforeEach(func() {
				authenticator = &mockAuthenticator{err: nil}
				interceptor = UnaryServerInterceptor(authenticator)
			})

			It("should call the handler and not return an error", func() {
				_, err := interceptor(context.Background(), nil, nil, func(ctx context.Context, req interface{}) (interface{}, error) {
					handlerCalled = true
					return "success", nil
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(handlerCalled).To(BeTrue())
			})
		})

		Context("when authentication fails", func() {
			BeforeEach(func() {
				authenticator = &mockAuthenticator{err: fakeError}
				interceptor = UnaryServerInterceptor(authenticator)
			})

			It("should not call the handler and return an error", func() {
				_, err := interceptor(context.Background(), nil, nil, func(ctx context.Context, req interface{}) (interface{}, error) {
					handlerCalled = true
					return nil, nil
				})
				Expect(err).To(MatchError(fakeError))
				Expect(handlerCalled).To(BeFalse())
			})
		})
	})

	Describe("StreamServerInterceptor", func() {
		var authenticator Authenticator
		var interceptor grpc.StreamServerInterceptor
		var handlerCalled bool
		var mockStream *mockServerStream

		BeforeEach(func() {
			handlerCalled = false
			mockStream = &mockServerStream{}
		})

		Context("when authentication is successful", func() {
			BeforeEach(func() {
				authenticator = &mockAuthenticator{err: nil}
				interceptor = StreamServerInterceptor(authenticator)
			})

			It("should call the handler and not return an error", func() {
				err := interceptor(nil, mockStream, nil, func(srv interface{}, stream grpc.ServerStream) error {
					handlerCalled = true
					return nil
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(handlerCalled).To(BeTrue())
			})
		})
	})

	Describe("authnWrapper", func() {
		var wrapper *authnWrapper
		var mockStream *mockServerStream
		var authenticator Authenticator

		BeforeEach(func() {
			mockStream = &mockServerStream{}
		})

		Context("when authentication is successful", func() {
			BeforeEach(func() {
				authenticator = &mockAuthenticator{err: nil}
				wrapper = &authnWrapper{ServerStream: mockStream, authenticator: authenticator}
			})

			It("should call the original RecvMsg and not return an error", func() {
				err := wrapper.RecvMsg(nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(mockStream.recvMsgCalled).To(BeTrue())
			})
		})

		Context("when authentication fails", func() {
			BeforeEach(func() {
				authenticator = &mockAuthenticator{err: fakeError}
				wrapper = &authnWrapper{ServerStream: mockStream, authenticator: authenticator}
			})

			It("should return an error without processing the message", func() {
				err := wrapper.RecvMsg(nil)
				Expect(err).To(MatchError(fakeError))
				Expect(mockStream.recvMsgCalled).To(BeTrue())
			})
		})
	})
})

// mockServerStream is a fake implementation of the grpc.ServerStream for testing.
type mockServerStream struct {
	recvMsgCalled bool
}

func (m *mockServerStream) SetHeader(md metadata.MD) error {
	return nil
}

func (m *mockServerStream) SendHeader(md metadata.MD) error {
	return nil
}

func (m *mockServerStream) SetTrailer(md metadata.MD) {}

func (m *mockServerStream) Context() context.Context {
	return context.Background()
}

func (m *mockServerStream) SendMsg(a any) error {
	return nil
}

func (m *mockServerStream) RecvMsg(x interface{}) error {
	m.recvMsgCalled = true
	return nil
}

type mockAuthenticator struct {
	err error
}

func (m *mockAuthenticator) Authenticate(ctx context.Context) error {
	return m.err
}
