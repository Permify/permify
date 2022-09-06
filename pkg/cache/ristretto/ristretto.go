package ristretto

import (
	"github.com/dgraph-io/ristretto"
)

// Ristretto -
type Ristretto struct {
	numCounters int64
	maxCost     int64
	bufferItems int64

	*ristretto.Cache
}

// New -
func New(opts ...Option) (*Ristretto, error) {
	rs := &Ristretto{
		numCounters: _defaultNumCounters,
		maxCost:     _defaultMaxCost,
		bufferItems: _defaultBufferItems,
	}

	// Custom options
	for _, opt := range opts {
		opt(rs)
	}

	c, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: rs.numCounters,
		MaxCost:     rs.maxCost,
		BufferItems: rs.bufferItems,
	})

	rs.Cache = c

	return rs, err
}
