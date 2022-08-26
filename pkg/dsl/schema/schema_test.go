package schema

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// TestSchema -
func TestSchema(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "schema-suite")
}
