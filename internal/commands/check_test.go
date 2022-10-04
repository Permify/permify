package commands

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Permify/permify/internal/repositories/entities"
	"github.com/Permify/permify/internal/repositories/mocks"
	"github.com/Permify/permify/pkg/logger"
	"github.com/Permify/permify/pkg/tuple"
)

var _ = Describe("check-command", func() {
	var checkCommand *CheckCommand
	l := logger.New("debug")

	// DRIVE SAMPLE

	driveConfigs := []entities.EntityConfig{
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
			SerializedConfig: []byte("entity doc {\n relation\tparent\t@organization\nrelation\towner\t@user\n  action read = (owner or parent.collaborator) or parent.admin\naction update = owner and parent.admin\n action delete = owner or parent.admin\n}"),
		},
	}

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
					UsersetEntity:   tuple.USER,
					UsersetObjectID: "1",
					UsersetRelation: "",
				},
			}

			getParentCollaborators := []entities.RelationTuple{
				{
					Entity:          "folder",
					ObjectID:        "1",
					Relation:        "collaborator",
					UsersetEntity:   tuple.USER,
					UsersetObjectID: "1",
					UsersetRelation: "",
				},
				{
					Entity:          "folder",
					ObjectID:        "1",
					Relation:        "collaborator",
					UsersetEntity:   tuple.USER,
					UsersetObjectID: "3",
					UsersetRelation: "",
				},
			}

			getDocOwners := []entities.RelationTuple{
				{
					Entity:          "doc",
					ObjectID:        "1",
					Relation:        "owner",
					UsersetEntity:   tuple.USER,
					UsersetObjectID: "2",
					UsersetRelation: "",
				},
			}

			relationTupleRepository.On("QueryTuples", "doc", "1", "parent").Return(getDocParent, nil).Times(2)
			relationTupleRepository.On("QueryTuples", "folder", "1", "collaborator").Return(getParentCollaborators, nil).Times(1)
			relationTupleRepository.On("QueryTuples", "doc", "1", "owner").Return(getDocOwners, nil).Times(1)
			relationTupleRepository.On("QueryTuples", "folder", "1", "admin").Return(getParentAdmins, nil).Times(1)

			checkCommand = NewCheckCommand(relationTupleRepository, l)

			re := &CheckQuery{
				Entity:  tuple.Entity{Type: "doc", ID: "1"},
				Subject: tuple.Subject{Type: tuple.USER, ID: "1"},
			}

			re.SetDepth(8)

			sch, err := entities.EntityConfigs(driveConfigs).ToSchema()
			Expect(err).ShouldNot(HaveOccurred())

			en, err := sch.GetEntityByName(re.Entity.Type)
			Expect(err).ShouldNot(HaveOccurred())

			ac, err := en.GetAction("read")
			Expect(err).ShouldNot(HaveOccurred())

			actualResult, err := checkCommand.Execute(context.Background(), re, ac.Child)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(re.loadDepth()).Should(Equal(int32(3)))
			Expect(true).Should(Equal(actualResult.Can))
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
					UsersetEntity:   tuple.USER,
					UsersetObjectID: "2",
					UsersetRelation: "",
				},
			}

			getDocOwners := []entities.RelationTuple{
				{
					Entity:          "doc",
					ObjectID:        "1",
					Relation:        "owner",
					UsersetEntity:   tuple.USER,
					UsersetObjectID: "1",
					UsersetRelation: "",
				},
			}

			relationTupleRepository.On("QueryTuples", "doc", "1", "parent").Return(getDocParent, nil).Times(1)
			relationTupleRepository.On("QueryTuples", "folder", "1", "admin").Return(getParentAdmins, nil).Times(1)
			relationTupleRepository.On("QueryTuples", "doc", "1", "owner").Return(getDocOwners, nil).Times(1)

			checkCommand = NewCheckCommand(relationTupleRepository, l)

			re := &CheckQuery{
				Entity:  tuple.Entity{Type: "doc", ID: "1"},
				Subject: tuple.Subject{Type: tuple.USER, ID: "1"},
			}

			re.SetDepth(8)

			sch, _ := entities.EntityConfigs(driveConfigs).ToSchema()

			en, _ := sch.GetEntityByName(re.Entity.Type)
			ac, _ := en.GetAction("update")

			actualResult, err := checkCommand.Execute(context.Background(), re, ac.Child)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(re.loadDepth()).Should(Equal(int32(5)))
			Expect(false).Should(Equal(actualResult.Can))
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
					UsersetEntity:   tuple.USER,
					UsersetObjectID: "2",
					UsersetRelation: "",
				},
			}

			getParentCollaborators := []entities.RelationTuple{
				{
					Entity:          "folder",
					ObjectID:        "1",
					Relation:        "collaborator",
					UsersetEntity:   tuple.USER,
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
					UsersetEntity:   tuple.USER,
					UsersetObjectID: "2",
					UsersetRelation: "",
				},
			}

			relationTupleRepository.On("QueryTuples", "doc", "1", "parent").Return(getDocParent, nil).Times(2)
			relationTupleRepository.On("QueryTuples", "folder", "1", "collaborator").Return(getParentCollaborators, nil).Times(1)
			relationTupleRepository.On("QueryTuples", "folder", "1", "admin").Return(getParentAdmins, nil).Times(2)
			relationTupleRepository.On("QueryTuples", "doc", "1", "owner").Return(getDocOwners, nil).Times(1)

			checkCommand = NewCheckCommand(relationTupleRepository, l)

			re := &CheckQuery{
				Entity:  tuple.Entity{Type: "doc", ID: "1"},
				Subject: tuple.Subject{Type: tuple.USER, ID: "1"},
			}

			re.SetDepth(8)

			sch, _ := entities.EntityConfigs(driveConfigs).ToSchema()

			en, _ := sch.GetEntityByName(re.Entity.Type)
			ac, _ := en.GetAction("read")

			actualResult, err := checkCommand.Execute(context.Background(), re, ac.Child)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(re.loadDepth()).Should(Equal(int32(2)))
			Expect(false).Should(Equal(actualResult.Can))
		})

	})

	// GITHUB SAMPLE

	githubConfigs := []entities.EntityConfig{
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

			getRepositoryOwner := []entities.RelationTuple{
				{
					Entity:          "repository",
					ObjectID:        "1",
					Relation:        "owner",
					UsersetEntity:   tuple.USER,
					UsersetObjectID: "2",
					UsersetRelation: "",
				},
			}

			relationTupleRepository.On("QueryTuples", "repository", "1", "owner").Return(getRepositoryOwner, nil).Times(1)

			checkCommand = NewCheckCommand(relationTupleRepository, l)

			re := &CheckQuery{
				Entity:  tuple.Entity{Type: "repository", ID: "1"},
				Subject: tuple.Subject{Type: tuple.USER, ID: "1"},
			}

			re.SetDepth(8)

			sch, _ := entities.EntityConfigs(githubConfigs).ToSchema()

			en, _ := sch.GetEntityByName(re.Entity.Type)
			ac, _ := en.GetAction("push")

			actualResult, err := checkCommand.Execute(context.Background(), re, ac.Child)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(re.loadDepth()).Should(Equal(int32(7)))
			Expect(false).Should(Equal(actualResult.Can))
		})

		It("Github Sample: Case 2", func() {
			relationTupleRepository := new(mocks.RelationTupleRepository)

			getRepositoryOwners := []entities.RelationTuple{
				{
					Entity:          "repository",
					ObjectID:        "1",
					Relation:        "owner",
					UsersetEntity:   "organization",
					UsersetObjectID: "2",
					UsersetRelation: "admin",
				},
			}

			getOrganizationAdmins := []entities.RelationTuple{
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
					UsersetEntity:   tuple.USER,
					UsersetObjectID: "3",
					UsersetRelation: "",
				},
				{
					Entity:          "organization",
					ObjectID:        "2",
					Relation:        "admin",
					UsersetEntity:   tuple.USER,
					UsersetObjectID: "8",
					UsersetRelation: "",
				},
			}

			getOrganizationMembers := []entities.RelationTuple{
				{
					Entity:          "organization",
					ObjectID:        "3",
					Relation:        "member",
					UsersetEntity:   tuple.USER,
					UsersetObjectID: "1",
					UsersetRelation: "",
				},
			}

			relationTupleRepository.On("QueryTuples", "repository", "1", "owner").Return(getRepositoryOwners, nil).Times(1)
			relationTupleRepository.On("QueryTuples", "organization", "2", "admin").Return(getOrganizationAdmins, nil).Times(1)
			relationTupleRepository.On("QueryTuples", "organization", "3", "member").Return(getOrganizationMembers, nil).Times(1)

			checkCommand = NewCheckCommand(relationTupleRepository, l)

			re := &CheckQuery{
				Entity:  tuple.Entity{Type: "repository", ID: "1"},
				Subject: tuple.Subject{Type: tuple.USER, ID: "1"},
			}

			re.SetDepth(8)
			sch, _ := entities.EntityConfigs(githubConfigs).ToSchema()

			en, _ := sch.GetEntityByName(re.Entity.Type)
			ac, _ := en.GetAction("push")

			actualResult, err := checkCommand.Execute(context.Background(), re, ac.Child)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(re.loadDepth()).Should(Equal(int32(5)))
			Expect(true).Should(Equal(actualResult.Can))
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
					UsersetEntity:   "user",
					UsersetObjectID: "1",
					UsersetRelation: "",
				},
			}

			getOrganizationAdmins := []entities.RelationTuple{
				{
					Entity:          "organization",
					ObjectID:        "8",
					Relation:        "admin",
					UsersetEntity:   "user",
					UsersetObjectID: "2",
					UsersetRelation: "",
				},
			}

			getRepositoryOwners := []entities.RelationTuple{
				{
					Entity:          "repository",
					ObjectID:        "1",
					Relation:        "owner",
					UsersetEntity:   "user",
					UsersetObjectID: "7",
					UsersetRelation: "",
				},
			}

			relationTupleRepository.On("QueryTuples", "repository", "1", "parent").Return(getRepositoryParent, nil).Times(2)
			relationTupleRepository.On("QueryTuples", "organization", "8", "member").Return(getOrganizationMembers, nil).Times(1)
			relationTupleRepository.On("QueryTuples", "organization", "8", "admin").Return(getOrganizationAdmins, nil).Times(1)
			relationTupleRepository.On("QueryTuples", "repository", "1", "owner").Return(getRepositoryOwners, nil).Times(1)

			checkCommand = NewCheckCommand(relationTupleRepository, l)

			re := &CheckQuery{
				Entity:  tuple.Entity{Type: "repository", ID: "1"},
				Subject: tuple.Subject{Type: tuple.USER, ID: "1"},
			}

			re.SetDepth(6)

			sch, _ := entities.EntityConfigs(githubConfigs).ToSchema()

			en, _ := sch.GetEntityByName(re.Entity.Type)
			ac, _ := en.GetAction("delete")

			actualResult, err := checkCommand.Execute(context.Background(), re, ac.Child)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(re.loadDepth()).Should(Equal(int32(1)))
			Expect(false).Should(Equal(actualResult.Can))
		})
	})
})
