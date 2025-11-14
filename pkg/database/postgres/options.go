package postgres

import (
	"time"
)

// Option - Option type
type Option func(*Postgres)

// MaxOpenConnections - Deprecated: use MaxConns instead for consistency with pgxpool.
// Kept for backward compatibility and internally forwards to MaxConns.
func MaxOpenConnections(size int) Option {
	return MaxConns(size)
}

// MaxConns - Defines maximum number of connections in the pool (maps to pgxpool MaxConns)
func MaxConns(size int) Option {
	return func(c *Postgres) {
		c.maxConns = size
	}
}

// MaxIdleConnections - Deprecated: use MinConns instead.
// Kept for backward compatibility and only used as a fallback for MinConns when
// MinConns is not set (0). MinIdleConns is only honored when explicitly configured.
func MaxIdleConnections(c int) Option {
	return func(p *Postgres) {
		p.maxIdleConnections = c
	}
}

// MinConns - Defines minimum number of connections in the pool
// If not set (0) and MaxIdleConnections is set, MaxIdleConnections will be used for backward compatibility (old behavior).
func MinConns(c int) Option {
	return func(p *Postgres) {
		p.minConns = c
	}
}

// MinIdleConns - Defines minimum number of idle connections in the pool.
// This is superior to MinConns for ensuring idle connections are always available.
// Note: MaxIdleConnections only affects MinConns when MinConns is not set; it does
// not implicitly set MinIdleConns.
func MinIdleConns(c int) Option {
	return func(p *Postgres) {
		p.minIdleConns = c
	}
}

// MaxConnectionIdleTime - Defines maximum connection idle for postgresql db
func MaxConnectionIdleTime(d time.Duration) Option {
	return func(p *Postgres) {
		p.maxConnectionIdleTime = d
	}
}

// MaxConnectionLifeTime - Defines maximum connection lifetime for postgresql db
func MaxConnectionLifeTime(d time.Duration) Option {
	return func(p *Postgres) {
		p.maxConnectionLifeTime = d
	}
}

// HealthCheckPeriod - Defines the period between health checks on idle connections
func HealthCheckPeriod(d time.Duration) Option {
	return func(p *Postgres) {
		p.healthCheckPeriod = d
	}
}

// MaxConnLifetimeJitter - Defines the jitter added to MaxConnLifetime to prevent all connections from expiring at once
func MaxConnLifetimeJitter(d time.Duration) Option {
	return func(p *Postgres) {
		p.maxConnLifetimeJitter = d
	}
}

// ConnectTimeout - Defines the maximum time to wait when establishing a new connection
func ConnectTimeout(d time.Duration) Option {
	return func(p *Postgres) {
		p.connectTimeout = d
	}
}

func MaxDataPerWrite(v int) Option {
	return func(c *Postgres) {
		c.maxDataPerWrite = v
	}
}

func WatchBufferSize(v int) Option {
	return func(c *Postgres) {
		c.watchBufferSize = v
	}
}

func MaxRetries(v int) Option {
	return func(c *Postgres) {
		c.maxRetries = v
	}
}
