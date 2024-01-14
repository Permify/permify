package memory

import (
	"sort"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestMemory(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "memory-suite")
}

// isSameArray - check if two arrays are the same
func isSameArray(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	sortedA := make([]string, len(a))
	copy(sortedA, a)
	sort.Strings(sortedA)

	sortedB := make([]string, len(b))
	copy(sortedB, b)
	sort.Strings(sortedB)

	for i := range sortedA {
		if sortedA[i] != sortedB[i] {
			return false
		}
	}

	return true
}
