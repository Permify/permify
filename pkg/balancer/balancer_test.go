package balancer

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// This is the entry point for the test suite for the "consistent" package.
// It registers a failure handler and runs the specifications (specs) for this package.
func TestBalancer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "balancer-suite")
}
