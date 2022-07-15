package services

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Permify/permify/internal/entities"
	"github.com/Permify/permify/internal/repositories/postgres/mocks"
	"github.com/Permify/permify/pkg/dsl/parser"
	`github.com/Permify/permify/pkg/tuple`
)

var _ = Describe("permission-service", func() {
	var permissionService *PermissionService

	driveSample := parser.TranslateToSchema("entity user {}\t`table:user|identifier:id`\nentity organization {\nrelation admin @user\t`rel:custom`\n} `table:organization|identifier:id`\nentity folder {\n relation\tparent\t@organization    `rel:belongs-to|cols:organization_id`\n    relation\tcreator\t@user  `rel:belongs-to|cols:creator_id`\nrelation\tcollaborator\t@user `rel:many-to-many|table:folder_collaborator|cols:folder_id,user_id`\n action read = collaborator\n    action update = collaborator\n    action delete = creator or parent.admin\n\n} `table:folder|identifier:id`\n\nentity doc {\n\n    relation\tparent\t\t\t@organization    `rel:belongs-to|cols:organization_id`\n    relation\towner\t\t\t@user            `rel:belongs-to|cols:owner_id`\n\n    action read = (owner or parent.collaborator) and parent.admin\n    action update = owner and parent.admin\n    action delete = owner or parent.admin\n\n} `table:doc|identifier:id`\n")

	Context("Check", func() {

		It("Drive Sample: Case 1", func() {

			relationTupleRepository := new(mocks.RelationTupleRepository)

			getDocParent := []entities.RelationTuple{
				{
					Entity:          "doc",
					ObjectID:        "1",
					Relation:        "parent",
					UsersetEntity:   "folder",
					UsersetObjectID: "1",
					UsersetRelation: tuple.ELLIPSIS,
				},
			}

			getParentAdmins := []entities.RelationTuple{
				{
					Entity:          "folder",
					ObjectID:        "1",
					Relation:        "admin",
					UsersetEntity:   "",
					UsersetObjectID: "1",
					UsersetRelation: "",
				},
			}

			getParentCollaborators := []entities.RelationTuple{
				{
					Entity:          "folder",
					ObjectID:        "1",
					Relation:        "collaborator",
					UsersetEntity:   "",
					UsersetObjectID: "1",
					UsersetRelation: "",
				},
				{
					Entity:          "folder",
					ObjectID:        "1",
					Relation:        "collaborator",
					UsersetEntity:   "",
					UsersetObjectID: "3",
					UsersetRelation: "",
				},
			}

			getDocOwners := []entities.RelationTuple{
				{
					Entity:          "doc",
					ObjectID:        "1",
					Relation:        "owner",
					UsersetEntity:   "",
					UsersetObjectID: "2",
					UsersetRelation: "",
				},
			}

			relationTupleRepository.On("QueryTuples", "doc", "1", "parent").Return(getDocParent, nil).Once()
			relationTupleRepository.On("QueryTuples", "folder", "1", "collaborator").Return(getParentCollaborators, nil).Once()
			relationTupleRepository.On("QueryTuples", "doc", "1", "owner").Return(getDocOwners, nil).Once()
			relationTupleRepository.On("QueryTuples", "doc", "1", "parent").Return(getDocParent, nil).Once()
			relationTupleRepository.On("QueryTuples", "folder", "1", "admin").Return(getParentAdmins, nil).Once()

			permissionService = &PermissionService{
				schema:     driveSample,
				repository: relationTupleRepository,
			}

			actualResult, err := permissionService.Check(context.Background(), "1", "read", "doc:1", 3)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(true).Should(Equal(actualResult))

		})

		It("Drive Sample: Case 2", func() {

			relationTupleRepository := new(mocks.RelationTupleRepository)

			getDocParent := []entities.RelationTuple{
				{
					Entity:          "doc",
					ObjectID:        "1",
					Relation:        "parent",
					UsersetEntity:   "folder",
					UsersetObjectID: "1",
					UsersetRelation: tuple.ELLIPSIS,
				},
			}

			getParentAdmins := []entities.RelationTuple{
				{
					Entity:          "folder",
					ObjectID:        "1",
					Relation:        "admin",
					UsersetEntity:   "",
					UsersetObjectID: "2",
					UsersetRelation: "",
				},
			}

			getDocOwners := []entities.RelationTuple{
				{
					Entity:          "doc",
					ObjectID:        "1",
					Relation:        "owner",
					UsersetEntity:   "",
					UsersetObjectID: "1",
					UsersetRelation: "",
				},
			}

			relationTupleRepository.On("QueryTuples", "doc", "1", "parent").Return(getDocParent, nil).Once()
			relationTupleRepository.On("QueryTuples", "doc", "1", "owner").Return(getDocOwners, nil).Once()
			relationTupleRepository.On("QueryTuples", "doc", "1", "parent").Return(getDocParent, nil).Once()
			relationTupleRepository.On("QueryTuples", "folder", "1", "admin").Return(getParentAdmins, nil).Once()

			permissionService = &PermissionService{
				schema:     driveSample,
				repository: relationTupleRepository,
			}

			actualResult, err := permissionService.Check(context.Background(), "1", "update", "doc:1", 3)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(false).Should(Equal(actualResult))

		})

		It("Drive Sample: Case 3", func() {

			relationTupleRepository := new(mocks.RelationTupleRepository)

			getDocParent := []entities.RelationTuple{
				{
					Entity:          "doc",
					ObjectID:        "1",
					Relation:        "parent",
					UsersetEntity:   "folder",
					UsersetObjectID: "1",
					UsersetRelation: tuple.ELLIPSIS,
				},
			}

			getParentAdmins := []entities.RelationTuple{
				{
					Entity:          "folder",
					ObjectID:        "1",
					Relation:        "admin",
					UsersetEntity:   "",
					UsersetObjectID: "1",
					UsersetRelation: "",
				},
			}

			getParentCollaborators := []entities.RelationTuple{
				{
					Entity:          "folder",
					ObjectID:        "1",
					Relation:        "collaborator",
					UsersetEntity:   "",
					UsersetObjectID: "2",
					UsersetRelation: "",
				},
				{
					Entity:          "folder",
					ObjectID:        "1",
					Relation:        "collaborator",
					UsersetEntity:   "folder",
					UsersetObjectID: "1",
					UsersetRelation: "admin",
				},
			}

			getDocOwners := []entities.RelationTuple{
				{
					Entity:          "doc",
					ObjectID:        "1",
					Relation:        "owner",
					UsersetEntity:   "",
					UsersetObjectID: "2",
					UsersetRelation: "",
				},
			}

			relationTupleRepository.On("QueryTuples", "doc", "1", "parent").Return(getDocParent, nil).Once()
			relationTupleRepository.On("QueryTuples", "folder", "1", "collaborator").Return(getParentCollaborators, nil).Once()
			relationTupleRepository.On("QueryTuples", "folder", "1", "admin").Return(getParentAdmins, nil).Once()
			relationTupleRepository.On("QueryTuples", "doc", "1", "owner").Return(getDocOwners, nil).Once()
			relationTupleRepository.On("QueryTuples", "doc", "1", "parent").Return(getDocParent, nil).Once()
			relationTupleRepository.On("QueryTuples", "folder", "1", "admin").Return(getParentAdmins, nil).Once()

			permissionService = &PermissionService{
				schema:     driveSample,
				repository: relationTupleRepository,
			}

			actualResult, err := permissionService.Check(context.Background(), "1", "read", "doc:1", 4)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(true).Should(Equal(actualResult))

		})
	})
})
