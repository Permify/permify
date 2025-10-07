package memory // Memory storage package tests
// Test suite for memory storage
import ( // Import statements
	"sort"    // Sorting operations
	"testing" // Testing framework

	// Test framework imports
	. "github.com/onsi/ginkgo/v2" // BDD test framework
	. "github.com/onsi/gomega"    // Assertion framework
) // End of imports
// TestMemory runs the memory storage test suite
func TestMemory(t *testing.T) { // Main test function
	RegisterFailHandler(Fail)   // Register failure handler
	RunSpecs(t, "memory-suite") // Run test specs
} // End of TestMemory

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
