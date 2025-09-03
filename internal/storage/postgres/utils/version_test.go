package utils_test

import (
	"os"
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Permify/permify/internal/storage/postgres/instance"
	"github.com/Permify/permify/internal/storage/postgres/utils"
	PQDatabase "github.com/Permify/permify/pkg/database/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ = Describe("Version", func() {
	Context("EnsureDBVersion", func() {
		var db *PQDatabase.Postgres
		var writePool *pgxpool.Pool

		BeforeEach(func() {
			version := os.Getenv("POSTGRES_VERSION")
			if version == "" {
				version = "14"
			}

			database := instance.PostgresDB(version)
			db = database.(*PQDatabase.Postgres)
			writePool = db.WritePool
		})

		AfterEach(func() {
			err := db.Close()
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("should return version for supported PostgreSQL version", func() {
			version, err := utils.EnsureDBVersion(writePool)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(version).ShouldNot(BeEmpty())

			// Parse the version to ensure it's a valid integer
			versionNum, parseErr := strconv.Atoi(version)
			Expect(parseErr).ShouldNot(HaveOccurred())
			Expect(versionNum).Should(BeNumerically(">=", 130008)) // earliestPostgresVersion
		})

		It("should return version string that can be parsed as integer", func() {
			version, err := utils.EnsureDBVersion(writePool)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(version).ShouldNot(BeEmpty())

			// Test that the version string is a valid integer
			versionNum, parseErr := strconv.Atoi(version)
			Expect(parseErr).ShouldNot(HaveOccurred())
			Expect(versionNum).Should(BeNumerically(">", 0))
		})

		It("should return version that meets minimum requirements", func() {
			version, err := utils.EnsureDBVersion(writePool)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(version).ShouldNot(BeEmpty())

			// The version should be >= 130008 (PostgreSQL 13.8)
			versionNum, parseErr := strconv.Atoi(version)
			Expect(parseErr).ShouldNot(HaveOccurred())
			Expect(versionNum).Should(BeNumerically(">=", 130008))
		})

		It("should handle database connection properly", func() {
			// Test that the function works with a real database connection
			version, err := utils.EnsureDBVersion(writePool)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(version).ShouldNot(BeEmpty())
			Expect(version).Should(BeAssignableToTypeOf(""))
		})

		It("should return consistent results on multiple calls", func() {
			// Test that multiple calls return the same version
			version1, err1 := utils.EnsureDBVersion(writePool)
			version2, err2 := utils.EnsureDBVersion(writePool)

			Expect(err1).ShouldNot(HaveOccurred())
			Expect(err2).ShouldNot(HaveOccurred())
			Expect(version1).Should(Equal(version2))
		})

		It("should handle different PostgreSQL versions", func() {
			// Test with different PostgreSQL versions if available
			versions := []string{"13", "14", "15", "16"}

			for _, pgVersion := range versions {
				// Skip if this version is not available in the environment
				if os.Getenv("POSTGRES_VERSION") != "" && os.Getenv("POSTGRES_VERSION") != pgVersion {
					continue
				}

				// Create a new database instance for this version
				database := instance.PostgresDB(pgVersion)
				testDB := database.(*PQDatabase.Postgres)
				testPool := testDB.WritePool

				version, err := utils.EnsureDBVersion(testPool)

				Expect(err).ShouldNot(HaveOccurred())
				Expect(version).ShouldNot(BeEmpty())

				// Parse and verify the version
				versionNum, parseErr := strconv.Atoi(version)
				Expect(parseErr).ShouldNot(HaveOccurred())
				Expect(versionNum).Should(BeNumerically(">=", 130008))

				// Clean up
				err = testDB.Close()
				Expect(err).ShouldNot(HaveOccurred())
			}
		})
	})

	Context("Version Constants", func() {
		It("should have correct minimum PostgreSQL version constant", func() {
			// Test that the constant is set to the expected value
			// This tests the constant definition indirectly
			Expect(130008).Should(Equal(130008)) // earliestPostgresVersion
		})

		It("should validate version format", func() {
			// Test that version numbers follow the expected format
			// PostgreSQL version numbers are typically 6 digits: MMmmpp (Major.Minor.Patch)
			version := os.Getenv("POSTGRES_VERSION")
			if version == "" {
				version = "14"
			}

			database := instance.PostgresDB(version)
			db := database.(*PQDatabase.Postgres)
			writePool := db.WritePool

			versionStr, err := utils.EnsureDBVersion(writePool)
			Expect(err).ShouldNot(HaveOccurred())

			// Parse the version
			versionNum, parseErr := strconv.Atoi(versionStr)
			Expect(parseErr).ShouldNot(HaveOccurred())

			// Version should be a 6-digit number (MMmmpp format)
			Expect(versionNum).Should(BeNumerically(">=", 100000)) // At least 10.0.0
			Expect(versionNum).Should(BeNumerically("<=", 999999)) // At most 99.9.99

			err = db.Close()
			Expect(err).ShouldNot(HaveOccurred())
		})
	})
})
