package cmd

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	basev1 "github.com/Permify/permify/pkg/pb/base/v1"
)

func TestRunPermissionCheck_Allowed(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	stub := &stubPermissionClient{
		checkFn: func(_ context.Context, in *basev1.PermissionCheckRequest, _ ...grpc.CallOption) (*basev1.PermissionCheckResponse, error) {
			assert.Equal(t, "t1", in.GetTenantId())
			assert.Equal(t, "view", in.GetPermission())
			return &basev1.PermissionCheckResponse{Can: basev1.CheckResult_CHECK_RESULT_ALLOWED}, nil
		},
	}
	ent := &basev1.Entity{Type: "document", Id: "1"}
	sub := &basev1.Subject{Type: "user", Id: "1"}
	rpcCtx, cancel := newGRPCCallContext(context.Background())
	defer cancel()
	err := runPermissionCheck(rpcCtx, &buf, stub, "t1", ent, sub, "view")
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "allowed")
}

func TestRunPermissionCheck_Denied(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	stub := &stubPermissionClient{
		checkFn: func(_ context.Context, _ *basev1.PermissionCheckRequest, _ ...grpc.CallOption) (*basev1.PermissionCheckResponse, error) {
			return &basev1.PermissionCheckResponse{Can: basev1.CheckResult_CHECK_RESULT_DENIED}, nil
		},
	}
	rpcCtx, cancel := newGRPCCallContext(context.Background())
	defer cancel()
	err := runPermissionCheck(rpcCtx, &buf, stub, "t1",
		&basev1.Entity{Type: "document", Id: "1"},
		&basev1.Subject{Type: "user", Id: "1"}, "edit")
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "denied")
}

func TestRunPermissionCheck_RPCError(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	stub := &stubPermissionClient{
		checkFn: func(_ context.Context, _ *basev1.PermissionCheckRequest, _ ...grpc.CallOption) (*basev1.PermissionCheckResponse, error) {
			return nil, status.Errorf(codes.FailedPrecondition, "schema missing")
		},
	}
	rpcCtx, cancel := newGRPCCallContext(context.Background())
	defer cancel()
	err := runPermissionCheck(rpcCtx, &buf, stub, "t1",
		&basev1.Entity{Type: "document", Id: "1"},
		&basev1.Subject{Type: "user", Id: "1"}, "view")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "schema missing")
}

func TestNewCheckCommand_RequiredFlags(t *testing.T) {
	t.Parallel()
	cmd := NewCheckCommand()
	cmd.SetArgs([]string{})
	cmd.SetOut(bytes.NewBuffer(nil))
	cmd.SetErr(bytes.NewBuffer(nil))
	err := cmd.Execute()
	require.Error(t, err)
}
