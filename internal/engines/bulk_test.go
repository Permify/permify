package engines

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Permify/permify/internal/invoke"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// mockCheckEngine is a mock implementation of invoke.Check for testing
type mockCheckEngine struct{}

func (m *mockCheckEngine) Check(ctx context.Context, request *base.PermissionCheckRequest) (*base.PermissionCheckResponse, error) {
	return &base.PermissionCheckResponse{
		Can: base.CheckResult_CHECK_RESULT_ALLOWED,
		Metadata: &base.PermissionCheckResponseMetadata{
			CheckCount: 1,
		},
	}, nil
}

// errorCheckEngine is a mock implementation that returns errors
type errorCheckEngine struct{}

func (e *errorCheckEngine) Check(ctx context.Context, request *base.PermissionCheckRequest) (*base.PermissionCheckResponse, error) {
	return &base.PermissionCheckResponse{
		Can: base.CheckResult_CHECK_RESULT_UNSPECIFIED,
		Metadata: &base.PermissionCheckResponseMetadata{
			CheckCount: 0,
		},
	}, errors.New("permission check failed")
}

var _ = Describe("Bulk", func() {
	Context("DefaultBulkCheckerConfig", func() {
		It("should return sensible default configuration", func() {
			config := DefaultBulkCheckerConfig()

			Expect(config.ConcurrencyLimit).Should(Equal(10))
			Expect(config.BufferSize).Should(Equal(1000))
		})
	})

	Context("NewBulkChecker", func() {
		var mockChecker invoke.Check

		BeforeEach(func() {
			mockChecker = &mockCheckEngine{}
		})

		It("should return error for nil context", func() {
			config := BulkCheckerConfig{
				ConcurrencyLimit: 5,
				BufferSize:       100,
			}

			_, err := NewBulkChecker(nil, mockChecker, BulkCheckerTypeEntity, func(entityID, ct string) {}, config)

			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(Equal("context cannot be nil"))
		})

		It("should return error for nil callback", func() {
			ctx := context.Background()
			config := BulkCheckerConfig{
				ConcurrencyLimit: 5,
				BufferSize:       100,
			}

			_, err := NewBulkChecker(ctx, mockChecker, BulkCheckerTypeEntity, nil, config)

			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(Equal("callback cannot be nil"))
		})

		It("should use default concurrency limit when not specified", func() {
			ctx := context.Background()
			config := BulkCheckerConfig{
				BufferSize: 100,
			}

			checker, err := NewBulkChecker(ctx, mockChecker, BulkCheckerTypeEntity, func(entityID, ct string) {}, config)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(checker.config.ConcurrencyLimit).Should(Equal(10)) // Default value
		})

		It("should use default buffer size when not specified", func() {
			ctx := context.Background()
			config := BulkCheckerConfig{
				ConcurrencyLimit: 5,
			}

			checker, err := NewBulkChecker(ctx, mockChecker, BulkCheckerTypeEntity, func(entityID, ct string) {}, config)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(checker.config.BufferSize).Should(Equal(1000)) // Default value
		})
	})

	Context("BulkChecker", func() {
		var checker *BulkChecker
		var ctx context.Context
		var mockChecker invoke.Check

		BeforeEach(func() {
			ctx = context.Background()
			config := BulkCheckerConfig{
				ConcurrencyLimit: 2,
				BufferSize:       10,
			}

			mockChecker = &mockCheckEngine{}

			var err error
			checker, err = NewBulkChecker(ctx, mockChecker, BulkCheckerTypeEntity, func(entityID, ct string) {}, config)
			Expect(err).ShouldNot(HaveOccurred())
		})

		Context("sortRequests", func() {
			It("should sort requests by entity ID for entity type", func() {
				requests := []BulkCheckerRequest{
					{
						Request: &base.PermissionCheckRequest{
							Entity: &base.Entity{Type: "user", Id: "user3"},
						},
					},
					{
						Request: &base.PermissionCheckRequest{
							Entity: &base.Entity{Type: "user", Id: "user1"},
						},
					},
					{
						Request: &base.PermissionCheckRequest{
							Entity: &base.Entity{Type: "user", Id: "user2"},
						},
					},
				}

				checker.sortRequests(requests)

				Expect(requests[0].Request.GetEntity().GetId()).Should(Equal("user1"))
				Expect(requests[1].Request.GetEntity().GetId()).Should(Equal("user2"))
				Expect(requests[2].Request.GetEntity().GetId()).Should(Equal("user3"))
			})

			It("should sort requests by subject ID for subject type", func() {
				// Create a new checker with subject type
				config := BulkCheckerConfig{
					ConcurrencyLimit: 2,
					BufferSize:       10,
				}

				mockChecker := &mockCheckEngine{}
				subjectChecker, err := NewBulkChecker(ctx, mockChecker, BulkCheckerTypeSubject, func(entityID, ct string) {}, config)
				Expect(err).ShouldNot(HaveOccurred())

				requests := []BulkCheckerRequest{
					{
						Request: &base.PermissionCheckRequest{
							Subject: &base.Subject{Type: "user", Id: "user3"},
						},
					},
					{
						Request: &base.PermissionCheckRequest{
							Subject: &base.Subject{Type: "user", Id: "user1"},
						},
					},
					{
						Request: &base.PermissionCheckRequest{
							Subject: &base.Subject{Type: "user", Id: "user2"},
						},
					},
				}

				subjectChecker.sortRequests(requests)

				Expect(requests[0].Request.GetSubject().GetId()).Should(Equal("user1"))
				Expect(requests[1].Request.GetSubject().GetId()).Should(Equal("user2"))
				Expect(requests[2].Request.GetSubject().GetId()).Should(Equal("user3"))
			})
		})

		Context("ExecuteRequests", func() {
			It("should return error for size 0", func() {
				err := checker.ExecuteRequests(0)
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).Should(Equal("size must be greater than 0"))
			})

			It("should handle context cancellation gracefully", func() {
				// Create a context that will be cancelled
				cancelCtx, cancel := context.WithCancel(context.Background())
				cancel()

				config := BulkCheckerConfig{
					ConcurrencyLimit: 2,
					BufferSize:       10,
				}

				cancelledChecker, err := NewBulkChecker(cancelCtx, mockChecker, BulkCheckerTypeEntity, func(entityID, ct string) {}, config)
				Expect(err).ShouldNot(HaveOccurred())

				// This should handle context cancellation gracefully
				err = cancelledChecker.ExecuteRequests(1)
				Expect(err).ShouldNot(HaveOccurred()) // Context errors should not be treated as errors
			})

			It("should return nil on successful execution", func() {
				err := checker.ExecuteRequests(1)
				Expect(err).ShouldNot(HaveOccurred())
			})
		})

		Context("getRequestResult", func() {
			It("should return result for successful request", func() {
				request := BulkCheckerRequest{
					Request: &base.PermissionCheckRequest{
						TenantId:   "t1",
						Entity:     &base.Entity{Type: "user", Id: "user1"},
						Subject:    &base.Subject{Type: "user", Id: "user2"},
						Permission: "read",
					},
					Result: base.CheckResult_CHECK_RESULT_ALLOWED,
				}

				result, err := checker.getRequestResult(ctx, request)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(result).Should(Equal(base.CheckResult_CHECK_RESULT_ALLOWED))
			})

			It("should return error for failed request", func() {
				// Create a checker with a callback that returns an error
				errorChecker := &errorCheckEngine{}

				config := BulkCheckerConfig{
					ConcurrencyLimit: 2,
					BufferSize:       10,
				}

				errorBulkChecker, err := NewBulkChecker(ctx, errorChecker, BulkCheckerTypeEntity, func(entityID, ct string) {}, config)
				Expect(err).ShouldNot(HaveOccurred())

				request := BulkCheckerRequest{
					Request: &base.PermissionCheckRequest{
						TenantId:   "t1",
						Entity:     &base.Entity{Type: "user", Id: "user1"},
						Subject:    &base.Subject{Type: "user", Id: "user2"},
						Permission: "read",
					},
					Result: base.CheckResult_CHECK_RESULT_UNSPECIFIED,
				}

				result, err := errorBulkChecker.getRequestResult(ctx, request)
				Expect(err).Should(HaveOccurred())
				Expect(result).Should(Equal(base.CheckResult_CHECK_RESULT_UNSPECIFIED))
			})
		})
	})

	Context("BulkEntityPublisher", func() {
		It("should handle context cancellation in Publish", func() {
			// Create a context that will be cancelled
			cancelCtx, cancel := context.WithCancel(context.Background())
			cancel()

			config := BulkCheckerConfig{
				ConcurrencyLimit: 2,
				BufferSize:       10,
			}

			mockChecker := &mockCheckEngine{}

			cancelledChecker, err := NewBulkChecker(cancelCtx, mockChecker, BulkCheckerTypeEntity, func(entityID, ct string) {}, config)
			Expect(err).ShouldNot(HaveOccurred())

			request := &base.PermissionLookupEntityRequest{
				TenantId:   "t1",
				Subject:    &base.Subject{Type: "user", Id: "user1"},
				Permission: "read",
			}

			cancelledPublisher := NewBulkEntityPublisher(cancelCtx, request, cancelledChecker)

			// This should handle context cancellation gracefully
			cancelledPublisher.Publish(&base.Entity{Type: "user", Id: "user2"}, &base.PermissionCheckRequestMetadata{}, &base.Context{}, base.CheckResult_CHECK_RESULT_ALLOWED)
			// The test passes if no panic occurs
		})
	})

	Context("BulkSubjectPublisher", func() {
		It("should handle context cancellation in Publish", func() {
			// Create a context that will be cancelled
			cancelCtx, cancel := context.WithCancel(context.Background())
			cancel()

			config := BulkCheckerConfig{
				ConcurrencyLimit: 2,
				BufferSize:       10,
			}

			mockChecker := &mockCheckEngine{}

			cancelledChecker, err := NewBulkChecker(cancelCtx, mockChecker, BulkCheckerTypeSubject, func(entityID, ct string) {}, config)
			Expect(err).ShouldNot(HaveOccurred())

			request := &base.PermissionLookupSubjectRequest{
				TenantId:   "t1",
				Entity:     &base.Entity{Type: "user", Id: "user1"},
				Permission: "read",
			}

			cancelledPublisher := NewBulkSubjectPublisher(cancelCtx, request, cancelledChecker)

			// This should handle context cancellation gracefully
			cancelledPublisher.Publish(&base.Subject{Type: "user", Id: "user2"}, &base.PermissionCheckRequestMetadata{}, &base.Context{}, base.CheckResult_CHECK_RESULT_ALLOWED)
			// The test passes if no panic occurs
		})
	})

	Context("Context Error Functions", func() {
		It("should identify context.Canceled as context error", func() {
			Expect(isContextError(context.Canceled)).Should(BeTrue())
		})

		It("should identify context.DeadlineExceeded as context error", func() {
			Expect(isContextError(context.DeadlineExceeded)).Should(BeTrue())
		})

		It("should not identify regular errors as context errors", func() {
			Expect(isContextError(errors.New("regular error"))).Should(BeFalse())
		})

		It("should identify context.Canceled as context related error", func() {
			ctx := context.Background()
			Expect(IsContextRelatedError(ctx, context.Canceled)).Should(BeTrue())
		})

		It("should identify context.DeadlineExceeded as context related error", func() {
			ctx := context.Background()
			Expect(IsContextRelatedError(ctx, context.DeadlineExceeded)).Should(BeTrue())
		})

		It("should not identify regular errors as context related errors", func() {
			ctx := context.Background()
			Expect(IsContextRelatedError(ctx, errors.New("regular error"))).Should(BeFalse())
		})
	})
})
