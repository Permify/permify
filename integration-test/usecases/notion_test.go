package usecases

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Permify/permify/integration-test/usecases/shapes"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/tuple"
)

var _ = Describe("notion-test", func() {
	ctx := context.Background()

	Context("Notion Sample: Assertions", func() {
		It("Notion Sample: Checks", func() {
			for _, scenario := range shapes.InitialNotionShape.Scenarios {
				for _, check := range scenario.Checks {

					entity, err := tuple.E(check.Entity)
					if err != nil {
						Expect(err).ShouldNot(HaveOccurred())
					}

					ear, err := tuple.EAR(check.Subject)
					if err != nil {
						Expect(err).ShouldNot(HaveOccurred())
					}

					subject := &base.Subject{
						Type:     ear.GetEntity().GetType(),
						Id:       ear.GetEntity().GetId(),
						Relation: ear.GetRelation(),
					}

					for permission, expected := range check.Assertions {
						exp := base.PermissionCheckResponse_RESULT_ALLOWED
						if !expected {
							exp = base.PermissionCheckResponse_RESULT_DENIED
						}

						res, err := permissionClient.Check(ctx, &base.PermissionCheckRequest{
							TenantId: "notion",
							Metadata: &base.PermissionCheckRequestMetadata{
								SchemaVersion: initialNotionSchemaVersion,
								SnapToken:     initialNotionSnapToken,
								Depth:         100,
							},
							Entity:     entity,
							Permission: permission,
							Subject:    subject,
						})

						Expect(err).ShouldNot(HaveOccurred())
						Expect(res.Can).Should(Equal(exp))
					}
				}
			}
		})

		It("Notion Sample: Entity Filtering", func() {
			for _, scenario := range shapes.InitialNotionShape.Scenarios {
				for _, filter := range scenario.EntityFilters {

					ear, err := tuple.EAR(filter.Subject)
					if err != nil {
						Expect(err).ShouldNot(HaveOccurred())
					}

					subject := &base.Subject{
						Type:     ear.GetEntity().GetType(),
						Id:       ear.GetEntity().GetId(),
						Relation: ear.GetRelation(),
					}

					for permission, expected := range filter.Assertions {
						res, err := permissionClient.LookupEntity(ctx, &base.PermissionLookupEntityRequest{
							TenantId: "notion",
							Metadata: &base.PermissionLookupEntityRequestMetadata{
								SchemaVersion: initialNotionSchemaVersion,
								SnapToken:     initialNotionSnapToken,
								Depth:         100,
							},
							EntityType: filter.EntityType,
							Permission: permission,
							Subject:    subject,
						})

						Expect(err).ShouldNot(HaveOccurred())
						Expect(IsSameArray(res.GetEntityIds(), expected)).Should(Equal(true))
					}
				}
			}
		})

		It("Notion Sample: Subject Filtering", func() {
			for _, scenario := range shapes.InitialNotionShape.Scenarios {
				for _, filter := range scenario.SubjectFilters {
					subjectReference := tuple.RelationReference(filter.SubjectReference)

					entity, err := tuple.E(filter.Entity)
					if err != nil {
						Expect(err).ShouldNot(HaveOccurred())
					}

					for permission, expected := range filter.Assertions {

						res, err := permissionClient.LookupSubject(ctx, &base.PermissionLookupSubjectRequest{
							TenantId: "notion",
							Metadata: &base.PermissionLookupSubjectRequestMetadata{
								SchemaVersion: initialNotionSchemaVersion,
								SnapToken:     initialNotionSnapToken,
							},
							SubjectReference: subjectReference,
							Permission:       permission,
							Entity:           entity,
						})

						Expect(err).ShouldNot(HaveOccurred())
						Expect(IsSameArray(res.GetSubjectIds(), expected)).Should(Equal(true))
					}
				}
			}
		})
	})
})
