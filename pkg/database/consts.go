package database

// Engine - Type declaration of engine
type Engine string

const (
	POSTGRES Engine = "postgres"
	MEMORY   Engine = "memory"
)

// String - Convert to string
func (c Engine) String() string {
	return string(c)
}
