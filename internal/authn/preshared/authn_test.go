package preshared

import (
	"context"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Permify/permify/internal/config"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

func TestPresharedKeyAuth(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "authentication preshared key suite")
}

var _ = Describe("KeyAuthn", func() {
	var (
		ctx           context.Context
		authenticator *KeyAuthn
		err           error
		keysConfig    config.Preshared
	)

	BeforeEach(func() {
		keysConfig = config.Preshared{Keys: []string{"key1", "key2"}}
		authenticator, err = NewKeyAuthn(context.Background(), keysConfig)
		Expect(err).ToNot(HaveOccurred())
	})

	Describe("Authenticate", func() {
		Context("with valid Bearer token", func() {
			BeforeEach(func() {
				md := metadata.New(map[string]string{"authorization": "Bearer key1"})
				ctx = metadata.NewIncomingContext(context.Background(), md)
			})

			It("should authenticate successfully", func() {
				err := authenticator.Authenticate(ctx)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("with invalid Bearer token", func() {
			BeforeEach(func() {
				md := metadata.New(map[string]string{"authorization": "Bearer invalidkey"})
				ctx = metadata.NewIncomingContext(context.Background(), md)
			})

			It("should return an error", func() {
				err := authenticator.Authenticate(ctx)
				Expect(err).To(HaveOccurred())
				Expect(status.Code(err)).To(Equal(codes.Unauthenticated))
			})
		})

		Context("with missing Bearer token", func() {
			BeforeEach(func() {
				ctx = context.Background()
			})

			It("should return an error", func() {
				err := authenticator.Authenticate(ctx)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).Should(Equal(base.ErrorCode_ERROR_CODE_MISSING_BEARER_TOKEN.String()))
			})
		})
	})
})
