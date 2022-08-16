package internal_errors

import (
	"errors"
)

var (
	ActionCannotFoundError       = errors.New("action cannot found")
	EntityConfigCannotFoundError = errors.New("entity config cannot found")
)
