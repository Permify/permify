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

	// DRIVE SAMPLE

	driveSchema := parser.TranslateToSchema("entity user {}\t`table:user|identifier:id`\nentity organization {\nrelation admin @user\t`rel:custom`\n} `table:organization|identifier:id`\nentity folder {\n relation\tparent\t@organization    `rel:belongs-to|cols:organization_id`\n    relation\tcreator\t@user  `rel:belongs-to|cols:creator_id`\nrelation\tcollaborator\t@user `rel:many-to-many|table:folder_collaborator|cols:folder_id,user_id`\n action read = collaborator\n    action update = collaborator\n    action delete = creator or parent.admin\n\n} `table:folder|identifier:id`\n\nentity doc {\n\n    relation\tparent\t\t\t@organization    `rel:belongs-to|cols:organization_id`\n    relation\towner\t\t\t@user            `rel:belongs-to|cols:owner_id`\n\n    action read = (owner or parent.collaborator) or parent.admin\n    action update = owner and parent.admin\n    action delete = owner or parent.admin\n\n} `table:doc|identifier:id`\n")

	Context("Drive Sample: Check", func() {

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

			relationTupleRepository.On("QueryTuples", "doc", "1", "parent").Return(getDocParent, nil).Times(2)
			relationTupleRepository.On("QueryTuples", "folder", "1", "collaborator").Return(getParentCollaborators, nil).Times(1)
			relationTupleRepository.On("QueryTuples", "doc", "1", "owner").Return(getDocOwners, nil).Times(1)
			relationTupleRepository.On("QueryTuples", "folder", "1", "admin").Return(getParentAdmins, nil).Times(1)

			permissionService = NewPermissionService(relationTupleRepository, driveSchema)

			actualResult, _, _, err := permissionService.Check(context.Background(), "1", "read", "doc:1", 8)
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

			relationTupleRepository.On("QueryTuples", "doc", "1", "parent").Return(getDocParent, nil).Times(2)
			relationTupleRepository.On("QueryTuples", "doc", "1", "owner").Return(getDocOwners, nil).Times(1)
			relationTupleRepository.On("QueryTuples", "folder", "1", "admin").Return(getParentAdmins, nil).Times(1)

			permissionService = NewPermissionService(relationTupleRepository, driveSchema)

			actualResult, _, _, err := permissionService.Check(context.Background(), "1", "update", "doc:1", 8)
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
					UsersetObjectID: "2",
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

			relationTupleRepository.On("QueryTuples", "doc", "1", "parent").Return(getDocParent, nil).Times(2)
			relationTupleRepository.On("QueryTuples", "folder", "1", "collaborator").Return(getParentCollaborators, nil).Times(1)
			relationTupleRepository.On("QueryTuples", "folder", "1", "admin").Return(getParentAdmins, nil).Times(1)
			relationTupleRepository.On("QueryTuples", "doc", "1", "owner").Return(getDocOwners, nil).Times(1)

			permissionService = NewPermissionService(relationTupleRepository, driveSchema)

			actualResult, _, _, err := permissionService.Check(context.Background(), "1", "read", "doc:1", 8)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(false).Should(Equal(actualResult))

		})

	})

	// GITHUB SAMPLE

	githubSchema := parser.TranslateToSchema("entity user {} `table:user|identifier:id`\n\nentity organization {\n\n    relation admin @user     `rel:custom`\n    relation member @user    `rel:many-to-many|table:organization_members|cols:organization_id,user_id`\n\n    action create_repository = admin or member\n    action delete = admin\n\n} `table:organization|identifier:id`\n\n\nentity repository {\n\n    relation    parent   @organization    `rel:belongs-to|cols:organization_id`\n    relation    owner    @user            `rel:belongs-to|cols:owner_id`\n\n    action push   = owner\n    action read   = owner and (parent.admin or parent.member)\n    action delete = parent.member and (parent.admin or owner)\n\n} `table:repository|identifier:id`")

	Context("Github Sample: Check", func() {

		It("Github Sample: Case 1", func() {

			relationTupleRepository := new(mocks.RelationTupleRepository)

			getDocParent := []entities.RelationTuple{
				{
					Entity:          "repository",
					ObjectID:        "1",
					Relation:        "owner",
					UsersetEntity:   "",
					UsersetObjectID: "2",
					UsersetRelation: "",
				},
			}

			relationTupleRepository.On("QueryTuples", "repository", "1", "owner").Return(getDocParent, nil).Times(2)

			permissionService = NewPermissionService(relationTupleRepository, githubSchema)

			actualResult, _, _, err := permissionService.Check(context.Background(), "1", "push", "repository:1", 8)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(false).Should(Equal(actualResult))

		})

		It("Github Sample: Case 2", func() {

			relationTupleRepository := new(mocks.RelationTupleRepository)

			getDocParent := []entities.RelationTuple{
				{
					Entity:          "repository",
					ObjectID:        "1",
					Relation:        "owner",
					UsersetEntity:   "organization",
					UsersetObjectID: "2",
					UsersetRelation: "admin",
				},
			}

			getOrgAdmins := []entities.RelationTuple{
				{
					Entity:          "organization",
					ObjectID:        "2",
					Relation:        "admin",
					UsersetEntity:   "organization",
					UsersetObjectID: "3",
					UsersetRelation: "member",
				},
				{
					Entity:          "organization",
					ObjectID:        "2",
					Relation:        "admin",
					UsersetEntity:   "",
					UsersetObjectID: "3",
					UsersetRelation: "",
				},
				{
					Entity:          "organization",
					ObjectID:        "2",
					Relation:        "admin",
					UsersetEntity:   "",
					UsersetObjectID: "8",
					UsersetRelation: "",
				},
			}

			getOrgMembers := []entities.RelationTuple{
				{
					Entity:          "organization",
					ObjectID:        "3",
					Relation:        "member",
					UsersetEntity:   "",
					UsersetObjectID: "1",
					UsersetRelation: "",
				},
			}

			relationTupleRepository.On("QueryTuples", "repository", "1", "owner").Return(getDocParent, nil).Times(1)
			relationTupleRepository.On("QueryTuples", "organization", "2", "admin").Return(getOrgAdmins, nil).Times(1)
			relationTupleRepository.On("QueryTuples", "organization", "3", "member").Return(getOrgMembers, nil).Times(1)

			permissionService = NewPermissionService(relationTupleRepository, githubSchema)

			actualResult, _, _, err := permissionService.Check(context.Background(), "1", "push", "repository:1", 4)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(true).Should(Equal(actualResult))

		})

		It("Github Sample: Case 3", func() {

			relationTupleRepository := new(mocks.RelationTupleRepository)

			getRepositoryParent := []entities.RelationTuple{
				{
					Entity:          "repository",
					ObjectID:        "1",
					Relation:        "parent",
					UsersetEntity:   "organization",
					UsersetObjectID: "8",
					UsersetRelation: tuple.ELLIPSIS,
				},
			}

			getOrganizationMembers := []entities.RelationTuple{
				{
					Entity:          "organization",
					ObjectID:        "8",
					Relation:        "member",
					UsersetEntity:   "",
					UsersetObjectID: "1",
					UsersetRelation: "",
				},
			}

			getOrganizationAdmins := []entities.RelationTuple{
				{
					Entity:          "organization",
					ObjectID:        "8",
					Relation:        "admin",
					UsersetEntity:   "",
					UsersetObjectID: "2",
					UsersetRelation: "",
				},
			}

			getRepositoryOwners := []entities.RelationTuple{
				{
					Entity:          "repository",
					ObjectID:        "1",
					Relation:        "owner",
					UsersetEntity:   "",
					UsersetObjectID: "7",
					UsersetRelation: "",
				},
			}

			relationTupleRepository.On("QueryTuples", "repository", "1", "parent").Return(getRepositoryParent, nil).Times(2)
			relationTupleRepository.On("QueryTuples", "organization", "8", "member").Return(getOrganizationMembers, nil).Times(1)
			relationTupleRepository.On("QueryTuples", "organization", "8", "admin").Return(getOrganizationAdmins, nil).Times(1)
			relationTupleRepository.On("QueryTuples", "repository", "1", "owner").Return(getRepositoryOwners, nil).Times(1)

			permissionService = NewPermissionService(relationTupleRepository, githubSchema)

			actualResult, _, _, err := permissionService.Check(context.Background(), "1", "delete", "repository:1", 6)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(false).Should(Equal(actualResult))

		})

	})
})
