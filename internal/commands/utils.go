package commands

import (
	"sync"
)

const (
	_defaultConcurrencyLimit = 100
)

// CheckOption - Option type
type CheckOption func(*CheckCommand)

// ConcurrencyLimit - Defines concurrency limit
func ConcurrencyLimit(limit int) CheckOption {
	return func(c *CheckCommand) {
		c.concurrencyLimit = limit
	}
}

// CheckMetadata - Metadata for check command
type CheckMetadata struct {
	mu        sync.Mutex
	callCount int32
}

// NewCheckMetadata it creates a new instance of CheckMetadata
func NewCheckMetadata() *CheckMetadata {
	return &CheckMetadata{
		callCount: 0,
	}
}

// AddCall -
func (r *CheckMetadata) AddCall() int32 {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.callCount++
	return r.callCount
}

// GetCallCount -
func (r *CheckMetadata) GetCallCount() int32 {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.callCount
}
