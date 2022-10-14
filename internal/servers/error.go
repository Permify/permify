package servers

import (
	"strconv"

	"google.golang.org/grpc/codes"
)

// GetStatus -
func GetStatus(err error) codes.Code {
	var code int
	code, err = strconv.Atoi(err.Error())
	if err != nil {
		return codes.Internal
	}
	switch {
	case code > 999 && code < 1999:
		return codes.Unauthenticated
	case code > 1999 && code < 2999:
		return codes.InvalidArgument
	case code > 2999 && code < 3999:
		return codes.NotFound
	case code > 3999 && code < 4999:
		return codes.Internal
	default:
		return codes.Internal
	}
}
