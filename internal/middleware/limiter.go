package middleware

import (
	"fmt"
	"time"

	"golang.org/x/net/context"

	"github.com/juju/ratelimit"
)

// RateLimiter struct is a wrapper around the juju Bucket struct
type RateLimiter struct {
	bucket *ratelimit.Bucket // bucket is the token bucket that forms the core of the rate limiter
}

// NewRateLimiter is a constructor function for RateLimiter.
// It creates a new RateLimiter that allows reqPerSec requests per second.
func NewRateLimiter(reqPerSec int64) *RateLimiter {
	// fillInterval is the amount of time between adding new tokens to the bucket.
	// We want to add a new token reqPerSec times per second, so fillInterval is the inverse of reqPerSec.
	fillInterval := time.Second / time.Duration(reqPerSec)

	// Create a new token bucket with a rate of reqPerSec tokens per second and a capacity of reqPerSec.
	bucket := ratelimit.NewBucket(fillInterval, reqPerSec)

	return &RateLimiter{
		bucket: bucket,
	}
}

// Limit checks if a request should be allowed based on the current state of the bucket.
// If no tokens are available (i.e., if TakeAvailable(1) returns 0), it means the rate limit has been hit,
// so it returns true. If a token is available, it returns false, meaning the request can proceed.
func (l *RateLimiter) Limit(_ context.Context) error {
	tokenRes := l.bucket.TakeAvailable(1)

	// When rate limit reached, return specific error for the clients.
	if tokenRes == 0 {
		return fmt.Errorf("reached Rate-Limiting %d", l.bucket.Available())
	}

	// Rate limit isn't reached.
	return nil
}
