package ristretto

// Option - Option types for cache
type Option func(ristretto *Ristretto)

// MaxCost - Defines maximum cost for ristretto
func MaxCost(n string) Option {
	return func(c *Ristretto) {
		c.maxCost = n
	}
}

// NumberOfCounters - Number of keys to track frequency.
func NumberOfCounters(n int64) Option {
	return func(c *Ristretto) {
		c.numCounters = n
	}
}
