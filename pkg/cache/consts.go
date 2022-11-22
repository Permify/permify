package cache

// Engine - Engine type for cache
type Engine string

const (
	RISTRETTO Engine = "ristretto"
)

// String - String converter
func (c Engine) String() string {
	return string(c)
}
