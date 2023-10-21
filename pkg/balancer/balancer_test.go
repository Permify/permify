package balancer

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestBalancer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "balancer-suite")
}
