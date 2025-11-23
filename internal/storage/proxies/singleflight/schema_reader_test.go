package singleflight

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/pkg/database"
)

// MockSchemaReader is a mock implementation of storage.SchemaReader for testing
type MockSchemaReader struct {
	storage.NoopSchemaReader
	headVersionCalls map[string]*int64
	mu               sync.Mutex
}

func NewMockSchemaReader() *MockSchemaReader {
	return &MockSchemaReader{
		headVersionCalls: make(map[string]*int64),
	}
}

func (m *MockSchemaReader) HeadVersion(ctx context.Context, tenantID string) (string, error) {
	// Track call count per tenant
	m.mu.Lock()
	counter, exists := m.headVersionCalls[tenantID]
	if !exists {
		counter = new(int64)
		m.headVersionCalls[tenantID] = counter
	}
	m.mu.Unlock()

	// Increment call count
	atomic.AddInt64(counter, 1)

	// Simulate some work
	time.Sleep(10 * time.Millisecond)

	return "version-" + tenantID, nil
}

// ErrorMockSchemaReader is a mock that returns errors for testing error handling
type ErrorMockSchemaReader struct {
	storage.NoopSchemaReader
}

func (m *ErrorMockSchemaReader) HeadVersion(ctx context.Context, tenantID string) (string, error) {
	return "", errors.New("delegate error")
}

func GetVersionCallCount(m *MockSchemaReader, tenantID string) int64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	if counter, exists := m.headVersionCalls[tenantID]; exists {
		return atomic.LoadInt64(counter)
	}
	return 0
}

