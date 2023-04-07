package cache

// Cache - Defines an interface for a generic cache.
type Cache interface {
	Get(key interface{}) (interface{}, bool)
	Set(key, value interface{}, cost int64) bool
	Wait()
	Close()
}

// noopCache -
type noopCache struct{}

func NewNoopCache() Cache { return &noopCache{} }

func (c *noopCache) Get(key any) (any, bool) {
	return nil, false
}

func (c *noopCache) Set(key, entry any, cost int64) bool {
	return false
}

func (c *noopCache) Wait() {}

func (c *noopCache) Close() {}
