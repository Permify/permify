package commands

import (
	"context"
	`github.com/doug-martin/goqu/v9`
	`strings`

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Permify/permify/internal/repositories/entities"
	"github.com/Permify/permify/internal/repositories/mocks"
	"github.com/Permify/permify/pkg/logger"
	"github.com/Permify/permify/pkg/tuple"
)

var _ = Describe("lookup-query-command", func() {
	var lookupQueryCommand *LookupQueryCommand
	l := logger.New("debug")

	githubConfigs := []entities.EntityConfig{
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

			getRepositoryParent := entities.RelationTuples{
				{
					Entity:          "repository",
					ObjectID:        "1",
					Relation:        "parent",
					UsersetEntity:   "organization",
					UsersetObjectID: "3",
					UsersetRelation: tuple.ELLIPSIS,
				},
			}

			getOrganizationMembers := entities.RelationTuples{
				{
					Entity:          "organization",
					ObjectID:        "6",
					Relation:        "member",
					UsersetEntity:   "user",
					UsersetObjectID: "1",
					UsersetRelation: "",
				},
			}

			getOrganizationAdmins := entities.RelationTuples{
				{
					Entity:          "organization",
					ObjectID:        "3",
					Relation:        "admin",
					UsersetEntity:   "user",
					UsersetObjectID: "1",
					UsersetRelation: "",
				},
			}

			relationTupleRepository.On("QueryTuples", "repository", "1", "parent").Return(getRepositoryParent, nil).Times(1)

			relationTupleRepository.On("ReverseQueryTuples", "organization", "member", "user", []string{"1"}, "").Return(getOrganizationMembers, nil).Times(1)
			relationTupleRepository.On("ReverseQueryTuples", "organization", "admin", "user", []string{"1"}, "").Return(getOrganizationAdmins, nil).Times(1)

			lookupQueryCommand = NewLookupQueryCommand(relationTupleRepository, l)

			lq := &LookupQueryQuery{
				EntityType: "repository",
				Action:     "read",
				Subject:    tuple.Subject{Type: tuple.USER, ID: "1"},
			}

			sch, _ := entities.EntityConfigs(githubConfigs).ToSchema()
			lq.SetSchema(sch)

			en, _ := sch.GetEntityByName(lq.EntityType)
			ac, _ := en.GetAction("read")

			actualResult, err := lookupQueryCommand.Execute(context.Background(), lq, ac.Child)
			query, _, e := goqu.Select("*").From("repositories").Where(goqu.And(goqu.I("repositories.owner_id").Eq("1")), goqu.And(goqu.And(goqu.I("repositories.organization_id").Eq("6"), goqu.I("repositories.organization_id").Eq("3")))).Prepared(true).ToSQL()

			Expect(e).ShouldNot(HaveOccurred())
			Expect(actualResult.Query).Should(Equal(strings.ReplaceAll(query, "\"", "")))
			Expect(actualResult.Args).Should(Equal([]interface{}{"1", "3", "6"}))
			Expect(err).ShouldNot(HaveOccurred())
		})
	})
})
