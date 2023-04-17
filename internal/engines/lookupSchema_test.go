package engines

import (
	"context"

	"golang.org/x/exp/slices"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Permify/permify/internal/repositories/mocks"
	"github.com/Permify/permify/internal/schema"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

var _ = Describe("lookup-schema-engine", func() {
	var lookupSchemaEngine *LookupSchemaEngine

	// DRIVE SAMPLE

	driveSchema := `
	entity user {}
	
	entity organization {
		relation admin @user
	}
	
	entity folder {
		relation org @organization
		relation creator @user
		relation collaborator @user
	
		action read = collaborator
		action update = collaborator
		action delete = creator or org.admin
	}
	
	entity doc {
		relation org @organization
		relation parent @folder
		relation owner @user
	
		action read = (owner or parent.collaborator) or org.admin
		action update = owner and org.admin
		action delete = owner or org.admin
	}
	`

	Context("Drive Sample: Lookup Schema", func() {
		It("Drive Sample: Case 1", func() {
			var err error

			// SCHEMA

			schemaReader := new(mocks.SchemaReader)

			var sch *base.SchemaDefinition
			sch, err = schema.NewSchemaFromStringDefinitions(true, driveSchema)
			Expect(err).ShouldNot(HaveOccurred())

			var en *base.EntityDefinition
			en, err = schema.GetEntityByName(sch, "folder")
			Expect(err).ShouldNot(HaveOccurred())

			schemaReader.On("ReadSchemaDefinition", "t1", "folder", "noop").Return(en, "noop", nil).Times(1)

			lookupSchemaEngine = NewLookupSchemaEngine(schemaReader)

			req := &base.PermissionLookupSchemaRequest{
				TenantId:      "t1",
				EntityType:    "folder",
				RelationNames: []string{"creator"},
				Metadata: &base.PermissionLookupSchemaRequestMetadata{
					SchemaVersion: "noop",
				},
			}

			actualResult, err := lookupSchemaEngine.Run(context.Background(), req)
			Expect(err).ShouldNot(HaveOccurred())
			for _, actionName := range actualResult.PermissionNames {
				Expect(slices.Contains([]string{"delete"}, actionName)).Should(Equal(slices.Contains([]string{"delete"}, actionName)))
			}
		})

		It("Drive Sample: Case 2", func() {
			var err error

			// SCHEMA

			schemaReader := new(mocks.SchemaReader)

			var sch *base.SchemaDefinition
			sch, err = schema.NewSchemaFromStringDefinitions(true, driveSchema)
			Expect(err).ShouldNot(HaveOccurred())

			var en *base.EntityDefinition
			en, err = schema.GetEntityByName(sch, "doc")
			Expect(err).ShouldNot(HaveOccurred())

			schemaReader.On("ReadSchemaDefinition", "t1", "doc", "noop").Return(en, "noop", nil).Times(2)

			lookupSchemaEngine = NewLookupSchemaEngine(schemaReader)

			req := &base.PermissionLookupSchemaRequest{
				TenantId:      "t1",
				EntityType:    "doc",
				RelationNames: []string{"owner", "org.admin"},
				Metadata: &base.PermissionLookupSchemaRequestMetadata{
					SchemaVersion: "noop",
				},
			}

			actualResult, err := lookupSchemaEngine.Run(context.Background(), req)
			Expect(err).ShouldNot(HaveOccurred())
			for _, actionName := range actualResult.PermissionNames {
				Expect(slices.Contains([]string{"read", "update", "delete"}, actionName)).Should(Equal(slices.Contains([]string{"read", "update", "delete"}, actionName)))
			}
		})
	})

	// GITHUB SAMPLE

	githubSchema := `
		entity user {}
	
		entity organization {
			relation admin @user
			relation member @user
	
			action create_repository = admin or member
			action delete = admin
		}
	
		entity repository {
			relation parent @organization
			relation owner @user
	
			action push   = owner
		    action read   = owner and (parent.admin or parent.member)
		    action delete = parent.member and (parent.admin or owner)
		}
		`

	Context("Github Sample: Lookup Schema", func() {
		It("Github Sample: Case 1", func() {
			var err error

			// SCHEMA

			schemaReader := new(mocks.SchemaReader)

			var sch *base.SchemaDefinition
			sch, err = schema.NewSchemaFromStringDefinitions(true, githubSchema)
			Expect(err).ShouldNot(HaveOccurred())

			var en *base.EntityDefinition
			en, err = schema.GetEntityByName(sch, "organization")
			Expect(err).ShouldNot(HaveOccurred())

			schemaReader.On("ReadSchemaDefinition", "t1", "organization", "noop").Return(en, "noop", nil).Times(2)

			lookupSchemaEngine = NewLookupSchemaEngine(schemaReader)

			req := &base.PermissionLookupSchemaRequest{
				TenantId:      "t1",
				EntityType:    "organization",
				RelationNames: []string{"admin"},
				Metadata: &base.PermissionLookupSchemaRequestMetadata{
					SchemaVersion: "noop",
				},
			}

			actualResult, err := lookupSchemaEngine.Run(context.Background(), req)
			Expect(err).ShouldNot(HaveOccurred())
			for _, actionName := range actualResult.PermissionNames {
				Expect(slices.Contains([]string{"create_repository", "delete"}, actionName)).Should(Equal(slices.Contains([]string{"create_repository", "delete"}, actionName)))
			}
		})

		It("Github Sample: Case 2", func() {
			var err error

			// SCHEMA

			schemaReader := new(mocks.SchemaReader)

			var sch *base.SchemaDefinition
			sch, err = schema.NewSchemaFromStringDefinitions(true, githubSchema)
			Expect(err).ShouldNot(HaveOccurred())

			var en *base.EntityDefinition
			en, err = schema.GetEntityByName(sch, "repository")
			Expect(err).ShouldNot(HaveOccurred())

			schemaReader.On("ReadSchemaDefinition", "t1", "repository", "noop").Return(en, "noop", nil).Times(2)

			lookupSchemaEngine = NewLookupSchemaEngine(schemaReader)

			req := &base.PermissionLookupSchemaRequest{
				TenantId:      "t1",
				EntityType:    "repository",
				RelationNames: []string{"parent.admin", "parent.member"},
				Metadata: &base.PermissionLookupSchemaRequestMetadata{
					SchemaVersion: "noop",
				},
			}

			actualResult, err := lookupSchemaEngine.Run(context.Background(), req)
			Expect(err).ShouldNot(HaveOccurred())
			for _, actionName := range actualResult.PermissionNames {
				Expect(slices.Contains([]string{"delete"}, actionName)).Should(Equal(slices.Contains([]string{"delete"}, actionName)))
			}
		})
	})
})
