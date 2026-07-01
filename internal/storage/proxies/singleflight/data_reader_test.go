package singleflight

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"google.golang.org/protobuf/types/known/anypb"

	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/pkg/database"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/token"
)

// MockDataReader is a mock implementation of storage.DataReader for testing
type MockDataReader struct {
	storage.NoopDataReader
	headSnapshotCalls         map[string]*int64
	queryRelationshipsCalls   map[string]*int64
	querySingleAttributeCalls map[string]*int64
	queryAttributesCalls      map[string]*int64
	mu                        sync.Mutex
}

func NewMockDataReader() *MockDataReader {
	return &MockDataReader{
		headSnapshotCalls:         make(map[string]*int64),
		queryRelationshipsCalls:   make(map[string]*int64),
		querySingleAttributeCalls: make(map[string]*int64),
		queryAttributesCalls:      make(map[string]*int64),
	}
}

func (m *MockDataReader) HeadSnapshot(_ context.Context, tenantID string) (token.SnapToken, error) {
	m.incrementCounter(m.headSnapshotCalls, tenantID)
	time.Sleep(10 * time.Millisecond)
	return token.NoopToken{Value: "snapshot-" + tenantID}, nil
}

func (m *MockDataReader) QueryRelationships(_ context.Context, tenantID string, filter *base.TupleFilter, _ string, _ database.CursorPagination) (*database.TupleIterator, error) {
	key := tenantID + "|" + filter.GetEntity().GetType() + "|" + filter.GetRelation()
	m.incrementCounter(m.queryRelationshipsCalls, key)
	time.Sleep(10 * time.Millisecond)
	return database.NewTupleIterator(
		&base.Tuple{
			Entity:   &base.Entity{Type: filter.GetEntity().GetType(), Id: "1"},
			Relation: filter.GetRelation(),
			Subject:  &base.Subject{Type: "user", Id: "alice"},
		},
		&base.Tuple{
			Entity:   &base.Entity{Type: filter.GetEntity().GetType(), Id: "1"},
			Relation: filter.GetRelation(),
			Subject:  &base.Subject{Type: "user", Id: "bob"},
		},
	), nil
}

func (m *MockDataReader) QuerySingleAttribute(_ context.Context, tenantID string, filter *base.AttributeFilter, _ string) (*base.Attribute, error) {
	key := tenantID + "|" + filter.GetEntity().GetType()
	m.incrementCounter(m.querySingleAttributeCalls, key)
	time.Sleep(10 * time.Millisecond)
	val, _ := anypb.New(&base.BooleanValue{Data: true})
	return &base.Attribute{
		Entity:    &base.Entity{Type: filter.GetEntity().GetType(), Id: "1"},
		Attribute: "is_public",
		Value:     val,
	}, nil
}

func (m *MockDataReader) QueryAttributes(_ context.Context, tenantID string, filter *base.AttributeFilter, _ string, _ database.CursorPagination) (*database.AttributeIterator, error) {
	key := tenantID + "|" + filter.GetEntity().GetType()
	m.incrementCounter(m.queryAttributesCalls, key)
	time.Sleep(10 * time.Millisecond)
	val, _ := anypb.New(&base.BooleanValue{Data: true})
	return database.NewAttributeIterator(
		&base.Attribute{
			Entity:    &base.Entity{Type: filter.GetEntity().GetType(), Id: "1"},
			Attribute: "is_public",
			Value:     val,
		},
	), nil
}

func (m *MockDataReader) incrementCounter(counters map[string]*int64, key string) {
	m.mu.Lock()
	counter, exists := counters[key]
	if !exists {
		counter = new(int64)
		counters[key] = counter
	}
	m.mu.Unlock()
	atomic.AddInt64(counter, 1)
}

func getCallCount(counters map[string]*int64, mu *sync.Mutex, key string) int64 {
	mu.Lock()
	defer mu.Unlock()
	if counter, exists := counters[key]; exists {
		return atomic.LoadInt64(counter)
	}
	return 0
}

// GetCallCount returns the HeadSnapshot call count for a tenant (for backward compat)
func GetCallCount(m *MockDataReader, tenantID string) int64 {
	return getCallCount(m.headSnapshotCalls, &m.mu, tenantID)
}

