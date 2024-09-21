package cache

// Cache - Defines an interface for a generic cache.
type Cache interface {
	Get(key string) (any, bool)
	Set(key string, value any, cost int64) bool
	Wait()
	Close()
}

// noopCache -
type noopCache struct{}

func NewNoopCache() Cache { return &noopCache{} }

func (c *noopCache) Get(_ string) (any, bool) {
	return nil, false
}

func (c *noopCache) Set(_ string, _ any, _ int64) bool {
	return false
}

func (c *noopCache) Wait() {}

func (c *noopCache) Close() {}
