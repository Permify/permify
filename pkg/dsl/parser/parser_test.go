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
		It("Case 1", func() {
			pr := NewParser(`
			entity repository {

			relation parent @organization
			relation owner  @user

			action read = owner and (parent.admin and not parent.member)

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

			a1 := st.ActionStatements[0].(*ast.ActionStatement)
			Expect(a1.Name.Literal).Should(Equal("read"))

			es := a1.ExpressionStatement.(*ast.ExpressionStatement)

			Expect(es.Expression.(*ast.InfixExpression).Left.(*ast.Identifier).String()).Should(Equal("owner"))
			Expect(es.Expression.(*ast.InfixExpression).Right.(*ast.InfixExpression).Left.(*ast.Identifier).String()).Should(Equal("parent.admin"))
			Expect(es.Expression.(*ast.InfixExpression).Right.(*ast.InfixExpression).Right.(*ast.Identifier).String()).Should(Equal("not parent.member"))
		})

		It("Case 2", func() {
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

			a1 := st.ActionStatements[0].(*ast.ActionStatement)
			Expect(a1.Name.Literal).Should(Equal("read"))

			es := a1.ExpressionStatement.(*ast.ExpressionStatement)

			Expect(es.Expression.(*ast.InfixExpression).Left.(*ast.InfixExpression).Left.(*ast.Identifier).String()).Should(Equal("owner"))
			Expect(es.Expression.(*ast.InfixExpression).Left.(*ast.InfixExpression).Right.(*ast.Identifier).String()).Should(Equal("parent.admin"))
			Expect(es.Expression.(*ast.InfixExpression).Right.(*ast.Identifier).String()).Should(Equal("parent.member"))
		})

		It("Case 3", func() {
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

			a1 := st.ActionStatements[0].(*ast.ActionStatement)
			Expect(a1.Name.Literal).Should(Equal("delete"))

			es := a1.ExpressionStatement.(*ast.ExpressionStatement)

			Expect(es.Expression.(*ast.Identifier).String()).Should(Equal("owner"))
		})

		It("Case 4", func() {
			pr := NewParser("entity organization {\n\nrelation owner @user\n\naction delete = not owner\n\n\n}\n\n")
			schema, err := pr.Parse()
			Expect(err).ShouldNot(HaveOccurred())

			st := schema.Statements[0].(*ast.EntityStatement)

			Expect(st.Name.Literal).Should(Equal("organization"))

			r1 := st.RelationStatements[0].(*ast.RelationStatement)
			Expect(r1.Name.Literal).Should(Equal("owner"))

			for _, a := range r1.RelationTypes {
				Expect(a.Type.Literal).Should(Equal("user"))
			}

			a1 := st.ActionStatements[0].(*ast.ActionStatement)
			Expect(a1.Name.Literal).Should(Equal("delete"))

			es := a1.ExpressionStatement.(*ast.ExpressionStatement)

			Expect(es.Expression.(*ast.Identifier).String()).Should(Equal("not owner"))
		})

		It("Case 5", func() {
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

			a1 := st.ActionStatements[0].(*ast.ActionStatement)
			Expect(a1.Name.Literal).Should(Equal("view"))

			es1 := a1.ExpressionStatement.(*ast.ExpressionStatement)
			Expect(es1.Expression.(*ast.Identifier).String()).Should(Equal("owner"))

			a2 := st.ActionStatements[1].(*ast.ActionStatement)
			Expect(a2.Name.Literal).Should(Equal("read"))

			es2 := a2.ExpressionStatement.(*ast.ExpressionStatement)

			Expect(es2.Expression.(*ast.InfixExpression).Left.(*ast.Identifier).String()).Should(Equal("view"))
			Expect(es2.Expression.(*ast.InfixExpression).Right.(*ast.InfixExpression).Left.(*ast.Identifier).String()).Should(Equal("parent.admin"))
			Expect(es2.Expression.(*ast.InfixExpression).Right.(*ast.InfixExpression).Right.(*ast.Identifier).String()).Should(Equal("parent.member"))
		})

		It("Case 6", func() {
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
    			action read = (owner and (parent.admin and not parent.member))
    
    			// parent.create_repository means user should be
    			// organization admin or organization member
    			action delete = (owner or (parent.create_repository))
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

			oa1 := organizationSt.ActionStatements[0].(*ast.ActionStatement)
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

			ra1 := repositorySt.ActionStatements[0].(*ast.ActionStatement)
			Expect(ra1.Name.Literal).Should(Equal("read"))

			res1 := ra1.ExpressionStatement.(*ast.ExpressionStatement)

			Expect(res1.Expression.(*ast.InfixExpression).Left.(*ast.Identifier).String()).Should(Equal("owner"))
			Expect(res1.Expression.(*ast.InfixExpression).Right.(*ast.InfixExpression).Left.(*ast.Identifier).String()).Should(Equal("parent.admin"))
			Expect(res1.Expression.(*ast.InfixExpression).Right.(*ast.InfixExpression).Right.(*ast.Identifier).String()).Should(Equal("not parent.member"))

			ra2 := repositorySt.ActionStatements[1].(*ast.ActionStatement)
			Expect(ra2.Name.Literal).Should(Equal("delete"))

			res2 := ra2.ExpressionStatement.(*ast.ExpressionStatement)

			Expect(res2.Expression.(*ast.InfixExpression).Left.(*ast.Identifier).String()).Should(Equal("owner"))
			Expect(res2.Expression.(*ast.InfixExpression).Right.(*ast.Identifier).String()).Should(Equal("parent.create_repository"))
		})
	})
})