// ErrorMockDataReader is a mock that returns errors for testing error handling
type ErrorMockDataReader struct {
	storage.NoopDataReader
}

func (m *ErrorMockDataReader) HeadSnapshot(_ context.Context, _ string) (token.SnapToken, error) {
	return nil, errors.New("delegate error")
}

func (m *ErrorMockDataReader) QueryRelationships(_ context.Context, _ string, _ *base.TupleFilter, _ string, _ database.CursorPagination) (*database.TupleIterator, error) {
	return nil, errors.New("delegate error")
}

func (m *ErrorMockDataReader) QuerySingleAttribute(_ context.Context, _ string, _ *base.AttributeFilter, _ string) (*base.Attribute, error) {
	return nil, errors.New("delegate error")
}

func (m *ErrorMockDataReader) QueryAttributes(_ context.Context, _ string, _ *base.AttributeFilter, _ string, _ database.CursorPagination) (*database.AttributeIterator, error) {
	return nil, errors.New("delegate error")
}

var _ = Describe("Singleflight DataReader", func() {
	var (
		mockDelegate *MockDataReader
		reader       *DataReader
		ctx          context.Context
	)

	BeforeEach(func() {
		mockDelegate = NewMockDataReader()
		reader = NewDataReader(mockDelegate)
		ctx = context.Background()
	})

	Describe("NewDataReader", func() {
		It("should create a new DataReader with delegate", func() {
			delegate := storage.NewNoopRelationshipReader()
			reader := NewDataReader(delegate)

			Expect(reader).ShouldNot(BeNil())
			Expect(reader.delegate).Should(Equal(delegate))
		})

		It("should create a new DataReader with nil delegate", func() {
			reader := NewDataReader(nil)

			Expect(reader).ShouldNot(BeNil())
			Expect(reader.delegate).Should(BeNil())
		})
	})

	Describe("ReadRelationships", func() {
		It("should delegate to underlying DataReader", func() {
			delegate := storage.NewNoopRelationshipReader()
			reader := NewDataReader(delegate)

			filter := &base.TupleFilter{}
			collection, ct, err := reader.ReadRelationships(ctx, "tenant1", filter, "token", database.Pagination{})

			Expect(err).ShouldNot(HaveOccurred())
			Expect(collection).ShouldNot(BeNil())
			Expect(ct).ShouldNot(BeNil())
		})
	})

	Describe("ReadAttributes", func() {
		It("should delegate to underlying DataReader", func() {
			delegate := storage.NewNoopRelationshipReader()
			reader := NewDataReader(delegate)

			filter := &base.AttributeFilter{}
			collection, ct, err := reader.ReadAttributes(ctx, "tenant1", filter, "token", database.Pagination{})

			Expect(err).ShouldNot(HaveOccurred())
			Expect(collection).ShouldNot(BeNil())
			Expect(ct).ShouldNot(BeNil())
		})
	})

	Describe("QueryUniqueSubjectReferences", func() {
		It("should delegate to underlying DataReader", func() {
			delegate := storage.NewNoopRelationshipReader()
			reader := NewDataReader(delegate)

			subjectRef := &base.RelationReference{
				Type:     "user",
				Relation: "member",
			}
			ids, ct, err := reader.QueryUniqueSubjectReferences(ctx, "tenant1", subjectRef, []string{}, "token", database.Pagination{})

			Expect(err).ShouldNot(HaveOccurred())
			Expect(ids).ShouldNot(BeNil())
			Expect(ct).ShouldNot(BeNil())
		})
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
			callCount := GetCallCount(mockDelegate, tenantID)
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

			// Each tenant should have exactly 1 call due to deduplication within the tenant
			Expect(GetCallCount(mockDelegate, tenant1)).To(Equal(int64(1)))
			Expect(GetCallCount(mockDelegate, tenant2)).To(Equal(int64(1)))
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

			// Should have 2 calls to the delegate
			callCount := GetCallCount(mockDelegate, tenantID)
			Expect(callCount).To(Equal(int64(2)))
		})

		It("should propagate errors from delegate", func() {
			errorDelegate := &ErrorMockDataReader{}
			errorReader := NewDataReader(errorDelegate)

			_, err := errorReader.HeadSnapshot(ctx, "tenant1")

			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(Equal("delegate error"))
		})

		It("should handle context cancellation", func() {
			cancelledCtx, cancel := context.WithCancel(ctx)
			cancel() // Cancel immediately

			_, err := reader.HeadSnapshot(cancelledCtx, "tenant1")

			// Context cancellation behavior depends on singleflight implementation
			// We just verify it doesn't panic
			_ = err
		})

		It("should handle concurrent requests with errors", func() {
			errorDelegate := &ErrorMockDataReader{}
			errorReader := NewDataReader(errorDelegate)
			tenantID := "tenant1"
			numConcurrentRequests := 5

			var wg sync.WaitGroup
			wg.Add(numConcurrentRequests)

			errorCount := int64(0)

			// Launch concurrent requests that will all fail
			for i := 0; i < numConcurrentRequests; i++ {
				go func() {
					defer wg.Done()
					_, err := errorReader.HeadSnapshot(ctx, tenantID)
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

	Describe("QueryRelationships", func() {
		filter := &base.TupleFilter{
			Entity: &base.EntityFilter{
				Type: "document",
				Ids:  []string{"1"},
			},
			Relation: "viewer",
		}
		snap := "snap-token-1"
		mockKey := "tenant1|document|viewer"

		It("should deduplicate concurrent requests with the same parameters", func() {
			numConcurrentRequests := 10

			var wg sync.WaitGroup
			wg.Add(numConcurrentRequests)

			for i := 0; i < numConcurrentRequests; i++ {
				go func() {
					defer wg.Done()
					it, err := reader.QueryRelationships(ctx, "tenant1", filter, snap, database.NewCursorPagination())
					Expect(err).ShouldNot(HaveOccurred())
					Expect(it).ShouldNot(BeNil())
				}()
			}

			wg.Wait()

			callCount := getCallCount(mockDelegate.queryRelationshipsCalls, &mockDelegate.mu, mockKey)
			Expect(callCount).To(Equal(int64(1)))
		})

		It("should return independent iterators to each caller", func() {
			numConcurrentRequests := 5

			var wg sync.WaitGroup
			wg.Add(numConcurrentRequests)

			iterators := make([]*database.TupleIterator, numConcurrentRequests)
			var mu sync.Mutex

			for i := 0; i < numConcurrentRequests; i++ {
				idx := i
				go func() {
					defer wg.Done()
					it, err := reader.QueryRelationships(ctx, "tenant1", filter, snap, database.NewCursorPagination())
					Expect(err).ShouldNot(HaveOccurred())
					mu.Lock()
					iterators[idx] = it
					mu.Unlock()
				}()
			}

			wg.Wait()

			// Each iterator should independently yield all tuples
			for i, it := range iterators {
				Expect(it.HasNext()).To(BeTrue(), "iterator %d should have tuples", i)
				t1 := it.GetNext()
				Expect(t1).ShouldNot(BeNil())
				Expect(t1.GetSubject().GetId()).To(Equal("alice"))
				t2 := it.GetNext()
				Expect(t2).ShouldNot(BeNil())
				Expect(t2.GetSubject().GetId()).To(Equal("bob"))
				Expect(it.HasNext()).To(BeFalse(), "iterator %d should be exhausted", i)
			}
		})

		It("should isolate requests for different tenants", func() {
			numConcurrentRequests := 5

			var wg sync.WaitGroup
			wg.Add(numConcurrentRequests * 2)

			for i := 0; i < numConcurrentRequests; i++ {
				go func() {
					defer wg.Done()
					_, err := reader.QueryRelationships(ctx, "tenant1", filter, snap, database.NewCursorPagination())
					Expect(err).ShouldNot(HaveOccurred())
				}()
			}

			for i := 0; i < numConcurrentRequests; i++ {
				go func() {
					defer wg.Done()
					_, err := reader.QueryRelationships(ctx, "tenant2", filter, snap, database.NewCursorPagination())
					Expect(err).ShouldNot(HaveOccurred())
				}()
			}

			wg.Wait()

			Expect(getCallCount(mockDelegate.queryRelationshipsCalls, &mockDelegate.mu, "tenant1|document|viewer")).To(Equal(int64(1)))
			Expect(getCallCount(mockDelegate.queryRelationshipsCalls, &mockDelegate.mu, "tenant2|document|viewer")).To(Equal(int64(1)))
		})

		It("should isolate requests with different filters", func() {
			filter2 := &base.TupleFilter{
				Entity: &base.EntityFilter{
					Type: "document",
					Ids:  []string{"1"},
				},
				Relation: "editor",
			}

			var wg sync.WaitGroup
			wg.Add(2)

			go func() {
				defer wg.Done()
				_, err := reader.QueryRelationships(ctx, "tenant1", filter, snap, database.NewCursorPagination())
				Expect(err).ShouldNot(HaveOccurred())
			}()

			go func() {
				defer wg.Done()
				_, err := reader.QueryRelationships(ctx, "tenant1", filter2, snap, database.NewCursorPagination())
				Expect(err).ShouldNot(HaveOccurred())
			}()

			wg.Wait()

			Expect(getCallCount(mockDelegate.queryRelationshipsCalls, &mockDelegate.mu, "tenant1|document|viewer")).To(Equal(int64(1)))
			Expect(getCallCount(mockDelegate.queryRelationshipsCalls, &mockDelegate.mu, "tenant1|document|editor")).To(Equal(int64(1)))
		})

		It("should not deduplicate sequential requests", func() {
			_, err := reader.QueryRelationships(ctx, "tenant1", filter, snap, database.NewCursorPagination())
			Expect(err).ShouldNot(HaveOccurred())

			time.Sleep(50 * time.Millisecond)

			_, err = reader.QueryRelationships(ctx, "tenant1", filter, snap, database.NewCursorPagination())
			Expect(err).ShouldNot(HaveOccurred())

			callCount := getCallCount(mockDelegate.queryRelationshipsCalls, &mockDelegate.mu, mockKey)
			Expect(callCount).To(Equal(int64(2)))
		})

		It("should propagate errors from delegate", func() {
			errorReader := NewDataReader(&ErrorMockDataReader{})

			_, err := errorReader.QueryRelationships(ctx, "tenant1", filter, snap, database.NewCursorPagination())

			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(Equal("delegate error"))
		})

		It("should handle concurrent requests with errors", func() {
			errorReader := NewDataReader(&ErrorMockDataReader{})
			numConcurrentRequests := 5

			var wg sync.WaitGroup
			wg.Add(numConcurrentRequests)

			errorCount := int64(0)

			for i := 0; i < numConcurrentRequests; i++ {
				go func() {
					defer wg.Done()
					_, err := errorReader.QueryRelationships(ctx, "tenant1", filter, snap, database.NewCursorPagination())
					if err != nil {
						atomic.AddInt64(&errorCount, 1)
					}
				}()
			}

			wg.Wait()

			Expect(atomic.LoadInt64(&errorCount)).To(Equal(int64(numConcurrentRequests)))
		})

		It("should handle empty result sets", func() {
			emptyReader := NewDataReader(storage.NewNoopRelationshipReader())

			it, err := emptyReader.QueryRelationships(ctx, "tenant1", filter, snap, database.NewCursorPagination())

			Expect(err).ShouldNot(HaveOccurred())
			Expect(it).ShouldNot(BeNil())
			Expect(it.HasNext()).To(BeFalse())
		})
	})

	Describe("QuerySingleAttribute", func() {
		filter := &base.AttributeFilter{
			Entity: &base.EntityFilter{
				Type: "document",
				Ids:  []string{"1"},
			},
			Attributes: []string{"is_public"},
		}
		snap := "snap-token-1"
		mockKey := "tenant1|document"

		It("should deduplicate concurrent requests with the same parameters", func() {
			numConcurrentRequests := 10

			var wg sync.WaitGroup
			wg.Add(numConcurrentRequests)

			for i := 0; i < numConcurrentRequests; i++ {
				go func() {
					defer wg.Done()
					attr, err := reader.QuerySingleAttribute(ctx, "tenant1", filter, snap)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(attr).ShouldNot(BeNil())
				}()
			}

			wg.Wait()

			callCount := getCallCount(mockDelegate.querySingleAttributeCalls, &mockDelegate.mu, mockKey)
			Expect(callCount).To(Equal(int64(1)))
		})

		It("should isolate requests for different tenants", func() {
			numConcurrentRequests := 5

			var wg sync.WaitGroup
			wg.Add(numConcurrentRequests * 2)

			for i := 0; i < numConcurrentRequests; i++ {
				go func() {
					defer wg.Done()
					_, err := reader.QuerySingleAttribute(ctx, "tenant1", filter, snap)
					Expect(err).ShouldNot(HaveOccurred())
				}()
			}

			for i := 0; i < numConcurrentRequests; i++ {
				go func() {
					defer wg.Done()
					_, err := reader.QuerySingleAttribute(ctx, "tenant2", filter, snap)
					Expect(err).ShouldNot(HaveOccurred())
				}()
			}

			wg.Wait()

			Expect(getCallCount(mockDelegate.querySingleAttributeCalls, &mockDelegate.mu, "tenant1|document")).To(Equal(int64(1)))
			Expect(getCallCount(mockDelegate.querySingleAttributeCalls, &mockDelegate.mu, "tenant2|document")).To(Equal(int64(1)))
		})

		It("should not deduplicate sequential requests", func() {
			_, err := reader.QuerySingleAttribute(ctx, "tenant1", filter, snap)
			Expect(err).ShouldNot(HaveOccurred())

			time.Sleep(50 * time.Millisecond)

			_, err = reader.QuerySingleAttribute(ctx, "tenant1", filter, snap)
			Expect(err).ShouldNot(HaveOccurred())

			callCount := getCallCount(mockDelegate.querySingleAttributeCalls, &mockDelegate.mu, mockKey)
			Expect(callCount).To(Equal(int64(2)))
		})

		It("should propagate errors from delegate", func() {
			errorReader := NewDataReader(&ErrorMockDataReader{})

			_, err := errorReader.QuerySingleAttribute(ctx, "tenant1", filter, snap)

			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(Equal("delegate error"))
		})

		It("should handle concurrent requests with errors", func() {
			errorReader := NewDataReader(&ErrorMockDataReader{})
			numConcurrentRequests := 5

			var wg sync.WaitGroup
			wg.Add(numConcurrentRequests)

			errorCount := int64(0)

			for i := 0; i < numConcurrentRequests; i++ {
				go func() {
					defer wg.Done()
					_, err := errorReader.QuerySingleAttribute(ctx, "tenant1", filter, snap)
					if err != nil {
						atomic.AddInt64(&errorCount, 1)
					}
				}()
			}

			wg.Wait()

			Expect(atomic.LoadInt64(&errorCount)).To(Equal(int64(numConcurrentRequests)))
		})

		It("should handle nil result", func() {
			emptyReader := NewDataReader(storage.NewNoopRelationshipReader())

			// NoopDataReader returns an empty Attribute, not nil
			attr, err := emptyReader.QuerySingleAttribute(ctx, "tenant1", filter, snap)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(attr).ShouldNot(BeNil())
		})
	})

	Describe("QueryAttributes", func() {
		filter := &base.AttributeFilter{
			Entity: &base.EntityFilter{
				Type: "document",
				Ids:  []string{"1"},
			},
			Attributes: []string{"is_public"},
		}
		snap := "snap-token-1"
		mockKey := "tenant1|document"

		It("should deduplicate concurrent requests with the same parameters", func() {
			numConcurrentRequests := 10

			var wg sync.WaitGroup
			wg.Add(numConcurrentRequests)

			for i := 0; i < numConcurrentRequests; i++ {
				go func() {
					defer wg.Done()
					it, err := reader.QueryAttributes(ctx, "tenant1", filter, snap, database.NewCursorPagination())
					Expect(err).ShouldNot(HaveOccurred())
					Expect(it).ShouldNot(BeNil())
				}()
			}

			wg.Wait()

			callCount := getCallCount(mockDelegate.queryAttributesCalls, &mockDelegate.mu, mockKey)
			Expect(callCount).To(Equal(int64(1)))
		})

		It("should return independent iterators to each caller", func() {
			numConcurrentRequests := 5

			var wg sync.WaitGroup
			wg.Add(numConcurrentRequests)

			iterators := make([]*database.AttributeIterator, numConcurrentRequests)
			var mu sync.Mutex

			for i := 0; i < numConcurrentRequests; i++ {
				idx := i
				go func() {
					defer wg.Done()
					it, err := reader.QueryAttributes(ctx, "tenant1", filter, snap, database.NewCursorPagination())
					Expect(err).ShouldNot(HaveOccurred())
					mu.Lock()
					iterators[idx] = it
					mu.Unlock()
				}()
			}

			wg.Wait()

			// Each iterator should independently yield all attributes
			for i, it := range iterators {
				Expect(it.HasNext()).To(BeTrue(), "iterator %d should have attributes", i)
				a := it.GetNext()
				Expect(a).ShouldNot(BeNil())
				Expect(a.GetAttribute()).To(Equal("is_public"))
				Expect(it.HasNext()).To(BeFalse(), "iterator %d should be exhausted", i)
			}
		})

		It("should isolate requests for different tenants", func() {
			numConcurrentRequests := 5

			var wg sync.WaitGroup
			wg.Add(numConcurrentRequests * 2)

			for i := 0; i < numConcurrentRequests; i++ {
				go func() {
					defer wg.Done()
					_, err := reader.QueryAttributes(ctx, "tenant1", filter, snap, database.NewCursorPagination())
					Expect(err).ShouldNot(HaveOccurred())
				}()
			}

			for i := 0; i < numConcurrentRequests; i++ {
				go func() {
					defer wg.Done()
					_, err := reader.QueryAttributes(ctx, "tenant2", filter, snap, database.NewCursorPagination())
					Expect(err).ShouldNot(HaveOccurred())
				}()
			}

			wg.Wait()

			Expect(getCallCount(mockDelegate.queryAttributesCalls, &mockDelegate.mu, "tenant1|document")).To(Equal(int64(1)))
			Expect(getCallCount(mockDelegate.queryAttributesCalls, &mockDelegate.mu, "tenant2|document")).To(Equal(int64(1)))
		})

		It("should not deduplicate sequential requests", func() {
			_, err := reader.QueryAttributes(ctx, "tenant1", filter, snap, database.NewCursorPagination())
			Expect(err).ShouldNot(HaveOccurred())

			time.Sleep(50 * time.Millisecond)

			_, err = reader.QueryAttributes(ctx, "tenant1", filter, snap, database.NewCursorPagination())
			Expect(err).ShouldNot(HaveOccurred())

			callCount := getCallCount(mockDelegate.queryAttributesCalls, &mockDelegate.mu, mockKey)
			Expect(callCount).To(Equal(int64(2)))
		})

		It("should propagate errors from delegate", func() {
			errorReader := NewDataReader(&ErrorMockDataReader{})

			_, err := errorReader.QueryAttributes(ctx, "tenant1", filter, snap, database.NewCursorPagination())

			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(Equal("delegate error"))
		})

		It("should handle concurrent requests with errors", func() {
			errorReader := NewDataReader(&ErrorMockDataReader{})
			numConcurrentRequests := 5

			var wg sync.WaitGroup
			wg.Add(numConcurrentRequests)

			errorCount := int64(0)

			for i := 0; i < numConcurrentRequests; i++ {
				go func() {
					defer wg.Done()
					_, err := errorReader.QueryAttributes(ctx, "tenant1", filter, snap, database.NewCursorPagination())
					if err != nil {
						atomic.AddInt64(&errorCount, 1)
					}
				}()
			}

			wg.Wait()

			Expect(atomic.LoadInt64(&errorCount)).To(Equal(int64(numConcurrentRequests)))
		})

		It("should handle empty result sets", func() {
			emptyReader := NewDataReader(storage.NewNoopRelationshipReader())

			it, err := emptyReader.QueryAttributes(ctx, "tenant1", filter, snap, database.NewCursorPagination())

			Expect(err).ShouldNot(HaveOccurred())
			Expect(it).ShouldNot(BeNil())
			Expect(it.HasNext()).To(BeFalse())
		})
	})
})
