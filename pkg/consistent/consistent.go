package hash

import (
	"log"
	"sort"
	"strconv"
	"sync"

	"github.com/Permify/permify/pkg/gossip"
)

// TopWeight is the maximum weight for a node in the consistent hashing
// ring. Higher weights indicate higher priority for the node.
const TopWeight = 100

// minReplicas is the default minimum number of replicas for a node in the
// consistent hashing ring. More replicas help distribute the load more
// evenly among the nodes.
const minReplicas = 100

// Consistent is an interface that represents a consistent hashing ring
// with support for node weights and replicas. Implementations of this
// interface should provide methods for adding, updating, and removing
// nodes, as well as retrieving the responsible node for a given key.
type Consistent interface {
	// Add inserts a node into the consistent hashing ring with the
	// default number of replicas and weight.
	Add(node string)

	// SyncNodes updates the consistent hashing ring with the current
	// state of the gossip cluster.
	SyncNodes(g *gossip.Gossip)

	// AddWithReplicas inserts a node into the consistent hashing ring
	// with the specified number of replicas and default weight.
	AddWithReplicas(node string, replicas int)

	// AddWithWeight inserts a node into the consistent hashing ring
	// with the specified weight and default number of replicas.
	AddWithWeight(node string, weight int)

	// Get finds the responsible node for the given key in the consistent
	// hashing ring. If a node is found, it returns the node and true;
	// otherwise, it returns an empty string and false.
	Get(v string) (string, bool)

	// Remove deletes a node from the consistent hashing ring.
	Remove(node string)

	// AddKey adds the given key to the consistent hashing ring and
	// returns true if the key was successfully added; otherwise, it
	// returns false.
	AddKey(key string) bool
}

// Func is a function type that takes a byte slice as input and returns
// a 64-bit unsigned integer hash value.
type Func func(data []byte) uint64

// ConsistentHash is a struct that implements a consistent hashing
// algorithm with support for node replicas. It contains the following fields:
//   - hashFunc: a Func type that represents the hash function to use for keys.
//   - replicas: the number of replicas for each node in the consistent hashing ring.
//   - keys: a slice of 64-bit unsigned integers representing the sorted hash keys.
//   - ring: a map that associates hash keys with their corresponding node names.
//   - nodes: a map that stores the nodes' names as keys with empty struct values
//     (used for quick membership tests).
//   - lock: a sync.RWMutex that provides synchronization for concurrent access.
type ConsistentHash struct {
	hashFunc Func
	replicas int
	keys     []uint64
	ring     map[uint64]string
	nodes    map[string]struct{}
	lock     sync.RWMutex
}

// NewConsistentHash creates a new ConsistentHash instance with the specified
// number of replicas, initial node list, and hash function. If the number of
// replicas is less than minReplicas, it will be set to minReplicas. If the
// hash function is not provided (nil), the default Hash function will be used.
// This function returns a pointer to the newly created ConsistentHash instance.
func NewConsistentHash(replicas int, nodes []string, fn Func) *ConsistentHash {
	// Ensure the number of replicas is not less than minReplicas.
	if replicas < minReplicas {
		replicas = minReplicas
	}

	// Set the default hash function if none is provided.
	if fn == nil {
		fn = Hash
	}

	// Create a new ConsistentHash instance with the specified hash function,
	// number of replicas, and empty ring and nodes maps.
	consistent := &ConsistentHash{
		hashFunc: fn,
		replicas: replicas,
		ring:     make(map[uint64]string),
		nodes:    make(map[string]struct{}),
	}

	// Add the initial nodes to the ConsistentHash instance.
	for i := range nodes {
		consistent.Add(nodes[i])
	}

	// Return the pointer to the new ConsistentHash instance.
	return consistent
}

// Add inserts a node into the ConsistentHash instance with the default
// number of replicas. It calls AddWithReplicas, passing the node and
// the number of replicas stored in the ConsistentHash instance.
func (h *ConsistentHash) Add(node string) {
	h.AddWithReplicas(node, h.replicas)
}

// SyncNodes updates the ConsistentHash instance with the current state
// of the gossip cluster. It adds new nodes from the gossip cluster with
// the default weight and removes nodes that are no longer in the cluster.
func (h *ConsistentHash) SyncNodes(g *gossip.Gossip) {
	// Get the current list of nodes from the gossip cluster.
	nodes := g.SyncMemberList()

	// Get the current list of nodes in the ConsistentHash instance.
	consistentNodes := h.nodes

	// Create a map of the nodes in the gossip cluster for easier comparison.
	nodeMap := make(map[string]struct{})
	for _, node := range nodes {
		nodeMap[node] = struct{}{}

		// If the node is not in the ConsistentHash instance, add it with the default weight.
		if _, exists := consistentNodes[node]; !exists {
			h.AddWithWeight(node, 100)
		}
	}

	// Remove nodes from the ConsistentHash instance that are no longer in the gossip cluster.
	for nodeName := range consistentNodes {
		if _, exists := nodeMap[nodeName]; !exists {
			h.Remove(nodeName)
		}
	}
}

