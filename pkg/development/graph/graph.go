package graph

// Node - Structure
type Node struct {
	Type  string `json:"type"`
	ID    string `json:"id"`
	Label string `json:"label"`
}

// Edge - Edge Structure
type Edge struct {
	From *Node `json:"from"`
	To   *Node `json:"to"`
}

// Graph - Graph Structure
type Graph struct {
	nodes []*Node
	edges []*Edge
}

// Nodes - Return Nodes Slice
func (g *Graph) Nodes() []*Node {
	return g.nodes
}

// Edges - Return Edge Slice
func (g *Graph) Edges() []*Edge {
	return g.edges
}

// AddNodes - Add nodes to graph
func (g *Graph) AddNodes(n []*Node) {
	g.nodes = append(g.nodes, n...)
}

// AddNode - Add node to graph
func (g *Graph) AddNode(n *Node) {
	g.nodes = append(g.nodes, n)
}

// AddEdges - Add edges to graph
func (g *Graph) AddEdges(e []*Edge) {
	g.edges = append(g.edges, e...)
}

// AddEdge - Add edge to graph
func (g *Graph) AddEdge(from, to *Node) {
	g.edges = append(g.edges, &Edge{
		From: from,
		To:   to,
	})
}
