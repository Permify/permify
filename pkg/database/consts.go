package database

import "math"

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

const (
	_defaultPageSize = math.MaxUint32
)
