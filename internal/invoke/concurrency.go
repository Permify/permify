package invoke

import (
	"context"
	"math"

	"golang.org/x/sync/semaphore"
)

// DefaultConcurrencyLimit is the default max concurrent DB operations per request.
const DefaultConcurrencyLimit = 100

type concurrencySemaphoreKeyType struct{}

var concurrencySemaphoreKey = concurrencySemaphoreKeyType{}

// noopSemaphore is a semaphore that never blocks — used as fallback
// when no request-scoped semaphore was set (e.g. in tests or dev mode).
var noopSemaphore = semaphore.NewWeighted(math.MaxInt64)

// WithConcurrencySemaphore stores a request-scoped semaphore in the context.
func WithConcurrencySemaphore(ctx context.Context, sem *semaphore.Weighted) context.Context {
	return context.WithValue(ctx, concurrencySemaphoreKey, sem)
}

// ConcurrencySemaphoreFromContext retrieves the request-scoped semaphore.
// Returns a no-op semaphore if none was set — never returns nil.
func ConcurrencySemaphoreFromContext(ctx context.Context) *semaphore.Weighted {
	if sem, ok := ctx.Value(concurrencySemaphoreKey).(*semaphore.Weighted); ok && sem != nil {
		return sem
	}
	return noopSemaphore
}
