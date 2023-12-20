package tuple

import (
	"errors"
)

var (
	ErrInvalidEntity            = errors.New("invalid entity")
	ErrInvalidTuple             = errors.New("invalid tuple")
	ErrInvalidEntityAndRelation = errors.New("invalid entity and relation")
)
