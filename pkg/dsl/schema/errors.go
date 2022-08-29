package schema

import (
	"errors"
)

var (
	// ActionCanNotFoundErr record not found error
	ActionCanNotFoundErr = errors.New("action not found")

	// RelationCanNotFoundErr record not found error
	RelationCanNotFoundErr = errors.New("relation not found")

	// EntityCanNotFoundErr record not found error
	EntityCanNotFoundErr = errors.New("entity not found")
)
