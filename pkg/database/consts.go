package database

// Engine -
type Engine string

const (
	POSTGRES Engine = "postgres"
	MEMORY   Engine = "memory"
)

// String -
func (c Engine) String() string {
	return string(c)
}
