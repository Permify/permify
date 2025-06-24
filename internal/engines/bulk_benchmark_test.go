package engines

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// MockChecker implements invoke.Check for benchmarking
type MockChecker struct {
	delay time.Duration
}

func (m *MockChecker) Check(ctx context.Context, request *base.PermissionCheckRequest) (*base.PermissionCheckResponse, error) {
	if m.delay > 0 {
		time.Sleep(m.delay)
	}
	return &base.PermissionCheckResponse{
		Can: base.CheckResult_CHECK_RESULT_ALLOWED,
	}, nil
}

// BenchmarkBulkChecker tests the performance of the BulkChecker
func BenchmarkBulkChecker(b *testing.B) {
	ctx := context.Background()
	checker := &MockChecker{delay: 1 * time.Millisecond}

	configs := []struct {
		name   string
		config BulkCheckerConfig
	}{
		{
			name:   "Default",
			config: DefaultBulkCheckerConfig(),
		},
		{
			name: "HighConcurrency",
			config: BulkCheckerConfig{
				ConcurrencyLimit: 50,
				BufferSize:       5000,
			},
		},
		{
			name: "SmallBatch",
			config: BulkCheckerConfig{
				ConcurrencyLimit: 10,
				BufferSize:       1000,
			},
		},
	}

	for _, cfg := range configs {
		b.Run(cfg.name, func(b *testing.B) {
			callback := func(id, ct string) {
				// Simulate callback work
			}

			b.ResetTimer()
			b.ReportAllocs()

			for b.Loop() {
				bc, err := NewBulkChecker(ctx, checker, BulkCheckerTypeEntity, callback, cfg.config)
				if err != nil {
					b.Fatal(err)
				}
				publisher := NewBulkEntityPublisher(ctx, &base.PermissionLookupEntityRequest{
					TenantId:   "test",
					Metadata:   &base.PermissionLookupEntityRequestMetadata{},
					EntityType: "document",
					Permission: "read",
					Subject:    &base.Subject{Id: "user1"},
				}, bc)

				// Publish 1000 requests
				for j := 0; j < 1000; j++ {
					publisher.Publish(
						&base.Entity{Id: fmt.Sprintf("doc%d", j)},
						&base.PermissionCheckRequestMetadata{},
						&base.Context{},
						base.CheckResult_CHECK_RESULT_UNSPECIFIED,
					)
				}

				// Execute requests
				err = bc.ExecuteRequests(100)
				if err != nil {
					b.Fatal(err)
				}
				bc.Close()
			}
		})
	}
}

// BenchmarkRequestCollection tests the performance of request collection
func BenchmarkRequestCollection(b *testing.B) {
	ctx := context.Background()
	checker := &MockChecker{}

	config := BulkCheckerConfig{
		ConcurrencyLimit: 10,
		BufferSize:       10000,
	}

	callback := func(id, ct string) {}

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		bc, err := NewBulkChecker(ctx, checker, BulkCheckerTypeEntity, callback, config)
		if err != nil {
			b.Fatal(err)
		}

		publisher := NewBulkEntityPublisher(ctx, &base.PermissionLookupEntityRequest{
			TenantId:   "test",
			Metadata:   &base.PermissionLookupEntityRequestMetadata{},
			EntityType: "document",
			Permission: "read",
			Subject:    &base.Subject{Id: "user1"},
		}, bc)

		// Publish requests rapidly
		for j := 0; j < 10000; j++ {
			publisher.Publish(
				&base.Entity{Id: fmt.Sprintf("doc%d", j)},
				&base.PermissionCheckRequestMetadata{},
				&base.Context{},
				base.CheckResult_CHECK_RESULT_UNSPECIFIED,
			)
		}

		bc.StopCollectingRequests()
		bc.Close()
	}
}

// BenchmarkSorting tests the performance of request sorting
func BenchmarkSorting(b *testing.B) {
	ctx := context.Background()
	checker := &MockChecker{}

	config := DefaultBulkCheckerConfig()
	callback := func(id, ct string) {}

	bc, err := NewBulkChecker(ctx, checker, BulkCheckerTypeEntity, callback, config)
	if err != nil {
		b.Fatal(err)
	}

	publisher := NewBulkEntityPublisher(ctx, &base.PermissionLookupEntityRequest{
		TenantId:   "test",
		Metadata:   &base.PermissionLookupEntityRequestMetadata{},
		EntityType: "document",
		Permission: "read",
		Subject:    &base.Subject{Id: "user1"},
	}, bc)

	// Pre-populate with requests
	for j := 0; j < 10000; j++ {
		publisher.Publish(
			&base.Entity{Id: fmt.Sprintf("doc%d", 10000-j)}, // Reverse order for worst case
			&base.PermissionCheckRequestMetadata{},
			&base.Context{},
			base.CheckResult_CHECK_RESULT_UNSPECIFIED,
		)
	}

	bc.StopCollectingRequests()

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		_ = bc.getSortedRequests()
	}

	bc.Close()
}

