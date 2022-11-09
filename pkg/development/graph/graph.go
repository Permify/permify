package graph

import (
	"sync"
)

// Node -
type Node struct {
	Type  string `json:"type"`
	ID    string `json:"id"`
	Label string `json:"label"`
}

// Edge -
type Edge struct {
	Extra any   `json:"extra"`
	From  *Node `json:"from"`
	To    *Node `json:"to"`
}

// Graph -
type Graph struct {
	nodes []*Node
	edges []*Edge
	lock  sync.RWMutex
}

// Nodes -
func (g *Graph) Nodes() []*Node {
	return g.nodes
}

// Edges -
func (g *Graph) Edges() []*Edge {
	return g.edges
}

// AddNodes -
func (g *Graph) AddNodes(n []*Node) {
	g.lock.Lock()
	g.nodes = append(g.nodes, n...)
	g.lock.Unlock()
}

// AddNode -
func (g *Graph) AddNode(n *Node) {
	g.lock.Lock()
	g.nodes = append(g.nodes, n)
	g.lock.Unlock()
}

// AddEdges -
func (g *Graph) AddEdges(e []*Edge) {
	g.lock.Lock()
	g.edges = append(g.edges, e...)
	g.lock.Unlock()
}

// AddEdge -
func (g *Graph) AddEdge(from, to *Node, extra any) {
	g.lock.Lock()
	g.edges = append(g.edges, &Edge{
		Extra: extra,
		From:  from,
		To:    to,
	})
	g.lock.Unlock()
}
