package commands

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Permify/permify/internal/repositories/entities"
	"github.com/Permify/permify/pkg/logger"
)

var _ = Describe("schema-lookup-command", func() {
	var schemaLookupCommand *SchemaLookupCommand
	l := logger.New("debug")

	// DRIVE SAMPLE

	driveConfigs := []entities.EntityConfig{
		{
			Entity:           "user",
			SerializedConfig: []byte("entity user {}"),
		},
		{
			Entity:           "organization",
			SerializedConfig: []byte("entity organization {\nrelation admin @user\n}"),
		},
		{
			Entity:           "folder",
			SerializedConfig: []byte("entity folder {\n relation\tparent\t@organization\nrelation\tcreator\t@user\nrelation\tcollaborator\t@user\n action read = collaborator\naction update = collaborator\naction delete = creator or parent.admin\n}"),
		},
		{
			Entity:           "doc",
			SerializedConfig: []byte("entity doc {\nelation\tparent\t@organization\nrelation\towner\t@user\n  action read = (owner or parent.collaborator) or parent.admin\naction update = owner and parent.admin\n action delete = owner or parent.admin\n}"),
		},
	}

	Context("Drive Sample: Schema Lookup", func() {
		It("Drive Sample: Case 1", func() {
			schemaLookupCommand = NewSchemaLookupCommand(l)
			re := &SchemaLookupQuery{
				Relations: []string{"creator"},
			}

			sch, _ := entities.EntityConfigs(driveConfigs).ToSchema()
			en, _ := sch.GetEntityByName("folder")

			actualResult, err := schemaLookupCommand.Execute(context.Background(), re, en.Actions)
			Expect(err).ShouldNot(HaveOccurred())
			Expect([]string{"delete"}).Should(Equal(actualResult.ActionNames))
		})

		It("Drive Sample: Case 2", func() {
			schemaLookupCommand = NewSchemaLookupCommand(l)
			re := &SchemaLookupQuery{
				Relations: []string{"owner", "parent.admin"},
			}

			sch, _ := entities.EntityConfigs(driveConfigs).ToSchema()
			en, _ := sch.GetEntityByName("doc")

			actualResult, err := schemaLookupCommand.Execute(context.Background(), re, en.Actions)
			Expect(err).ShouldNot(HaveOccurred())
			Expect([]string{"read", "update", "delete"}).Should(Equal(actualResult.ActionNames))
		})
	})

	// GITHUB SAMPLE

	githubConfigs := []entities.EntityConfig{
		{
			Entity:           "user",
			SerializedConfig: []byte("entity user {}"),
		},
		{
			Entity:           "organization",
			SerializedConfig: []byte("entity organization {\nrelation admin @user\nrelation member @user\naction create_repository = admin or member\naction delete = admin\n}"),
		},
		{
			Entity:           "repository",
			SerializedConfig: []byte("entity repository {\nrelation parent @organization\n relation owner @user\n  action push   = owner\n    action read   = owner and (parent.admin or parent.member)\n    action delete = parent.member and (parent.admin or owner)\n}"),
		},
	}

	Context("Github Sample: Schema Lookup", func() {
		It("Github Sample: Case 1", func() {
			schemaLookupCommand = NewSchemaLookupCommand(l)
			re := &SchemaLookupQuery{
				Relations: []string{"admin"},
			}

			sch, _ := entities.EntityConfigs(githubConfigs).ToSchema()
			en, _ := sch.GetEntityByName("organization")

			actualResult, err := schemaLookupCommand.Execute(context.Background(), re, en.Actions)
			Expect(err).ShouldNot(HaveOccurred())
			Expect([]string{"create_repository", "delete"}).Should(Equal(actualResult.ActionNames))
		})

		It("Github Sample: Case 2", func() {
			schemaLookupCommand = NewSchemaLookupCommand(l)
			re := &SchemaLookupQuery{
				Relations: []string{"parent.admin", "parent.member"},
			}

			sch, _ := entities.EntityConfigs(githubConfigs).ToSchema()
			en, _ := sch.GetEntityByName("repository")

			actualResult, err := schemaLookupCommand.Execute(context.Background(), re, en.Actions)
			Expect(err).ShouldNot(HaveOccurred())
			Expect([]string{"delete"}).Should(Equal(actualResult.ActionNames))
		})
	})
})