// AddWithReplicas inserts a node into the ConsistentHash instance with the
// specified number of replicas. If the number of replicas is greater than
// the maximum allowed by the instance, it will be capped. It also updates
// the keys and ring with the hash values for the node's replicas and sorts
// the keys slice.
func (h *ConsistentHash) AddWithReplicas(node string, replicas int) {
	h.Remove(node)

	// Cap the number of replicas if it's greater than the maximum allowed.
	if replicas > h.replicas {
		replicas = h.replicas
	}

	h.lock.Lock()
	defer h.lock.Unlock()
	h.addNode(node)

	// Add the hash values for the node's replicas to the keys and ring.
	for i := 0; i < replicas; i++ {
		hash := h.hashFunc([]byte(node + strconv.Itoa(i)))
		h.keys = append(h.keys, hash)
		h.ring[hash] = node
	}

	// Sort the keys slice.
	sort.Slice(h.keys, func(i, j int) bool {
		return h.keys[i] < h.keys[j]
	})
}

// AddWithWeight inserts a node into the ConsistentHash instance with a
// specified weight. The number of replicas is calculated based on the
// weight and the maximum allowed replicas. If the resulting number of
// replicas is less than 1, it will be set to 1. It calls AddWithReplicas
// with the calculated number of replicas.
func (h *ConsistentHash) AddWithWeight(node string, weight int) {
	replicas := h.replicas * weight / TopWeight
	if replicas < 1 {
		replicas = 1
	}
	h.AddWithReplicas(node, replicas)
}

// Get retrieves the node that is responsible for the specified key
// from the ConsistentHash instance. It calculates the hash of the key,
// finds the index of the smallest hash in the keys slice that is greater
// than or equal to the key hash, and returns the corresponding node
// from the ring. If the ring is empty, it returns an empty string and false.
func (h *ConsistentHash) Get(v string) (string, bool) {
	h.lock.RLock()
	defer h.lock.RUnlock()

	// If the ring is empty, return an empty string and false.
	if len(h.ring) == 0 {
		log.Printf("ring is empty")
		return "", false
	}

	// Calculate the hash of the key.
	hash := h.hashFunc([]byte(v))
	// Find the index of the smallest hash in the keys slice that is greater
	// than or equal to the key hash.
	index := sort.Search(len(h.keys), func(i int) bool {
		return h.keys[i] >= hash
	})

	// If the index is equal to the length of the keys slice, wrap around to the beginning.
	if index == len(h.keys) {
		index = 0
	}

	// Get the node corresponding to the hash in the ring.
	node := h.ring[h.keys[index]]
	if node != "" {
		return node, true
	}

	return "", false
}

// Remove deletes a node from the ConsistentHash instance. It removes
// the node's replicas from the keys slice and the ring. It also removes
// the node from the nodes map.
func (h *ConsistentHash) Remove(node string) {
	h.lock.Lock()
	defer h.lock.Unlock()

	// If the node is not in the nodes map, return early.
	if !h.containsNode(node) {
		return
	}

	// Remove the node's replicas from the keys slice and the ring.
	for i := 0; i < h.replicas; i++ {
		hash := h.hashFunc([]byte(node + strconv.Itoa(i)))
		index := sort.Search(len(h.keys), func(i int) bool {
			return h.keys[i] >= hash
		})
		if index < len(h.keys) && h.keys[index] == hash {
			h.keys = append(h.keys[:index], h.keys[index+1:]...)
		}
		delete(h.ring, hash)
	}

	// Remove the node from the nodes map.
	h.removeNode(node)
}

// AddKey inserts a key into the ConsistentHash instance, associating it
// with the appropriate node. It calculates the hash of the key, finds
// the responsible node using the Get method, and adds the hash and node
// to the keys slice and the ring. It then sorts the keys slice and returns true.
func (h *ConsistentHash) AddKey(key string) bool {
	node, ok := h.Get(key)
	if !ok {
		return false
	}
	hash := h.hashFunc([]byte(key))
	h.keys = append(h.keys, hash)
	h.ring[hash] = node

	sort.Slice(h.keys, func(i, j int) bool {
		return h.keys[i] < h.keys[j]
	})

	return true
}

// addNode adds a node to the nodes map in the ConsistentHash instance.
func (h *ConsistentHash) addNode(node string) {
	h.nodes[node] = struct{}{}
}

// removeNode removes a node from the nodes map in the ConsistentHash instance.
func (h *ConsistentHash) removeNode(node string) {
	delete(h.nodes, node)
}

// containsNode checks if a node exists in the nodes map in the ConsistentHash instance.
func (h *ConsistentHash) containsNode(node string) bool {
	_, ok := h.nodes[node]
	return ok
}
