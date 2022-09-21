package errors

import (
	"github.com/Permify/permify/pkg/errors"
)

var (
	// service errors

	CanceledError = errors.ServiceError.SetMessage("canceled error")

	// validation errors

	DepthError              = errors.ValidationError.AddParam("depth", "depth not enough")
	UndefinedChildTypeError = errors.ValidationError.AddParam("schema", "undefined child type")
	UndefinedChildKindError = errors.ValidationError.AddParam("schema", "undefined child kind")
	ActionCannotFoundError  = errors.ValidationError.AddParam("action", "action can not found")
)
