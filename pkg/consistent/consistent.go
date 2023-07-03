package hash

import (
	"github.com/Permify/permify/internal/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"sort"
	"strconv"
	"sync"
)

const (
	// TopWeight is the maximum weight for a node in the consistent hashing ring.
	TopWeight = 100

	// minReplicas is the minimum number of replicas for a node in the ring.
	minReplicas = 100
)

type Func func(data []byte) uint64

type Consistent interface {
	Add(node string)
	AddWithReplicas(node string, replicas int)
	AddWithWeight(node string, weight int)
	Get(v string) (string, *grpc.ClientConn, bool)
	Remove(node string)
	AddKey(key string) bool
}

type ConsistentHash struct {
	hashFunc          Func
	replicas          int
	keys              []uint64
	ring              map[uint64]string
	Nodes             map[string]*grpc.ClientConn
	lock              sync.RWMutex
	connectionOptions []grpc.DialOption
}

func NewConsistentHash(replicas int, fn Func, connOpts config.GRPC) *ConsistentHash {
	if replicas < minReplicas {
		replicas = minReplicas
	}
	if fn == nil {
		fn = Hash
	}
	options := []grpc.DialOption{
		grpc.WithBlock(),
	}
	if connOpts.TLSConfig.Enabled {
		c, err := credentials.NewClientTLSFromFile(connOpts.TLSConfig.CertPath, "")
		if err != nil {
			return nil
		}
		options = append(options, grpc.WithTransportCredentials(c))
	} else {
		options = append(options, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	return &ConsistentHash{
		hashFunc:          fn,
		replicas:          replicas,
		ring:              make(map[uint64]string),
		Nodes:             make(map[string]*grpc.ClientConn),
		connectionOptions: options,
	}
}

func (h *ConsistentHash) Add(node string) {
	h.AddWithReplicas(node, h.replicas)
}

func (h *ConsistentHash) AddWithReplicas(node string, replicas int) {
	if h.containsNode(node) {
		return
	}

	if replicas > h.replicas {
		replicas = h.replicas
	}

	h.lock.Lock()
	defer h.lock.Unlock()

	h.addNode(node)
	for i := 0; i < replicas; i++ {
		hash := h.hashFunc([]byte(node + strconv.Itoa(i)))
		h.keys = append(h.keys, hash)
		h.ring[hash] = node
	}

	sort.Slice(h.keys, func(i, j int) bool {
		return h.keys[i] < h.keys[j]
	})
}

func (h *ConsistentHash) AddWithWeight(node string, weight int) {
	replicas := h.replicas * weight / TopWeight
	if replicas < 1 {
		replicas = 1
	}
	h.AddWithReplicas(node, replicas)
}

func (h *ConsistentHash) Remove(node string) {
	if !h.containsNode(node) {
		return
	}

	h.lock.Lock()
	defer h.lock.Unlock()

	h.removeNode(node)
	newKeys := make([]uint64, 0, len(h.keys))
	for _, key := range h.keys {
		if h.ring[key] == node {
			delete(h.ring, key)
		} else {
			newKeys = append(newKeys, key)
		}
	}

	h.keys = newKeys
}

func (h *ConsistentHash) Get(v string) (string, *grpc.ClientConn, bool) {
	h.lock.RLock()
	defer h.lock.RUnlock()
	return h.get(v)
}

func (h *ConsistentHash) get(v string) (string, *grpc.ClientConn, bool) {
	if len(h.ring) == 0 {
		log.Printf("ring is empty")
		return "", nil, false
	}

	hash := h.hashFunc([]byte(v))
	index := sort.Search(len(h.keys), func(i int) bool {
		return h.keys[i] >= hash
	})

	if index == len(h.keys) {
		index = 0
	}

	node := h.ring[h.keys[index]]
	if node != "" {
		return node, h.Nodes[node], true
	}

	return "", nil, false
}

func (h *ConsistentHash) AddKey(key string) bool {
	h.lock.Lock()
	defer h.lock.Unlock()

	node, _, ok := h.get(key)
	if !ok {
		return false
	}

	// Check if key already exists
	_, ok = h.ring[h.hashFunc([]byte(key))]
	if ok {
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

func (h *ConsistentHash) containsNode(node string) bool {
	_, exists := h.Nodes[node]
	return exists
}

func (h *ConsistentHash) addNode(node string) {
	conn, err := grpc.Dial(node, h.connectionOptions...)
	if err != nil {
		log.Printf("failed to dial: %v", err)
		return
	}
	h.Nodes[node] = conn
}

func (h *ConsistentHash) removeNode(node string) {
	err := h.Nodes[node].Close()
	if err != nil {
		log.Printf("failed to close connection: %v", err)
		return
	}
	delete(h.Nodes, node)
}
