package engines

import (
	"sort"
	"testing"

	"github.com/rs/xid"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/pkg/dsl/compiler"
	"github.com/Permify/permify/pkg/dsl/parser"
)

// This is the entry point for the test suite for the "engine" package.
// It registers a failure handler and runs the specifications (specs) for this package.
func TestEngines(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "engine-suite")
}

// newSchema -
func newSchema(model string) ([]storage.SchemaDefinition, error) {
	sch, err := parser.NewParser(model).Parse()
	if err != nil {
		return nil, err
	}

	_, _, err = compiler.NewCompiler(false, sch).Compile()
	if err != nil {
		return nil, err
	}

	version := xid.New().String()

	cnf := make([]storage.SchemaDefinition, 0, len(sch.Statements))
	for _, st := range sch.Statements {
		cnf = append(cnf, storage.SchemaDefinition{
			TenantID:             "t1",
			Version:              version,
			Name:                 st.GetName(),
			SerializedDefinition: []byte(st.String()),
		})
	}

	return cnf, err
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
