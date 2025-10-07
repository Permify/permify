package memory // Memory storage package tests
// Test file for bundle writer functionality
import ( // Import statements
	"context" // Context for request handling
	// Test framework imports
	. "github.com/onsi/ginkgo/v2" // BDD test framework
	. "github.com/onsi/gomega"    // Assertion framework

	// Internal imports
	"github.com/Permify/permify/internal/storage"                   // Storage interfaces
	"github.com/Permify/permify/internal/storage/memory/migrations" // Database migrations
	"github.com/Permify/permify/pkg/database/memory"                // Memory database
	base "github.com/Permify/permify/pkg/pb/base/v1"                // Protocol buffers
) // End of imports
// BundleWriter test suite
var _ = Describe("BundleWriter", func() { // Test suite for bundle writer
	var db *memory.Memory          // Database instance
	var bundleWriter *BundleWriter // Writer instance
	var bundleReader *BundleReader // Reader instance
	// Setup test environment
	BeforeEach(func() { // Initialize before each test
		database, err := memory.New(migrations.Schema) // Create database
		Expect(err).ShouldNot(HaveOccurred())          // No error expected
		db = database                                  // Assign database
		bundleWriter = NewBundleWriter(db)             // Create writer
		bundleReader = NewBundleReader(db)             // Create reader
	}) // End of BeforeEach
	// Teardown test environment
	AfterEach(func() { // Cleanup after each test
		err := db.Close()                     // Close database
		Expect(err).ShouldNot(HaveOccurred()) // No error expected
	}) // End of AfterEach
	// Test cases for Write operation
	Context("Write", func() { // Write test context
		It("should write and read DataBundles with correct relationships and attributes", func() { // Test write operation
			ctx := context.Background() // Create context
			// Define first bundle set
			bundles1 := []*base.DataBundle{ // First bundle set
				{ // User created bundle
					Name: "user_created", // Bundle name
					Arguments: []string{ // Bundle arguments
						"organizationID", // Organization ID
						"companyID",      // Company ID
						"userID",         // User ID
					}, // End of arguments
					Operations: []*base.Operation{ // Bundle operations
						{ // Operation definition
							RelationshipsWrite: []string{ // Relationships to write
								"organization:{{.organizationID}}#member@company:{{.companyID}}#admin", // Member relationship
								"organization:{{.organizationID}}#member@user:{{.userID}}",             // Member relationship
								"organization:{{.organizationID}}#admin@user:{{.userID}}",              // Admin relationship
							}, // End of relationships write
							RelationshipsDelete: []string{ // Relationships to delete
								"company:{{.companyID}}#admin@user:{{.userID}}", // Admin relationship
							}, // End of relationships delete
							AttributesWrite: []string{ // Attributes to write
								"organization:{{.organizationID}}$public|boolean:true", // Public attribute
							}, // End of attributes write
							AttributesDelete: []string{ // Attributes to delete
								"organization:{{.organizationID}}$balance|double:120.900", // Balance attribute
							}, // End of attributes delete
						}, // End of operation
					}, // End of operations
				}, // End of bundle
			} // End of bundles1
			// Convert to storage bundles
			var sBundles1 []storage.Bundle // Storage bundles
			for _, b := range bundles1 {   // Iterate bundles
				sBundles1 = append(sBundles1, storage.Bundle{ // Append bundle
					Name:       b.Name, // Bundle name
					DataBundle: b,      // Data bundle
					TenantID:   "t1",   // Tenant ID
				}) // End of append
			} // End of iteration
			// Write first bundle set
			names1, err := bundleWriter.Write(ctx, sBundles1) // Write bundles
			// Verify write operation
			Expect(err).ShouldNot(HaveOccurred())                  // No error expected
			Expect(names1).Should(Equal([]string{"user_created"})) // Verify names
			// Read and verify first bundle
			bundle1, err := bundleReader.Read(ctx, "t1", "user_created") // Read bundle
			Expect(err).ShouldNot(HaveOccurred())                        // No error expected
			// Verify bundle properties
			Expect(bundle1.GetName()).Should(Equal("user_created")) // Verify name
			Expect(bundle1.GetArguments()).Should(Equal([]string{   // Verify arguments
				"organizationID", // Organization ID
				"companyID",      // Company ID
				"userID",         // User ID
			})) // End of arguments verification
			// Verify relationships write
			Expect(bundle1.GetOperations()[0].RelationshipsWrite).Should(Equal([]string{ // Verify relationships write
				"organization:{{.organizationID}}#member@company:{{.companyID}}#admin", // Member relationship
				"organization:{{.organizationID}}#member@user:{{.userID}}",             // Member relationship
				"organization:{{.organizationID}}#admin@user:{{.userID}}",              // Admin relationship
			})) // End of relationships write verification
			// Verify relationships delete
			Expect(bundle1.GetOperations()[0].RelationshipsDelete).Should(Equal([]string{ // Verify relationships delete
				"company:{{.companyID}}#admin@user:{{.userID}}", // Admin relationship
			})) // End of relationships delete verification
			// Verify attributes write
			Expect(bundle1.GetOperations()[0].AttributesWrite).Should(Equal([]string{ // Verify attributes write
				"organization:{{.organizationID}}$public|boolean:true", // Public attribute
			})) // End of attributes write verification
			// Verify attributes delete
			Expect(bundle1.GetOperations()[0].AttributesDelete).Should(Equal([]string{ // Verify attributes delete
				"organization:{{.organizationID}}$balance|double:120.900", // Balance attribute
			})) // End of attributes delete verification
			// Define second bundle set
			bundles2 := []*base.DataBundle{ // Second bundle set
				{ // User created bundle update
					Name: "user_created", // Bundle name
					Arguments: []string{ // Bundle arguments
						"organizationID", // Organization ID
						"companyID",      // Company ID
						"userID",         // User ID
					}, // End of arguments
					Operations: []*base.Operation{ // Bundle operations
						{ // Operation definition
							RelationshipsWrite: []string{ // Relationships to write
								"organization:{{.organizationID}}#member@company:{{.companyID}}#admin", // Member relationship
								"organization:{{.organizationID}}#admin@user:{{.userID}}",              // Admin relationship
							}, // End of relationships write
							RelationshipsDelete: []string{ // Relationships to delete
								"company:{{.companyID}}#admin@user:{{.userID}}", // Admin relationship
							}, // End of relationships delete
							AttributesWrite:  []string{}, // Empty attributes write
							AttributesDelete: []string{}, // Empty attributes delete
						}, // End of operation
					}, // End of operations
				}, // End of bundle
			} // End of bundles2
			// Convert to storage bundles
			var sBundles2 []storage.Bundle // Storage bundles
			for _, b := range bundles2 {   // Iterate bundles
				sBundles2 = append(sBundles2, storage.Bundle{ // Append bundle
					Name:       b.Name, // Bundle name
					DataBundle: b,      // Data bundle
					TenantID:   "t1",   // Tenant ID
				}) // End of append
			} // End of iteration
			// Write second bundle set
			names2, err := bundleWriter.Write(ctx, sBundles2) // Write bundles
			// Verify write operation
			Expect(err).ShouldNot(HaveOccurred())                  // No error expected
			Expect(names2).Should(Equal([]string{"user_created"})) // Verify names
			// Read and verify second bundle
			bundle2, err := bundleReader.Read(ctx, "t1", "user_created") // Read bundle
			Expect(err).ShouldNot(HaveOccurred())                        // No error expected
			// Verify bundle properties
			Expect(bundle2.GetName()).Should(Equal("user_created")) // Verify name
			Expect(bundle2.GetArguments()).Should(Equal([]string{   // Verify arguments
				"organizationID", // Organization ID
				"companyID",      // Company ID
				"userID",         // User ID
			})) // End of arguments verification
			// Verify relationships write
			Expect(bundle2.GetOperations()[0].RelationshipsWrite).Should(Equal([]string{ // Verify relationships write
				"organization:{{.organizationID}}#member@company:{{.companyID}}#admin", // Member relationship
				"organization:{{.organizationID}}#admin@user:{{.userID}}",              // Admin relationship
			})) // End of relationships write verification
			// Verify relationships delete
			Expect(bundle2.GetOperations()[0].RelationshipsDelete).Should(Equal([]string{ // Verify relationships delete
				"company:{{.companyID}}#admin@user:{{.userID}}", // Admin relationship
			})) // End of relationships delete verification
			// Verify attributes write empty
			Expect(bundle2.GetOperations()[0].AttributesWrite).Should(BeEmpty()) // Verify empty
			// Verify attributes delete empty
			Expect(bundle2.GetOperations()[0].AttributesDelete).Should(BeEmpty()) // Verify empty
		}) // End of test case
	}) // End of Write context
	// Test cases for Delete operation
	Context("Delete", func() { // Delete test context
		It("should delete DataBundles Correctly", func() { // Test delete operation
			ctx := context.Background() // Create context
			// Define bundles for delete test
			bundles := []*base.DataBundle{ // Test bundles
				{ // First bundle
					Name: "user_created", // Bundle name
					Arguments: []string{ // Bundle arguments
						"organizationID", // Organization ID
						"companyID",      // Company ID
						"userID",         // User ID
					}, // End of arguments
					Operations: []*base.Operation{ // Bundle operations
						{ // Operation definition
							RelationshipsWrite: []string{ // Relationships to write
								"organization:{{.organizationID}}#member@company:{{.companyID}}#admin", // Member relationship
								"organization:{{.organizationID}}#member@user:{{.userID}}",             // Member relationship
								"organization:{{.organizationID}}#admin@user:{{.userID}}",              // Admin relationship
							}, // End of relationships write
							RelationshipsDelete: []string{ // Relationships to delete
								"company:{{.companyID}}#admin@user:{{.userID}}", // Admin relationship
							}, // End of relationships delete
							AttributesWrite: []string{ // Attributes to write
								"organization:{{.organizationID}}$public|boolean:true", // Public attribute
							}, // End of attributes write
							AttributesDelete: []string{ // Attributes to delete
								"organization:{{.organizationID}}$balance|double:120.900", // Balance attribute
							}, // End of attributes delete
						}, // End of operation
					}, // End of operations
				}, // End of first bundle
				{ // Second bundle
					Name: "user_deleted", // Bundle name
					Arguments: []string{ // Bundle arguments
						"organizationID", // Organization ID
						"companyID",      // Company ID
					}, // End of arguments
					Operations: []*base.Operation{ // Bundle operations
						{ // Operation definition
							RelationshipsWrite: []string{ // Relationships to write
								"organization:{{.organizationID}}#member@company:{{.companyID}}#admin", // Member relationship
							}, // End of relationships write
							RelationshipsDelete: []string{}, // Empty relationships delete
							AttributesWrite:     []string{}, // Empty attributes write
							AttributesDelete:    []string{}, // Empty attributes delete
						}, // End of operation
					}, // End of operations
				}, // End of second bundle
			} // End of bundles
			// Convert to storage bundles
			var sBundles []storage.Bundle // Storage bundles
			for _, b := range bundles {   // Iterate bundles
				sBundles = append(sBundles, storage.Bundle{ // Append bundle
					Name:       b.Name, // Bundle name
					DataBundle: b,      // Data bundle
					TenantID:   "t1",   // Tenant ID
				}) // End of append
			} // End of iteration
			// Write bundles
			names, err := bundleWriter.Write(ctx, sBundles)                       // Write bundles
			Expect(err).ShouldNot(HaveOccurred())                                 // No error expected
			Expect(names).Should(Equal([]string{"user_created", "user_deleted"})) // Verify names
			// Verify bundles exist before delete
			_, err = bundleReader.Read(ctx, "t1", "user_created") // Read first bundle
			Expect(err).ShouldNot(HaveOccurred())                 // Should exist
			// Delete first bundle
			err = bundleWriter.Delete(ctx, "t1", "user_created") // Delete bundle
			Expect(err).ShouldNot(HaveOccurred())                // No error expected
			// Verify first bundle deleted
			_, err = bundleReader.Read(ctx, "t1", "user_created")                                  // Try to read deleted bundle
			Expect(err.Error()).Should(Equal(base.ErrorCode_ERROR_CODE_BUNDLE_NOT_FOUND.String())) // Should not exist
			// Verify second bundle still exists
			_, err = bundleReader.Read(ctx, "t1", "user_deleted") // Read second bundle
			Expect(err).ShouldNot(HaveOccurred())                 // Should still exist
		}) // End of test case
	}) // End of Delete context
}) // End of BundleWriter test suite
