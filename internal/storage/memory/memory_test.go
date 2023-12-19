package memory

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestMemory(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "memory-suite")
}
