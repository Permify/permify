package commands

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	`github.com/Permify/permify/internal/repositories`
	"github.com/Permify/permify/internal/repositories/mocks"
	`github.com/Permify/permify/pkg/dsl/schema`
	`github.com/Permify/permify/pkg/dsl/translator`
	"github.com/Permify/permify/pkg/logger"
	base `github.com/Permify/permify/pkg/pb/base/v1`
	"github.com/Permify/permify/pkg/tuple"
)

var _ = Describe("check-command", func() {
	var checkCommand *CheckCommand
	l := logger.New("debug")

	// DRIVE SAMPLE

	driveConfigs := []repositories.EntityConfig{
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

			getDocParent := []*base.Tuple{
				{
					Entity: &base.Entity{
						Type: "doc",
						Id:   "1",
					},
					Relation: "parent",
					Subject: &base.Subject{
						Type:     "folder",
						Id:       "1",
						Relation: tuple.ELLIPSIS,
					},
				},
			}

			getParentAdmins := []*base.Tuple{
				{
					Entity: &base.Entity{
						Type: "folder",
						Id:   "1",
					},
					Relation: "admin",
					Subject: &base.Subject{
						Type:     tuple.USER,
						Id:       "1",
						Relation: "",
					},
				},
			}

			getParentCollaborators := []*base.Tuple{
				{
					Entity: &base.Entity{
						Type: "folder",
						Id:   "1",
					},
					Relation: "collaborator",
					Subject: &base.Subject{
						Type:     tuple.USER,
						Id:       "1",
						Relation: "",
					},
				},
				{
					Entity: &base.Entity{
						Type: "folder",
						Id:   "1",
					},
					Relation: "collaborator",
					Subject: &base.Subject{
						Type:     tuple.USER,
						Id:       "3",
						Relation: "",
					},
				},
			}

			getDocOwners := []*base.Tuple{
				{
					Entity: &base.Entity{
						Type: "doc",
						Id:   "1",
					},
					Relation: "owner",
					Subject: &base.Subject{
						Type:     tuple.USER,
						Id:       "2",
						Relation: "",
					},
				},
			}

			relationTupleRepository.On("QueryTuples", "doc", "1", "parent").Return(tuple.NewTupleCollection(getDocParent...).CreateTupleIterator(), nil).Times(2)
			relationTupleRepository.On("QueryTuples", "folder", "1", "collaborator").Return(tuple.NewTupleCollection(getParentCollaborators...).CreateTupleIterator(), nil).Times(1)
			relationTupleRepository.On("QueryTuples", "doc", "1", "owner").Return(tuple.NewTupleCollection(getDocOwners...).CreateTupleIterator(), nil).Times(1)
			relationTupleRepository.On("QueryTuples", "folder", "1", "admin").Return(tuple.NewTupleCollection(getParentAdmins...).CreateTupleIterator(), nil).Times(1)

			checkCommand = NewCheckCommand(relationTupleRepository, l)

			re := &CheckQuery{
				Entity:  &base.Entity{Type: "doc", Id: "1"},
				Subject: &base.Subject{Type: tuple.USER, Id: "1"},
			}

			re.SetDepth(8)

			var serializedConfigs []string
			for _, c := range driveConfigs {
				serializedConfigs = append(serializedConfigs, c.Serialized())
			}

			sch, err := translator.StringToSchema(serializedConfigs...)
			Expect(err).ShouldNot(HaveOccurred())

			en, err := schema.GetEntityByName(sch, re.Entity.Type)
			Expect(err).ShouldNot(HaveOccurred())

			ac, err := schema.GetAction(en, "read")
			Expect(err).ShouldNot(HaveOccurred())

			actualResult, err := checkCommand.Execute(context.Background(), re, ac.Child)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(re.loadDepth()).Should(Equal(int32(4)))
			Expect(true).Should(Equal(actualResult.Can))
		})

		It("Drive Sample: Case 2", func() {
			relationTupleRepository := new(mocks.RelationTupleRepository)

			getDocParent := []*base.Tuple{
				{
					Entity: &base.Entity{
						Type: "doc",
						Id:   "1",
					},
					Relation: "parent",
					Subject: &base.Subject{
						Type:     "folder",
						Id:       "1",
						Relation: tuple.ELLIPSIS,
					},
				},
			}

			getParentAdmins := []*base.Tuple{
				{
					Entity: &base.Entity{
						Type: "folder",
						Id:   "1",
					},
					Relation: "admin",
					Subject: &base.Subject{
						Type:     tuple.USER,
						Id:       "2",
						Relation: "",
					},
				},
			}

			getDocOwners := []*base.Tuple{
				{
					Entity: &base.Entity{
						Type: "doc",
						Id:   "1",
					},
					Relation: "owner",
					Subject: &base.Subject{
						Type:     tuple.USER,
						Id:       "1",
						Relation: "",
					},
				},
			}

			relationTupleRepository.On("QueryTuples", "doc", "1", "parent").Return(tuple.NewTupleCollection(getDocParent...).CreateTupleIterator(), nil).Times(1)
			relationTupleRepository.On("QueryTuples", "folder", "1", "admin").Return(tuple.NewTupleCollection(getParentAdmins...).CreateTupleIterator(), nil).Times(1)
			relationTupleRepository.On("QueryTuples", "doc", "1", "owner").Return(tuple.NewTupleCollection(getDocOwners...).CreateTupleIterator(), nil).Times(1)

			checkCommand = NewCheckCommand(relationTupleRepository, l)

			re := &CheckQuery{
				Entity:  &base.Entity{Type: "doc", Id: "1"},
				Subject: &base.Subject{Type: tuple.USER, Id: "1"},
			}

			re.SetDepth(8)

			var serializedConfigs []string
			for _, c := range driveConfigs {
				serializedConfigs = append(serializedConfigs, c.Serialized())
			}

			sch, err := translator.StringToSchema(serializedConfigs...)
			Expect(err).ShouldNot(HaveOccurred())

			en, err := schema.GetEntityByName(sch, re.Entity.Type)
			Expect(err).ShouldNot(HaveOccurred())

			ac, err := schema.GetAction(en, "update")
			Expect(err).ShouldNot(HaveOccurred())

			actualResult, err := checkCommand.Execute(context.Background(), re, ac.Child)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(re.loadDepth()).Should(Equal(int32(5)))
			Expect(false).Should(Equal(actualResult.Can))
		})

		It("Drive Sample: Case 3", func() {
			relationTupleRepository := new(mocks.RelationTupleRepository)

			getDocParent := []*base.Tuple{
				{
					Entity: &base.Entity{
						Type: "doc",
						Id:   "1",
					},
					Relation: "parent",
					Subject: &base.Subject{
						Type:     "folder",
						Id:       "1",
						Relation: tuple.ELLIPSIS,
					},
				},
			}

			getParentAdmins := []*base.Tuple{
				{
					Entity: &base.Entity{
						Type: "folder",
						Id:   "1",
					},
					Relation: "admin",
					Subject: &base.Subject{
						Type:     tuple.USER,
						Id:       "2",
						Relation: "",
					},
				},
			}

			getParentCollaborators := []*base.Tuple{
				{
					Entity: &base.Entity{
						Type: "folder",
						Id:   "1",
					},
					Relation: "collaborator",
					Subject: &base.Subject{
						Type:     tuple.USER,
						Id:       "2",
						Relation: "",
					},
				},
				{
					Entity: &base.Entity{
						Type: "folder",
						Id:   "1",
					},
					Relation: "collaborator",
					Subject: &base.Subject{
						Type:     "folder",
						Id:       "1",
						Relation: "admin",
					},
				},
			}

			getDocOwners := []*base.Tuple{
				{
					Entity: &base.Entity{
						Type: "doc",
						Id:   "1",
					},
					Relation: "owner",
					Subject: &base.Subject{
						Type:     tuple.USER,
						Id:       "2",
						Relation: "",
					},
				},
			}

			relationTupleRepository.On("QueryTuples", "doc", "1", "parent").Return(tuple.NewTupleCollection(getDocParent...).CreateTupleIterator(), nil).Times(2)
			relationTupleRepository.On("QueryTuples", "folder", "1", "collaborator").Return(tuple.NewTupleCollection(getParentCollaborators...).CreateTupleIterator(), nil).Times(1)
			relationTupleRepository.On("QueryTuples", "folder", "1", "admin").Return(tuple.NewTupleCollection(getParentAdmins...).CreateTupleIterator(), nil).Times(2)
			relationTupleRepository.On("QueryTuples", "doc", "1", "owner").Return(tuple.NewTupleCollection(getDocOwners...).CreateTupleIterator(), nil).Times(1)

			checkCommand = NewCheckCommand(relationTupleRepository, l)

			re := &CheckQuery{
				Entity:  &base.Entity{Type: "doc", Id: "1"},
				Subject: &base.Subject{Type: tuple.USER, Id: "1"},
			}

			re.SetDepth(8)

			var serializedConfigs []string
			for _, c := range driveConfigs {
				serializedConfigs = append(serializedConfigs, c.Serialized())
			}

			sch, err := translator.StringToSchema(serializedConfigs...)
			Expect(err).ShouldNot(HaveOccurred())

			en, err := schema.GetEntityByName(sch, re.Entity.Type)
			Expect(err).ShouldNot(HaveOccurred())

			ac, err := schema.GetAction(en, "read")
			Expect(err).ShouldNot(HaveOccurred())

			actualResult, err := checkCommand.Execute(context.Background(), re, ac.Child)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(re.loadDepth()).Should(Equal(int32(4)))
			Expect(false).Should(Equal(actualResult.Can))
		})
	})

	// GITHUB SAMPLE

	githubConfigs := []repositories.EntityConfig{
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

			getRepositoryOwner := []*base.Tuple{
				{
					Entity: &base.Entity{
						Type: "repository",
						Id:   "1",
					},
					Relation: "owner",
					Subject: &base.Subject{
						Type:     tuple.USER,
						Id:       "2",
						Relation: "",
					},
				},
			}

			relationTupleRepository.On("QueryTuples", "repository", "1", "owner").Return(tuple.NewTupleCollection(getRepositoryOwner...).CreateTupleIterator(), nil).Times(1)

			checkCommand = NewCheckCommand(relationTupleRepository, l)

			re := &CheckQuery{
				Entity:  &base.Entity{Type: "repository", Id: "1"},
				Subject: &base.Subject{Type: tuple.USER, Id: "1"},
			}

			re.SetDepth(8)

			var serializedConfigs []string
			for _, c := range githubConfigs {
				serializedConfigs = append(serializedConfigs, c.Serialized())
			}

			sch, err := translator.StringToSchema(serializedConfigs...)
			Expect(err).ShouldNot(HaveOccurred())

			en, err := schema.GetEntityByName(sch, re.Entity.Type)
			Expect(err).ShouldNot(HaveOccurred())

			ac, err := schema.GetAction(en, "push")
			Expect(err).ShouldNot(HaveOccurred())

			actualResult, err := checkCommand.Execute(context.Background(), re, ac.Child)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(re.loadDepth()).Should(Equal(int32(7)))
			Expect(false).Should(Equal(actualResult.Can))
		})

		It("Github Sample: Case 2", func() {
			relationTupleRepository := new(mocks.RelationTupleRepository)

			getRepositoryOwners := []*base.Tuple{
				{
					Entity: &base.Entity{
						Type: "repository",
						Id:   "1",
					},
					Relation: "owner",
					Subject: &base.Subject{
						Type:     "organization",
						Id:       "2",
						Relation: "admin",
					},
				},
			}

			getOrganizationAdmins := []*base.Tuple{
				{
					Entity: &base.Entity{
						Type: "organization",
						Id:   "2",
					},
					Relation: "admin",
					Subject: &base.Subject{
						Type:     "organization",
						Id:       "3",
						Relation: "member",
					},
				},
				{
					Entity: &base.Entity{
						Type: "organization",
						Id:   "2",
					},
					Relation: "admin",
					Subject: &base.Subject{
						Type:     tuple.USER,
						Id:       "3",
						Relation: "",
					},
				},
				{
					Entity: &base.Entity{
						Type: "organization",
						Id:   "2",
					},
					Relation: "admin",
					Subject: &base.Subject{
						Type:     tuple.USER,
						Id:       "8",
						Relation: "",
					},
				},
			}

			getOrganizationMembers := []*base.Tuple{
				{
					Entity: &base.Entity{
						Type: "organization",
						Id:   "3",
					},
					Relation: "member",
					Subject: &base.Subject{
						Type:     tuple.USER,
						Id:       "1",
						Relation: "",
					},
				},
			}

			relationTupleRepository.On("QueryTuples", "repository", "1", "owner").Return(tuple.NewTupleCollection(getRepositoryOwners...).CreateTupleIterator(), nil).Times(1)
			relationTupleRepository.On("QueryTuples", "organization", "2", "admin").Return(tuple.NewTupleCollection(getOrganizationAdmins...).CreateTupleIterator(), nil).Times(1)
			relationTupleRepository.On("QueryTuples", "organization", "3", "member").Return(tuple.NewTupleCollection(getOrganizationMembers...).CreateTupleIterator(), nil).Times(1)

			checkCommand = NewCheckCommand(relationTupleRepository, l)

			re := &CheckQuery{
				Entity:  &base.Entity{Type: "repository", Id: "1"},
				Subject: &base.Subject{Type: tuple.USER, Id: "1"},
			}

			re.SetDepth(8)

			var serializedConfigs []string
			for _, c := range githubConfigs {
				serializedConfigs = append(serializedConfigs, c.Serialized())
			}

			sch, err := translator.StringToSchema(serializedConfigs...)
			Expect(err).ShouldNot(HaveOccurred())

			en, err := schema.GetEntityByName(sch, re.Entity.Type)
			Expect(err).ShouldNot(HaveOccurred())

			ac, err := schema.GetAction(en, "push")
			Expect(err).ShouldNot(HaveOccurred())

			actualResult, err := checkCommand.Execute(context.Background(), re, ac.Child)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(re.loadDepth()).Should(Equal(int32(5)))
			Expect(true).Should(Equal(actualResult.Can))
		})

		It("Github Sample: Case 3", func() {
			relationTupleRepository := new(mocks.RelationTupleRepository)

			getRepositoryParent := []*base.Tuple{
				{
					Entity: &base.Entity{
						Type: "repository",
						Id:   "1",
					},
					Relation: "parent",
					Subject: &base.Subject{
						Type:     "organization",
						Id:       "8",
						Relation: tuple.ELLIPSIS,
					},
				},
			}

			getOrganizationMembers := []*base.Tuple{
				{
					Entity: &base.Entity{
						Type: "organization",
						Id:   "8",
					},
					Relation: "member",
					Subject: &base.Subject{
						Type:     tuple.USER,
						Id:       "1",
						Relation: "",
					},
				},
			}

			getOrganizationAdmins := []*base.Tuple{
				{
					Entity: &base.Entity{
						Type: "organization",
						Id:   "8",
					},
					Relation: "admin",
					Subject: &base.Subject{
						Type:     tuple.USER,
						Id:       "2",
						Relation: "",
					},
				},
			}

			getRepositoryOwners := []*base.Tuple{
				{
					Entity: &base.Entity{
						Type: "repository",
						Id:   "1",
					},
					Relation: "owner",
					Subject: &base.Subject{
						Type:     tuple.USER,
						Id:       "7",
						Relation: "",
					},
				},
			}

			relationTupleRepository.On("QueryTuples", "repository", "1", "parent").Return(tuple.NewTupleCollection(getRepositoryParent...).CreateTupleIterator(), nil).Times(2)
			relationTupleRepository.On("QueryTuples", "organization", "8", "member").Return(tuple.NewTupleCollection(getOrganizationMembers...).CreateTupleIterator(), nil).Times(1)
			relationTupleRepository.On("QueryTuples", "organization", "8", "admin").Return(tuple.NewTupleCollection(getOrganizationAdmins...).CreateTupleIterator(), nil).Times(1)
			relationTupleRepository.On("QueryTuples", "repository", "1", "owner").Return(tuple.NewTupleCollection(getRepositoryOwners...).CreateTupleIterator(), nil).Times(1)

			checkCommand = NewCheckCommand(relationTupleRepository, l)

			re := &CheckQuery{
				Entity:  &base.Entity{Type: "repository", Id: "1"},
				Subject: &base.Subject{Type: tuple.USER, Id: "1"},
			}

			re.SetDepth(6)

			var serializedConfigs []string
			for _, c := range githubConfigs {
				serializedConfigs = append(serializedConfigs, c.Serialized())
			}

			sch, err := translator.StringToSchema(serializedConfigs...)
			Expect(err).ShouldNot(HaveOccurred())

			en, err := schema.GetEntityByName(sch, re.Entity.Type)
			Expect(err).ShouldNot(HaveOccurred())

			ac, err := schema.GetAction(en, "delete")
			Expect(err).ShouldNot(HaveOccurred())

			actualResult, err := checkCommand.Execute(context.Background(), re, ac.Child)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(re.loadDepth()).Should(Equal(int32(2)))
			Expect(false).Should(Equal(actualResult.Can))
		})
	})
})
