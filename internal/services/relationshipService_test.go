package services

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Permify/permify/internal/entities"
	"github.com/Permify/permify/internal/repositories/mocks"
)

var _ = Describe("relationship-service", func() {
	var relationshipService *RelationshipService

	Context("WriteRelationship", func() {
		It("Write", func() {
			relationTupleRepository := new(mocks.RelationTupleRepository)

			tuple := entities.RelationTuple{
				Entity:          "organization",
				ObjectID:        "1",
				Relation:        "admin",
				UsersetEntity:   "user",
				UsersetObjectID: "1",
				UsersetRelation: "",
			}

			relationTupleRepository.On("Write", tuple).Return(nil).Once()

			relationshipService = &RelationshipService{
				rt: relationTupleRepository,
			}

			err := relationshipService.WriteRelationship(context.Background(), tuple)
			Expect(err).ShouldNot(HaveOccurred())
		})
	})

	Context("DeleteRelationship", func() {
		It("Write", func() {
			relationTupleRepository := new(mocks.RelationTupleRepository)

			tuple := entities.RelationTuple{
				Entity:          "organization",
				ObjectID:        "1",
				Relation:        "member",
				UsersetEntity:   "user",
				UsersetObjectID: "2",
				UsersetRelation: "",
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
