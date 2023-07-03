package hash

import (
	"github.com/Permify/permify/internal/config"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"net"
	"testing"
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
	c := NewConsistentHash(100, hashFunc, config.GRPC{})

	// Test Add
	node := lis.Addr().String()
	c.Add(node)
	_, conn, ok := c.Get("key")
	assert.True(t, ok)
	assert.NotNil(t, conn)

	// Test Remove
	c.Remove(node)
	_, _, ok = c.Get("key")
	assert.False(t, ok)

	// Test AddWithReplicas
	c.AddWithReplicas(node, 100)
	_, conn, ok = c.Get("key")
	assert.True(t, ok)
	assert.NotNil(t, conn)
}
