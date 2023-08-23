package usecases

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Permify/permify/integration-test/usecases/shapes"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/tuple"
)

var _ = Describe("facebook-groups-test", func() {
	ctx := context.Background()

	Context("Facebook Groups Sample: Assertions", func() {
		It("Facebook Groups Sample: Checks", func() {
			for _, scenario := range shapes.InitialFacebookGroupsShape.Scenarios {
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

					var contextTuples []*base.Tuple

					for _, t := range check.Context.Tuples {
						tup, err := tuple.Tuple(t)
						if err != nil {
							Expect(err).ShouldNot(HaveOccurred())
						}

						contextTuples = append(contextTuples, tup)
					}

					for permission, expected := range check.Assertions {
						exp := base.CheckResult_CHECK_RESULT_ALLOWED
						if !expected {
							exp = base.CheckResult_CHECK_RESULT_DENIED
						}

						res, err := permissionClient.Check(ctx, &base.PermissionCheckRequest{
							TenantId: "facebook-groups",
							Metadata: &base.PermissionCheckRequestMetadata{
								SchemaVersion: initialFacebookGroupsSchemaVersion,
								SnapToken:     initialFacebookGroupsSnapToken,
								Depth:         100,
							},
							Context: &base.Context{
								Tuples: contextTuples,
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

		It("Facebook Groups Sample: Entity Filtering", func() {
			for _, scenario := range shapes.InitialFacebookGroupsShape.Scenarios {
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

					var contextTuples []*base.Tuple

					for _, t := range filter.Context.Tuples {
						tup, err := tuple.Tuple(t)
						if err != nil {
							Expect(err).ShouldNot(HaveOccurred())
						}

						contextTuples = append(contextTuples, tup)
					}

					for permission, expected := range filter.Assertions {
						res, err := permissionClient.LookupEntity(ctx, &base.PermissionLookupEntityRequest{
							TenantId: "facebook-groups",
							Metadata: &base.PermissionLookupEntityRequestMetadata{
								SchemaVersion: initialFacebookGroupsSchemaVersion,
								SnapToken:     initialFacebookGroupsSnapToken,
								Depth:         100,
							},
							Context: &base.Context{
								Tuples: contextTuples,
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

		It("Facebook Groups Sample: Subject Filtering", func() {
			for _, scenario := range shapes.InitialFacebookGroupsShape.Scenarios {
				for _, filter := range scenario.SubjectFilters {
					subjectReference := tuple.RelationReference(filter.SubjectReference)

					entity, err := tuple.E(filter.Entity)
					if err != nil {
						Expect(err).ShouldNot(HaveOccurred())
					}

					var contextTuples []*base.Tuple

					for _, t := range filter.Context.Tuples {
						tup, err := tuple.Tuple(t)
						if err != nil {
							Expect(err).ShouldNot(HaveOccurred())
						}

						contextTuples = append(contextTuples, tup)
					}

					for permission, expected := range filter.Assertions {

						res, err := permissionClient.LookupSubject(ctx, &base.PermissionLookupSubjectRequest{
							TenantId: "facebook-groups",
							Metadata: &base.PermissionLookupSubjectRequestMetadata{
								SchemaVersion: initialFacebookGroupsSchemaVersion,
								SnapToken:     initialFacebookGroupsSnapToken,
								Depth:         100,
							},
							Context: &base.Context{
								Tuples: contextTuples,
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
