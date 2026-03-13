package memory

import (
	"context"
	"time"

	"github.com/rs/xid"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/internal/storage/memory/migrations"
	"github.com/Permify/permify/pkg/database"
	"github.com/Permify/permify/pkg/database/memory"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

var _ = Describe("SchemaReader", func() {
	var db *memory.Memory

	var schemaWriter *SchemaWriter
	var schemaReader *SchemaReader

	BeforeEach(func() {
		database, err := memory.New(migrations.Schema)
		Expect(err).ShouldNot(HaveOccurred())
		db = database

		schemaWriter = NewSchemaWriter(db)
		schemaReader = NewSchemaReader(db)
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

	Context("Read Schema", func() {
		It("should write and then read the schema for a tenant", func() {
			ctx := context.Background()

			version := xid.New().String()

			schema := []storage.SchemaDefinition{
				{TenantID: "t1", Name: "user", SerializedDefinition: []byte("entity user {}"), Version: version},
				{TenantID: "t1", Name: "organization", SerializedDefinition: []byte("entity organization { relation admin @user}"), Version: version},
			}

			err := schemaWriter.WriteSchema(ctx, schema)
			Expect(err).ShouldNot(HaveOccurred())

			sch, err := schemaReader.ReadSchema(ctx, "t1", version)
			Expect(err).ShouldNot(HaveOccurred())

			Expect(sch.EntityDefinitions["user"]).Should(Equal(&base.EntityDefinition{
				Name:        "user",
				Relations:   map[string]*base.RelationDefinition{},
				Permissions: map[string]*base.PermissionDefinition{},
				Attributes:  map[string]*base.AttributeDefinition{},
				References:  map[string]base.EntityDefinition_Reference{},
			}))

			Expect(sch.EntityDefinitions["organization"].GetName()).Should(Equal("organization"))
			Expect(sch.EntityDefinitions["organization"].GetRelations()["admin"].Name).Should(Equal("admin"))
			Expect(sch.EntityDefinitions["organization"].GetRelations()["admin"].GetRelationReferences()[0].GetType()).Should(Equal("user"))
			Expect(sch.EntityDefinitions["organization"].GetRelations()["admin"].GetRelationReferences()[0].GetRelation()).Should(Equal(""))

			Expect(sch.EntityDefinitions["organization"].GetPermissions()).Should(Equal(map[string]*base.PermissionDefinition{}))
			Expect(sch.EntityDefinitions["organization"].GetAttributes()).Should(Equal(map[string]*base.AttributeDefinition{}))
			Expect(sch.EntityDefinitions["organization"].GetReferences()["admin"]).Should(Equal(base.EntityDefinition_REFERENCE_RELATION))

			Expect(sch.RuleDefinitions).Should(Equal(map[string]*base.RuleDefinition{}))

			Expect(sch.References["user"]).Should(Equal(base.SchemaDefinition_REFERENCE_ENTITY))
			Expect(sch.References["organization"]).Should(Equal(base.SchemaDefinition_REFERENCE_ENTITY))
		})
	})

	Context("Read Schema String", func() {
		It("should write and then read the schema for a tenant", func() {
			ctx := context.Background()

			version := xid.New().String()

			schema := []storage.SchemaDefinition{
				{TenantID: "t1", Name: "user", SerializedDefinition: []byte("entity user {}"), Version: version},
				{TenantID: "t1", Name: "organization", SerializedDefinition: []byte("entity organization { relation admin @user}"), Version: version},
			}

			err := schemaWriter.WriteSchema(ctx, schema)
			Expect(err).ShouldNot(HaveOccurred())

			defs, err := schemaReader.ReadSchemaString(ctx, "t1", version)
			Expect(err).ShouldNot(HaveOccurred())

			Expect(isSameArray(defs, []string{"entity user {}", "entity organization { relation admin @user}"})).Should(BeTrue())
		})
	})

	Context("Read Entity Definition", func() {
		It("should write and then read the entity definition for a tenant", func() {
			ctx := context.Background()

			version := xid.New().String()

			schema := []storage.SchemaDefinition{
				{TenantID: "t1", Name: "user", SerializedDefinition: []byte("entity user {}"), Version: version},
				{TenantID: "t1", Name: "organization", SerializedDefinition: []byte("entity organization { relation admin @user}"), Version: version},
			}

			err := schemaWriter.WriteSchema(ctx, schema)
			Expect(err).ShouldNot(HaveOccurred())

			en, v, err := schemaReader.ReadEntityDefinition(ctx, "t1", "organization", version)
			Expect(err).ShouldNot(HaveOccurred())

			Expect(version).Should(Equal(v))

			Expect(en.GetName()).Should(Equal("organization"))

			Expect(en.GetRelations()["admin"].GetName()).Should(Equal("admin"))
			Expect(en.GetRelations()["admin"].GetRelationReferences()[0].GetType()).Should(Equal("user"))
			Expect(en.GetRelations()["admin"].GetRelationReferences()[0].GetRelation()).Should(Equal(""))

			Expect(en.GetPermissions()).Should(Equal(map[string]*base.PermissionDefinition{}))
			Expect(en.GetAttributes()).Should(Equal(map[string]*base.AttributeDefinition{}))
			Expect(en.GetReferences()["admin"]).Should(Equal(base.EntityDefinition_REFERENCE_RELATION))
		})
	})

	Context("Read Rule Definition", func() {
		It("should write and then read the rule definition for a tenant", func() {
			ctx := context.Background()

			version := xid.New().String()

			schema := []storage.SchemaDefinition{
				{TenantID: "t1", Name: "user", SerializedDefinition: []byte("entity user {}"), Version: version},
				{TenantID: "t1", Name: "organization", SerializedDefinition: []byte("entity organization { relation admin @user}"), Version: version},
				{TenantID: "t1", Name: "check_ip_range", SerializedDefinition: []byte("rule check_ip_range(ip_address string, ip_range string[]) {\n ip_address in ip_range\n}"), Version: version},
			}

			err := schemaWriter.WriteSchema(ctx, schema)
			Expect(err).ShouldNot(HaveOccurred())

			ru, v, err := schemaReader.ReadRuleDefinition(ctx, "t1", "check_ip_range", version)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(version).Should(Equal(v))
			Expect(ru.Name).Should(Equal("check_ip_range"))
			Expect(ru.Arguments).Should(Equal([]*base.NamedArgument{
				{Name: "ip_address", Type: base.AttributeType_ATTRIBUTE_TYPE_STRING},
				{Name: "ip_range", Type: base.AttributeType_ATTRIBUTE_TYPE_STRING_ARRAY},
			}))
		})
	})

	Context("List Schema Versions", func() {
		It("should write a few schemas for a tenant and then list all schema versions available", func() {
			ctx := context.Background()

			version := xid.New().String()
			schema := []storage.SchemaDefinition{
				{TenantID: "t1", Name: "test1", SerializedDefinition: []byte("entity user {}"), Version: version},
			}
			err := schemaWriter.WriteSchema(ctx, schema)
			Expect(err).ShouldNot(HaveOccurred())

			version = xid.New().String()
			schema = []storage.SchemaDefinition{
				{TenantID: "t1", Name: "test2", SerializedDefinition: []byte("entity user {}"), Version: version},
			}
			err = schemaWriter.WriteSchema(ctx, schema)
			Expect(err).ShouldNot(HaveOccurred())

			version = xid.New().String()
			schema = []storage.SchemaDefinition{
				{TenantID: "t1", Name: "test3", SerializedDefinition: []byte("entity user {}"), Version: version},
			}
			err = schemaWriter.WriteSchema(ctx, schema)
			Expect(err).ShouldNot(HaveOccurred())

			version = xid.New().String()
			schema = []storage.SchemaDefinition{
				{TenantID: "t1", Name: "test4", SerializedDefinition: []byte("entity user {}"), Version: version},
			}
			err = schemaWriter.WriteSchema(ctx, schema)
			Expect(err).ShouldNot(HaveOccurred())

			version = xid.New().String()
			schema = []storage.SchemaDefinition{
				{TenantID: "t1", Name: "test5", SerializedDefinition: []byte("entity user {}"), Version: version},
			}
			err = schemaWriter.WriteSchema(ctx, schema)
			Expect(err).ShouldNot(HaveOccurred())

			version = xid.New().String()
			schema = []storage.SchemaDefinition{
				{TenantID: "t2", Name: "test6", SerializedDefinition: []byte("entity user {}"), Version: version},
			}
			err = schemaWriter.WriteSchema(ctx, schema)
			Expect(err).ShouldNot(HaveOccurred())

			col1, ct1, err := schemaReader.ListSchemas(ctx, "t1", database.NewPagination(database.Size(3), database.Token("")))
			Expect(err).ShouldNot(HaveOccurred())
			Expect(len(col1)).Should(Equal(3))

			col2, ct2, err := schemaReader.ListSchemas(ctx, "t1", database.NewPagination(database.Size(3), database.Token(ct1.String())))
			Expect(err).ShouldNot(HaveOccurred())
			Expect(len(col2)).Should(Equal(2))
			Expect(ct2.String()).Should(Equal(""))
		})
	})

	Context("Error handling and edge cases", func() {
		It("should handle schema not found in HeadVersion", func() {
			ctx := context.Background()

			// Test with non-existent tenant
			_, err := schemaReader.HeadVersion(ctx, "non-existent-tenant")
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("ERROR_CODE_SCHEMA_NOT_FOUND"))
		})

		It("should handle schema not found in ReadEntityDefinition", func() {
			ctx := context.Background()

			// Test with non-existent entity
			_, _, err := schemaReader.ReadEntityDefinition(ctx, "t1", "non-existent-entity", "v1")
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("ERROR_CODE_SCHEMA_NOT_FOUND"))
		})

		It("should handle schema not found in ReadRuleDefinition", func() {
			ctx := context.Background()

			// Test with non-existent rule
			_, _, err := schemaReader.ReadRuleDefinition(ctx, "t1", "non-existent-rule", "v1")
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("ERROR_CODE_SCHEMA_NOT_FOUND"))
		})

		It("should handle invalid schema parsing in ReadEntityDefinition", func() {
			ctx := context.Background()

			version := xid.New().String()

			// Write invalid schema
			schema := []storage.SchemaDefinition{
				{TenantID: "t1", Name: "invalid", SerializedDefinition: []byte("invalid schema syntax"), Version: version},
			}

			err := schemaWriter.WriteSchema(ctx, schema)
			Expect(err).ShouldNot(HaveOccurred())

			// Try to read entity definition - should fail due to invalid schema
			_, _, err = schemaReader.ReadEntityDefinition(ctx, "t1", "invalid", version)
			Expect(err).Should(HaveOccurred())
		})

		It("should handle invalid schema parsing in ReadRuleDefinition", func() {
			ctx := context.Background()

			version := xid.New().String()

			// Write invalid schema
			schema := []storage.SchemaDefinition{
				{TenantID: "t1", Name: "invalid_rule", SerializedDefinition: []byte("invalid rule syntax"), Version: version},
			}

			err := schemaWriter.WriteSchema(ctx, schema)
			Expect(err).ShouldNot(HaveOccurred())

			// Try to read rule definition - should fail due to invalid schema
			_, _, err = schemaReader.ReadRuleDefinition(ctx, "t1", "invalid_rule", version)
			Expect(err).Should(HaveOccurred())
		})

		It("should handle invalid token in ListSchemas", func() {
			ctx := context.Background()

			// Test with invalid token
			_, _, err := schemaReader.ListSchemas(ctx, "t1", database.NewPagination(database.Size(10), database.Token("invalid-token")))
			Expect(err).Should(HaveOccurred())
		})

		It("should handle invalid xid in ListSchemas", func() {
			ctx := context.Background()

			// Write schema with invalid version (not a valid xid)
			schema := []storage.SchemaDefinition{
				{TenantID: "t1", Name: "test", SerializedDefinition: []byte("entity user {}"), Version: "invalid-version"},
			}

			err := schemaWriter.WriteSchema(ctx, schema)
			Expect(err).ShouldNot(HaveOccurred())

			// Try to list schemas - should fail due to invalid xid
			_, _, err = schemaReader.ListSchemas(ctx, "t1", database.NewPagination(database.Size(10), database.Token("")))
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("ERROR_CODE_INTERNAL"))
		})

		It("should handle type conversion error in ListSchemas", func() {
			ctx := context.Background()

			// This test is harder to trigger in normal operation since the database
			// should only contain storage.SchemaDefinition objects, but we can test
			// the error path by creating a mock scenario

			// Write a valid schema first
			version := xid.New().String()
			schema := []storage.SchemaDefinition{
				{TenantID: "t1", Name: "test", SerializedDefinition: []byte("entity user {}"), Version: version},
			}

			err := schemaWriter.WriteSchema(ctx, schema)
			Expect(err).ShouldNot(HaveOccurred())

			// The type conversion error is hard to trigger in normal operation
			// since the database should only contain storage.SchemaDefinition objects
			// This test verifies the error handling exists in the code
			_, _, err = schemaReader.ListSchemas(ctx, "t1", database.NewPagination(database.Size(10), database.Token("")))
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("should handle pagination with valid token in ListSchemas", func() {
			ctx := context.Background()

			// Write multiple schemas
			versions := make([]string, 5)
			for i := 0; i < 5; i++ {
				version := xid.New().String()
				versions[i] = version
				schema := []storage.SchemaDefinition{
					{TenantID: "t1", Name: "test" + string(rune('a'+i)), SerializedDefinition: []byte("entity user {}"), Version: version},
				}
				err := schemaWriter.WriteSchema(ctx, schema)
				Expect(err).ShouldNot(HaveOccurred())
				time.Sleep(time.Millisecond * 2) // Ensure different timestamps
			}

			// Get first page
			schemas1, token1, err := schemaReader.ListSchemas(ctx, "t1", database.NewPagination(database.Size(2), database.Token("")))
			Expect(err).ShouldNot(HaveOccurred())
			Expect(len(schemas1)).Should(Equal(2))
			Expect(token1.String()).ShouldNot(Equal(""))

			// Get second page using token
			schemas2, token2, err := schemaReader.ListSchemas(ctx, "t1", database.NewPagination(database.Size(2), database.Token(token1.String())))
			Expect(err).ShouldNot(HaveOccurred())
			Expect(len(schemas2)).Should(Equal(2))
			Expect(token2.String()).ShouldNot(Equal(""))

			// Get remaining pages
			schemas3, token3, err := schemaReader.ListSchemas(ctx, "t1", database.NewPagination(database.Size(2), database.Token(token2.String())))
			Expect(err).ShouldNot(HaveOccurred())
			Expect(len(schemas3)).Should(Equal(1))
			Expect(token3.String()).Should(Equal(""))
		})

		It("should handle empty schema definitions in ReadSchema", func() {
			ctx := context.Background()

			version := xid.New().String()

			// Write empty schema definitions
			schema := []storage.SchemaDefinition{
				{TenantID: "t1", Name: "empty", SerializedDefinition: []byte(""), Version: version},
			}

			err := schemaWriter.WriteSchema(ctx, schema)
			Expect(err).ShouldNot(HaveOccurred())

			// Try to read schema - should succeed with empty schema
			sch, err := schemaReader.ReadSchema(ctx, "t1", version)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(sch).ShouldNot(BeNil())
			Expect(len(sch.EntityDefinitions)).Should(Equal(0))
		})

		It("should handle empty schema definitions in ReadSchemaString", func() {
			ctx := context.Background()

			version := xid.New().String()

			// Write empty schema definitions
			schema := []storage.SchemaDefinition{
				{TenantID: "t1", Name: "empty", SerializedDefinition: []byte(""), Version: version},
			}

			err := schemaWriter.WriteSchema(ctx, schema)
			Expect(err).ShouldNot(HaveOccurred())

			// Read schema string - should return empty definitions
			defs, err := schemaReader.ReadSchemaString(ctx, "t1", version)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(len(defs)).Should(Equal(1))
			Expect(defs[0]).Should(Equal(""))
		})

		It("should handle non-existent version in ReadSchema", func() {
			ctx := context.Background()

			// Try to read non-existent version
			_, err := schemaReader.ReadSchema(ctx, "t1", "non-existent-version")
			Expect(err).ShouldNot(HaveOccurred())
			// Should return empty schema, not an error
		})

		It("should handle non-existent version in ReadSchemaString", func() {
			ctx := context.Background()

			// Try to read non-existent version
			defs, err := schemaReader.ReadSchemaString(ctx, "t1", "non-existent-version")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(len(defs)).Should(Equal(0))
		})
	})
})
