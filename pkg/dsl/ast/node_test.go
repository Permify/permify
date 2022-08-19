package ast

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// TestNode -
func TestNode(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "node-suite")
}
