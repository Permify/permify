package schema // Schema package tests
import (       // Package imports
	"strings" // String utilities for test assertions
	// Test frameworks
	. "github.com/onsi/ginkgo/v2" // BDD test framework
	. "github.com/onsi/gomega"    // Matcher library
)                                   // End imports
var _ = Describe("Loader", func() { // Loader test suite
	Context("LoadSchema function", func() { // LoadSchema tests
		It("should load schema from URL", func() { // Test URL loading
			schemaLoader := NewSchemaLoader()                                                                                                                                                                                                                                                                                                                                                                                                                                                                   // Create loader instance
			loadedSchema, loadErr := schemaLoader.LoadSchema("https://gist.githubusercontent.com/neo773/d50f089c141bf61776c22157413ddbac/raw/ed2eb12108e49fce11be27d0387b8b01912b9d98/gistfile1.txt")                                                                                                                                                                                                                                                                                                           // Load from URL
			Expect(loadErr).ShouldNot(HaveOccurred())                                                                                                                                                                                                                                                                                                                                                                                                                                                           // No error expected
			expectedSchemaDefinition := "entity userhttp {}\n\nentity organization {\nrelation admin @userhttp\nrelation member @userhttp\naction create_repository = (admin or member)\naction delete = admin\n}\n\nentity repository {\nrelation owner @userhttp @organization#member\nrelation parent @organization\naction push = owner\naction read = (owner and (parent.admin and parent.member))\naction delete = (parent.member and (parent.admin or owner))\naction edit = parent.member not owner\n}" // Expected schema
			Expect(strings.Join(strings.Fields(loadedSchema), "")).To(Equal(strings.Join(strings.Fields(expectedSchemaDefinition), "")))                                                                                                                                                                                                                                                                                                                                                                        // Compare schemas
		}) // End URL test
		// File loading test
		It("should load schema from file", func() { // Test file loading
			fileLoader := NewSchemaLoader()                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           // Create file loader
			fileSchema, fileErr := fileLoader.LoadSchema("./schema.txt")                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              // Load from file
			Expect(fileErr).ShouldNot(HaveOccurred())                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 // No error expected
			expectedFileSchema := "entity userfs {} // User entity for file schema tests\n// Organization section\nentity organization { // Organization entity\n // Relations\n  relation admin @userfs // Admin relation\n  relation member @userfs // Member relation\n // Actions\n  action create_repository = (admin or member) // Create repository action\n  action delete = admin // Delete action\n} // End organization\n// Repository section\nentity repository { // Repository entity\n // Relations\n  relation owner @userfs @organization#member // Owner relation\n  relation parent @organization // Parent relation\n // Actions\n  action push = owner // Push action\n  action read = (owner and (parent.admin and parent.member)) // Read action\n  action delete = (parent.member and (parent.admin or owner)) // Delete action\n  action edit = parent.member not owner // Edit action\n} // End repository" // Expected file schema with comments
			Expect(strings.Join(strings.Fields(fileSchema), "")).To(Equal(strings.Join(strings.Fields(expectedFileSchema), "")))                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      // Compare file schemas
		}) // End file test
		// Inline loading test
		It("should load inline schema", func() { // Test inline loading
			inlineLoader := NewSchemaLoader()                                                                                                                                                                                                                                                                                                                                                                                                                                                                       // Create inline loader
			inlineInput := "entity userinline {}\n\nentity organization {\nrelation admin @userinline\nrelation member @userinline\naction create_repository = (admin or member)\naction delete = admin\n}\n\nentity repository {\nrelation owner @userinline @organization#member\nrelation parent @organization\naction push = owner\naction read = (owner and (parent.admin and parent.member))\naction delete = (parent.member and (parent.admin or owner))\naction edit = parent.member not owner\n}"          // Inline schema input
			inlineSchema, inlineErr := inlineLoader.LoadSchema(inlineInput)                                                                                                                                                                                                                                                                                                                                                                                                                                         // Load inline schema
			Expect(inlineErr).ShouldNot(HaveOccurred())                                                                                                                                                                                                                                                                                                                                                                                                                                                             // No error expected
			expectedInlineSchema := "entity userinline {}\n\nentity organization {\nrelation admin @userinline\nrelation member @userinline\naction create_repository = (admin or member)\naction delete = admin\n}\n\nentity repository {\nrelation owner @userinline @organization#member\nrelation parent @organization\naction push = owner\naction read = (owner and (parent.admin and parent.member))\naction delete = (parent.member and (parent.admin or owner))\naction edit = parent.member not owner\n}" // Expected inline
			Expect(strings.Join(strings.Fields(inlineSchema), "")).To(Equal(strings.Join(strings.Fields(expectedInlineSchema), "")))                                                                                                                                                                                                                                                                                                                                                                                // Compare inline
		}) // End inline test

		It("should return error for empty schema", func() { // Test empty schema
			loader := NewSchemaLoader()
			_, err := loader.LoadSchema("")
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(Equal("schema is empty"))
		}) // End test
		It("should return error for invalid URL scheme", func() { // Test invalid URL
			loader := NewSchemaLoader()
			_, err := loader.LoadSchema("ftp://example.com/schema.txt")
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(Equal("invalid URL scheme"))
		})

		It("should treat invalid file paths as inline schema", func() {
			loader := NewSchemaLoader()
			// Since determineSchemaType falls back to Inline when file check fails,
			// this will be treated as inline schema and return the input as-is
			result, err := loader.LoadSchema("/absolute/path/schema.txt")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(result).Should(Equal("/absolute/path/schema.txt"))
		})

		It("should treat directory traversal paths as inline schema", func() {
			loader := NewSchemaLoader()
			// Since determineSchemaType falls back to Inline when file check fails,
			// this will be treated as inline schema and return the input as-is
			result, err := loader.LoadSchema("../../../etc/passwd")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(result).Should(Equal("../../../etc/passwd"))
		})

		It("should treat non-existent files as inline schema", func() {
			loader := NewSchemaLoader()
			// Since determineSchemaType falls back to Inline when file check fails,
			// this will be treated as inline schema and return the input as-is
			result, err := loader.LoadSchema("nonexistent.txt")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(result).Should(Equal("nonexistent.txt"))
		})

		It("should return error when loader function not found", func() {
			// Create a loader with missing functions to test the error path
			loader := &Loader{
				loaders: map[Type]func(string) (string, error){
					// Only include URL, missing File and Inline
					URL: loadFromURL,
				},
			}

			// This should trigger the "loader function not found" error
			// when determineSchemaType returns File or Inline
			_, err := loader.LoadSchema("entity user {}")
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("loader function not found"))
		})
	})

	Context("determineSchemaType function", func() {
		It("should return URL type for valid URL", func() {
			schemaType, err := determineSchemaType("https://example.com/schema.txt")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(schemaType).Should(Equal(URL))
		})

		It("should return File type for valid file path", func() {
			schemaType, err := determineSchemaType("./schema.txt")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(schemaType).Should(Equal(File))
		})

		It("should return Inline type for non-URL, non-file input", func() {
			schemaType, err := determineSchemaType("entity user {}")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(schemaType).Should(Equal(Inline))
		})

		It("should return Inline type when file check fails", func() {
			schemaType, err := determineSchemaType("nonexistent.txt")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(schemaType).Should(Equal(Inline))
		})
	})

	Context("isURL function", func() {
		It("should return true for valid HTTP URL", func() {
			Expect(isURL("http://example.com")).Should(BeTrue())
		})

		It("should return true for valid HTTPS URL", func() {
			Expect(isURL("https://example.com")).Should(BeTrue())
		})

		It("should return false for invalid URL", func() {
			Expect(isURL("not-a-url")).Should(BeFalse())
		})

		It("should return false for URL without scheme", func() {
			Expect(isURL("example.com")).Should(BeFalse())
		})

		It("should return false for URL without host", func() {
			Expect(isURL("http://")).Should(BeFalse())
		})
	})

	Context("isFilePath function", func() {
		It("should return true for existing file", func() {
			valid, err := isFilePath("./schema.txt")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(valid).Should(BeTrue())
		})

		It("should return false with 'file does not exist' error for non-existent file", func() {
			valid, err := isFilePath("nonexistent.txt")
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(Equal("file does not exist"))
			Expect(valid).Should(BeFalse())
		})

		It("should return false with 'permission denied' error for permission issues", func() {
			// This test might not work on all systems, but it's good to have
			valid, err := isFilePath("/root/restricted_file.txt")
			if err != nil && err.Error() == "permission denied" {
				Expect(valid).Should(BeFalse())
			}
		})
	})

	Context("loadFromURL function", func() {
		It("should return error for invalid URL", func() {
			_, err := loadFromURL("not-a-url")
			Expect(err).Should(HaveOccurred())
		})

		It("should return error for invalid URL scheme", func() {
			_, err := loadFromURL("ftp://example.com/schema.txt")
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(Equal("invalid URL scheme"))
		})

		It("should return error for non-existent URL", func() {
			_, err := loadFromURL("https://nonexistent-domain-12345.com/schema.txt")
			Expect(err).Should(HaveOccurred())
		})
	})

	Context("loadFromFile function", func() {
		It("should return error for absolute path", func() {
			_, err := loadFromFile("/absolute/path/schema.txt")
			Expect(err).Should(HaveOccurred())
			// Unix returns "invalid file path"; Windows may return path-not-found from os.ReadFile
			Expect(err.Error()).Should(Or(
				Equal("invalid file path"),
				ContainSubstring("cannot find the path"),
				ContainSubstring("no such file"),
			))
		})

		It("should return error for directory traversal", func() {
			_, err := loadFromFile("../../../etc/passwd")
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(Equal("invalid file path"))
		})

		It("should return error for non-existent file", func() {
			_, err := loadFromFile("nonexistent.txt")
			Expect(err).Should(HaveOccurred())
		})
	})

	Context("loadInline function", func() {
		It("should return error for empty schema", func() {
			_, err := loadInline("")
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(Equal("schema is empty"))
		})

		It("should return schema for non-empty input", func() {
			schema := "entity user {}"
			result, err := loadInline(schema)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(result).Should(Equal(schema))
		}) // End test case
	}) // End LoadSchema tests
}) // End Loader suite
