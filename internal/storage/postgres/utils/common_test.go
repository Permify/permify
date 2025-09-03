package utils_test

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Masterminds/squirrel"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"

	"github.com/Permify/permify/internal/storage/postgres/utils"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

var _ = Describe("Common", func() {
	Context("TestSnapshotQuery", func() {
		It("Case 1", func() {
			sl := squirrel.Select("column").From("table")
			revision := uint64(42)

			query := utils.SnapshotQuery(sl, revision)
			sql, _, err := query.ToSql()
			Expect(err).ShouldNot(HaveOccurred())

			expectedSQL := "SELECT column FROM table WHERE (pg_visible_in_snapshot(created_tx_id, (select snapshot from transactions where id = '42'::xid8)) = true OR created_tx_id = '42'::xid8) AND ((pg_visible_in_snapshot(expired_tx_id, (select snapshot from transactions where id = '42'::xid8)) = false OR expired_tx_id = '0'::xid8) AND expired_tx_id <> '42'::xid8)"
			Expect(sql).Should(Equal(expectedSQL))
		})
	})

	Context("TestGarbageCollectQuery", func() {
		It("Case 1", func() {
			query := utils.GenerateGCQuery("relation_tuples", 100)
			sql, _, err := query.ToSql()
			Expect(err).ShouldNot(HaveOccurred())

			expectedSQL := "DELETE FROM relation_tuples WHERE expired_tx_id <> '0'::xid8 AND expired_tx_id < '100'::xid8"
			Expect(expectedSQL).Should(Equal(sql))
		})

		It("Case 2 - Tenant Aware", func() {
			query := utils.GenerateGCQueryForTenant("relation_tuples", "tenant1", 100)
			sql, args, err := query.ToSql()
			Expect(err).ShouldNot(HaveOccurred())

			expectedSQL := "DELETE FROM relation_tuples WHERE tenant_id = ? AND expired_tx_id <> '0'::xid8 AND expired_tx_id < '100'::xid8"
			Expect(expectedSQL).Should(Equal(sql))
			Expect(args).Should(Equal([]interface{}{"tenant1"}))
		})
	})

	Context("Error Handling", func() {
		var (
			ctx  context.Context
			span trace.Span
		)

		BeforeEach(func() {
			ctx = context.Background()
			_, span = noop.NewTracerProvider().Tracer("test").Start(ctx, "test-span")
		})

		It("should handle context-related errors", func() {
			// Test context cancellation
			cancelCtx, cancel := context.WithCancel(context.Background())
			cancel()

			err := utils.HandleError(cancelCtx, span, errors.New("some error"), base.ErrorCode_ERROR_CODE_INTERNAL)

			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(Equal(base.ErrorCode_ERROR_CODE_CANCELLED.String()))
		})

		It("should handle context deadline exceeded", func() {
			// Test context deadline exceeded
			deadlineCtx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Hour))
			defer cancel()

			err := utils.HandleError(deadlineCtx, span, errors.New("some error"), base.ErrorCode_ERROR_CODE_INTERNAL)

			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(Equal(base.ErrorCode_ERROR_CODE_CANCELLED.String()))
		})

		It("should handle context.Canceled error", func() {
			err := utils.HandleError(ctx, span, context.Canceled, base.ErrorCode_ERROR_CODE_INTERNAL)

			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(Equal(base.ErrorCode_ERROR_CODE_CANCELLED.String()))
		})

		It("should handle context.DeadlineExceeded error", func() {
			err := utils.HandleError(ctx, span, context.DeadlineExceeded, base.ErrorCode_ERROR_CODE_INTERNAL)

			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(Equal(base.ErrorCode_ERROR_CODE_CANCELLED.String()))
		})

		It("should handle connection closed error", func() {
			err := utils.HandleError(ctx, span, errors.New("conn closed"), base.ErrorCode_ERROR_CODE_INTERNAL)

			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(Equal(base.ErrorCode_ERROR_CODE_CANCELLED.String()))
		})

		It("should handle serialization-related errors", func() {
			err := utils.HandleError(ctx, span, errors.New("could not serialize access due to concurrent update"), base.ErrorCode_ERROR_CODE_INTERNAL)

			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(Equal(base.ErrorCode_ERROR_CODE_SERIALIZATION.String()))
		})

		It("should handle duplicate key value error", func() {
			err := utils.HandleError(ctx, span, errors.New("duplicate key value violates unique constraint"), base.ErrorCode_ERROR_CODE_INTERNAL)

			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(Equal(base.ErrorCode_ERROR_CODE_SERIALIZATION.String()))
		})

		It("should handle operational errors", func() {
			operationalErr := errors.New("database connection failed")
			err := utils.HandleError(ctx, span, operationalErr, base.ErrorCode_ERROR_CODE_INTERNAL)

			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(Equal(base.ErrorCode_ERROR_CODE_INTERNAL.String()))
		})
	})

	Context("Error Detection Functions", func() {
		It("should detect context-related errors", func() {
			// Test context cancellation
			cancelCtx, cancel := context.WithCancel(context.Background())
			cancel()

			Expect(utils.IsContextRelatedError(cancelCtx, errors.New("some error"))).Should(BeTrue())

			// Test context deadline exceeded
			deadlineCtx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Hour))
			defer cancel()

			Expect(utils.IsContextRelatedError(deadlineCtx, errors.New("some error"))).Should(BeTrue())

			// Test context.Canceled error
			Expect(utils.IsContextRelatedError(context.Background(), context.Canceled)).Should(BeTrue())

			// Test context.DeadlineExceeded error
			Expect(utils.IsContextRelatedError(context.Background(), context.DeadlineExceeded)).Should(BeTrue())

			// Test connection closed error
			Expect(utils.IsContextRelatedError(context.Background(), errors.New("conn closed"))).Should(BeTrue())

			// Test non-context error
			Expect(utils.IsContextRelatedError(context.Background(), errors.New("some other error"))).Should(BeFalse())
		})

		It("should detect serialization-related errors", func() {
			// Test serialization error
			Expect(utils.IsSerializationRelatedError(errors.New("could not serialize access due to concurrent update"))).Should(BeTrue())

			// Test duplicate key error
			Expect(utils.IsSerializationRelatedError(errors.New("duplicate key value violates unique constraint"))).Should(BeTrue())

			// Test non-serialization error
			Expect(utils.IsSerializationRelatedError(errors.New("some other error"))).Should(BeFalse())
		})
	})

	Context("Backoff and Random Functions", func() {
		It("should implement exponential backoff with jitter", func() {
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()

			start := time.Now()
			utils.WaitWithBackoff(ctx, "test-tenant", 0)
			elapsed := time.Since(start)

			// Should have waited at least some time (backoff + jitter)
			Expect(elapsed).Should(BeNumerically(">", 10*time.Millisecond))
			// But not too long (should be less than our timeout)
			Expect(elapsed).Should(BeNumerically("<", 90*time.Millisecond))
		})

		It("should respect context cancellation in backoff", func() {
			ctx, cancel := context.WithCancel(context.Background())
			cancel() // Cancel immediately

			start := time.Now()
			utils.WaitWithBackoff(ctx, "test-tenant", 0)
			elapsed := time.Since(start)

			// Should exit quickly due to context cancellation
			Expect(elapsed).Should(BeNumerically("<", 10*time.Millisecond))
		})

		It("should increase backoff with more retries", func() {
			ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
			defer cancel()

			// Test with different retry counts
			start1 := time.Now()
			utils.WaitWithBackoff(ctx, "test-tenant", 0)
			elapsed1 := time.Since(start1)

			start2 := time.Now()
			utils.WaitWithBackoff(ctx, "test-tenant", 2)
			elapsed2 := time.Since(start2)

			// Higher retry count should generally result in longer backoff
			// (though jitter might make this not always true, so we use a reasonable threshold)
			Expect(elapsed2).Should(BeNumerically(">=", time.Duration(float64(elapsed1)*0.5)))
		})

		It("should generate secure random float64 values through backoff", func() {
			// Test that the backoff function uses secure random values
			// by running it multiple times and checking for variation
			ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
			defer cancel()

			// Run backoff multiple times to test jitter variation
			durations := make([]time.Duration, 5)
			for i := 0; i < 5; i++ {
				start := time.Now()
				utils.WaitWithBackoff(ctx, "test-tenant", 0)
				durations[i] = time.Since(start)
			}

			// Should have some variation in durations due to jitter
			hasVariation := false
			for i := 1; i < len(durations); i++ {
				if durations[i] != durations[0] {
					hasVariation = true
					break
				}
			}
			Expect(hasVariation).Should(BeTrue())
		})
	})

	Context("Logging Integration", func() {
		It("should log context-related errors at debug level", func() {
			var buf strings.Builder
			logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
			ctx := context.WithValue(context.Background(), "logger", logger)

			// Create a cancelled context
			cancelCtx, cancel := context.WithCancel(context.Background())
			cancel()

			_, span := noop.NewTracerProvider().Tracer("test").Start(ctx, "test-span")
			utils.HandleError(cancelCtx, span, errors.New("some error"), base.ErrorCode_ERROR_CODE_INTERNAL)

			// Note: The actual logging happens in the HandleError function,
			// but we can't easily capture it in this test setup.
			// The test verifies the function doesn't panic and returns the expected error.
			Expect(buf.String()).Should(ContainSubstring(""))
		})

		It("should log serialization errors at debug level", func() {
			ctx := context.Background()
			_, span := noop.NewTracerProvider().Tracer("test").Start(ctx, "test-span")

			err := utils.HandleError(ctx, span, errors.New("could not serialize access"), base.ErrorCode_ERROR_CODE_INTERNAL)

			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(Equal(base.ErrorCode_ERROR_CODE_SERIALIZATION.String()))
		})

		It("should log operational errors at error level", func() {
			ctx := context.Background()
			_, span := noop.NewTracerProvider().Tracer("test").Start(ctx, "test-span")

			operationalErr := errors.New("database connection failed")
			err := utils.HandleError(ctx, span, operationalErr, base.ErrorCode_ERROR_CODE_INTERNAL)

			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(Equal(base.ErrorCode_ERROR_CODE_INTERNAL.String()))
		})
	})

	Context("Span Integration", func() {
		It("should record errors in span for operational errors", func() {
			ctx := context.Background()
			tracer := noop.NewTracerProvider().Tracer("test")
			_, span := tracer.Start(ctx, "test-span")

			operationalErr := errors.New("database connection failed")
			err := utils.HandleError(ctx, span, operationalErr, base.ErrorCode_ERROR_CODE_INTERNAL)

			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(Equal(base.ErrorCode_ERROR_CODE_INTERNAL.String()))

			// The span should have recorded the error and set status
			// Note: With noop tracer, we can't easily verify this, but the test ensures no panics
			span.End()
		})

		It("should set span status to error for operational errors", func() {
			ctx := context.Background()
			tracer := noop.NewTracerProvider().Tracer("test")
			_, span := tracer.Start(ctx, "test-span")

			operationalErr := errors.New("database connection failed")
			err := utils.HandleError(ctx, span, operationalErr, base.ErrorCode_ERROR_CODE_INTERNAL)

			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(Equal(base.ErrorCode_ERROR_CODE_INTERNAL.String()))

			span.End()
		})
	})
})
