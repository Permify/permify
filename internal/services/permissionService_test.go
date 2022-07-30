package services

import (
	"context"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Permify/permify/internal/entities"
	"github.com/Permify/permify/internal/repositories/mocks"
	"github.com/Permify/permify/pkg/tuple"
)

var _ = Describe("permission-service", func() {
	var permissionService *PermissionService
	var schemaService *SchemaService

	// DRIVE SAMPLE

	var driveConfigs = []entities.EntityConfig{
		{
			Entity:           "user",
			SerializedConfig: []byte("entity user {}"),
		},
		{
			Entity:           "organization",
			SerializedConfig: []byte("entity organization {\nrelation admin @user\n}"),
		},
		{
			Entity:           "folder",
			SerializedConfig: []byte("entity folder {\n relation\tparent\t@organization\nrelation\tcreator\t@user\nrelation\tcollaborator\t@user\n action read = collaborator\naction update = collaborator\naction delete = creator or parent.admin\n}"),
		},
		{
			Entity:           "doc",
			SerializedConfig: []byte("entity doc {\nelation\tparent\t@organization\nrelation\towner\t@user\n  action read = (owner or parent.collaborator) or parent.admin\naction update = owner and parent.admin\n action delete = owner or parent.admin\n}"),
		},
	}

	Context("Drive Sample: Check", func() {
		It("Drive Sample: Case 1", func() {
			relationTupleRepository := new(mocks.RelationTupleRepository)
			entityConfigRepository := new(mocks.EntityConfigRepository)

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

			entityConfigRepository.On("All").Return(driveConfigs, nil)

			schemaService = NewSchemaService(entityConfigRepository)
			permissionService = NewPermissionService(relationTupleRepository, schemaService)

			actualResult, _, _, err := permissionService.Check(context.Background(), "1", "read", "doc:1", 8)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(true).Should(Equal(actualResult))
		})

		It("Drive Sample: Case 2", func() {
			relationTupleRepository := new(mocks.RelationTupleRepository)
			entityConfigRepository := new(mocks.EntityConfigRepository)

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

			entityConfigRepository.On("All").Return(driveConfigs, nil)

			schemaService = NewSchemaService(entityConfigRepository)
			permissionService = NewPermissionService(relationTupleRepository, schemaService)

			actualResult, _, _, err := permissionService.Check(context.Background(), "1", "update", "doc:1", 8)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(false).Should(Equal(actualResult))
		})

		It("Drive Sample: Case 3", func() {
			relationTupleRepository := new(mocks.RelationTupleRepository)
			entityConfigRepository := new(mocks.EntityConfigRepository)

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

			entityConfigRepository.On("All").Return(driveConfigs, nil)

			schemaService = NewSchemaService(entityConfigRepository)
			permissionService = NewPermissionService(relationTupleRepository, schemaService)

			actualResult, _, _, err := permissionService.Check(context.Background(), "1", "read", "doc:1", 8)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(false).Should(Equal(actualResult))
		})
	})

	// GITHUB SAMPLE

	var githubConfigs = []entities.EntityConfig{
		{
			Entity:           "user",
			SerializedConfig: []byte("entity user {}"),
		},
		{
			Entity:           "organization",
			SerializedConfig: []byte("entity organization {\nrelation admin @user\nrelation member @user\naction create_repository = admin or member\naction delete = admin\n}"),
		},
		{
			Entity:           "repository",
			SerializedConfig: []byte("entity repository {\nrelation parent @organization\n relation owner @user\n  action push   = owner\n    action read   = owner and (parent.admin or parent.member)\n    action delete = parent.member and (parent.admin or owner)\n}"),
		},
	}

	Context("Github Sample: Check", func() {
		It("Github Sample: Case 1", func() {
			relationTupleRepository := new(mocks.RelationTupleRepository)
			entityConfigRepository := new(mocks.EntityConfigRepository)

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

			entityConfigRepository.On("All").Return(githubConfigs, nil)

			schemaService = NewSchemaService(entityConfigRepository)
			permissionService = NewPermissionService(relationTupleRepository, schemaService)

			actualResult, _, _, err := permissionService.Check(context.Background(), "1", "push", "repository:1", 8)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(false).Should(Equal(actualResult))
		})

		It("Github Sample: Case 2", func() {
			relationTupleRepository := new(mocks.RelationTupleRepository)
			entityConfigRepository := new(mocks.EntityConfigRepository)

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

			entityConfigRepository.On("All").Return(githubConfigs, nil)

			schemaService = NewSchemaService(entityConfigRepository)
			permissionService = NewPermissionService(relationTupleRepository, schemaService)

			actualResult, _, _, err := permissionService.Check(context.Background(), "1", "push", "repository:1", 4)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(true).Should(Equal(actualResult))
		})

		It("Github Sample: Case 3", func() {
			relationTupleRepository := new(mocks.RelationTupleRepository)
			entityConfigRepository := new(mocks.EntityConfigRepository)

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

			entityConfigRepository.On("All").Return(githubConfigs, nil)

			schemaService = NewSchemaService(entityConfigRepository)
			permissionService = NewPermissionService(relationTupleRepository, schemaService)

			actualResult, _, _, err := permissionService.Check(context.Background(), "1", "delete", "repository:1", 6)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(false).Should(Equal(actualResult))
		})
	})
})