// BenchmarkConcurrentPublishing tests concurrent publishing performance
func BenchmarkConcurrentPublishing(b *testing.B) {
	ctx := context.Background()
	checker := &MockChecker{}

	config := BulkCheckerConfig{
		ConcurrencyLimit: 50,
		BufferSize:       50000,
	}

	callback := func(id, ct string) {}

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		bc, err := NewBulkChecker(ctx, checker, BulkCheckerTypeEntity, callback, config)
		if err != nil {
			b.Fatal(err)
		}

		publisher := NewBulkEntityPublisher(ctx, &base.PermissionLookupEntityRequest{
			TenantId:   "test",
			Metadata:   &base.PermissionLookupEntityRequestMetadata{},
			EntityType: "document",
			Permission: "read",
			Subject:    &base.Subject{Id: "user1"},
		}, bc)

		// Publish requests concurrently
		var wg sync.WaitGroup
		for g := 0; g < 10; g++ { // 10 goroutines
			wg.Add(1)
			go func(goroutineID int) {
				defer wg.Done()
				for j := 0; j < 1000; j++ {
					publisher.Publish(
						&base.Entity{Id: fmt.Sprintf("doc%d_%d", goroutineID, j)},
						&base.PermissionCheckRequestMetadata{},
						&base.Context{},
						base.CheckResult_CHECK_RESULT_UNSPECIFIED,
					)
				}
			}(g)
		}

		wg.Wait()
		bc.StopCollectingRequests()
		bc.Close()
	}
}

// BenchmarkHighLoadRequestCollection tests the performance under high load
func BenchmarkHighLoadRequestCollection(b *testing.B) {
	ctx := context.Background()
	checker := &MockChecker{}

	config := BulkCheckerConfig{
		ConcurrencyLimit: 100,
		BufferSize:       100000,
	}

	callback := func(id, ct string) {}

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		bc, err := NewBulkChecker(ctx, checker, BulkCheckerTypeEntity, callback, config)
		if err != nil {
			b.Fatal(err)
		}

		publisher := NewBulkEntityPublisher(ctx, &base.PermissionLookupEntityRequest{
			TenantId:   "test",
			Metadata:   &base.PermissionLookupEntityRequestMetadata{},
			EntityType: "document",
			Permission: "read",
			Subject:    &base.Subject{Id: "user1"},
		}, bc)

		// Publish 100,000 requests rapidly
		for j := 0; j < 100000; j++ {
			publisher.Publish(
				&base.Entity{Id: fmt.Sprintf("doc%d", j)},
				&base.PermissionCheckRequestMetadata{},
				&base.Context{},
				base.CheckResult_CHECK_RESULT_UNSPECIFIED,
			)
		}

		bc.StopCollectingRequests()
		bc.Close()
	}
}

// BenchmarkExtremeConcurrency tests with extreme concurrency to show batch processing benefits
func BenchmarkExtremeConcurrency(b *testing.B) {
	ctx := context.Background()
	checker := &MockChecker{}

	config := BulkCheckerConfig{
		ConcurrencyLimit: 200,
		BufferSize:       200000,
	}

	callback := func(id, ct string) {}

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		bc, err := NewBulkChecker(ctx, checker, BulkCheckerTypeEntity, callback, config)
		if err != nil {
			b.Fatal(err)
		}

		publisher := NewBulkEntityPublisher(ctx, &base.PermissionLookupEntityRequest{
			TenantId:   "test",
			Metadata:   &base.PermissionLookupEntityRequestMetadata{},
			EntityType: "document",
			Permission: "read",
			Subject:    &base.Subject{Id: "user1"},
		}, bc)

		// Publish 50,000 requests with extreme concurrency (50 goroutines)
		var wg sync.WaitGroup
		for g := 0; g < 50; g++ {
			wg.Add(1)
			go func(goroutineID int) {
				defer wg.Done()
				for j := 0; j < 1000; j++ {
					publisher.Publish(
						&base.Entity{Id: fmt.Sprintf("doc%d_%d", goroutineID, j)},
						&base.PermissionCheckRequestMetadata{},
						&base.Context{},
						base.CheckResult_CHECK_RESULT_UNSPECIFIED,
					)
				}
			}(g)
		}

		wg.Wait()
		bc.StopCollectingRequests()
		bc.Close()
	}
}

