package specific

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSpecific(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "specific-suite")
}
