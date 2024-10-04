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
			Expect(ru.Arguments).Should(Equal(map[string]base.AttributeType{
				"ip_address": base.AttributeType_ATTRIBUTE_TYPE_STRING,
				"ip_range":   base.AttributeType_ATTRIBUTE_TYPE_STRING_ARRAY,
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
})
