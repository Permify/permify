package schema

import (
	"errors"
)

var (
	ErrUndefinedLeafType = errors.New("undefined leaf type")
	ErrUnimplemented     = errors.New("unimplemented")
)
