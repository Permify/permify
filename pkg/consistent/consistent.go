package hash

import (
	"errors"
	"fmt"
	"log"
	"sort"
	"strconv"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/Permify/permify/internal/config"
)

const (
	// TopWeight is the maximum weight for a node in the consistent hashing ring.
	TopWeight = 100

	// minReplicas is the minimum number of replicas for a node in the ring.
	minReplicas = 100
)

// Func is a type representing a hashing function that accepts a byte slice and returns an uint64 hash
type Func func(data []byte) uint64

// Consistent is an interface for a consistent hash ring
type Consistent interface {
	Add(node string) error
	AddWithReplicas(node string, replicas int) error
	AddWithWeight(node string, weight int) error
	Get(v string) (string, *grpc.ClientConn, bool)
	Remove(node string) error
	AddKey(key string) bool
}

// ConsistentHash implements a consistent hashing ring
type ConsistentHash struct {
	hashFunc          Func                        // Hash function used to hash keys
	replicas          int                         // Number of virtual nodes per physical node
	keys              []uint64                    // Sorted hash ring keys
	ring              map[uint64]string           // Hash ring
	Nodes             map[string]*grpc.ClientConn // Node addresses mapped to their gRPC client connections
	lock              sync.RWMutex                // Lock to ensure the thread safety of the hash ring
	connectionOptions []grpc.DialOption           // Connection options for gRPC
}

// NewConsistentHash creates a new consistent hash ring
func NewConsistentHash(replicas int, fn Func, connOpts config.GRPC) (*ConsistentHash, error) {
	// If the number of replicas is less than the minimum, set it to the minimum
	if replicas < minReplicas {
		replicas = minReplicas
	}

	// If no hash function is provided, use the default hash function
	if fn == nil {
		fn = Hash
	}

	// Set up the default gRPC dial options
	options := []grpc.DialOption{
		grpc.WithBlock(),
	}

	// If TLS is enabled, load the certificate from the specified path and add it to the dial options
	if connOpts.TLSConfig.Enabled {
		c, err := credentials.NewClientTLSFromFile(connOpts.TLSConfig.CertPath, "")
		if err != nil {
			return nil, fmt.Errorf("failed to create client TLS from file: %w", err)
		}
		options = append(options, grpc.WithTransportCredentials(c))
	} else {
		// If TLS is not enabled, add insecure transport credentials to the dial options
		options = append(options, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	// Return a new ConsistentHash
	return &ConsistentHash{
		hashFunc:          fn,
		replicas:          replicas,
		ring:              make(map[uint64]string),
		Nodes:             make(map[string]*grpc.ClientConn),
		connectionOptions: options,
	}, nil
}

// Add adds a node to the hash ring with the default number of replicas
func (h *ConsistentHash) Add(node string) error {
	return h.AddWithReplicas(node, h.replicas)
}

// AddWithReplicas adds a node to the hash ring with a specified number of replicas
func (h *ConsistentHash) AddWithReplicas(node string, replicas int) error {
	// Check if the node already exists in the ring
	if h.containsNode(node) {
		return errors.New("node already exists in the ring")
	}

	// Limit the number of replicas to the default if necessary
	if replicas > h.replicas {
		replicas = h.replicas
	}

	// Lock the hash ring to prevent data races
	h.lock.Lock()
	defer h.lock.Unlock()

	// Add the node and its replicas to the ring
	err := h.addNode(node)
	if err != nil {
		return err
	}

	for i := 0; i < replicas; i++ {
		hash := h.hashFunc([]byte(node + strconv.Itoa(i)))
		h.keys = append(h.keys, hash)
		h.ring[hash] = node
	}

	// Sort the keys for binary search
	sort.Slice(h.keys, func(i, j int) bool {
		return h.keys[i] < h.keys[j]
	})

	return nil
}

// AddWithWeight adds a node to the hash ring with a number of replicas proportional to its weight
func (h *ConsistentHash) AddWithWeight(node string, weight int) error {
	replicas := h.replicas * weight / TopWeight
	if replicas < 1 {
		replicas = 1
	}
	return h.AddWithReplicas(node, replicas)
}

// Remove removes a node from the hash ring
func (h *ConsistentHash) Remove(node string) error {
	// Check if the node exists in the ring
	if !h.containsNode(node) {
		return fmt.Errorf("node %s does not exist in the ring", node)
	}

	// Lock the hash ring to prevent data races
	h.lock.Lock()
	defer h.lock.Unlock()

	// Remove the node and its replicas from the ring
	err := h.removeNode(node)
	if err != nil {
		return err
	}
	newKeys := make([]uint64, 0, len(h.keys))
	for _, key := range h.keys {
		if h.ring[key] == node {
			delete(h.ring, key)
		} else {
			newKeys = append(newKeys, key)
		}
	}

	h.keys = newKeys
	return nil
}

// Get returns the node responsible for the hash of the given key and its gRPC connection
func (h *ConsistentHash) Get(v string) (string, *grpc.ClientConn, bool) {
	h.lock.RLock()
	defer h.lock.RUnlock()
	return h.get(v)
}

// get is the unprotected version of Get, which requires a pre-acquired read lock
func (h *ConsistentHash) get(v string) (string, *grpc.ClientConn, bool) {
	// If the ring is empty, return nothing
	if len(h.ring) == 0 {
		log.Printf("ring is empty")
		return "", nil, false
	}

	// Find the first hash that is not less than the hash of the key
	hash := h.hashFunc([]byte(v))
	index := sort.Search(len(h.keys), func(i int) bool {
		return h.keys[i] >= hash
	})

	// If there's no such hash, wrap around to the beginning of the ring
	if index == len(h.keys) {
		index = 0
	}

	// Lookup for the node responsible for the hash in the ring.
	node, ok := h.ring[h.keys[index]]
	if !ok {
		log.Printf("failed to get node for key %v", h.keys[index])
		return "", nil, false
	}

	// Lookup for the connection associated with the node in the Nodes map.
	conn, ok := h.Nodes[node]
	if !ok || conn == nil {
		log.Printf("failed to get connection for node %s", node)
		return "", nil, false
	}

	return node, conn, true
}

// AddKey adds a key to the hash ring and assigns it to a node
func (h *ConsistentHash) AddKey(key string) bool {
	h.lock.Lock()
	defer h.lock.Unlock()

	// Find the node responsible for the key
	node, _, ok := h.get(key)
	if !ok {
		return false
	}

	// Check if the key already exists
	_, ok = h.ring[h.hashFunc([]byte(key))]
	if ok {
		return false
	}

	// Add the key to the ring
	hash := h.hashFunc([]byte(key))
	h.keys = append(h.keys, hash)
	h.ring[hash] = node

	// Sort the keys for binary search
	sort.Slice(h.keys, func(i, j int) bool {
		return h.keys[i] < h.keys[j]
	})

	return true
}

// containsNode checks if a node exists in the ring
func (h *ConsistentHash) containsNode(node string) bool {
	_, exists := h.Nodes[node]
	return exists
}

// addNode adds a node to the ring by creating a gRPC connection to it
func (h *ConsistentHash) addNode(node string) error {
	conn, err := grpc.Dial(node, h.connectionOptions...)
	if err != nil {
		return fmt.Errorf("failed to dial: %w", err)
	}
	h.Nodes[node] = conn
	return nil
}

// removeNode removes a node from the ring by closing its gRPC connection
func (h *ConsistentHash) removeNode(node string) error {
	err := h.Nodes[node].Close()
	if err != nil {
		return fmt.Errorf("failed to remove node: %v", err)
	}
	delete(h.Nodes, node)
	return nil
}
