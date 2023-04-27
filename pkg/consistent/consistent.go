package hash

import (
	"github.com/Permify/permify/pkg/gossip"
	"log"
	"sort"
	"strconv"
	"sync"
)

const (
	TopWeight   = 100
	minReplicas = 100
)

type ConsistentEngine interface {
	Add(node string)
	SyncNodes(g *gossip.Engine)
	AddWithReplicas(node string, replicas int)
	AddWithWeight(node string, weight int)
	Get(v string) (string, bool)
	Remove(node string)
	AddKey(key string) bool
}

type (
	Func func(data []byte) uint64

	ConsistentHash struct {
		hashFunc Func
		replicas int
		keys     []uint64
		ring     map[uint64]string
		nodes    map[string]struct{}
		lock     sync.RWMutex
	}
)

func NewConsistentHash(replicas int, seedNodes []string, fn Func) *ConsistentHash {
	if replicas < minReplicas {
		replicas = minReplicas
	}

	if fn == nil {
		fn = Hash
	}

	consistent := &ConsistentHash{
		hashFunc: fn,
		replicas: replicas,
		ring:     make(map[uint64]string),
		nodes:    make(map[string]struct{}),
	}

	for i := range seedNodes {
		consistent.Add(seedNodes[i])
	}

	return consistent
}

func (h *ConsistentHash) Add(node string) {
	h.AddWithReplicas(node, h.replicas)
}

func (h *ConsistentHash) SyncNodes(g *gossip.Engine) {
	nodes := g.SyncMemberList()

	consistentNodes := h.nodes

	nodeMap := make(map[string]struct{})
	for _, node := range nodes {
		nodeMap[node] = struct{}{}

		if _, exists := consistentNodes[node]; !exists {
			h.AddWithWeight(node, 100)
		}
	}

	for nodeName := range consistentNodes {
		if _, exists := nodeMap[nodeName]; !exists {
			h.Remove(nodeName)
		}
	}
}

func (h *ConsistentHash) AddWithReplicas(node string, replicas int) {
	h.Remove(node)

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

func (h *ConsistentHash) Get(v string) (string, bool) {
	h.lock.RLock()
	defer h.lock.RUnlock()

	if len(h.ring) == 0 {
		log.Printf("ring is empty")
		return "", false
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
		return node, true
	}

	return "", false
}

func (h *ConsistentHash) Remove(node string) {
	h.lock.Lock()
	defer h.lock.Unlock()

	if !h.containsNode(node) {
		return
	}

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

	h.removeNode(node)
}

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

func (h *ConsistentHash) addNode(node string) {
	h.nodes[node] = struct{}{}
}

func (h *ConsistentHash) removeNode(node string) {
	delete(h.nodes, node)
}

func (h *ConsistentHash) containsNode(node string) bool {
	_, ok := h.nodes[node]
	return ok
}
