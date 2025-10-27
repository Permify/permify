package singleflight

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/pkg/token"
)

// MockDataReader is a mock implementation of storage.DataReader for testing
type MockDataReader struct {
	storage.NoopDataReader
	headSnapshotCalls map[string]*int64
	mu                sync.Mutex
}

func NewMockDataReader() *MockDataReader {
	return &MockDataReader{
		headSnapshotCalls: make(map[string]*int64),
	}
}

func (m *MockDataReader) HeadSnapshot(ctx context.Context, tenantID string) (token.SnapToken, error) {
	// Track call count per tenant
	m.mu.Lock()
	counter, exists := m.headSnapshotCalls[tenantID]
	if !exists {
		counter = new(int64)
		m.headSnapshotCalls[tenantID] = counter
	}
	m.mu.Unlock()

	// Increment call count
	atomic.AddInt64(counter, 1)

	// Simulate some work
	time.Sleep(10 * time.Millisecond)

	return token.NoopToken{Value: "snapshot-" + tenantID}, nil
}

func GetCallCount(m *MockDataReader, tenantID string) int64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	if counter, exists := m.headSnapshotCalls[tenantID]; exists {
		return atomic.LoadInt64(counter)
	}
	return 0
}

var _ = Describe("Singleflight DataReader", func() {
	var (
		mockDelegate storage.DataReader
		reader       *DataReader
		ctx          context.Context
	)

	BeforeEach(func() {
		mockDelegate = NewMockDataReader()
		reader = NewDataReader(mockDelegate)
		ctx = context.Background()
	})

	Describe("HeadSnapshot", func() {
		It("should deduplicate concurrent requests for the same tenant", func() {
			tenantID := "tenant1"
			numConcurrentRequests := 10

			var wg sync.WaitGroup
			wg.Add(numConcurrentRequests)

			// Launch concurrent requests for the same tenant
			for i := 0; i < numConcurrentRequests; i++ {
				go func() {
					defer wg.Done()
					_, err := reader.HeadSnapshot(ctx, tenantID)
					Expect(err).ShouldNot(HaveOccurred())
				}()
			}

			wg.Wait()

			// Only 1 call should reach the delegate due to deduplication
			mock := mockDelegate.(*MockDataReader)
			callCount := GetCallCount(mock, tenantID)
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
					_, err := reader.HeadSnapshot(ctx, tenant1)
					Expect(err).ShouldNot(HaveOccurred())
				}()
			}

			// Launch concurrent requests for tenant2
			for i := 0; i < numConcurrentRequests; i++ {
				go func() {
					defer wg.Done()
					_, err := reader.HeadSnapshot(ctx, tenant2)
					Expect(err).ShouldNot(HaveOccurred())
				}()
			}

			wg.Wait()

			mock := mockDelegate.(*MockDataReader)

			// Each tenant should have exactly 1 call due to deduplication within the tenant
			Expect(GetCallCount(mock, tenant1)).To(Equal(int64(1)))
			Expect(GetCallCount(mock, tenant2)).To(Equal(int64(1)))
		})

		It("should return correct snapshot for each tenant", func() {
			tenant1 := "tenant1"
			tenant2 := "tenant2"

			var wg sync.WaitGroup
			wg.Add(2)

			var result1, result2 token.SnapToken

			go func() {
				defer wg.Done()
				var err error
				result1, err = reader.HeadSnapshot(ctx, tenant1)
				Expect(err).ShouldNot(HaveOccurred())
			}()

			go func() {
				defer wg.Done()
				var err error
				result2, err = reader.HeadSnapshot(ctx, tenant2)
				Expect(err).ShouldNot(HaveOccurred())
			}()

			wg.Wait()

			// Verify that each tenant gets its own snapshot
			Expect(result1.(token.NoopToken).Value).To(Equal("snapshot-" + tenant1))
			Expect(result2.(token.NoopToken).Value).To(Equal("snapshot-" + tenant2))
		})

		It("should not deduplicate sequential requests", func() {
			tenantID := "tenant1"

			// First request
			_, err := reader.HeadSnapshot(ctx, tenantID)
			Expect(err).ShouldNot(HaveOccurred())

			// Small delay to ensure first request completes
			time.Sleep(50 * time.Millisecond)

			// Second request (should trigger another call to delegate)
			_, err = reader.HeadSnapshot(ctx, tenantID)
			Expect(err).ShouldNot(HaveOccurred())

			mock := mockDelegate.(*MockDataReader)

			// Should have 2 calls to the delegate
			callCount := GetCallCount(mock, tenantID)
			Expect(callCount).To(Equal(int64(2)))
		})
	})
})
