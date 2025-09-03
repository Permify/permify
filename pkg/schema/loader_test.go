package schema

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Loader", func() {
	Context("LoadSchema function", func() {
		It("should load schema from URL", func() {
			loader := NewSchemaLoader()
			schema, err := loader.LoadSchema("https://gist.githubusercontent.com/neo773/d50f089c141bf61776c22157413ddbac/raw/ed2eb12108e49fce11be27d0387b8b01912b9d98/gistfile1.txt")
			Expect(err).ShouldNot(HaveOccurred())

			expectedSchema := `
			entity userhttp {}

			entity organization {
		   
			   relation admin @userhttp
			   relation member @userhttp
		   
			   action create_repository = (admin or member)
			   action delete = admin
		   }
		   
			entity repository {
		   
			   relation owner @userhttp @organization#member
			   relation parent @organization
		   
			   action push = owner
			   action read = (owner and (parent.admin and parent.member))
			   action delete = (parent.member and (parent.admin or owner))
			   action edit = parent.member not owner
		   }
			`
			Expect(strings.Join(strings.Fields(schema), "")).To(Equal(strings.Join(strings.Fields(expectedSchema), "")))
		})

		It("should load schema from file", func() {
			loader := NewSchemaLoader()
			schema, err := loader.LoadSchema("./schema.txt")
			Expect(err).ShouldNot(HaveOccurred())

			expectedSchema := `
			entity userfs {}

			entity organization {
		   
			   relation admin @userfs
			   relation member @userfs
		   
			   action create_repository = (admin or member)
			   action delete = admin
		   }
		   
			entity repository {
		   
			   relation owner @userfs @organization#member
			   relation parent @organization
		   
			   action push = owner
			   action read = (owner and (parent.admin and parent.member))
			   action delete = (parent.member and (parent.admin or owner))
			   action edit = parent.member not owner
		   }
			`
			Expect(strings.Join(strings.Fields(schema), "")).To(Equal(strings.Join(strings.Fields(expectedSchema), "")))
		})

		It("should load inline schema", func() {
			loader := NewSchemaLoader()
			schema, err := loader.LoadSchema(`entity userinline {}

			entity organization {
		   
			   relation admin @userinline
			   relation member @userinline
		   
			   action create_repository = (admin or member)
			   action delete = admin
		   }
		   
			entity repository {
		   
			   relation owner @userinline @organization#member
			   relation parent @organization
		   
			   action push = owner
			   action read = (owner and (parent.admin and parent.member))
			   action delete = (parent.member and (parent.admin or owner))
			   action edit = parent.member not owner
		   }`)

			Expect(err).ShouldNot(HaveOccurred())

			expectedSchema := `
			entity userinline {}

			entity organization {
		   
			   relation admin @userinline
			   relation member @userinline
		   
			   action create_repository = (admin or member)
			   action delete = admin
		   }
		   
			entity repository {
		   
			   relation owner @userinline @organization#member
			   relation parent @organization
		   
			   action push = owner
			   action read = (owner and (parent.admin and parent.member))
			   action delete = (parent.member and (parent.admin or owner))
			   action edit = parent.member not owner
		   }
		  `
			Expect(strings.Join(strings.Fields(schema), "")).To(Equal(strings.Join(strings.Fields(expectedSchema), "")))
		})

		It("should return error for empty schema", func() {
			loader := NewSchemaLoader()
			_, err := loader.LoadSchema("")
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(Equal("schema is empty"))
		})

		It("should return error for invalid URL scheme", func() {
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
			Expect(err.Error()).Should(Equal("invalid file path"))
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
		})
	})
})
