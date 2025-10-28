package singleflight

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Permify/permify/internal/storage"
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
	})
})
