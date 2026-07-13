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

var _ = Describe("SharedSchemaReader", func() {
	var db *memory.Memory

	var sharedSchemaWriter *SharedSchemaWriter
	var sharedSchemaReader *SharedSchemaReader
	var tenantWriter *TenantWriter
	var schemaReader *SchemaReader

	BeforeEach(func() {
		database, err := memory.New(migrations.Schema)
		Expect(err).ShouldNot(HaveOccurred())
		db = database

		sharedSchemaWriter = NewSharedSchemaWriter(db)
		sharedSchemaReader = NewSharedSchemaReader(db)
		tenantWriter = NewTenantWriter(db)
		schemaReader = NewSchemaReader(db)
	})

	AfterEach(func() {
		err := db.Close()
		Expect(err).ShouldNot(HaveOccurred())
	})

	Context("ReadSharedSchema", func() {
		It("should read a shared schema by ID and version", func() {
			ctx := context.Background()
			version := xid.New().String()

			defs := []storage.SharedSchemaDefinition{
				{SharedSchemaID: "shared-1", Name: "user", SerializedDefinition: []byte("entity user {}"), Version: version},
				{SharedSchemaID: "shared-1", Name: "organization", SerializedDefinition: []byte("entity organization { relation admin @user}"), Version: version},
			}

			err := sharedSchemaWriter.WriteSharedSchema(ctx, defs)
			Expect(err).ShouldNot(HaveOccurred())

			sch, err := sharedSchemaReader.ReadSharedSchema(ctx, "shared-1", version)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(sch.EntityDefinitions).Should(HaveLen(2))
			Expect(sch.EntityDefinitions).Should(HaveKey("user"))
			Expect(sch.EntityDefinitions).Should(HaveKey("organization"))
		})
	})

	Context("ReadSharedSchemaString", func() {
		It("should read a shared schema as string definitions", func() {
			ctx := context.Background()
			version := xid.New().String()

			defs := []storage.SharedSchemaDefinition{
				{SharedSchemaID: "shared-1", Name: "user", SerializedDefinition: []byte("entity user {}"), Version: version},
				{SharedSchemaID: "shared-1", Name: "organization", SerializedDefinition: []byte("entity organization { relation admin @user}"), Version: version},
			}

			err := sharedSchemaWriter.WriteSharedSchema(ctx, defs)
			Expect(err).ShouldNot(HaveOccurred())

			strs, err := sharedSchemaReader.ReadSharedSchemaString(ctx, "shared-1", version)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(isSameArray(strs, []string{"entity user {}", "entity organization { relation admin @user}"})).Should(BeTrue())
		})
	})

	Context("ReadSharedEntityDefinition", func() {
		It("should read a single entity definition from a shared schema", func() {
			ctx := context.Background()
			version := xid.New().String()

			defs := []storage.SharedSchemaDefinition{
				{SharedSchemaID: "shared-1", Name: "user", SerializedDefinition: []byte("entity user {}"), Version: version},
				{SharedSchemaID: "shared-1", Name: "organization", SerializedDefinition: []byte("entity organization { relation admin @user}"), Version: version},
			}

			err := sharedSchemaWriter.WriteSharedSchema(ctx, defs)
			Expect(err).ShouldNot(HaveOccurred())

			en, v, err := sharedSchemaReader.ReadSharedEntityDefinition(ctx, "shared-1", "organization", version)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(v).Should(Equal(version))
			Expect(en.GetName()).Should(Equal("organization"))
			Expect(en.GetRelations()["admin"].GetName()).Should(Equal("admin"))
		})

		It("should return error for non-existent entity", func() {
			ctx := context.Background()
			_, _, err := sharedSchemaReader.ReadSharedEntityDefinition(ctx, "shared-1", "non-existent", "v1")
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("ERROR_CODE_SCHEMA_NOT_FOUND"))
		})
	})

	Context("ReadSharedRuleDefinition", func() {
		It("should read a single rule definition from a shared schema", func() {
			ctx := context.Background()
			version := xid.New().String()

			defs := []storage.SharedSchemaDefinition{
				{SharedSchemaID: "shared-1", Name: "user", SerializedDefinition: []byte("entity user {}"), Version: version},
				{SharedSchemaID: "shared-1", Name: "check_ip_range", SerializedDefinition: []byte("rule check_ip_range(ip_address string, ip_range string[]) {\n ip_address in ip_range\n}"), Version: version},
			}

			err := sharedSchemaWriter.WriteSharedSchema(ctx, defs)
			Expect(err).ShouldNot(HaveOccurred())

			ru, v, err := sharedSchemaReader.ReadSharedRuleDefinition(ctx, "shared-1", "check_ip_range", version)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(v).Should(Equal(version))
			Expect(ru.Name).Should(Equal("check_ip_range"))
			Expect(ru.Arguments).Should(Equal(map[string]base.AttributeType{
				"ip_address": base.AttributeType_ATTRIBUTE_TYPE_STRING,
				"ip_range":   base.AttributeType_ATTRIBUTE_TYPE_STRING_ARRAY,
			}))
		})

		It("should return error for non-existent rule", func() {
			ctx := context.Background()
			_, _, err := sharedSchemaReader.ReadSharedRuleDefinition(ctx, "shared-1", "non-existent", "v1")
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("ERROR_CODE_SCHEMA_NOT_FOUND"))
		})
	})

	Context("SharedHeadVersion", func() {
		It("should return the latest version for a shared schema", func() {
			ctx := context.Background()

			var latestVersion string
			for i := 0; i < 3; i++ {
				version := xid.New().String()
				defs := []storage.SharedSchemaDefinition{
					{SharedSchemaID: "shared-1", Name: "user", SerializedDefinition: []byte("entity user {}"), Version: version},
				}
				err := sharedSchemaWriter.WriteSharedSchema(ctx, defs)
				Expect(err).ShouldNot(HaveOccurred())
				latestVersion = version
				time.Sleep(time.Millisecond * 2)
			}

			headVersion, err := sharedSchemaReader.SharedHeadVersion(ctx, "shared-1")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(headVersion).Should(Equal(latestVersion))
		})

		It("should return error for non-existent shared schema", func() {
			ctx := context.Background()
			_, err := sharedSchemaReader.SharedHeadVersion(ctx, "non-existent")
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("ERROR_CODE_SCHEMA_NOT_FOUND"))
		})
	})

	Context("ListSharedSchemas", func() {
		It("should list shared schemas with pagination", func() {
			ctx := context.Background()

			// Write 4 shared schemas
			for i := 0; i < 4; i++ {
				version := xid.New().String()
				schemaID := "shared-" + string(rune('a'+i))
				defs := []storage.SharedSchemaDefinition{
					{SharedSchemaID: schemaID, Name: "user", SerializedDefinition: []byte("entity user {}"), Version: version},
				}
				err := sharedSchemaWriter.WriteSharedSchema(ctx, defs)
				Expect(err).ShouldNot(HaveOccurred())
				time.Sleep(time.Millisecond * 2)
			}

			// First page
			page1, ct1, err := sharedSchemaReader.ListSharedSchemas(ctx, database.NewPagination(database.Size(2), database.Token("")))
			Expect(err).ShouldNot(HaveOccurred())
			Expect(page1).Should(HaveLen(2))
			Expect(ct1.String()).ShouldNot(Equal(""))

			// Second page
			page2, ct2, err := sharedSchemaReader.ListSharedSchemas(ctx, database.NewPagination(database.Size(2), database.Token(ct1.String())))
			Expect(err).ShouldNot(HaveOccurred())
			Expect(page2).Should(HaveLen(2))
			Expect(ct2.String()).Should(Equal(""))
		})

		It("should return all schemas when page size exceeds count", func() {
			ctx := context.Background()

			version := xid.New().String()
			defs := []storage.SharedSchemaDefinition{
				{SharedSchemaID: "shared-only-one", Name: "user", SerializedDefinition: []byte("entity user {}"), Version: version},
			}
			err := sharedSchemaWriter.WriteSharedSchema(ctx, defs)
			Expect(err).ShouldNot(HaveOccurred())

			schemas, ct, err := sharedSchemaReader.ListSharedSchemas(ctx, database.NewPagination(database.Size(100), database.Token("")))
			Expect(err).ShouldNot(HaveOccurred())
			Expect(len(schemas)).Should(BeNumerically(">=", 1))
			Expect(ct.String()).Should(Equal(""))
		})
	})

	Context("GetTenantSharedSchemaID", func() {
		It("should return empty string for tenant without shared schema", func() {
			ctx := context.Background()

			_, err := tenantWriter.CreateTenant(ctx, "t1", "Tenant 1")
			Expect(err).ShouldNot(HaveOccurred())

			id, err := sharedSchemaReader.GetTenantSharedSchemaID(ctx, "t1")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(id).Should(Equal(""))
		})

		It("should return shared schema ID after assignment", func() {
			ctx := context.Background()

			_, err := tenantWriter.CreateTenant(ctx, "t1", "Tenant 1")
			Expect(err).ShouldNot(HaveOccurred())

			version := xid.New().String()
			defs := []storage.SharedSchemaDefinition{
				{SharedSchemaID: "shared-1", Name: "user", SerializedDefinition: []byte("entity user {}"), Version: version},
			}
			err = sharedSchemaWriter.WriteSharedSchema(ctx, defs)
			Expect(err).ShouldNot(HaveOccurred())

			err = sharedSchemaWriter.AssignSharedSchema(ctx, "shared-1", []string{"t1"})
			Expect(err).ShouldNot(HaveOccurred())

			id, err := sharedSchemaReader.GetTenantSharedSchemaID(ctx, "t1")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(id).Should(Equal("shared-1"))
		})

		It("should return empty string for non-existent tenant", func() {
			ctx := context.Background()

			id, err := sharedSchemaReader.GetTenantSharedSchemaID(ctx, "non-existent")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(id).Should(Equal(""))
		})
	})

	Context("HeadVersion with shared schema", func() {
		It("should return shared schema version when tenant has shared schema assigned", func() {
			ctx := context.Background()

			// Create tenant
			_, err := tenantWriter.CreateTenant(ctx, "t1", "Tenant 1")
			Expect(err).ShouldNot(HaveOccurred())

			// Write shared schema
			version := xid.New().String()
			defs := []storage.SharedSchemaDefinition{
				{SharedSchemaID: "shared-1", Name: "user", SerializedDefinition: []byte("entity user {}"), Version: version},
			}
			err = sharedSchemaWriter.WriteSharedSchema(ctx, defs)
			Expect(err).ShouldNot(HaveOccurred())

			// Assign shared schema to tenant
			err = sharedSchemaWriter.AssignSharedSchema(ctx, "shared-1", []string{"t1"})
			Expect(err).ShouldNot(HaveOccurred())

			// HeadVersion should return shared schema ID and version
			sharedID, headVer, err := schemaReader.HeadVersion(ctx, "t1")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(sharedID).Should(Equal("shared-1"))
			Expect(headVer).Should(Equal(version))
		})
	})
})
