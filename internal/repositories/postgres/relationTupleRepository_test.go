package postgres

import (
	"github.com/DATA-DOG/go-sqlmock"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Relation Tuple Repository", func() {
	var mock sqlmock.Sqlmock

	BeforeEach(func() {
		var err error
		Expect(err).ShouldNot(HaveOccurred())
	})

	AfterEach(func() {
		err := mock.ExpectationsWereMet()
		Expect(err).ShouldNot(HaveOccurred())
	})
})
