package commands

import (
	`context`
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	`github.com/Permify/permify/internal/repositories/mocks`
	`github.com/Permify/permify/pkg/dsl/compiler`
	`github.com/Permify/permify/pkg/dsl/schema`
	`github.com/Permify/permify/pkg/logger`
	base `github.com/Permify/permify/pkg/pb/base/v1`
)

var _ = Describe("lookup-schema-command", func() {
	var schemaLookupCommand *LookupSchemaCommand
	l := logger.New("debug")

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

	Context("Drive Sample: Schema Lookup", func() {
		It("Drive Sample: Case 1", func() {
			var err error

			// SCHEMA

			schemaReader := new(mocks.SchemaReader)

			var sch *base.IndexedSchema
			sch, err = compiler.NewSchema(driveSchema)
			Expect(err).ShouldNot(HaveOccurred())

			var en *base.EntityDefinition
			en, err = schema.GetEntityByName(sch, "folder")
			Expect(err).ShouldNot(HaveOccurred())

			schemaReader.On("ReadSchemaDefinition", "folder", "noop").Return(en, "noop", nil).Times(1)

			schemaLookupCommand = NewLookupSchemaCommand(schemaReader, l)

			req := &base.PermissionLookupSchemaRequest{
				EntityType:    "folder",
				RelationNames: []string{"creator"},
				SchemaVersion: "noop",
			}

			actualResult, err := schemaLookupCommand.Execute(context.Background(), req)
			Expect(err).ShouldNot(HaveOccurred())
			Expect([]string{"delete"}).Should(Equal(actualResult.ActionNames))
		})

		It("Drive Sample: Case 2", func() {
			var err error

			// SCHEMA

			schemaReader := new(mocks.SchemaReader)

			var sch *base.IndexedSchema
			sch, err = compiler.NewSchema(driveSchema)
			Expect(err).ShouldNot(HaveOccurred())

			var en *base.EntityDefinition
			en, err = schema.GetEntityByName(sch, "doc")
			Expect(err).ShouldNot(HaveOccurred())

			schemaReader.On("ReadSchemaDefinition", "doc", "noop").Return(en, "noop", nil).Times(1)

			schemaLookupCommand = NewLookupSchemaCommand(schemaReader, l)

			req := &base.PermissionLookupSchemaRequest{
				EntityType:    "doc",
				RelationNames: []string{"owner", "org.admin"},
				SchemaVersion: "noop",
			}

			actualResult, err := schemaLookupCommand.Execute(context.Background(), req)
			Expect(err).ShouldNot(HaveOccurred())
			Expect([]string{"read", "update", "delete"}).Should(Equal(actualResult.ActionNames))

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

	Context("Github Sample: Schema Lookup", func() {
		It("Github Sample: Case 1", func() {
			var err error

			// SCHEMA

			schemaReader := new(mocks.SchemaReader)

			var sch *base.IndexedSchema
			sch, err = compiler.NewSchema(githubSchema)
			Expect(err).ShouldNot(HaveOccurred())

			var en *base.EntityDefinition
			en, err = schema.GetEntityByName(sch, "organization")
			Expect(err).ShouldNot(HaveOccurred())

			schemaReader.On("ReadSchemaDefinition", "organization", "noop").Return(en, "noop", nil).Times(1)

			schemaLookupCommand = NewLookupSchemaCommand(schemaReader, l)

			req := &base.PermissionLookupSchemaRequest{
				EntityType:    "organization",
				RelationNames: []string{"admin"},
				SchemaVersion: "noop",
			}

			actualResult, err := schemaLookupCommand.Execute(context.Background(), req)
			Expect(err).ShouldNot(HaveOccurred())
			Expect([]string{"create_repository", "delete"}).Should(Equal(actualResult.ActionNames))
		})

		It("Github Sample: Case 2", func() {
			var err error

			// SCHEMA

			schemaReader := new(mocks.SchemaReader)

			var sch *base.IndexedSchema
			sch, err = compiler.NewSchema(githubSchema)
			Expect(err).ShouldNot(HaveOccurred())

			var en *base.EntityDefinition
			en, err = schema.GetEntityByName(sch, "repository")
			Expect(err).ShouldNot(HaveOccurred())

			schemaReader.On("ReadSchemaDefinition", "repository", "noop").Return(en, "noop", nil).Times(1)

			schemaLookupCommand = NewLookupSchemaCommand(schemaReader, l)

			req := &base.PermissionLookupSchemaRequest{
				EntityType:    "repository",
				RelationNames: []string{"parent.admin", "parent.member"},
				SchemaVersion: "noop",
			}

			actualResult, err := schemaLookupCommand.Execute(context.Background(), req)
			Expect(err).ShouldNot(HaveOccurred())
			Expect([]string{"delete"}).Should(Equal(actualResult.ActionNames))
		})
	})
})
