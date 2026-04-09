package cmd

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"

	basev1 "github.com/Permify/permify/pkg/pb/base/v1"
)

func TestRunTenantCreate(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	stub := &stubTenancyClient{
		createFn: func(_ context.Context, in *basev1.TenantCreateRequest, _ ...grpc.CallOption) (*basev1.TenantCreateResponse, error) {
			assert.Equal(t, "t1", in.GetId())
			assert.Equal(t, "One", in.GetName())
			return &basev1.TenantCreateResponse{
				Tenant: &basev1.Tenant{Id: "t1", Name: "One", CreatedAt: timestamppb.Now()},
			}, nil
		},
	}
	rpcCtx, cancel := newGRPCCallContext(context.Background())
	defer cancel()
	err := runTenantCreate(rpcCtx, &buf, stub, "t1", "One")
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "t1")
}

func TestRunTenantList(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	stub := &stubTenancyClient{
		listFn: func(_ context.Context, in *basev1.TenantListRequest, _ ...grpc.CallOption) (*basev1.TenantListResponse, error) {
			assert.EqualValues(t, 50, in.GetPageSize())
			return &basev1.TenantListResponse{
				Tenants: []*basev1.Tenant{{Id: "a", Name: "A"}},
			}, nil
		},
	}
	rpcCtx, cancel := newGRPCCallContext(context.Background())
	defer cancel()
	err := runTenantList(rpcCtx, &buf, stub, 50, "")
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "a")
}

func TestRunTenantDelete(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	stub := &stubTenancyClient{
		deleteFn: func(_ context.Context, in *basev1.TenantDeleteRequest, _ ...grpc.CallOption) (*basev1.TenantDeleteResponse, error) {
			assert.Equal(t, "t1", in.GetId())
			return &basev1.TenantDeleteResponse{TenantId: "t1"}, nil
		},
	}
	rpcCtx, cancel := newGRPCCallContext(context.Background())
	defer cancel()
	err := runTenantDelete(rpcCtx, &buf, stub, "t1")
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "t1")
}

func TestNewTenantCreateCommand_RequiredFlags(t *testing.T) {
	t.Parallel()
	cmd := newTenantCreateCommand()
	cmd.SetArgs([]string{})
	cmd.SetOut(bytes.NewBuffer(nil))
	cmd.SetErr(bytes.NewBuffer(nil))
	require.Error(t, cmd.Execute())
}
