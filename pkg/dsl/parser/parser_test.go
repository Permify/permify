package parser

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Permify/permify/pkg/dsl/ast"
)

// TestParser -
func TestParser(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "parser-suite")
}

var _ = Describe("parser", func() {
	Context("Statement", func() {
		It("Case 1 - Repository with parent and owner relations and read action", func() {
			pr := NewParser(`
			entity repository {
		
			relation parent @organization
			relation owner  @user
		
			action read = owner and (parent.admin not parent.member)
		
			}`)

			schema, err := pr.Parse()
			Expect(err).ShouldNot(HaveOccurred())
			st := schema.Statements[0].(*ast.EntityStatement)

			Expect(st.Name.Literal).Should(Equal("repository"))

			r1 := st.RelationStatements[0].(*ast.RelationStatement)
			Expect(r1.Name.Literal).Should(Equal("parent"))

			for _, a := range r1.RelationTypes {
				Expect(a.Type.Literal).Should(Equal("organization"))
			}

			r2 := st.RelationStatements[1].(*ast.RelationStatement)
			Expect(r2.Name.Literal).Should(Equal("owner"))

			for _, a := range r2.RelationTypes {
				Expect(a.Type.Literal).Should(Equal("user"))
			}

			a1 := st.PermissionStatements[0].(*ast.PermissionStatement)
			Expect(a1.Name.Literal).Should(Equal("read"))

			es := a1.ExpressionStatement.(*ast.ExpressionStatement)

			Expect(es.Expression.(*ast.InfixExpression).Left.(*ast.Identifier).String()).Should(Equal("owner"))
			Expect(es.Expression.(*ast.InfixExpression).Right.(*ast.InfixExpression).Left.(*ast.Identifier).String()).Should(Equal("parent.admin"))
			Expect(es.Expression.(*ast.InfixExpression).Right.(*ast.InfixExpression).Right.(*ast.Identifier).String()).Should(Equal("parent.member"))
		})

		It("Case 2 - Repository with parent and owner relations and read action", func() {
			pr := NewParser(`
			entity repository {
				relation parent   @organization
				relation owner  @user
		
				action read = (owner and parent.admin) and parent.member
			}`)

			schema, err := pr.Parse()
			Expect(err).ShouldNot(HaveOccurred())
			st := schema.Statements[0].(*ast.EntityStatement)

			Expect(st.Name.Literal).Should(Equal("repository"))

			r1 := st.RelationStatements[0].(*ast.RelationStatement)
			Expect(r1.Name.Literal).Should(Equal("parent"))

			for _, a := range r1.RelationTypes {
				Expect(a.Type.Literal).Should(Equal("organization"))
			}

			r2 := st.RelationStatements[1].(*ast.RelationStatement)
			Expect(r2.Name.Literal).Should(Equal("owner"))

			for _, a := range r2.RelationTypes {
				Expect(a.Type.Literal).Should(Equal("user"))
			}

			a1 := st.PermissionStatements[0].(*ast.PermissionStatement)
			Expect(a1.Name.Literal).Should(Equal("read"))

			es := a1.ExpressionStatement.(*ast.ExpressionStatement)

			Expect(es.Expression.(*ast.InfixExpression).Left.(*ast.InfixExpression).Left.(*ast.Identifier).String()).Should(Equal("owner"))
			Expect(es.Expression.(*ast.InfixExpression).Left.(*ast.InfixExpression).Right.(*ast.Identifier).String()).Should(Equal("parent.admin"))
			Expect(es.Expression.(*ast.InfixExpression).Right.(*ast.Identifier).String()).Should(Equal("parent.member"))
		})

		It("Case 3 - Organization with owner relation and delete action", func() {
			pr := NewParser(`
			entity organization {
				relation owner @user
				action delete = owner
			}
			`)
			schema, err := pr.Parse()
			Expect(err).ShouldNot(HaveOccurred())
			st := schema.Statements[0].(*ast.EntityStatement)

			Expect(st.Name.Literal).Should(Equal("organization"))

			r1 := st.RelationStatements[0].(*ast.RelationStatement)
			Expect(r1.Name.Literal).Should(Equal("owner"))

			for _, a := range r1.RelationTypes {
				Expect(a.Type.Literal).Should(Equal("user"))
			}

			a1 := st.PermissionStatements[0].(*ast.PermissionStatement)
			Expect(a1.Name.Literal).Should(Equal("delete"))

			es := a1.ExpressionStatement.(*ast.ExpressionStatement)

			Expect(es.Expression.(*ast.Identifier).String()).Should(Equal("owner"))
		})

		It("Case 4: Organization with owner relation and delete action", func() {
			pr := NewParser("entity organization {\n\nrelation owner @user\n\naction delete = owner\n\n\n}\n\n")
			schema, err := pr.Parse()
			Expect(err).ShouldNot(HaveOccurred())

			st := schema.Statements[0].(*ast.EntityStatement)

			Expect(st.Name.Literal).Should(Equal("organization"))

			r1 := st.RelationStatements[0].(*ast.RelationStatement)
			Expect(r1.Name.Literal).Should(Equal("owner"))

			for _, a := range r1.RelationTypes {
				Expect(a.Type.Literal).Should(Equal("user"))
			}

			a1 := st.PermissionStatements[0].(*ast.PermissionStatement)
			Expect(a1.Name.Literal).Should(Equal("delete"))

			es := a1.ExpressionStatement.(*ast.ExpressionStatement)

			Expect(es.Expression.(*ast.Identifier).String()).Should(Equal("owner"))
		})

		It("Case 5 - Repository view and read actions with ownership and parent organization", func() {
			pr := NewParser(`
			entity repository {
		
				relation parent  @organization
				relation owner  @user @organization#member
		
				action view = owner
				action read = view and (parent.admin and parent.member)
			}
			`)

			schema, err := pr.Parse()
			Expect(err).ShouldNot(HaveOccurred())

			st := schema.Statements[0].(*ast.EntityStatement)

			Expect(st.Name.Literal).Should(Equal("repository"))

			r1 := st.RelationStatements[0].(*ast.RelationStatement)
			Expect(r1.Name.Literal).Should(Equal("parent"))

			for _, a := range r1.RelationTypes {
				Expect(a.Type.Literal).Should(Equal("organization"))
			}

			r2 := st.RelationStatements[1].(*ast.RelationStatement)
			Expect(r2.Name.Literal).Should(Equal("owner"))

			Expect(r2.RelationTypes[0].Type.Literal).Should(Equal("user"))
			Expect(r2.RelationTypes[1].Type.Literal).Should(Equal("organization"))
			Expect(r2.RelationTypes[1].Relation.Literal).Should(Equal("member"))

			a1 := st.PermissionStatements[0].(*ast.PermissionStatement)
			Expect(a1.Name.Literal).Should(Equal("view"))

			es1 := a1.ExpressionStatement.(*ast.ExpressionStatement)
			Expect(es1.Expression.(*ast.Identifier).String()).Should(Equal("owner"))

			a2 := st.PermissionStatements[1].(*ast.PermissionStatement)
			Expect(a2.Name.Literal).Should(Equal("read"))

			es2 := a2.ExpressionStatement.(*ast.ExpressionStatement)

			Expect(es2.Expression.(*ast.InfixExpression).Left.(*ast.Identifier).String()).Should(Equal("view"))
			Expect(es2.Expression.(*ast.InfixExpression).Right.(*ast.InfixExpression).Left.(*ast.Identifier).String()).Should(Equal("parent.admin"))
			Expect(es2.Expression.(*ast.InfixExpression).Right.(*ast.InfixExpression).Right.(*ast.Identifier).String()).Should(Equal("parent.member"))
		})

		It("Case 6 - Complex organization and repository permissions", func() {
			pr := NewParser(`
			entity user {}

			entity organization {
    			// relations
				relation admin @user
    			relation member @user

				// actions
    			action create_repository = (admin or member)
			}

			entity repository {
    			// relations
    			relation owner @user @organization#member
    			relation parent @organization
    
    			// actions
    			permission read = (owner and (parent.admin not parent.member)) or owner
    
    			// parent.create_repository means user should be
    			// organization admin or organization member
    			permission delete = (owner or (parent.create_repository))
			}
			`)

			schema, err := pr.Parse()
			Expect(err).ShouldNot(HaveOccurred())

			// USER
			userSt := schema.Statements[0].(*ast.EntityStatement)
			Expect(userSt.Name.Literal).Should(Equal("user"))

			// ORGANIZATION
			organizationSt := schema.Statements[1].(*ast.EntityStatement)

			Expect(organizationSt.Name.Literal).Should(Equal("organization"))

			or1 := organizationSt.RelationStatements[0].(*ast.RelationStatement)
			Expect(or1.Name.Literal).Should(Equal("admin"))

			Expect(or1.RelationTypes[0].Type.Literal).Should(Equal("user"))

			or2 := organizationSt.RelationStatements[1].(*ast.RelationStatement)
			Expect(or2.Name.Literal).Should(Equal("member"))

			Expect(or2.RelationTypes[0].Type.Literal).Should(Equal("user"))

			oa1 := organizationSt.PermissionStatements[0].(*ast.PermissionStatement)
			Expect(oa1.Name.Literal).Should(Equal("create_repository"))

			oes1 := oa1.ExpressionStatement.(*ast.ExpressionStatement)

			Expect(oes1.Expression.(*ast.InfixExpression).Left.(*ast.Identifier).String()).Should(Equal("admin"))
			Expect(oes1.Expression.(*ast.InfixExpression).Right.(*ast.Identifier).String()).Should(Equal("member"))

			// REPOSITORY

			repositorySt := schema.Statements[2].(*ast.EntityStatement)

			Expect(repositorySt.Name.Literal).Should(Equal("repository"))

			rr1 := repositorySt.RelationStatements[0].(*ast.RelationStatement)
			Expect(rr1.Name.Literal).Should(Equal("owner"))

			Expect(rr1.RelationTypes[0].Type.Literal).Should(Equal("user"))
			Expect(rr1.RelationTypes[1].Type.Literal).Should(Equal("organization"))
			Expect(rr1.RelationTypes[1].Relation.Literal).Should(Equal("member"))

			rr2 := repositorySt.RelationStatements[1].(*ast.RelationStatement)
			Expect(rr2.Name.Literal).Should(Equal("parent"))

			Expect(rr2.RelationTypes[0].Type.Literal).Should(Equal("organization"))

			ra1 := repositorySt.PermissionStatements[0].(*ast.PermissionStatement)
			Expect(ra1.Name.Literal).Should(Equal("read"))

			res1 := ra1.ExpressionStatement.(*ast.ExpressionStatement)

			Expect(res1.Expression.(*ast.InfixExpression).Left.(*ast.InfixExpression).Left.(*ast.Identifier).String()).Should(Equal("owner"))
			Expect(res1.Expression.(*ast.InfixExpression).Left.(*ast.InfixExpression).Right.(*ast.InfixExpression).Left.(*ast.Identifier).String()).Should(Equal("parent.admin"))
			Expect(res1.Expression.(*ast.InfixExpression).Left.(*ast.InfixExpression).Right.(*ast.InfixExpression).Right.(*ast.Identifier).String()).Should(Equal("parent.member"))
			Expect(res1.Expression.(*ast.InfixExpression).Right.(*ast.Identifier).String()).Should(Equal("owner"))

			ra2 := repositorySt.PermissionStatements[1].(*ast.PermissionStatement)
			Expect(ra2.Name.Literal).Should(Equal("delete"))

			res2 := ra2.ExpressionStatement.(*ast.ExpressionStatement)

			Expect(res2.Expression.(*ast.InfixExpression).Left.(*ast.Identifier).String()).Should(Equal("owner"))
			Expect(res2.Expression.(*ast.InfixExpression).Right.(*ast.Identifier).String()).Should(Equal("parent.create_repository"))
		})

		It("Case 7 - Multiple actions", func() {
			pr := NewParser(`
		entity user {}

		entity organization {
			// relations
			relation admin @user
			relation member @user

			// actions
			action create_repository = (admin or member)
			action manage_team = (admin)
		}

		entity team {
			// relations
			relation leader @user
			relation member @user

			// actions
			permission add_member = (leader or (parent.manage_team))
		}
		`)

			schema, err := pr.Parse()
			Expect(err).ShouldNot(HaveOccurred())

			// USER
			userSt := schema.Statements[0].(*ast.EntityStatement)
			Expect(userSt.Name.Literal).Should(Equal("user"))

			// ORGANIZATION
			organizationSt := schema.Statements[1].(*ast.EntityStatement)
			Expect(organizationSt.Name.Literal).Should(Equal("organization"))

			oa2 := organizationSt.PermissionStatements[1].(*ast.PermissionStatement)
			Expect(oa2.Name.Literal).Should(Equal("manage_team"))

			oes2 := oa2.ExpressionStatement.(*ast.ExpressionStatement)
			Expect(oes2.Expression.(*ast.Identifier).String()).Should(Equal("admin"))

			// TEAM
			teamSt := schema.Statements[2].(*ast.EntityStatement)
			Expect(teamSt.Name.Literal).Should(Equal("team"))

			tperm1 := teamSt.PermissionStatements[0].(*ast.PermissionStatement)
			Expect(tperm1.Name.Literal).Should(Equal("add_member"))

			tes1 := tperm1.ExpressionStatement.(*ast.ExpressionStatement)
			Expect(tes1.Expression.(*ast.InfixExpression).Left.(*ast.Identifier).String()).Should(Equal("leader"))
			Expect(tes1.Expression.(*ast.InfixExpression).Right.(*ast.Identifier).String()).Should(Equal("parent.manage_team"))
		})

		It("Case 8 - Complex nested expressions", func() {
			pr := NewParser(`
	entity user {}

	entity organization {
		// relations
		relation admin @user
		relation member @user

		// actions
		action manage_organization = ((admin not member) or (member not admin))
	}

	entity team {
		// relations
		relation leader @user
		relation member @user

		// actions
		permission manage_team = ((leader and parent.manage_organization) or (member not parent.manage_organization))
	}
	`)

			schema, err := pr.Parse()
			Expect(err).ShouldNot(HaveOccurred())

			// USER
			userSt := schema.Statements[0].(*ast.EntityStatement)
			Expect(userSt.Name.Literal).Should(Equal("user"))

			// ORGANIZATION
			organizationSt := schema.Statements[1].(*ast.EntityStatement)
			Expect(organizationSt.Name.Literal).Should(Equal("organization"))

			oa1 := organizationSt.PermissionStatements[0].(*ast.PermissionStatement)
			Expect(oa1.Name.Literal).Should(Equal("manage_organization"))

			oes1 := oa1.ExpressionStatement.(*ast.ExpressionStatement)
			Expect(oes1.Expression.(*ast.InfixExpression).Left.(*ast.InfixExpression).Left.(*ast.Identifier).String()).Should(Equal("admin"))
			Expect(oes1.Expression.(*ast.InfixExpression).Left.(*ast.InfixExpression).Right.(*ast.Identifier).String()).Should(Equal("member"))
			Expect(oes1.Expression.(*ast.InfixExpression).Right.(*ast.InfixExpression).Left.(*ast.Identifier).String()).Should(Equal("member"))
			Expect(oes1.Expression.(*ast.InfixExpression).Right.(*ast.InfixExpression).Right.(*ast.Identifier).String()).Should(Equal("admin"))

			// TEAM
			teamSt := schema.Statements[2].(*ast.EntityStatement)
			Expect(teamSt.Name.Literal).Should(Equal("team"))

			tperm1 := teamSt.PermissionStatements[0].(*ast.PermissionStatement)
			Expect(tperm1.Name.Literal).Should(Equal("manage_team"))

			tes1 := tperm1.ExpressionStatement.(*ast.ExpressionStatement)
			Expect(tes1.Expression.(*ast.InfixExpression).Left.(*ast.InfixExpression).Left.(*ast.Identifier).String()).Should(Equal("leader"))
			Expect(tes1.Expression.(*ast.InfixExpression).Left.(*ast.InfixExpression).Right.(*ast.Identifier).String()).Should(Equal("parent.manage_organization"))
			Expect(tes1.Expression.(*ast.InfixExpression).Right.(*ast.InfixExpression).Left.(*ast.Identifier).String()).Should(Equal("member"))
			Expect(tes1.Expression.(*ast.InfixExpression).Right.(*ast.InfixExpression).Right.(*ast.Identifier).String()).Should(Equal("parent.manage_organization"))
		})

		It("Case 9 - More complex nested expressions", func() {
			pr := NewParser(`
	entity user {}

	entity organization {
		// relations
		relation admin @user
		relation member @user

		// actions
		action manage_organization = (((admin not member) or member) not (admin and member))
	}

	entity project {
		// relations
		relation owner @user
		relation contributor @user

		// actions
		permission manage_project = ((owner and (parent.admin or parent.member)) or (contributor not parent.manage_organization not (parent.admin and parent.member)))
	}
	`)

			schema, err := pr.Parse()
			Expect(err).ShouldNot(HaveOccurred())

			// USER
			userSt := schema.Statements[0].(*ast.EntityStatement)
			Expect(userSt.Name.Literal).Should(Equal("user"))

			// ORGANIZATION
			organizationSt := schema.Statements[1].(*ast.EntityStatement)
			Expect(organizationSt.Name.Literal).Should(Equal("organization"))

			oa1 := organizationSt.PermissionStatements[0].(*ast.PermissionStatement)
			Expect(oa1.Name.Literal).Should(Equal("manage_organization"))

			oes1 := oa1.ExpressionStatement.(*ast.ExpressionStatement)
			Expect(oes1.Expression.(*ast.InfixExpression).Left.(*ast.InfixExpression).Left.(*ast.InfixExpression).Left.(*ast.Identifier).String()).Should(Equal("admin"))
			Expect(oes1.Expression.(*ast.InfixExpression).Left.(*ast.InfixExpression).Left.(*ast.InfixExpression).Right.(*ast.Identifier).String()).Should(Equal("member"))
			Expect(oes1.Expression.(*ast.InfixExpression).Left.(*ast.InfixExpression).Right.(*ast.Identifier).String()).Should(Equal("member"))
			Expect(oes1.Expression.(*ast.InfixExpression).Right.(*ast.InfixExpression).Left.(*ast.Identifier).String()).Should(Equal("admin"))
			Expect(oes1.Expression.(*ast.InfixExpression).Right.(*ast.InfixExpression).Right.(*ast.Identifier).String()).Should(Equal("member"))

			// PROJECT
			projectSt := schema.Statements[2].(*ast.EntityStatement)
			Expect(projectSt.Name.Literal).Should(Equal("project"))

			p1 := projectSt.PermissionStatements[0].(*ast.PermissionStatement)
			Expect(p1.Name.Literal).Should(Equal("manage_project"))

			eps1 := p1.ExpressionStatement.(*ast.ExpressionStatement)
			Expect(eps1.Expression.(*ast.InfixExpression).Left.(*ast.InfixExpression).Left.(*ast.Identifier).String()).Should(Equal("owner"))
			Expect(eps1.Expression.(*ast.InfixExpression).Left.(*ast.InfixExpression).Right.(*ast.InfixExpression).Left.(*ast.Identifier).String()).Should(Equal("parent.admin"))
			Expect(eps1.Expression.(*ast.InfixExpression).Left.(*ast.InfixExpression).Right.(*ast.InfixExpression).Right.(*ast.Identifier).String()).Should(Equal("parent.member"))
			Expect(eps1.Expression.(*ast.InfixExpression).Right.(*ast.InfixExpression).Left.(*ast.InfixExpression).Left.(*ast.Identifier).String()).Should(Equal("contributor"))
			Expect(eps1.Expression.(*ast.InfixExpression).Right.(*ast.InfixExpression).Right.(*ast.InfixExpression).Left.(*ast.Identifier).String()).Should(Equal("parent.admin"))
			Expect(eps1.Expression.(*ast.InfixExpression).Right.(*ast.InfixExpression).Right.(*ast.InfixExpression).Left.(*ast.Identifier).String()).Should(Equal("parent.admin"))
			Expect(eps1.Expression.(*ast.InfixExpression).Right.(*ast.InfixExpression).Right.(*ast.InfixExpression).Right.(*ast.Identifier).String()).Should(Equal("parent.member"))
		})

		It("Case 10 - Duplicate entity", func() {
			pr := NewParser(`
        entity user {}
        entity user {}
    `)

			_, err := pr.Parse()

			// Ensure an error is returned
			Expect(err).Should(HaveOccurred())

			// Ensure the error message contains the expected string
			Expect(err.Error()).Should(ContainSubstring("3:23:duplication found for user"))
		})

		It("Case 11 - Duplicate Relation", func() {
			pr := NewParser(`
				entity organization {
					relation member @user
					relation member @user
				} `)

			_, err := pr.Parse()

			// Ensure an error is returned
			Expect(err).Should(HaveOccurred())

			// Ensure the error message contains the expected string
			Expect(err.Error()).Should(ContainSubstring("5:2:duplication found for organization#member"))
		})

		It("Case 12 - Duplicate action", func() {
			pr := NewParser(`
			entity organization {
				relation admin @user
				action delete = admin
				permission delete = admin
			}`)

			_, err := pr.Parse()

			// Ensure an error is returned
			Expect(err).Should(HaveOccurred())

			// Ensure the error message contains the expected string
			Expect(err.Error()).Should(ContainSubstring("5:25:duplication found for organization#delete"))
		})

		It("Case 13 - Attribute", func() {
			pr := NewParser(`
			entity repository {
		
				relation parent  @organization
				relation owner  @user @organization#member
		
				attribute is_public boolean
		
				action view = owner
				action read = view and (parent.admin and parent.member)
			}
			`)

			schema, err := pr.Parse()
			Expect(err).ShouldNot(HaveOccurred())

			st := schema.Statements[0].(*ast.EntityStatement)

			Expect(st.Name.Literal).Should(Equal("repository"))

			r1 := st.RelationStatements[0].(*ast.RelationStatement)
			Expect(r1.Name.Literal).Should(Equal("parent"))

			for _, a := range r1.RelationTypes {
				Expect(a.Type.Literal).Should(Equal("organization"))
			}

			r2 := st.RelationStatements[1].(*ast.RelationStatement)
			Expect(r2.Name.Literal).Should(Equal("owner"))

			Expect(r2.RelationTypes[0].Type.Literal).Should(Equal("user"))
			Expect(r2.RelationTypes[1].Type.Literal).Should(Equal("organization"))
			Expect(r2.RelationTypes[1].Relation.Literal).Should(Equal("member"))

			at1 := st.AttributeStatements[0].(*ast.AttributeStatement)
			Expect(at1.Name.Literal).Should(Equal("is_public"))
			Expect(at1.AttributeType.String()).Should(Equal("boolean"))

			a1 := st.PermissionStatements[0].(*ast.PermissionStatement)
			Expect(a1.Name.Literal).Should(Equal("view"))

			es1 := a1.ExpressionStatement.(*ast.ExpressionStatement)
			Expect(es1.Expression.(*ast.Identifier).String()).Should(Equal("owner"))

			a2 := st.PermissionStatements[1].(*ast.PermissionStatement)
			Expect(a2.Name.Literal).Should(Equal("read"))

			es2 := a2.ExpressionStatement.(*ast.ExpressionStatement)

			Expect(es2.Expression.(*ast.InfixExpression).Left.(*ast.Identifier).String()).Should(Equal("view"))
			Expect(es2.Expression.(*ast.InfixExpression).Right.(*ast.InfixExpression).Left.(*ast.Identifier).String()).Should(Equal("parent.admin"))
			Expect(es2.Expression.(*ast.InfixExpression).Right.(*ast.InfixExpression).Right.(*ast.Identifier).String()).Should(Equal("parent.member"))
		})

		It("Case 14 - Rule", func() {
			pr := NewParser(`
			entity account {
    			relation owner @user
    			attribute balance float

    			permission withdraw = check_balance(request.amount, balance) and owner
			}
	
			rule check_balance(amount float, balance float) {
				balance >= amount && amount <= 5000
			}
			`)

			schema, err := pr.Parse()
			Expect(err).ShouldNot(HaveOccurred())

			st := schema.Statements[0].(*ast.EntityStatement)

			Expect(st.Name.Literal).Should(Equal("account"))

			r1 := st.RelationStatements[0].(*ast.RelationStatement)
			Expect(r1.Name.Literal).Should(Equal("owner"))

			for _, a := range r1.RelationTypes {
				Expect(a.Type.Literal).Should(Equal("user"))
			}

			a1 := st.AttributeStatements[0].(*ast.AttributeStatement)
			Expect(a1.Name.Literal).Should(Equal("balance"))
			Expect(a1.AttributeType.Type.Literal).Should(Equal("float"))

			p1 := st.PermissionStatements[0].(*ast.PermissionStatement)
			Expect(p1.Name.Literal).Should(Equal("withdraw"))

			es1 := p1.ExpressionStatement.(*ast.ExpressionStatement)

			Expect(es1.Expression.(*ast.InfixExpression).Left.(*ast.Call).String()).Should(Equal("check_balance(request.amount, balance)"))
			Expect(es1.Expression.(*ast.InfixExpression).Right.(*ast.Identifier).String()).Should(Equal("owner"))

			rs1 := schema.Statements[1].(*ast.RuleStatement)

			Expect(rs1.Name.Literal).Should(Equal("check_balance"))
			Expect(rs1.Expression).Should(Equal("\nbalance >= amount && amount <= 5000\n\t\t"))
		})

		It("Case 15 - Array", func() {
			pr := NewParser(`
				entity organization {
					
					relation admin @user
				
					attribute location string[]
				
					permission view = check_location(request.current_location, location) or admin
				}
				
				rule check_location(current_location string, location string[]) {
					current_location in location
				}
			`)

			schema, err := pr.Parse()
			Expect(err).ShouldNot(HaveOccurred())

			st := schema.Statements[0].(*ast.EntityStatement)

			Expect(st.Name.Literal).Should(Equal("organization"))

			r1 := st.RelationStatements[0].(*ast.RelationStatement)
			Expect(r1.Name.Literal).Should(Equal("admin"))
			for _, a := range r1.RelationTypes {
				Expect(a.Type.Literal).Should(Equal("user"))
			}

			a1 := st.AttributeStatements[0].(*ast.AttributeStatement)
			Expect(a1.Name.Literal).Should(Equal("location"))
			Expect(a1.AttributeType.Type.Literal).Should(Equal("string"))
			Expect(a1.AttributeType.IsArray).Should(Equal(true))

			p1 := st.PermissionStatements[0].(*ast.PermissionStatement)
			Expect(p1.Name.Literal).Should(Equal("view"))

			es1 := p1.ExpressionStatement.(*ast.ExpressionStatement)

			Expect(es1.Expression.(*ast.InfixExpression).Left.(*ast.Call).String()).Should(Equal("check_location(request.current_location, location)"))
			Expect(es1.Expression.(*ast.InfixExpression).Right.(*ast.Identifier).String()).Should(Equal("admin"))

			rs1 := schema.Statements[1].(*ast.RuleStatement)

			Expect(rs1.Name.Literal).Should(Equal("check_location"))
			Expect(rs1.Expression).Should(Equal("\ncurrent_location in location\n\t\t\t"))
		})

		It("Case 16", func() {
			pr := NewParser(`
			entity & {
			`)

			_, err := pr.Parse()

			// Ensure an error is returned
			Expect(err).Should(HaveOccurred())

			// Ensure the error message contains the expected string
			Expect(err.Error()).Should(ContainSubstring("2:13:expected next token to be IDENT, got AMPERSAND instead"))
		})

		It("Case 17", func() {
			pr := NewParser(`
			entity asd 
			`)

			_, err := pr.Parse()

			// Ensure an error is returned
			Expect(err).Should(HaveOccurred())

			// Ensure the error message contains the expected string
			Expect(err.Error()).Should(ContainSubstring("3:2:expected next token to be LCB, got NEWLINE instead"))
		})

		It("Case 18", func() {
			pr := NewParser(`
			entity asd {
			`)

			_, err := pr.Parse()

			// Ensure an error is returned
			Expect(err).Should(HaveOccurred())

			// Ensure the error message contains the expected string
			Expect(err.Error()).Should(ContainSubstring("3:7:expected token to be RCB, got EOF instead"))
		})

		It("Case 19", func() {
			pr := NewParser(`
			entity asd {

				attribute 987d bool				

			`)

			_, err := pr.Parse()

			// Ensure an error is returned
			Expect(err).Should(HaveOccurred())

			// Ensure the error message contains the expected string
			Expect(err.Error()).Should(ContainSubstring("4:19:expected next token to be IDENT, got INTEGER instead"))
		})

		It("Case 20", func() {
			pr := NewParser(`
			entity asd {

				relation user user				

			`)

			_, err := pr.Parse()

			// Ensure an error is returned
			Expect(err).Should(HaveOccurred())

			// Ensure the error message contains the expected string
			Expect(err.Error()).Should(ContainSubstring("4:24:expected next token to be SIGN, got IDENT instead"))
		})

		It("Case 21", func() {
			pr := NewParser(`
			entity asd {

				relation user @
				relation admin @user

			`)

			_, err := pr.Parse()

			// Ensure an error is returned
			Expect(err).Should(HaveOccurred())

			// Ensure the error message contains the expected string
			Expect(err.Error()).Should(ContainSubstring("5:2:expected next token to be IDENT, got NEWLINE instead"))
		})

		It("Case 22", func() {
			pr := NewParser(`
			entity asd {

				relation admin @user

				permission = admin
			`)

			_, err := pr.Parse()

			// Ensure an error is returned
			Expect(err).Should(HaveOccurred())

			// Ensure the error message contains the expected string
			Expect(err.Error()).Should(ContainSubstring("6:18:expected next token to be IDENT, got ASSIGN instead"))
		})

		It("Case 23", func() {
			p := NewParser(`
				entity repository {
					
					relation admin @user
				    relation member @user

					action read = admin or member
				}
			`)

			schema, err := p.Parse()
			Expect(err).ShouldNot(HaveOccurred())

			p1 := NewParser(`
				relation parent @organization
			`)

			stmt1, err := p1.ParsePartial("repository")
			Expect(err).ShouldNot(HaveOccurred())

			err = schema.AddStatement("repository", stmt1)
			Expect(err).ShouldNot(HaveOccurred())

			p2 := NewParser(`
				relation owner  @user
			`)

			stmt2, err := p2.ParsePartial("repository")
			Expect(err).ShouldNot(HaveOccurred())

			err = schema.AddStatement("repository", stmt2)
			Expect(err).ShouldNot(HaveOccurred())

			err = schema.DeleteStatement("repository", "admin")
			Expect(err).ShouldNot(HaveOccurred())

			err = schema.DeleteStatement("repository", "member")
			Expect(err).ShouldNot(HaveOccurred())

			p3 := NewParser(`
				action read = owner and (parent.admin not parent.member)
			`)

			stmt3, err := p3.ParsePartial("repository")
			Expect(err).ShouldNot(HaveOccurred())

			err = schema.UpdateStatement("repository", stmt3)
			Expect(err).ShouldNot(HaveOccurred())

			st := schema.Statements[0].(*ast.EntityStatement)

			Expect(st.Name.Literal).Should(Equal("repository"))

			r1 := st.RelationStatements[0].(*ast.RelationStatement)
			Expect(r1.Name.Literal).Should(Equal("parent"))

			for _, a := range r1.RelationTypes {
				Expect(a.Type.Literal).Should(Equal("organization"))
			}

			r2 := st.RelationStatements[1].(*ast.RelationStatement)
			Expect(r2.Name.Literal).Should(Equal("owner"))

			for _, a := range r2.RelationTypes {
				Expect(a.Type.Literal).Should(Equal("user"))
			}

			a1 := st.PermissionStatements[0].(*ast.PermissionStatement)
			Expect(a1.Name.Literal).Should(Equal("read"))

			es := a1.ExpressionStatement.(*ast.ExpressionStatement)

			Expect(es.Expression.(*ast.InfixExpression).Left.(*ast.Identifier).String()).Should(Equal("owner"))
			Expect(es.Expression.(*ast.InfixExpression).Right.(*ast.InfixExpression).Left.(*ast.Identifier).String()).Should(Equal("parent.admin"))
			Expect(es.Expression.(*ast.InfixExpression).Right.(*ast.InfixExpression).Right.(*ast.Identifier).String()).Should(Equal("parent.member"))
		})

		It("Case 24 - Multi-line Permission Expression w/ Rule", func() {
			pr := NewParser(`
			entity account {
    			relation owner @user
    			attribute balance float

    			permission withdraw = check_balance(request.amount, balance) and 
					owner
			}
	
			rule check_balance(amount float, balance float) {
				balance >= amount && amount <= 5000
			}
			`)

			schema, err := pr.Parse()
			Expect(err).ShouldNot(HaveOccurred())

			st := schema.Statements[0].(*ast.EntityStatement)

			Expect(st.Name.Literal).Should(Equal("account"))

			r1 := st.RelationStatements[0].(*ast.RelationStatement)
			Expect(r1.Name.Literal).Should(Equal("owner"))

			for _, a := range r1.RelationTypes {
				Expect(a.Type.Literal).Should(Equal("user"))
			}

			a1 := st.AttributeStatements[0].(*ast.AttributeStatement)
			Expect(a1.Name.Literal).Should(Equal("balance"))
			Expect(a1.AttributeType.Type.Literal).Should(Equal("float"))

			p1 := st.PermissionStatements[0].(*ast.PermissionStatement)
			Expect(p1.Name.Literal).Should(Equal("withdraw"))

			es1 := p1.ExpressionStatement.(*ast.ExpressionStatement)

			Expect(es1.Expression.(*ast.InfixExpression).Left.(*ast.Call).String()).Should(Equal("check_balance(request.amount, balance)"))
			Expect(es1.Expression.(*ast.InfixExpression).Right.(*ast.Identifier).String()).Should(Equal("owner"))

			rs1 := schema.Statements[1].(*ast.RuleStatement)

			Expect(rs1.Name.Literal).Should(Equal("check_balance"))
			Expect(rs1.Expression).Should(Equal("\nbalance >= amount && amount <= 5000\n\t\t"))
		})

		It("Case 25 - Multi-line Permission Expression w/ Rule", func() {
			pr := NewParser(`
			entity account {
    			relation owner @user
    			attribute balance float

    			permission withdraw = 
					check_balance(request.amount, balance) and owner
			}
	
			rule check_balance(amount float, balance float) {
				balance >= amount && amount <= 5000
			}
			`)

			schema, err := pr.Parse()
			Expect(err).ShouldNot(HaveOccurred())

			st := schema.Statements[0].(*ast.EntityStatement)

			Expect(st.Name.Literal).Should(Equal("account"))

			r1 := st.RelationStatements[0].(*ast.RelationStatement)
			Expect(r1.Name.Literal).Should(Equal("owner"))

			for _, a := range r1.RelationTypes {
				Expect(a.Type.Literal).Should(Equal("user"))
			}

			a1 := st.AttributeStatements[0].(*ast.AttributeStatement)
			Expect(a1.Name.Literal).Should(Equal("balance"))
			Expect(a1.AttributeType.Type.Literal).Should(Equal("float"))

			p1 := st.PermissionStatements[0].(*ast.PermissionStatement)
			Expect(p1.Name.Literal).Should(Equal("withdraw"))

			es1 := p1.ExpressionStatement.(*ast.ExpressionStatement)

			Expect(es1.Expression.(*ast.InfixExpression).Left.(*ast.Call).String()).Should(Equal("check_balance(request.amount, balance)"))
			Expect(es1.Expression.(*ast.InfixExpression).Right.(*ast.Identifier).String()).Should(Equal("owner"))

			rs1 := schema.Statements[1].(*ast.RuleStatement)

			Expect(rs1.Name.Literal).Should(Equal("check_balance"))
			Expect(rs1.Expression).Should(Equal("\nbalance >= amount && amount <= 5000\n\t\t"))
		})

		It("Case 26 - Multi-line Permission Expression w/ Rule - should fail", func() {
			pr := NewParser(`
			entity account {
    			relation owner @user
    			attribute balance float

    			permission withdraw = check_balance(request.amount, balance)
					owner
			}
	
			rule check_balance(amount float, balance float) {
				balance >= amount && amount <= 5000
			}
			`)

			_, err := pr.Parse()
			Expect(err).Should(HaveOccurred())

			// Ensure an error is returned
			Expect(err).Should(HaveOccurred())

			// Ensure the error message contains the expected string
			Expect(err.Error()).Should(ContainSubstring("8:2:expected token to be RELATION, PERMISSION, ATTRIBUTE, got IDENT instead"))
		})

		It("Case 27 - Multi-line Permission Complex Expression w/ Rule", func() {
			pr := NewParser(`
entity report {
    relation parent @organization
    relation team @team
    attribute confidentiality_level double

    permission view = 
		confidentiality_level_high(confidentiality_level) and 
		parent.director or 
		confidentiality_level_medium_high(confidentiality_level) and 
		(parent.director or team.lead) or 
		confidentiality_level_medium(confidentiality_level) and (team.lead or team.member) or 
		confidentiality_level_low(confidentiality_level) and 
		parent.member
    permission edit = team.lead
}

rule confidentiality_level_high(confidentiality_level double) {
    confidentiality_level == 4.0
}

rule confidentiality_level_medium_high(confidentiality_level double) {
    confidentiality_level == 3.0
}

rule confidentiality_level_medium(confidentiality_level double) {
    confidentiality_level == 2.0
}

rule confidentiality_level_low(confidentiality_level double) {
    confidentiality_level == 1.0
}
			`)

			_, err := pr.Parse()
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("Case 28 - Multi-line Permission Expression w/ Rule", func() {
			pr := NewParser(`
			entity account {
    			relation owner @user
    			relation admin @user

    			permission withdraw = admin or 
					owner
			}
			`)

			_, err := pr.Parse()
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("Case 29 - Multi-line Permission Expression w/ Rule - should fail", func() {
			pr := NewParser(`
			entity account {
    			relation owner @user
    			relation admin @user

    			permission withdraw = admin 
					or owner
			}
			`)

			_, err := pr.Parse()
			// Ensure an error is returned
			Expect(err).Should(HaveOccurred())

			// Ensure the error message contains the expected string
			Expect(err.Error()).Should(ContainSubstring("7:15:expected token to be RELATION, PERMISSION, ATTRIBUTE, got OR instead"))
		})
	})
})
