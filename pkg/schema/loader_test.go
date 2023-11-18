package schema

import (
	"strings"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestLoader(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Loader Suite")
}

var _ = Describe("Loader", func() {
	Describe("LoadSchema function", func() {
		It("should load schema from URL", func() {
			loader := NewSchemaLoader()
			schema, _ := loader.LoadSchema("https://gist.githubusercontent.com/neo773/d50f089c141bf61776c22157413ddbac/raw/ed2eb12108e49fce11be27d0387b8b01912b9d98/gistfile1.txt")
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
			schema, _ := loader.LoadSchema("./schema.txt")
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
			schema, _ := loader.LoadSchema(`entity userinline {}

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
