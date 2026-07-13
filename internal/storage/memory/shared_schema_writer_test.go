package memory

import (
	"context"

	"github.com/rs/xid"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/internal/storage/memory/migrations"
	"github.com/Permify/permify/pkg/database/memory"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

var _ = Describe("SharedSchemaWriter", func() {
	var db *memory.Memory

	var sharedSchemaWriter *SharedSchemaWriter
	var sharedSchemaReader *SharedSchemaReader
	var tenantWriter *TenantWriter

	BeforeEach(func() {
		database, err := memory.New(migrations.Schema)
		Expect(err).ShouldNot(HaveOccurred())
		db = database

		sharedSchemaWriter = NewSharedSchemaWriter(db)
		sharedSchemaReader = NewSharedSchemaReader(db)
		tenantWriter = NewTenantWriter(db)
	})

	AfterEach(func() {
		err := db.Close()
		Expect(err).ShouldNot(HaveOccurred())
	})

	Context("WriteSharedSchema", func() {
		It("should write a shared schema and read it back", func() {
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

			Expect(sch.EntityDefinitions["user"]).Should(Equal(&base.EntityDefinition{
				Name:        "user",
				Relations:   map[string]*base.RelationDefinition{},
				Permissions: map[string]*base.PermissionDefinition{},
				Attributes:  map[string]*base.AttributeDefinition{},
				References:  map[string]base.EntityDefinition_Reference{},
			}))

			Expect(sch.EntityDefinitions["organization"].GetName()).Should(Equal("organization"))
			Expect(sch.EntityDefinitions["organization"].GetRelations()["admin"].GetName()).Should(Equal("admin"))
		})

		It("should update head version on write", func() {
			ctx := context.Background()
			version := xid.New().String()

			defs := []storage.SharedSchemaDefinition{
				{SharedSchemaID: "shared-1", Name: "user", SerializedDefinition: []byte("entity user {}"), Version: version},
			}

			err := sharedSchemaWriter.WriteSharedSchema(ctx, defs)
			Expect(err).ShouldNot(HaveOccurred())

			headVersion, err := sharedSchemaReader.SharedHeadVersion(ctx, "shared-1")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(headVersion).Should(Equal(version))
		})
	})

	Context("AssignSharedSchema", func() {
		It("should assign a shared schema to tenants", func() {
			ctx := context.Background()

			// Create tenants
			_, err := tenantWriter.CreateTenant(ctx, "t1", "Tenant 1")
			Expect(err).ShouldNot(HaveOccurred())
			_, err = tenantWriter.CreateTenant(ctx, "t2", "Tenant 2")
			Expect(err).ShouldNot(HaveOccurred())

			// Write shared schema
			version := xid.New().String()
			defs := []storage.SharedSchemaDefinition{
				{SharedSchemaID: "shared-1", Name: "user", SerializedDefinition: []byte("entity user {}"), Version: version},
			}
			err = sharedSchemaWriter.WriteSharedSchema(ctx, defs)
			Expect(err).ShouldNot(HaveOccurred())

			// Assign to both tenants
			err = sharedSchemaWriter.AssignSharedSchema(ctx, "shared-1", []string{"t1", "t2"})
			Expect(err).ShouldNot(HaveOccurred())

			// Verify assignment
			id1, err := sharedSchemaReader.GetTenantSharedSchemaID(ctx, "t1")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(id1).Should(Equal("shared-1"))

			id2, err := sharedSchemaReader.GetTenantSharedSchemaID(ctx, "t2")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(id2).Should(Equal("shared-1"))
		})

		It("should skip non-existent tenants without error", func() {
			ctx := context.Background()

			err := sharedSchemaWriter.AssignSharedSchema(ctx, "shared-1", []string{"non-existent"})
			Expect(err).ShouldNot(HaveOccurred())
		})
	})
})
