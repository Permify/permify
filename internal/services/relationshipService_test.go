package services

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Permify/permify/internal/entities"
	"github.com/Permify/permify/internal/repositories/postgres/mocks"
)

var _ = Describe("relationship-service", func() {
	var relationshipService *RelationshipService

	Context("WriteRelationship", func() {
		It("Write", func() {
			relationTupleRepository := new(mocks.RelationTupleRepository)

			tuples := []entities.RelationTuple{
				{
					Entity:          "organization",
					ObjectID:        "1",
					Relation:        "admin",
					UsersetEntity:   "",
					UsersetObjectID: "1",
					UsersetRelation: "",
				},
				{
					Entity:          "organization",
					ObjectID:        "1",
					Relation:        "member",
					UsersetEntity:   "",
					UsersetObjectID: "2",
					UsersetRelation: "",
				},
			}

			relationTupleRepository.On("Write", tuples).Return(nil).Once()

			relationshipService = &RelationshipService{
				repository: relationTupleRepository,
			}

			err := relationshipService.WriteRelationship(context.Background(), tuples)
			Expect(err).ShouldNot(HaveOccurred())
		})
	})

	Context("DeleteRelationship", func() {
		It("Write", func() {
			relationTupleRepository := new(mocks.RelationTupleRepository)

			tuples := []entities.RelationTuple{
				{
					Entity:          "organization",
					ObjectID:        "1",
					Relation:        "admin",
					UsersetEntity:   "",
					UsersetObjectID: "1",
					UsersetRelation: "",
				},
				{
					Entity:          "organization",
					ObjectID:        "1",
					Relation:        "member",
					UsersetEntity:   "",
					UsersetObjectID: "2",
					UsersetRelation: "",
				},
			}

			relationTupleRepository.On("Delete", tuples).Return(nil).Once()

			relationshipService = &RelationshipService{
				repository: relationTupleRepository,
			}

			err := relationshipService.DeleteRelationship(context.Background(), tuples)
			Expect(err).ShouldNot(HaveOccurred())
		})
	})
})
