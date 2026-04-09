package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGRPCStatusError(t *testing.T) {
	t.Parallel()
	err := GRPCStatusError(status.Errorf(codes.InvalidArgument, "bad request"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "bad request")
}

func TestGRPCStatusError_NonGRPC(t *testing.T) {
	t.Parallel()
	err := GRPCStatusError(assert.AnError)
	require.Error(t, err)
	assert.ErrorIs(t, err, assert.AnError)
}
