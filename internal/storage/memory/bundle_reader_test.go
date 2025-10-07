package memory // Memory storage package tests
// Test file for bundle reader functionality
import ( // Import statements
	"context" // Context for request handling
	// Internal imports
	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/internal/storage/memory/migrations" // Database migrations
	"github.com/Permify/permify/pkg/database/memory"                // Memory database

	// Test framework section begins
	// Test framework section start
	// Test framework section
	// Test framework imports
	. "github.com/onsi/ginkgo/v2" // BDD test framework
	. "github.com/onsi/gomega"    // Assertion framework

	// Protocol buffer section
	// Protocol buffer imports
	base "github.com/Permify/permify/pkg/pb/base/v1"
) // End of imports
// BundleReader test suite
var _ = Describe("BundleReader", func() { // Test suite for bundle reader
	var db *memory.Memory          // Database instance
	var bundleWriter *BundleWriter // Writer instance
	var bundleReader *BundleReader // Reader instance
	// Setup test environment
	BeforeEach(func() { // Initialize test environment before each test
		database, err := memory.New(migrations.Schema) // Create new in-memory database
		Expect(err).ShouldNot(HaveOccurred())          // Ensure no error occurred
		db = database                                  // Assign database instance
		bundleWriter = NewBundleWriter(db)             // Create bundle writer
		bundleReader = NewBundleReader(db)             // Create bundle reader
	}) // End of BeforeEach
	// Teardown test environment
	AfterEach(func() { // Cleanup after each test
		err := db.Close()                     // Close database connection
		Expect(err).ShouldNot(HaveOccurred()) // Ensure no error occurred
	}) // End of AfterEach
	// Test cases for Read operation
	Context("Read", func() { // Test context for Read method
		It("should write and read DataBundles with correct relationships and attributes", func() { // Test case for successful read
			ctx := context.Background() // Create background context
			// Define test bundles
			bundles := []*base.DataBundle{ // Test data bundles
				{ // User creation bundle
					Name: "user_created", // Bundle name
					Arguments: []string{ // Bundle arguments
						"organizationID", // Organization identifier
						"userID",         // User identifier
					}, // End of arguments
					Operations: []*base.Operation{ // Bundle operations
						{ // Bundle operation definition
							RelationshipsWrite: []string{ // Relationships to write
								"organization:{{.organizationID}}#member@user:{{.userID}}", // Member relationship
								"organization:{{.organizationID}}#admin@user:{{.userID}}",  // Admin relationship
							}, // End of relationships write
							RelationshipsDelete: []string{}, // Relationships to delete
							AttributesWrite: []string{ // Attributes to write
								"organization:{{.organizationID}}$public|boolean:true", // Public attribute
							}, // End of attributes write
							AttributesDelete: []string{ // Attributes to delete
								"organization:{{.organizationID}}$balance|integer[]:120,568", // Balance attribute
							}, // End of attributes delete
						}, // End of operation
					}, // End of operations
				}, // End of bundle
			} // End of bundles
			// Convert to storage bundles
			var sBundles []storage.Bundle // Storage bundles
			for _, b := range bundles {   // Iterate over bundles
				sBundles = append(sBundles, storage.Bundle{ // Append storage bundle
					Name:       b.Name, // Bundle name
					DataBundle: b,      // Data bundle
					TenantID:   "t1",   // Tenant identifier
				}) // End of bundle append
			} // End of bundle iteration
			// Write bundles to database
			names, err := bundleWriter.Write(ctx, sBundles) // Write bundles
			// Verify write operation
			Expect(err).ShouldNot(HaveOccurred())                 // No error expected
			Expect(names).Should(Equal([]string{"user_created"})) // Verify written bundle names
			// Read bundle from database
			bundle, err := bundleReader.Read(ctx, "t1", "user_created") // Read bundle
			Expect(err).ShouldNot(HaveOccurred())                       // No error expected
			// Verify bundle properties
			Expect(bundle.GetName()).Should(Equal("user_created")) // Verify bundle name
			Expect(bundle.GetArguments()).Should(Equal([]string{   // Verify bundle arguments
				"organizationID", // Organization ID argument
				"userID",         // User ID argument
			})) // End of arguments verification
			// Verify relationships write operations
			Expect(bundle.GetOperations()[0].RelationshipsWrite).Should(Equal([]string{ // Verify relationships write
				"organization:{{.organizationID}}#member@user:{{.userID}}", // Member relationship
				"organization:{{.organizationID}}#admin@user:{{.userID}}",  // Admin relationship
			})) // End of relationships write verification
			// Verify relationships delete operations
			Expect(bundle.GetOperations()[0].RelationshipsDelete).Should(BeEmpty()) // Verify relationships delete
			// Verify attributes write operations
			Expect(bundle.GetOperations()[0].AttributesWrite).Should(Equal([]string{ // Verify attributes write
				"organization:{{.organizationID}}$public|boolean:true", // Public attribute
			})) // End of attributes write verification
			// Verify attributes delete operations
			Expect(bundle.GetOperations()[0].AttributesDelete).Should(Equal([]string{ // Verify attributes delete
				"organization:{{.organizationID}}$balance|integer[]:120,568", // Balance attribute
			})) // End of attributes delete verification
		}) // End of test case
		// Test case for non-existing bundle
		It("should get error on non-existing bundle", func() { // Test case for error handling
			ctx := context.Background() // Create background context
			// Attempt to read non-existing bundle
			_, err := bundleReader.Read(ctx, "t1", "user_created")                                 // Read non-existing bundle
			Expect(err.Error()).Should(Equal(base.ErrorCode_ERROR_CODE_BUNDLE_NOT_FOUND.String())) // Verify error
		}) // End of test case
	}) // End of Read context
}) // End of BundleReader test suite
