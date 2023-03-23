package engines

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// This is the entry point for the test suite for the "engine" package.
// It registers a failure handler and runs the specifications (specs) for this package.
func TestEngines(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "engine-suite")
}
