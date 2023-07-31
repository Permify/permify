package hash

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"

	"github.com/Permify/permify/internal/config"
)

func TestAddAndRemove(t *testing.T) {
	// Mock grpc server
	lis, err := net.Listen("tcp", "localhost:0")
	assert.NoError(t, err)
	grpcServer := grpc.NewServer()
	go func() {
		_ = grpcServer.Serve(lis)
	}()
	defer grpcServer.Stop()

	// Create ConsistentHash instance
	hashFunc := Hash
	c, err := NewConsistentHash(100, hashFunc, config.GRPC{})
	assert.NoError(t, err)

	// Test Add
	node := lis.Addr().String()
	err = c.Add(node)
	assert.NoError(t, err)

	_, conn, ok := c.Get("key")
	assert.True(t, ok)
	assert.NotNil(t, conn)

	// Test Remove
	err = c.Remove(node)
	assert.NoError(t, err)

	_, _, ok = c.Get("key")
	assert.False(t, ok)

	// Test AddWithReplicas
	err = c.AddWithReplicas(node, 100)
	assert.NoError(t, err)

	_, conn, ok = c.Get("key")
	assert.True(t, ok)
	assert.NotNil(t, conn)
}
