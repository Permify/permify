package cache

// Engine -
type Engine string

const (
	RISTRETTO Engine = "ristretto"
)

// String -
func (c Engine) String() string {
	return string(c)
}
