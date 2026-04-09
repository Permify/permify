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

func TestParseDataYAML(t *testing.T) {
	t.Parallel()
	raw := []byte(`
metadata:
  schema_version: ""
tuples:
  - entity: {type: organization, id: "1"}
    relation: admin
    subject: {type: user, id: "3"}
`)
	tuples, meta, err := parseDataYAML(raw)
	require.NoError(t, err)
	require.Len(t, tuples, 1)
	assert.Equal(t, "organization", tuples[0].GetEntity().GetType())
	assert.Equal(t, "admin", tuples[0].GetRelation())
	require.NotNil(t, meta)
}

func TestParseDataYAML_NoTuples(t *testing.T) {
	t.Parallel()
	_, _, err := parseDataYAML([]byte(`tuples: []`))
	require.Error(t, err)
}

func TestRunDataWrite(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	stub := &stubDataClient{
		writeFn: func(_ context.Context, in *basev1.DataWriteRequest, _ ...grpc.CallOption) (*basev1.DataWriteResponse, error) {
			assert.Equal(t, "t1", in.GetTenantId())
			assert.Len(t, in.GetTuples(), 1)
			return &basev1.DataWriteResponse{SnapToken: "snap-1"}, nil
		},
	}
	tuples := []*basev1.Tuple{
		{
			Entity:   &basev1.Entity{Type: "document", Id: "1"},
			Relation: "viewer",
			Subject:  &basev1.Subject{Type: "user", Id: "1"},
		},
	}
	rpcCtx, cancel := newGRPCCallContext(context.Background())
	defer cancel()
	err := runDataWrite(rpcCtx, &buf, stub, "t1", &basev1.DataWriteRequestMetadata{}, tuples)
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "snap-1")
}

func TestRunDataReadRelationships(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	stub := &stubDataClient{
		readRelFn: func(_ context.Context, in *basev1.RelationshipReadRequest, _ ...grpc.CallOption) (*basev1.RelationshipReadResponse, error) {
			assert.Equal(t, "document", in.GetFilter().GetEntity().GetType())
			return &basev1.RelationshipReadResponse{
				Tuples: []*basev1.Tuple{
					{
						Entity:   &basev1.Entity{Type: "document", Id: "1"},
						Relation: "viewer",
						Subject:  &basev1.Subject{Type: "user", Id: "2"},
					},
				},
			}, nil
		},
	}
	rpcCtx, cancel := newGRPCCallContext(context.Background())
	defer cancel()
	err := runDataReadRelationships(rpcCtx, &buf, stub, "t1", &basev1.Entity{Type: "document", Id: "1"}, 10)
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "document:1")
}

func TestNewDataWriteCommand_RequiredFlags(t *testing.T) {
	t.Parallel()
	cmd := newDataWriteCommand()
	cmd.SetArgs([]string{})
	cmd.SetOut(bytes.NewBuffer(nil))
	cmd.SetErr(bytes.NewBuffer(nil))
	require.Error(t, cmd.Execute())
}
