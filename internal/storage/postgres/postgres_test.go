package postgres

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPostgres(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "postgres-suite")
}
