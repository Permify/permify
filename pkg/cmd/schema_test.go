package cmd

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	basev1 "github.com/Permify/permify/pkg/pb/base/v1"
)

func TestRunSchemaWrite(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	stub := &stubSchemaClient{
		writeFn: func(_ context.Context, in *basev1.SchemaWriteRequest, _ ...grpc.CallOption) (*basev1.SchemaWriteResponse, error) {
			assert.Equal(t, "t1", in.GetTenantId())
			assert.Contains(t, in.GetSchema(), "entity")
			return &basev1.SchemaWriteResponse{SchemaVersion: "v1"}, nil
		},
	}
	rpcCtx, cancel := newGRPCCallContext(context.Background())
	defer cancel()
	err := runSchemaWrite(rpcCtx, &buf, stub, "t1", "entity user {}")
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "v1")
}

func TestRunSchemaRead(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	stub := &stubSchemaClient{
		readFn: func(_ context.Context, in *basev1.SchemaReadRequest, _ ...grpc.CallOption) (*basev1.SchemaReadResponse, error) {
			assert.Equal(t, "t1", in.GetTenantId())
			return &basev1.SchemaReadResponse{
				Schema: &basev1.SchemaDefinition{
					EntityDefinitions: map[string]*basev1.EntityDefinition{
						"user": {Name: "user"},
					},
				},
			}, nil
		},
	}
	rpcCtx, cancel := newGRPCCallContext(context.Background())
	defer cancel()
	err := runSchemaRead(rpcCtx, &buf, stub, "t1", "")
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "user")
}

func TestNewSchemaWriteCommand_RequiredFlags(t *testing.T) {
	t.Parallel()
	cmd := newSchemaWriteCommand()
	cmd.SetArgs([]string{})
	cmd.SetOut(bytes.NewBuffer(nil))
	cmd.SetErr(bytes.NewBuffer(nil))
	require.Error(t, cmd.Execute())
}
