package oidc

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestOIDC(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "authentication oidc suite")
}
