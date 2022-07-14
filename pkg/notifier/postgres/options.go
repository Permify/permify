package postgres

import (
	"time"
)

// Option -.
type Option func(subscriber *Notifier)

// MinReconnectInterval -.
func MinReconnectInterval(timeout time.Duration) Option {
	return func(c *Notifier) {
		c.minReconnect = timeout
	}
}

// MaxReconnectInterval -.
func MaxReconnectInterval(timeout time.Duration) Option {
	return func(c *Notifier) {
		c.maxReconnect = timeout
	}
}

// SslMode -.
func SslMode(mode string) Option {
	return func(c *Notifier) {
		c.sslMode = mode
	}
}
