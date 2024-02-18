package postgres

import (
	"context"
	"os"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/pkg/database"
	PQDatabase "github.com/Permify/permify/pkg/database/postgres"
	"github.com/Permify/permify/pkg/dsl/compiler"
	"github.com/Permify/permify/pkg/dsl/parser"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/rs/xid"
)


var _ = Describe("SchemaUpdater", func() {
	var db database.Database
	var schemaWriter *SchemaWriter
	var schemaReader *SchemaReader
	var schemaUpdater *SchemaUpdater

	BeforeEach(func() {
		version := os.Getenv("POSTGRES_VERSION")

		if version == "" {
			version = "14"
		}

		db = postgresDB(version)
		schemaWriter = NewSchemaWriter(db.(*PQDatabase.Postgres))
		schemaReader = NewSchemaReader(db.(*PQDatabase.Postgres))
		schemaUpdater = NewSchemaUpdater(db.(*PQDatabase.Postgres))
	})

	AfterEach(func() {
		err := db.Close()
		Expect(err).ShouldNot(HaveOccurred())
	})

	Context("Head Version", func() {
		It("should retrieve the most recent schema version for a tenant", func() {
			ctx := context.Background()

			var mostRecentVersion string

			// Insert multiple schema versions for a single tenant
			for i := 0; i < 3; i++ {
				version := xid.New().String()
				schema := []storage.SchemaDefinition{
					{TenantID: "t1", Name: "organization", SerializedDefinition: []byte("entity organization {}"), Version: version},
					{TenantID: "t1", Name: "user", SerializedDefinition: []byte("entity user {}"), Version: version},
				}
				err := schemaWriter.WriteSchema(ctx, schema)
				Expect(err).ShouldNot(HaveOccurred())
				mostRecentVersion = version // Keep track of the last inserted version
				// Sleep to ensure the version is different (if versions are time-based)
				time.Sleep(time.Millisecond * 2)
			}

			// Attempt to retrieve the head version from SchemaReader
			headVersion, err := schemaReader.HeadVersion(ctx, "t1")
			Expect(err).ShouldNot(HaveOccurred())

			// Validate that the retrieved head version matches the most recently inserted version
			Expect(headVersion).Should(Equal(mostRecentVersion), "The retrieved head version should be the most recently written one.")
		})
	})

	Context("Update Schema", func() {
		It("should write, update and read the schema for a tenant", func() {
			ctx := context.Background()

			version := xid.New().String()

			schema := []storage.SchemaDefinition{
				{TenantID: "t1", Name: "user", SerializedDefinition: []byte("entity user {}"), Version: version},
				{TenantID: "t1", Name: "organization", SerializedDefinition: []byte("entity organization { relation admin @user \n relation deleter @user}"), Version: version},
			}

			err := schemaWriter.WriteSchema(ctx, schema)
			Expect(err).ShouldNot(HaveOccurred())

			definitions := map[string]map[string][]string{
				"organization": {
					"write": []string{"relation writer @user"},
					"update": []string{"relation admin @user @organization"},
					"delete": []string{"relation deleter @user"},
				},
			}

			sch, err := schemaUpdater.UpdateSchema(ctx, "t1", version, definitions)
			Expect(err).ShouldNot(HaveOccurred())

			schemaAst, err := parser.NewParser(strings.Join(sch, "\n")).Parse()
			Expect(err).ShouldNot(HaveOccurred())

			_, _, err = compiler.NewCompiler(true, schemaAst).Compile()
			Expect(err).ShouldNot(HaveOccurred())

			newVersion := xid.New().String()
			newSchema := make([]storage.SchemaDefinition, 0, len(schemaAst.Statements))
			for _, st := range schemaAst.Statements {
				newSchema = append(newSchema, storage.SchemaDefinition{
					TenantID:             "t1",
					Version:              newVersion,
					Name:                 st.GetName(),
					SerializedDefinition: []byte(st.String()),
				})
			}
			
			err = schemaWriter.WriteSchema(ctx, newSchema)
			Expect(err).ShouldNot(HaveOccurred())

			newSch, err := schemaReader.ReadSchema(ctx, "t1", newVersion)
			Expect(err).ShouldNot(HaveOccurred())

			Expect(newSch.EntityDefinitions["user"]).Should(Equal(&base.EntityDefinition{
				Name:        "user",
				Relations:   map[string]*base.RelationDefinition{},
				Permissions: map[string]*base.PermissionDefinition{},
				Attributes:  map[string]*base.AttributeDefinition{},
				References:  map[string]base.EntityDefinition_Reference{},
			}))

			Expect(newSch.EntityDefinitions["organization"].GetName()).Should(Equal("organization"))
			Expect(newSch.EntityDefinitions["organization"].GetRelations()["admin"].Name).Should(Equal("admin"))
			Expect(newSch.EntityDefinitions["organization"].GetRelations()["admin"].GetRelationReferences()[0].GetType()).Should(Equal("user"))
			Expect(newSch.EntityDefinitions["organization"].GetRelations()["admin"].GetRelationReferences()[1].GetType()).Should(Equal("organization"))
			Expect(newSch.EntityDefinitions["organization"].GetRelations()["admin"].GetRelationReferences()[0].GetRelation()).Should(Equal(""))
			Expect(newSch.EntityDefinitions["organization"].GetRelations()["writer"].Name).Should(Equal("writer"))
			Expect(newSch.EntityDefinitions["organization"].GetRelations()["writer"].GetRelationReferences()[0].GetType()).Should(Equal("user"))
			Expect(newSch.EntityDefinitions["organization"].GetRelations()["writer"].GetRelationReferences()[0].GetRelation()).Should(Equal(""))

			Expect(newSch.EntityDefinitions["organization"].GetPermissions()).Should(Equal(map[string]*base.PermissionDefinition{}))
			Expect(newSch.EntityDefinitions["organization"].GetAttributes()).Should(Equal(map[string]*base.AttributeDefinition{}))
			Expect(newSch.EntityDefinitions["organization"].GetReferences()["admin"]).Should(Equal(base.EntityDefinition_REFERENCE_RELATION))

			Expect(newSch.RuleDefinitions).Should(Equal(map[string]*base.RuleDefinition{}))

			Expect(newSch.References["user"]).Should(Equal(base.SchemaDefinition_REFERENCE_ENTITY))
			Expect(newSch.References["organization"]).Should(Equal(base.SchemaDefinition_REFERENCE_ENTITY))
		})
	})
})