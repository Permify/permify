package services

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Permify/permify/internal/repositories/mocks"
	"github.com/Permify/permify/pkg/tuple"
)

var _ = Describe("relationship-service", func() {
	var relationshipService *RelationshipService

	Context("WriteRelationship", func() {
		It("Write", func() {
			relationTupleRepository := new(mocks.RelationTupleRepository)

			tuple := tuple.Tuple{
				Entity: tuple.Entity{
					Type: "organization",
					ID:   "1",
				},
				Relation: "admin",
				Subject: tuple.Subject{
					Type: "user",
					ID:   "1",
				},
			}

			relationTupleRepository.On("Write", tuple).Return(nil).Once()

			relationshipService = &RelationshipService{
				rt: relationTupleRepository,
			}

			err := relationshipService.WriteRelationship(context.Background(), tuple, "")
			Expect(err).ShouldNot(HaveOccurred())
		})
	})

	Context("DeleteRelationship", func() {
		It("Write", func() {
			relationTupleRepository := new(mocks.RelationTupleRepository)

			tuple := tuple.Tuple{
				Entity: tuple.Entity{
					Type: "organization",
					ID:   "1",
				},
				Relation: "member",
				Subject: tuple.Subject{
					Type: "user",
					ID:   "2",
				},
			}

			relationTupleRepository.On("Delete", tuple).Return(nil).Once()

			relationshipService = &RelationshipService{
				rt: relationTupleRepository,
			}

			err := relationshipService.DeleteRelationship(context.Background(), tuple)
			Expect(err).ShouldNot(HaveOccurred())
		})
	})
})
