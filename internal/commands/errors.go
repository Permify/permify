package commands

import (
	"errors"
)

var (
	DepthError              = errors.New("depth error")
	CanceledError           = errors.New("canceled error")
	UndefinedChildTypeError = errors.New("undefined child type")
	UndefinedChildKindError = errors.New("undefined child kind")
)
