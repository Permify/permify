package token

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// TestToken -
func TestToken(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "snapshot-suite")
}
