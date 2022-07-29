package mongo

import (
	"time"
)

// Option -.
type Option func(*Mongo)

// MaxPoolSize -.
func MaxPoolSize(size int) Option {
	return func(c *Mongo) {
		c.maxPoolSize = size
	}
}

// ConnTimeout -.
func ConnTimeout(timeout time.Duration) Option {
	return func(c *Mongo) {
		c.connTimeout = timeout
	}
}