var _ = Describe("Singleflight SchemaReader", func() {
	var (
		mockDelegate storage.SchemaReader
		reader       *SchemaReader
		ctx          context.Context
	)

	BeforeEach(func() {
		mockDelegate = NewMockSchemaReader()
		reader = NewSchemaReader(mockDelegate)
		ctx = context.Background()
	})

	Describe("NewSchemaReader", func() {
		It("should create a new SchemaReader with delegate", func() {
			delegate := storage.NewNoopSchemaReader()
			reader := NewSchemaReader(delegate)

			Expect(reader).ShouldNot(BeNil())
			Expect(reader.delegate).Should(Equal(delegate))
		})

		It("should create a new SchemaReader with nil delegate", func() {
			reader := NewSchemaReader(nil)

			Expect(reader).ShouldNot(BeNil())
			Expect(reader.delegate).Should(BeNil())
		})
	})

	Describe("ReadSchema", func() {
		It("should delegate to underlying SchemaReader", func() {
			delegate := storage.NewNoopSchemaReader()
			reader := NewSchemaReader(delegate)

			schema, err := reader.ReadSchema(ctx, "tenant1", "v1.0")

			Expect(err).ShouldNot(HaveOccurred())
			Expect(schema).ShouldNot(BeNil())
		})
	})

	Describe("ReadSchemaString", func() {
		It("should delegate to underlying SchemaReader", func() {
			delegate := storage.NewNoopSchemaReader()
			reader := NewSchemaReader(delegate)

			definitions, err := reader.ReadSchemaString(ctx, "tenant1", "v1.0")

			Expect(err).ShouldNot(HaveOccurred())
			Expect(definitions).ShouldNot(BeNil())
		})
	})

	Describe("ReadEntityDefinition", func() {
		It("should delegate to underlying SchemaReader", func() {
			delegate := storage.NewNoopSchemaReader()
			reader := NewSchemaReader(delegate)

			definition, version, err := reader.ReadEntityDefinition(ctx, "tenant1", "user", "v1.0")

			Expect(err).ShouldNot(HaveOccurred())
			Expect(definition).ShouldNot(BeNil())
			// NoopSchemaReader returns empty string for version
			Expect(version).Should(Equal(""))
		})
	})

	Describe("ReadRuleDefinition", func() {
		It("should delegate to underlying SchemaReader", func() {
			delegate := storage.NewNoopSchemaReader()
			reader := NewSchemaReader(delegate)

			definition, version, err := reader.ReadRuleDefinition(ctx, "tenant1", "check_balance", "v1.0")

			Expect(err).ShouldNot(HaveOccurred())
			Expect(definition).ShouldNot(BeNil())
			// NoopSchemaReader returns empty string for version
			Expect(version).Should(Equal(""))
		})
	})

	Describe("ListSchemas", func() {
		It("should delegate to underlying SchemaReader", func() {
			delegate := storage.NewNoopSchemaReader()
			reader := NewSchemaReader(delegate)

			schemas, ct, err := reader.ListSchemas(ctx, "tenant1", database.Pagination{})

			Expect(err).ShouldNot(HaveOccurred())
			// NoopSchemaReader returns nil for both schemas and ct
			Expect(schemas).Should(BeNil())
			Expect(ct).Should(BeNil())
		})
	})

	Describe("HeadVersion", func() {
		It("should deduplicate concurrent requests for the same tenant", func() {
			tenantID := "tenant1"
			numConcurrentRequests := 10

			var wg sync.WaitGroup
			wg.Add(numConcurrentRequests)

			// Launch concurrent requests for the same tenant
			for i := 0; i < numConcurrentRequests; i++ {
				go func() {
					defer wg.Done()
					_, err := reader.HeadVersion(ctx, tenantID)
					Expect(err).ShouldNot(HaveOccurred())
				}()
			}

			wg.Wait()

			// Only 1 call should reach the delegate due to deduplication
			mock := mockDelegate.(*MockSchemaReader)
			callCount := GetVersionCallCount(mock, tenantID)
			Expect(callCount).To(Equal(int64(1)))
		})

		It("should isolate requests for different tenants", func() {
			tenant1 := "tenant1"
			tenant2 := "tenant2"
			numConcurrentRequests := 5

			var wg sync.WaitGroup
			wg.Add(numConcurrentRequests * 2)

			// Launch concurrent requests for tenant1
			for i := 0; i < numConcurrentRequests; i++ {
				go func() {
					defer wg.Done()
					_, err := reader.HeadVersion(ctx, tenant1)
					Expect(err).ShouldNot(HaveOccurred())
				}()
			}

			// Launch concurrent requests for tenant2
			for i := 0; i < numConcurrentRequests; i++ {
				go func() {
					defer wg.Done()
					_, err := reader.HeadVersion(ctx, tenant2)
					Expect(err).ShouldNot(HaveOccurred())
				}()
			}

			wg.Wait()

			mock := mockDelegate.(*MockSchemaReader)

			// Each tenant should have exactly 1 call due to deduplication within the tenant
			Expect(GetVersionCallCount(mock, tenant1)).To(Equal(int64(1)))
			Expect(GetVersionCallCount(mock, tenant2)).To(Equal(int64(1)))
		})

		It("should return correct version for each tenant", func() {
			tenant1 := "tenant1"
			tenant2 := "tenant2"

			var wg sync.WaitGroup
			wg.Add(2)

			var result1, result2 string

			go func() {
				defer wg.Done()
				var err error
				result1, err = reader.HeadVersion(ctx, tenant1)
				Expect(err).ShouldNot(HaveOccurred())
			}()

			go func() {
				defer wg.Done()
				var err error
				result2, err = reader.HeadVersion(ctx, tenant2)
				Expect(err).ShouldNot(HaveOccurred())
			}()

			wg.Wait()

			// Verify that each tenant gets its own version
			Expect(result1).To(Equal("version-" + tenant1))
			Expect(result2).To(Equal("version-" + tenant2))
		})

		It("should not deduplicate sequential requests", func() {
			tenantID := "tenant1"

			// First request
			_, err := reader.HeadVersion(ctx, tenantID)
			Expect(err).ShouldNot(HaveOccurred())

			// Small delay to ensure first request completes
			time.Sleep(50 * time.Millisecond)

			// Second request (should trigger another call to delegate)
			_, err = reader.HeadVersion(ctx, tenantID)
			Expect(err).ShouldNot(HaveOccurred())

			mock := mockDelegate.(*MockSchemaReader)

			// Should have 2 calls to the delegate
			callCount := GetVersionCallCount(mock, tenantID)
			Expect(callCount).To(Equal(int64(2)))
		})

		It("should propagate errors from delegate", func() {
			errorDelegate := &ErrorMockSchemaReader{}
			errorReader := NewSchemaReader(errorDelegate)

			_, err := errorReader.HeadVersion(ctx, "tenant1")

			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(Equal("delegate error"))
		})

		It("should handle context cancellation", func() {
			cancelledCtx, cancel := context.WithCancel(ctx)
			cancel() // Cancel immediately

			_, err := reader.HeadVersion(cancelledCtx, "tenant1")

			// Context cancellation behavior depends on singleflight implementation
			// We just verify it doesn't panic
			_ = err
		})

		It("should handle concurrent requests with errors", func() {
			errorDelegate := &ErrorMockSchemaReader{}
			errorReader := NewSchemaReader(errorDelegate)
			tenantID := "tenant1"
			numConcurrentRequests := 5

			var wg sync.WaitGroup
			wg.Add(numConcurrentRequests)

			errorCount := int64(0)

			// Launch concurrent requests that will all fail
			for i := 0; i < numConcurrentRequests; i++ {
				go func() {
					defer wg.Done()
					_, err := errorReader.HeadVersion(ctx, tenantID)
					if err != nil {
						atomic.AddInt64(&errorCount, 1)
					}
				}()
			}

			wg.Wait()

			// All requests should receive the error
			Expect(atomic.LoadInt64(&errorCount)).To(Equal(int64(numConcurrentRequests)))
		})
	})
})
