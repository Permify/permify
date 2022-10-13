package database

import (
	"net/http"

	"google.golang.org/grpc/codes"

	"github.com/Permify/permify/pkg/errors"
)

var (
	// ErrRecordNotFound record not found error
	ErrRecordNotFound errors.Kind = "record not found"

	// ErrUniqueConstraint duplicate key value violates
	ErrUniqueConstraint errors.Kind = "unique constraint"

	// ErrBuilder query builder error
	ErrBuilder errors.Kind = "query builder"

	// ErrExecution -
	ErrExecution errors.Kind = "execution"

	// ErrScan -
	ErrScan errors.Kind = "scan"

	// ErrMigration -
	ErrMigration errors.Kind = "migration"
)

// KindToHttpStatus -
var KindToHttpStatus = map[errors.Kind]int{
	ErrRecordNotFound:   http.StatusUnprocessableEntity,
	ErrUniqueConstraint: http.StatusUnprocessableEntity,
	ErrBuilder:          http.StatusInternalServerError,
	ErrExecution:        http.StatusInternalServerError,
	ErrScan:             http.StatusInternalServerError,
	ErrMigration:        http.StatusInternalServerError,
}

// GetKindToHttpStatus -
func GetKindToHttpStatus(kind errors.Kind) int {
	status, ok := KindToHttpStatus[kind]
	if !ok {
		return http.StatusInternalServerError
	}
	return status
}

// KindToGrpcStatus -
var KindToGrpcStatus = map[errors.Kind]codes.Code{
	ErrRecordNotFound:   codes.InvalidArgument,
	ErrUniqueConstraint: codes.InvalidArgument,
	ErrBuilder:          codes.Internal,
	ErrExecution:        codes.Internal,
	ErrScan:             codes.Internal,
	ErrMigration:        codes.Internal,
}

// GetKindToGRPCStatus -
func GetKindToGRPCStatus(kind errors.Kind) codes.Code {
	status, ok := KindToGrpcStatus[kind]
	if !ok {
		return http.StatusInternalServerError
	}
	return status
}
