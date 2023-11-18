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
	})
})
