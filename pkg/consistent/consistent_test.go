package hash_test

import (
	hash "github.com/Permify/permify/pkg/consistent"
	"github.com/gookit/color"
	"testing"
)

func TestConsistentHash_Updated(t *testing.T) {
	ch := hash.NewConsistentHash(100, []string{"node1", "node2", "node3", "node5"}, nil)

	keys := []string{"key1", "key2", "key3", "key4", "key5", "key6", "key7", "key8", "key9", "key10"}

	// Test initial assignment
	assignment := make(map[string]string)
	for _, key := range keys {
		node, ok := ch.Get(key)
		if !ok {
			t.Errorf("Failed to get node for key '%s'", key)
		}
		assignment[key] = node
	}

	ok := ch.AddKey("wrqw")
	if !ok {
		t.Errorf("Failed to get node for key '%s'", "")
	}

	nodes, _ := ch.Get("wrqw")
	t.Log(nodes)
	// Test consistency after adding a new node with weight
	ch.AddWithWeight("node4", 100)
	for _, key := range keys {
		node, ok := ch.Get(key)
		if !ok {
			t.Errorf("Failed to get node for key '%s'", key)
		}
		if node != assignment[key] {
			color.Info.Printf("Key '%s' was reassigned from '%s' to '%s' after removing a node", key, assignment[key], node)
		}
	}

	// Test consistency after removing a node
	ch.Remove("node4")
	for _, key := range keys {
		node, ok := ch.Get(key)
		if !ok {
			t.Errorf("Failed to get node for key '%s'", key)
		}
		if node != assignment[key] {
			color.Info.Printf("Key '%s' was reassigned from '%s' to '%s' after removing a node", key, assignment[key], node)
		}
	}

	// Test consistency after removing a node
	ch.Remove("node1")
	for _, key := range keys {
		node, ok := ch.Get(key)
		if !ok {
			t.Errorf("Failed to get node for key '%s'", key)
		}
		if node != assignment[key] {
			color.Info.Printf("Key '%s' was reassigned from '%s' to '%s' after removing a node", key, assignment[key], node)
		}
	}

	for _, key := range keys {
		node, _ := ch.Get(key)

		t.Log(node, key)
	}
}
