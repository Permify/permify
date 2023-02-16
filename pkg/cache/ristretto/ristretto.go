package ristretto

import (
	"github.com/dgraph-io/ristretto"
	"github.com/dustin/go-humanize"
)

// Ristretto - Structure for Ristretto
type Ristretto struct {
	numCounters int64
	maxCost     string
	bufferItems int64

	*ristretto.Cache
}

// New - Creates new ristretto cache
func New(opts ...Option) (*Ristretto, error) {
	rs := &Ristretto{
		numCounters: _defaultNumCounters,
		maxCost:     _defaultMaxCost,
		bufferItems: 64,
	}

	// Custom options
	for _, opt := range opts {
		opt(rs)
	}

	mc, err := humanize.ParseBytes(rs.maxCost)
	if err != nil {
		return nil, err
	}

	c, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: rs.numCounters,
		MaxCost:     int64(mc),
		BufferItems: rs.bufferItems,
	})

	rs.Cache = c

	return rs, err
}
