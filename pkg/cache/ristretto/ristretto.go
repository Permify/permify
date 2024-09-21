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

	c *ristretto.Cache[string, any]
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

	c, err := ristretto.NewCache(&ristretto.Config[string, any]{
		NumCounters: rs.numCounters,
		MaxCost:     int64(mc),
		BufferItems: rs.bufferItems,
	})

	rs.c = c

	return rs, err
}

// Get - Gets value from cache
func (r *Ristretto) Get(key string) (interface{}, bool) {
	return r.c.Get(key)
}

// Set - Sets value to cache
func (r *Ristretto) Set(key string, value any, cost int64) bool {
	return r.c.Set(key, value, cost)
}

// Wait - Waits for cache to be ready
func (r *Ristretto) Wait() {
	r.c.Wait()
}

// Close - Closes cache
func (r *Ristretto) Close() {
	r.c.Close()
}
