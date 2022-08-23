package internal_errors

import (
	"errors"
)

var (
	ActionCannotFoundError       = errors.New("internal:repository, action cannot found")
	EntityConfigCannotFoundError = errors.New("internal:repository, entity config cannot found")
)
