package attribute

import (
	"errors"
)

var (
	ErrInvalidEntity    = errors.New("invalid entity")
	ErrInvalidAttribute = errors.New("invalid attribute")
	ErrInvalidValue     = errors.New("invalid value")
)
