package commands

import (
	"context"
	"strings"

	"github.com/doug-martin/goqu/v9"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Permify/permify/internal/repositories"
	"github.com/Permify/permify/internal/repositories/mocks"
	"github.com/Permify/permify/pkg/dsl/compiler"
	"github.com/Permify/permify/pkg/dsl/schema"
	"github.com/Permify/permify/pkg/logger"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/tuple"
)

var _ = Describe("lookup-query-command", func() {
	var lookupQueryCommand *LookupQueryCommand
	l := logger.New("debug")

	// GITHUB SAMPLE

	githubConfigs := []repositories.EntityConfig{
		{
			Entity:           "user",
			SerializedConfig: []byte("entity user {}"),
		},
		{
			Entity:           "organization",
			SerializedConfig: []byte("entity organization {\n\n    relation admin @user\n    relation member @user\n\n    action create_repository = (admin or member)\n    action delete = admin\n} `table:organizations`"),
		},
		{
			Entity:           "repository",
			SerializedConfig: []byte("entity repository {\n\n    relation owner @user  `column:owner_id`\n    relation parent @organization `column:organization_id`\n\n    action push = owner\n    action read = (owner and (parent.admin and parent.member))\n    action delete = (parent.member and (parent.admin or owner))\n\n} `table:repositories`"),
		},
	}

	Context("Github Sample: Lookup Query", func() {
		It("Github Sample: Case 1", func() {
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
						Id:       "6",
						Relation: tuple.ELLIPSIS,
					},
				},
			}

			getOrganizationMembers := []*base.Tuple{
				{
					Entity: &base.Entity{
						Type: "organization",
						Id:   "6",
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
						Id:   "6",
					},
					Relation: "admin",
					Subject: &base.Subject{
						Type:     tuple.USER,
						Id:       "1",
						Relation: "",
					},
				},
			}

			relationTupleRepository.On("QueryTuples", "repository", "1", "parent").Return(tuple.NewTupleCollection(getRepositoryParent...).CreateTupleIterator(), nil).Times(1)
			relationTupleRepository.On("ReverseQueryTuples", "organization", "member", "user", []string{"1"}, "").Return(tuple.NewTupleCollection(getOrganizationMembers...).CreateTupleIterator(), nil).Times(1)
			relationTupleRepository.On("ReverseQueryTuples", "organization", "admin", "user", []string{"1"}, "").Return(tuple.NewTupleCollection(getOrganizationAdmins...).CreateTupleIterator(), nil).Times(1)

			lookupQueryCommand = NewLookupQueryCommand(relationTupleRepository, l)

			lq := &LookupQueryQuery{
				EntityType: "repository",
				Action:     "read",
				Subject:    &base.Subject{Type: tuple.USER, Id: "1"},
			}

			var serializedConfigs []string
			for _, c := range githubConfigs {
				serializedConfigs = append(serializedConfigs, c.Serialized())
			}

			sch, err := compiler.NewSchema(serializedConfigs...)
			Expect(err).ShouldNot(HaveOccurred())

			en, err := schema.GetEntityByName(sch, lq.EntityType)
			Expect(err).ShouldNot(HaveOccurred())

			ac, err := schema.GetActionByNameInEntityDefinition(en, "read")
			Expect(err).ShouldNot(HaveOccurred())

			lq.SetSchema(sch)

			actualResult, err := lookupQueryCommand.Execute(context.Background(), lq, ac.Child)
			query, _, e := goqu.Select("*").From("repositories").Where(goqu.And(goqu.I("repositories.owner_id").Eq("1")), goqu.And(goqu.And(goqu.I("repositories.organization_id").Eq("6"), goqu.I("repositories.organization_id").Eq("3")))).Prepared(true).ToSQL()

			Expect(e).ShouldNot(HaveOccurred())
			Expect(actualResult.Query).Should(Equal(strings.ReplaceAll(query, "\"", "")))
			Expect(actualResult.Args).Should(Equal([]string{"1", "6", "6"}))
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("Github Sample: Case 2", func() {
			relationTupleRepository := new(mocks.RelationTupleRepository)

			lookupQueryCommand = NewLookupQueryCommand(relationTupleRepository, l)

			lq := &LookupQueryQuery{
				EntityType: "repository",
				Action:     "read",
				Subject:    &base.Subject{Type: tuple.USER, Id: "1"},
			}

			var serializedConfigs []string
			for _, c := range githubConfigs {
				serializedConfigs = append(serializedConfigs, c.Serialized())
			}

			sch, err := compiler.NewSchema(serializedConfigs...)
			Expect(err).ShouldNot(HaveOccurred())

			en, err := schema.GetEntityByName(sch, lq.EntityType)
			Expect(err).ShouldNot(HaveOccurred())

			ac, err := schema.GetActionByNameInEntityDefinition(en, "push")
			Expect(err).ShouldNot(HaveOccurred())

			lq.SetSchema(sch)

			actualResult, err := lookupQueryCommand.Execute(context.Background(), lq, ac.Child)
			query, _, e := goqu.Select("*").From("repositories").Where(goqu.And(goqu.I("repositories.owner_id").Eq("1"))).Prepared(true).ToSQL()

			Expect(e).ShouldNot(HaveOccurred())
			Expect(actualResult.Query).Should(Equal(strings.ReplaceAll(query, "\"", "")))
			Expect(actualResult.Args).Should(Equal([]string{"1"}))
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("Github Sample: Case 3", func() {
			relationTupleRepository := new(mocks.RelationTupleRepository)

			getOrganizationAdmins := []*base.Tuple{
				{
					Entity: &base.Entity{
						Type: "organization",
						Id:   "6",
					},
					Relation: "admin",
					Subject: &base.Subject{
						Type:     tuple.USER,
						Id:       "1",
						Relation: "",
					},
				},
				{
					Entity: &base.Entity{
						Type: "organization",
						Id:   "6",
					},
					Relation: "admin",
					Subject: &base.Subject{
						Type:     "organization",
						Id:       "3",
						Relation: "member",
					},
				},
			}

			getOrganizationMembers1 := []*base.Tuple{
				{
					Entity: &base.Entity{
						Type: "organization",
						Id:   "6",
					},
					Relation: "member",
					Subject: &base.Subject{
						Type:     tuple.USER,
						Id:       "1",
						Relation: "",
					},
				},
			}

			getOrganizationMembers2 := []*base.Tuple{
				{
					Entity: &base.Entity{
						Type: "organization",
						Id:   "3",
					},
					Relation: "member",
					Subject: &base.Subject{
						Type:     tuple.USER,
						Id:       "9",
						Relation: "",
					},
				},
			}

			relationTupleRepository.On("QueryTuples", "organization", "6", "admin").Return(tuple.NewTupleCollection(getOrganizationAdmins...).CreateTupleIterator(), nil).Times(1)
			relationTupleRepository.On("QueryTuples", "organization", "3", "member").Return(tuple.NewTupleCollection(getOrganizationMembers2...).CreateTupleIterator(), nil).Times(1)
			relationTupleRepository.On("ReverseQueryTuples", "organization", "member", "organization", []string{"6"}, "admin").Return(tuple.NewTupleCollection(getOrganizationMembers1...).CreateTupleIterator(), nil).Times(1)
			relationTupleRepository.On("ReverseQueryTuples", "organization", "admin", "organization", []string{"6"}, "admin").Return(tuple.NewTupleCollection(getOrganizationAdmins...).CreateTupleIterator(), nil).Times(1)

			lookupQueryCommand = NewLookupQueryCommand(relationTupleRepository, l)

			lq := &LookupQueryQuery{
				EntityType: "repository",
				Action:     "read",
				Subject:    &base.Subject{Type: "organization", Id: "6", Relation: "admin"},
			}

			var serializedConfigs []string
			for _, c := range githubConfigs {
				serializedConfigs = append(serializedConfigs, c.Serialized())
			}

			sch, err := compiler.NewSchema(serializedConfigs...)
			Expect(err).ShouldNot(HaveOccurred())

			en, err := schema.GetEntityByName(sch, lq.EntityType)
			Expect(err).ShouldNot(HaveOccurred())

			ac, err := schema.GetActionByNameInEntityDefinition(en, "read")
			Expect(err).ShouldNot(HaveOccurred())

			lq.SetSchema(sch)

			actualResult, err := lookupQueryCommand.Execute(context.Background(), lq, ac.Child)
			query, _, e := goqu.Select("*").From("repositories").Where(goqu.And(goqu.I("repositories.owner_id").In("1", "3")), goqu.And(goqu.And(goqu.I("repositories.organization_id").Eq("6"), goqu.I("repositories.organization_id").Eq("6")))).Prepared(true).ToSQL()

			Expect(e).ShouldNot(HaveOccurred())
			Expect(actualResult.Query).Should(Equal(strings.ReplaceAll(query, "\"", "")))
			Expect(actualResult.Args).Should(Equal([]string{"1", "9", "6", "6"}))
			Expect(err).ShouldNot(HaveOccurred())
		})
	})
})
