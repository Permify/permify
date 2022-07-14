package repositories

import (
	"errors"
)

var (
	// ErrRecordNotFound record not found error
	ErrRecordNotFound = errors.New("record not found")

	// ErrUniqueConstraint duplicate key value violates
	ErrUniqueConstraint = errors.New("unique constraint")
)
