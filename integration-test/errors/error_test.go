package errors

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestError(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "error-suite")
}
