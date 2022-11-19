package commands

import (
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("lookup-query-command", func() {
	//var lookupQueryCommand *LookupQueryCommand
	//l := logger.New("debug")
	//
	//// GITHUB SAMPLE
	//
	//githubConfigs := []repositories.SchemaDefinition{
	//	{
	//		EntityType:           "user",
	//		SerializedDefinition: []byte("entity user {}"),
	//	},
	//	{
	//		EntityType:           "organization",
	//		SerializedDefinition: []byte("entity organization {\n\n    relation admin @user\n    relation member @user\n\n    action create_repository = (admin or member)\n    action delete = admin\n} `table:organizations`"),
	//	},
	//	{
	//		EntityType:           "repository",
	//		SerializedDefinition: []byte("entity repository {\n\n    relation owner @user  `column:owner_id`\n    relation parent @organization `column:organization_id`\n\n    action push = owner\n    action read = (owner and (parent.admin and parent.member))\n    action delete = (parent.member and (parent.admin or owner))\n\n} `table:repositories`"),
	//	},
	//}
	//
	//Context("Github Sample: Lookup Query", func() {
	//	It("Github Sample: Case 1", func() {
	//		relationTupleRepository := new(mocks.RelationshipReader)
	//
	//		getRepositoryParent := []*base.Tuple{
	//			{
	//				Entity: &base.Entity{
	//					Type: "repository",
	//					Id:   "1",
	//				},
	//				Relation: "parent",
	//				Subject: &base.Subject{
	//					Type:     "organization",
	//					Id:       "6",
	//					Relation: tuple.ELLIPSIS,
	//				},
	//			},
	//		}
	//
	//		getOrganizationMembers := []*base.Tuple{
	//			{
	//				Entity: &base.Entity{
	//					Type: "organization",
	//					Id:   "6",
	//				},
	//				Relation: "member",
	//				Subject: &base.Subject{
	//					Type:     tuple.USER,
	//					Id:       "1",
	//					Relation: "",
	//				},
	//			},
	//		}
	//
	//		getOrganizationAdmins := []*base.Tuple{
	//			{
	//				Entity: &base.Entity{
	//					Type: "organization",
	//					Id:   "6",
	//				},
	//				Relation: "admin",
	//				Subject: &base.Subject{
	//					Type:     tuple.USER,
	//					Id:       "1",
	//					Relation: "",
	//				},
	//			},
	//		}
	//
	//		relationTupleRepository.On("QueryTuples", "repository", "1", "parent").Return(database.NewTupleCollection(getRepositoryParent...).CreateTupleIterator(), nil).Times(1)
	//		relationTupleRepository.On("ReverseQueryTuples", "organization", "member", "user", []string{"1"}, "").Return(database.NewTupleCollection(getOrganizationMembers...).CreateTupleIterator(), nil).Times(1)
	//		relationTupleRepository.On("ReverseQueryTuples", "organization", "admin", "user", []string{"1"}, "").Return(database.NewTupleCollection(getOrganizationAdmins...).CreateTupleIterator(), nil).Times(1)
	//
	//		lookupQueryCommand = NewLookupQueryCommand(relationTupleRepository, l)
	//
	//		lq := &LookupQueryQuery{
	//			EntityType: "repository",
	//			Action:     "read",
	//			Subject:    &base.Subject{Type: tuple.USER, Id: "1"},
	//		}
	//
	//		var serializedConfigs []string
	//		for _, c := range githubConfigs {
	//			serializedConfigs = append(serializedConfigs, c.Serialized())
	//		}
	//
	//		sch, err := compiler.NewSchema(serializedConfigs...)
	//		Expect(err).ShouldNot(HaveOccurred())
	//
	//		en, err := schema.GetEntityByName(sch, lq.EntityType)
	//		Expect(err).ShouldNot(HaveOccurred())
	//
	//		ac, err := schema.GetActionByNameInEntityDefinition(en, "read")
	//		Expect(err).ShouldNot(HaveOccurred())
	//
	//		lq.SetSchema(sch)
	//
	//		actualResult, err := lookupQueryCommand.Execute(context.Background(), lq, ac.Child)
	//		query, _, e := goqu.Select("*").From("repositories").Where(goqu.And(goqu.I("repositories.owner_id").Eq("1")), goqu.And(goqu.And(goqu.I("repositories.organization_id").Eq("6"), goqu.I("repositories.organization_id").Eq("3")))).Prepared(true).ToSQL()
	//
	//		Expect(e).ShouldNot(HaveOccurred())
	//		Expect(actualResult.Query).Should(Equal(strings.ReplaceAll(query, "\"", "")))
	//		Expect(err).ShouldNot(HaveOccurred())
	//	})
	//
	//	It("Github Sample: Case 2", func() {
	//		relationTupleRepository := new(mocks.RelationshipReader)
	//
	//		lookupQueryCommand = NewLookupQueryCommand(relationTupleRepository, l)
	//
	//		lq := &LookupQueryQuery{
	//			EntityType: "repository",
	//			Action:     "read",
	//			Subject:    &base.Subject{Type: tuple.USER, Id: "1"},
	//		}
	//
	//		var serializedConfigs []string
	//		for _, c := range githubConfigs {
	//			serializedConfigs = append(serializedConfigs, c.Serialized())
	//		}
	//
	//		sch, err := compiler.NewSchema(serializedConfigs...)
	//		Expect(err).ShouldNot(HaveOccurred())
	//
	//		en, err := schema.GetEntityByName(sch, lq.EntityType)
	//		Expect(err).ShouldNot(HaveOccurred())
	//
	//		ac, err := schema.GetActionByNameInEntityDefinition(en, "push")
	//		Expect(err).ShouldNot(HaveOccurred())
	//
	//		lq.SetSchema(sch)
	//
	//		actualResult, err := lookupQueryCommand.Execute(context.Background(), lq, ac.Child)
	//		query, _, e := goqu.Select("*").From("repositories").Where(goqu.And(goqu.I("repositories.owner_id").Eq("1"))).Prepared(true).ToSQL()
	//
	//		Expect(e).ShouldNot(HaveOccurred())
	//		Expect(actualResult.Query).Should(Equal(strings.ReplaceAll(query, "\"", "")))
	//		Expect(err).ShouldNot(HaveOccurred())
	//	})
	//
	//	It("Github Sample: Case 3", func() {
	//		relationTupleRepository := new(mocks.RelationshipReader)
	//
	//		getOrganizationAdmins := []*base.Tuple{
	//			{
	//				Entity: &base.Entity{
	//					Type: "organization",
	//					Id:   "6",
	//				},
	//				Relation: "admin",
	//				Subject: &base.Subject{
	//					Type:     tuple.USER,
	//					Id:       "1",
	//					Relation: "",
	//				},
	//			},
	//			{
	//				Entity: &base.Entity{
	//					Type: "organization",
	//					Id:   "6",
	//				},
	//				Relation: "admin",
	//				Subject: &base.Subject{
	//					Type:     "organization",
	//					Id:       "3",
	//					Relation: "member",
	//				},
	//			},
	//		}
	//
	//		getOrganizationMembers1 := []*base.Tuple{
	//			{
	//				Entity: &base.Entity{
	//					Type: "organization",
	//					Id:   "6",
	//				},
	//				Relation: "member",
	//				Subject: &base.Subject{
	//					Type:     tuple.USER,
	//					Id:       "1",
	//					Relation: "",
	//				},
	//			},
	//		}
	//
	//		getOrganizationMembers2 := []*base.Tuple{
	//			{
	//				Entity: &base.Entity{
	//					Type: "organization",
	//					Id:   "3",
	//				},
	//				Relation: "member",
	//				Subject: &base.Subject{
	//					Type:     tuple.USER,
	//					Id:       "9",
	//					Relation: "",
	//				},
	//			},
	//		}
	//
	//		relationTupleRepository.On("QueryTuples", "organization", "6", "admin").Return(database.NewTupleCollection(getOrganizationAdmins...).CreateTupleIterator(), nil).Times(1)
	//		relationTupleRepository.On("QueryTuples", "organization", "3", "member").Return(database.NewTupleCollection(getOrganizationMembers2...).CreateTupleIterator(), nil).Times(1)
	//		relationTupleRepository.On("ReverseQueryTuples", "organization", "member", "organization", []string{"6"}, "admin").Return(database.NewTupleCollection(getOrganizationMembers1...).CreateTupleIterator(), nil).Times(1)
	//		relationTupleRepository.On("ReverseQueryTuples", "organization", "admin", "organization", []string{"6"}, "admin").Return(database.NewTupleCollection(getOrganizationAdmins...).CreateTupleIterator(), nil).Times(1)
	//
	//		lookupQueryCommand = NewLookupQueryCommand(relationTupleRepository, l)
	//
	//		lq := &LookupQueryQuery{
	//			EntityType: "repository",
	//			Action:     "read",
	//			Subject:    &base.Subject{Type: "organization", Id: "6", Relation: "admin"},
	//		}
	//
	//		var serializedConfigs []string
	//		for _, c := range githubConfigs {
	//			serializedConfigs = append(serializedConfigs, c.Serialized())
	//		}
	//
	//		sch, err := compiler.NewSchema(serializedConfigs...)
	//		Expect(err).ShouldNot(HaveOccurred())
	//
	//		en, err := schema.GetEntityByName(sch, lq.EntityType)
	//		Expect(err).ShouldNot(HaveOccurred())
	//
	//		ac, err := schema.GetActionByNameInEntityDefinition(en, "read")
	//		Expect(err).ShouldNot(HaveOccurred())
	//
	//		lq.SetSchema(sch)
	//
	//		actualResult, err := lookupQueryCommand.Execute(context.Background(), lq, ac.Child)
	//		query, _, e := goqu.Select("*").From("repositories").Where(goqu.And(goqu.I("repositories.owner_id").In("1", "3")), goqu.And(goqu.And(goqu.I("repositories.organization_id").Eq("6"), goqu.I("repositories.organization_id").Eq("6")))).Prepared(true).ToSQL()
	//
	//		Expect(e).ShouldNot(HaveOccurred())
	//		Expect(actualResult.Query).Should(Equal(strings.ReplaceAll(query, "\"", "")))
	//		Expect(err).ShouldNot(HaveOccurred())
	//	})
	//})
})
