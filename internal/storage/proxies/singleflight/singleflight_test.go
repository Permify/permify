package singleflight

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSingleflight(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "singleflight-suite")
}
