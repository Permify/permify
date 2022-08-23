package internal_errors

import (
	"errors"
)

var (
	DepthError              = errors.New("internal:command, depth error")
	CanceledError           = errors.New("internal:command, canceled error")
	UndefinedChildTypeError = errors.New("internal:command, undefined child type")
	UndefinedChildKindError = errors.New("internal:command, undefined child kind")
)
