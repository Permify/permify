package hash_test

import (
	"testing"

	hash "github.com/Permify/permify/pkg/consistent"
	"github.com/gookit/color"
)

func TestConsistentHash_Updated(t *testing.T) {
	ch := hash.NewConsistentHash(100, nil)

	keys := []string{"key1", "key2", "key3", "key4", "key5", "key6", "key7", "key8", "key9", "key10"}
	ch.Add("node1")
	ch.Add("node2")

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
		t.Errorf("Failed to add key '%s'", "wrqw")
	}

	// Test consistency after adding a new node
	ch.Add("node4")
	for _, key := range keys {
		node, ok := ch.Get(key)
		if !ok {
			t.Errorf("Failed to get node for key '%s'", key)
		}
		if node != assignment[key] {
			color.Info.Printf("Key '%s' was reassigned from '%s' to '%s' after adding a node", key, assignment[key], node)
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
	t.Log("TestConsistentHash_Updated passed")
}
