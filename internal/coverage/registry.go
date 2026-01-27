package coverage

import (
	"context"
	"fmt"
	"sync"
)

type contextKey struct{}

var registryKey = contextKey{}

// SourceInfo represents the position of a logic node in the schema source.
type SourceInfo struct {
	Line   int32
	Column int32
}

// Registry tracks coverage for logic nodes using deterministic paths.
type Registry struct {
	mu    sync.RWMutex
	nodes map[string]*NodeInfo
}

// NodeInfo contains coverage details for a specific logic node.
type NodeInfo struct {
	Path       string
	SourceInfo SourceInfo
	VisitCount int
	Type       string // e.g., "UNION", "INTERSECTION", "LEAF"
}

// LogicNodeCoverage represents coverage information for a logical node
type LogicNodeCoverage struct {
	Path       string
	SourceInfo SourceInfo
	Type       string
}

// EntityCoverageInfo represents coverage information for a single entity
type EntityCoverageInfo struct {
	EntityName string

	UncoveredRelationships       []string
	CoverageRelationshipsPercent int

	UncoveredAttributes       []string
	CoverageAttributesPercent int

	UncoveredAssertions       map[string][]string
	CoverageAssertionsPercent map[string]int

	UncoveredLogicNodes  []LogicNodeCoverage
	CoverageLogicPercent int
}

// SchemaCoverageInfo represents the overall coverage information for a schema
type SchemaCoverageInfo struct {
	EntityCoverageInfo         []EntityCoverageInfo
	TotalRelationshipsCoverage int
	TotalAttributesCoverage    int
	TotalAssertionsCoverage    int
	TotalLogicCoverage         int
}

// NewRegistry creates a new Coverage Registry.
func NewRegistry() *Registry {
	return &Registry{
		nodes: make(map[string]*NodeInfo),
	}
}

// ReportAll returns all logic nodes regardless of visit count.
func (r *Registry) ReportAll() (nodes []NodeInfo) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, node := range r.nodes {
		nodes = append(nodes, *node)
	}
	return nodes
}

// Register adds a node to the registry.
func (r *Registry) Register(path string, info SourceInfo, nodeType string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.nodes[path]; !ok {
		r.nodes[path] = &NodeInfo{
			Path:       path,
			SourceInfo: info,
			Type:       nodeType,
		}
	}
}

// Visit increments the visit count for a path.
func (r *Registry) Visit(path string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if node, ok := r.nodes[path]; ok {
		node.VisitCount++
	}
}

// Report returns all logic nodes and their coverage status.
func (r *Registry) Report() (uncovered []LogicNodeCoverage) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, node := range r.nodes {
		if node.VisitCount == 0 {
			uncovered = append(uncovered, LogicNodeCoverage{
				Path:       node.Path,
				SourceInfo: node.SourceInfo,
				Type:       node.Type,
			})
		}
	}
	return uncovered
}

// ContextWithRegistry returns a new context with the given registry and initial path.
func ContextWithRegistry(ctx context.Context, r *Registry) context.Context {
	return context.WithValue(ctx, registryKey, r)
}

// RegistryFromContext retrieves the registry from the context.
func RegistryFromContext(ctx context.Context) *Registry {
	if r, ok := ctx.Value(registryKey).(*Registry); ok {
		return r
	}
	return nil
}

type pathKey struct{}

// ContextWithPath returns a new context with an updated path.
func ContextWithPath(ctx context.Context, path string) context.Context {
	return context.WithValue(ctx, pathKey{}, path)
}

// PathFromContext retrieves the current path from the context.
func PathFromContext(ctx context.Context) string {
	if p, ok := ctx.Value(pathKey{}).(string); ok {
		return p
	}
	return ""
}

// Track marks the current path as visited if a registry is present in the context.
func Track(ctx context.Context) {
	if r := RegistryFromContext(ctx); r != nil {
		if p := PathFromContext(ctx); p != "" {
			r.Visit(p)
		}
	}
}

// AppendPath helper to build the deterministic path.
func AppendPath(curr, segment string) string {
	if curr == "" {
		return segment
	}
	return fmt.Sprintf("%s.%s", curr, segment)
}

// GetPath helper to build the full coverage path from permission and exclusion.
func GetPath(permission, exclusion string) string {
	if exclusion == "" {
		return permission
	}
	return fmt.Sprintf("%s.%s", permission, exclusion)
}
