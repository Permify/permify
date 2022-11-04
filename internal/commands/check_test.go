package commands

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Permify/permify/internal/repositories/mocks"
	"github.com/Permify/permify/pkg/database"
	"github.com/Permify/permify/pkg/dsl/compiler"
	"github.com/Permify/permify/pkg/dsl/schema"
	"github.com/Permify/permify/pkg/logger"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/tuple"
)

var _ = Describe("check-command", func() {
	var checkCommand *CheckCommand
	l := logger.New("debug")

	// DRIVE SAMPLE

	driveSchema := `
entity user {}

entity organization {
	relation admin @user
}

entity folder {
	relation org @organization
	relation creator @user
	relation collaborator @user

	action read = collaborator
	action update = collaborator
	action delete = creator or org.admin
}

entity doc {
	relation org @organization
	relation parent @folder
	relation owner @user
	
	action read = (owner or parent.collaborator) or org.admin
	action update = owner and org.admin
	action delete = owner or org.admin
}
`

	Context("Drive Sample: Check", func() {
		It("Drive Sample: Case 1", func() {
			relationTupleRepository := new(mocks.RelationTupleRepository)

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

			getDocOrg := []*base.Tuple{
				{
					Entity: &base.Entity{
						Type: "doc",
						Id:   "1",
					},
					Relation: "org",
					Subject: &base.Subject{
						Type:     "organization",
						Id:       "1",
						Relation: tuple.ELLIPSIS,
					},
				},
			}

			getOrgAdmins := []*base.Tuple{
				{
					Entity: &base.Entity{
						Type: "organization",
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

			relationTupleRepository.On("QueryTuples", "doc", "1", "owner").Return(database.NewTupleCollection(getDocOwners...).CreateTupleIterator(), nil).Times(1)
			relationTupleRepository.On("QueryTuples", "doc", "1", "parent").Return(database.NewTupleCollection(getDocParent...).CreateTupleIterator(), nil).Times(1)
			relationTupleRepository.On("QueryTuples", "folder", "1", "collaborator").Return(database.NewTupleCollection(getParentCollaborators...).CreateTupleIterator(), nil).Times(1)
			relationTupleRepository.On("QueryTuples", "doc", "1", "org").Return(database.NewTupleCollection(getDocOrg...).CreateTupleIterator(), nil).Times(1)
			relationTupleRepository.On("QueryTuples", "organization", "1", "admin").Return(database.NewTupleCollection(getOrgAdmins...).CreateTupleIterator(), nil).Times(1)

			checkCommand = NewCheckCommand(relationTupleRepository, l)

			re := &CheckQuery{
				Entity:  &base.Entity{Type: "doc", Id: "1"},
				Subject: &base.Subject{Type: tuple.USER, Id: "1"},
			}

			re.SetDepth(8)

			sch, err := compiler.NewSchema(driveSchema)
			Expect(err).ShouldNot(HaveOccurred())

			en, err := schema.GetEntityByName(sch, re.Entity.Type)
			Expect(err).ShouldNot(HaveOccurred())

			ac, err := schema.GetActionByNameInEntityDefinition(en, "read")
			Expect(err).ShouldNot(HaveOccurred())

			actualResult, err := checkCommand.Execute(context.Background(), re, ac.Child)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(re.loadDepth()).Should(Equal(int32(3)))
			Expect(true).Should(Equal(actualResult.Can))
		})

		It("Drive Sample: Case 2", func() {
			relationTupleRepository := new(mocks.RelationTupleRepository)

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

			getDocOrg := []*base.Tuple{
				{
					Entity: &base.Entity{
						Type: "doc",
						Id:   "1",
					},
					Relation: "org",
					Subject: &base.Subject{
						Type:     "organization",
						Id:       "1",
						Relation: tuple.ELLIPSIS,
					},
				},
			}

			getOrgAdmins := []*base.Tuple{
				{
					Entity: &base.Entity{
						Type: "organization",
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

			relationTupleRepository.On("QueryTuples", "doc", "1", "owner").Return(database.NewTupleCollection(getDocOwners...).CreateTupleIterator(), nil).Times(1)
			relationTupleRepository.On("QueryTuples", "doc", "1", "org").Return(database.NewTupleCollection(getDocOrg...).CreateTupleIterator(), nil).Times(1)
			relationTupleRepository.On("QueryTuples", "organization", "1", "admin").Return(database.NewTupleCollection(getOrgAdmins...).CreateTupleIterator(), nil).Times(1)

			checkCommand = NewCheckCommand(relationTupleRepository, l)

			re := &CheckQuery{
				Entity:  &base.Entity{Type: "doc", Id: "1"},
				Subject: &base.Subject{Type: tuple.USER, Id: "1"},
			}

			re.SetDepth(8)

			sch, err := compiler.NewSchema(driveSchema)
			Expect(err).ShouldNot(HaveOccurred())

			en, err := schema.GetEntityByName(sch, re.Entity.Type)
			Expect(err).ShouldNot(HaveOccurred())

			ac, err := schema.GetActionByNameInEntityDefinition(en, "update")
			Expect(err).ShouldNot(HaveOccurred())

			actualResult, err := checkCommand.Execute(context.Background(), re, ac.Child)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(re.loadDepth()).Should(Equal(int32(5)))
			Expect(false).Should(Equal(actualResult.Can))
		})

		It("Drive Sample: Case 3", func() {
			relationTupleRepository := new(mocks.RelationTupleRepository)

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

			getParentCollaborators := []*base.Tuple{
				{
					Entity: &base.Entity{
						Type: "folder",
						Id:   "1",
					},
					Relation: "collaborator",
					Subject: &base.Subject{
						Type:     tuple.USER,
						Id:       "7",
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

			getDocOrg := []*base.Tuple{
				{
					Entity: &base.Entity{
						Type: "doc",
						Id:   "1",
					},
					Relation: "org",
					Subject: &base.Subject{
						Type:     "organization",
						Id:       "1",
						Relation: tuple.ELLIPSIS,
					},
				},
			}

			getOrgAdmins := []*base.Tuple{
				{
					Entity: &base.Entity{
						Type: "organization",
						Id:   "1",
					},
					Relation: "admin",
					Subject: &base.Subject{
						Type:     tuple.USER,
						Id:       "7",
						Relation: "",
					},
				},
			}

			relationTupleRepository.On("QueryTuples", "doc", "1", "owner").Return(database.NewTupleCollection(getDocOwners...).CreateTupleIterator(), nil).Times(1)
			relationTupleRepository.On("QueryTuples", "doc", "1", "parent").Return(database.NewTupleCollection(getDocParent...).CreateTupleIterator(), nil).Times(1)
			relationTupleRepository.On("QueryTuples", "folder", "1", "collaborator").Return(database.NewTupleCollection(getParentCollaborators...).CreateTupleIterator(), nil).Times(1)
			relationTupleRepository.On("QueryTuples", "doc", "1", "org").Return(database.NewTupleCollection(getDocOrg...).CreateTupleIterator(), nil).Times(1)
			relationTupleRepository.On("QueryTuples", "organization", "1", "admin").Return(database.NewTupleCollection(getOrgAdmins...).CreateTupleIterator(), nil).Times(1)

			checkCommand = NewCheckCommand(relationTupleRepository, l)

			re := &CheckQuery{
				Entity:  &base.Entity{Type: "doc", Id: "1"},
				Subject: &base.Subject{Type: tuple.USER, Id: "1"},
			}

			re.SetDepth(8)

			sch, err := compiler.NewSchema(driveSchema)
			Expect(err).ShouldNot(HaveOccurred())

			en, err := schema.GetEntityByName(sch, re.Entity.Type)
			Expect(err).ShouldNot(HaveOccurred())

			ac, err := schema.GetActionByNameInEntityDefinition(en, "read")
			Expect(err).ShouldNot(HaveOccurred())

			actualResult, err := checkCommand.Execute(context.Background(), re, ac.Child)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(re.loadDepth()).Should(Equal(int32(3)))
			Expect(false).Should(Equal(actualResult.Can))
		})
	})

	// GITHUB SAMPLE

	githubSchema := `
	entity user {}
	
	entity organization {
		relation admin @user
		relation member @user
	
		action create_repository = admin or member
		action delete = admin
	}
	
	entity repository {
		relation parent @organization
		relation owner @user
	
	 	action push   = owner
	   action read   = owner and (parent.admin or parent.member)
	   action delete = parent.member and (parent.admin or owner)
	}
	`

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

			relationTupleRepository.On("QueryTuples", "repository", "1", "owner").Return(database.NewTupleCollection(getRepositoryOwner...).CreateTupleIterator(), nil).Times(1)

			checkCommand = NewCheckCommand(relationTupleRepository, l)

			re := &CheckQuery{
				Entity:  &base.Entity{Type: "repository", Id: "1"},
				Subject: &base.Subject{Type: tuple.USER, Id: "1"},
			}

			re.SetDepth(8)

			sch, err := compiler.NewSchema(githubSchema)
			Expect(err).ShouldNot(HaveOccurred())

			en, err := schema.GetEntityByName(sch, re.Entity.Type)
			Expect(err).ShouldNot(HaveOccurred())

			ac, err := schema.GetActionByNameInEntityDefinition(en, "push")
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

			relationTupleRepository.On("QueryTuples", "repository", "1", "owner").Return(database.NewTupleCollection(getRepositoryOwners...).CreateTupleIterator(), nil).Times(1)
			relationTupleRepository.On("QueryTuples", "organization", "2", "admin").Return(database.NewTupleCollection(getOrganizationAdmins...).CreateTupleIterator(), nil).Times(1)
			relationTupleRepository.On("QueryTuples", "organization", "3", "member").Return(database.NewTupleCollection(getOrganizationMembers...).CreateTupleIterator(), nil).Times(1)

			checkCommand = NewCheckCommand(relationTupleRepository, l)

			re := &CheckQuery{
				Entity:  &base.Entity{Type: "repository", Id: "1"},
				Subject: &base.Subject{Type: tuple.USER, Id: "1"},
			}

			re.SetDepth(8)

			sch, err := compiler.NewSchema(githubSchema)
			Expect(err).ShouldNot(HaveOccurred())

			en, err := schema.GetEntityByName(sch, re.Entity.Type)
			Expect(err).ShouldNot(HaveOccurred())

			ac, err := schema.GetActionByNameInEntityDefinition(en, "push")
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

			relationTupleRepository.On("QueryTuples", "repository", "1", "parent").Return(database.NewTupleCollection(getRepositoryParent...).CreateTupleIterator(), nil).Times(2)
			relationTupleRepository.On("QueryTuples", "organization", "8", "member").Return(database.NewTupleCollection(getOrganizationMembers...).CreateTupleIterator(), nil).Times(1)
			relationTupleRepository.On("QueryTuples", "organization", "8", "admin").Return(database.NewTupleCollection(getOrganizationAdmins...).CreateTupleIterator(), nil).Times(1)
			relationTupleRepository.On("QueryTuples", "repository", "1", "owner").Return(database.NewTupleCollection(getRepositoryOwners...).CreateTupleIterator(), nil).Times(1)

			checkCommand = NewCheckCommand(relationTupleRepository, l)

			re := &CheckQuery{
				Entity:  &base.Entity{Type: "repository", Id: "1"},
				Subject: &base.Subject{Type: tuple.USER, Id: "1"},
			}

			re.SetDepth(6)

			sch, err := compiler.NewSchema(githubSchema)
			Expect(err).ShouldNot(HaveOccurred())

			en, err := schema.GetEntityByName(sch, re.Entity.Type)
			Expect(err).ShouldNot(HaveOccurred())

			ac, err := schema.GetActionByNameInEntityDefinition(en, "delete")
			Expect(err).ShouldNot(HaveOccurred())

			actualResult, err := checkCommand.Execute(context.Background(), re, ac.Child)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(re.loadDepth()).Should(Equal(int32(2)))
			Expect(false).Should(Equal(actualResult.Can))
		})
	})
})