// BenchmarkBurstRequests tests handling of burst requests
func BenchmarkBurstRequests(b *testing.B) {
	ctx := context.Background()
	checker := &MockChecker{}

	config := BulkCheckerConfig{
		ConcurrencyLimit: 50,
		BufferSize:       50000,
	}

	callback := func(id, ct string) {}

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		bc, err := NewBulkChecker(ctx, checker, BulkCheckerTypeEntity, callback, config)
		if err != nil {
			b.Fatal(err)
		}

		publisher := NewBulkEntityPublisher(ctx, &base.PermissionLookupEntityRequest{
			TenantId:   "test",
			Metadata:   &base.PermissionLookupEntityRequestMetadata{},
			EntityType: "document",
			Permission: "read",
			Subject:    &base.Subject{Id: "user1"},
		}, bc)

		// Simulate burst requests - publish in waves
		for wave := 0; wave < 10; wave++ {
			var wg sync.WaitGroup
			for g := 0; g < 20; g++ { // 20 goroutines per wave
				wg.Add(1)
				go func(goroutineID, waveID int) {
					defer wg.Done()
					for j := 0; j < 500; j++ {
						publisher.Publish(
							&base.Entity{Id: fmt.Sprintf("doc%d_%d_%d", waveID, goroutineID, j)},
							&base.PermissionCheckRequestMetadata{},
							&base.Context{},
							base.CheckResult_CHECK_RESULT_UNSPECIFIED,
						)
					}
				}(g, wave)
			}
			wg.Wait()
		}

		bc.StopCollectingRequests()
		bc.Close()
	}
}

// BenchmarkMixedWorkload tests mixed workload with different request patterns
func BenchmarkMixedWorkload(b *testing.B) {
	ctx := context.Background()
	checker := &MockChecker{}

	config := BulkCheckerConfig{
		ConcurrencyLimit: 75,
		BufferSize:       75000,
	}

	callback := func(id, ct string) {}

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		bc, err := NewBulkChecker(ctx, checker, BulkCheckerTypeEntity, callback, config)
		if err != nil {
			b.Fatal(err)
		}

		publisher := NewBulkEntityPublisher(ctx, &base.PermissionLookupEntityRequest{
			TenantId:   "test",
			Metadata:   &base.PermissionLookupEntityRequestMetadata{},
			EntityType: "document",
			Permission: "read",
			Subject:    &base.Subject{Id: "user1"},
		}, bc)

		// Mixed workload: some sequential, some concurrent, some burst
		var wg sync.WaitGroup

		// Sequential requests
		for j := 0; j < 10000; j++ {
			publisher.Publish(
				&base.Entity{Id: fmt.Sprintf("seq_%d", j)},
				&base.PermissionCheckRequestMetadata{},
				&base.Context{},
				base.CheckResult_CHECK_RESULT_UNSPECIFIED,
			)
		}

		// Concurrent requests
		for g := 0; g < 25; g++ {
			wg.Add(1)
			go func(goroutineID int) {
				defer wg.Done()
				for j := 0; j < 2000; j++ {
					publisher.Publish(
						&base.Entity{Id: fmt.Sprintf("concurrent_%d_%d", goroutineID, j)},
						&base.PermissionCheckRequestMetadata{},
						&base.Context{},
						base.CheckResult_CHECK_RESULT_UNSPECIFIED,
					)
				}
			}(g)
		}

		// Burst requests
		for burst := 0; burst < 5; burst++ {
			for g := 0; g < 10; g++ {
				wg.Add(1)
				go func(goroutineID, burstID int) {
					defer wg.Done()
					for j := 0; j < 1000; j++ {
						publisher.Publish(
							&base.Entity{Id: fmt.Sprintf("burst_%d_%d_%d", burstID, goroutineID, j)},
							&base.PermissionCheckRequestMetadata{},
							&base.Context{},
							base.CheckResult_CHECK_RESULT_UNSPECIFIED,
						)
					}
				}(g, burst)
			}
		}

		wg.Wait()
		bc.StopCollectingRequests()
		bc.Close()
	}
}
