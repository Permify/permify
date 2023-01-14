package postgres

// Option - Option type
type Option func(*Postgres)

// MaxOpenConnections - Defines maximum open connections for postgresql db
func MaxOpenConnections(size int) Option {
	return func(c *Postgres) {
		c.maxOpenConnections = size
	}
}
