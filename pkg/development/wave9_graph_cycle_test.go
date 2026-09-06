package development

import (
	"testing"
)

func TestWave9GraphCycleDetectionGuard(t *testing.T) {
	visited := make(map[string]bool)
	recursionStack := make(map[string]bool)

	graph := map[string][]string{
		"user:1":  {"group:engineering"},
		"group:engineering": {"group:admins"},
		"group:admins":      {"group:engineering"}, // cycle
	}

	var hasCycle func(node string) bool
	hasCycle = func(node string) bool {
		visited[node] = true
		recursionStack[node] = true

		for _, neighbor := range graph[node] {
			if !visited[neighbor] {
				if hasCycle(neighbor) {
					return true
				}
			} else if recursionStack[neighbor] {
				return true
			}
		}

		recursionStack[node] = false
		return false
	}

	cycleDetected := hasCycle("user:1")
	if !cycleDetected {
		t.Errorf("expected cycle detection in recursive group hierarchy")
	}
}
