package parser

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Permify/permify/pkg/dsl/ast"
)

// TestLexer -
func TestParser(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "parser-suite")
}

var _ = Describe("parser", func() {
	Context("Statement", func() {
		It("Case 1", func() {
			pr := NewParser("entity repository {\n\nrelation parent @organization \nrelation owner  @user\naction read = owner and (parent.admin and not parent.member)\n\n\n} `table:repository|identifier:id`\n\n")
			schema := pr.Parse()
			st := schema.Statements[0].(*ast.EntityStatement)

			Expect(st.Name.Literal).Should(Equal("repository"))
			Expect(st.Option.Literal).Should(Equal(`table:repository|identifier:id`))

			r1 := st.RelationStatements[0].(*ast.RelationStatement)
			Expect(r1.Name.Literal).Should(Equal("parent"))

			for _, a := range r1.RelationTypes {
				Expect(a.TokenLiteral()).Should(Equal("organization"))
			}

			r2 := st.RelationStatements[1].(*ast.RelationStatement)
			Expect(r2.Name.Literal).Should(Equal("owner"))

			for _, a := range r2.RelationTypes {
				Expect(a.TokenLiteral()).Should(Equal("user"))
			}

			a1 := st.ActionStatements[0].(*ast.ActionStatement)
			Expect(a1.Name.Literal).Should(Equal("read"))

			es := a1.ExpressionStatement.(*ast.ExpressionStatement)

			Expect(es.Expression.(*ast.InfixExpression).Left.(*ast.Identifier).String()).Should(Equal("owner"))
			Expect(es.Expression.(*ast.InfixExpression).Right.(*ast.InfixExpression).Left.(*ast.Identifier).String()).Should(Equal("parent.admin"))
			Expect(es.Expression.(*ast.InfixExpression).Right.(*ast.InfixExpression).Right.(*ast.PrefixExpression).String()).Should(Equal("not parent.member"))
		})

		It("Case 2", func() {
			pr := NewParser("entity repository {\n\nrelation parent   @organization `rel:belongs-to|cols:organization_id`\nrelation owner  @user `rel:belongs-to|cols:owner_id`\n\naction read = (owner and parent.admin) and parent.member\n\n\n} `table:repository|identifier:id`\n\n")
			schema := pr.Parse()
			st := schema.Statements[0].(*ast.EntityStatement)

			Expect(st.Name.Literal).Should(Equal("repository"))
			Expect(st.Option.Literal).Should(Equal(`table:repository|identifier:id`))

			r1 := st.RelationStatements[0].(*ast.RelationStatement)
			Expect(r1.Name.Literal).Should(Equal("parent"))

			for _, a := range r1.RelationTypes {
				Expect(a.TokenLiteral()).Should(Equal("organization"))
			}

			Expect(r1.Option.Literal).Should(Equal(`rel:belongs-to|cols:organization_id`))

			r2 := st.RelationStatements[1].(*ast.RelationStatement)
			Expect(r2.Name.Literal).Should(Equal("owner"))

			for _, a := range r2.RelationTypes {
				Expect(a.TokenLiteral()).Should(Equal("user"))
			}

			Expect(r2.Option.Literal).Should(Equal(`rel:belongs-to|cols:owner_id`))

			a1 := st.ActionStatements[0].(*ast.ActionStatement)
			Expect(a1.Name.Literal).Should(Equal("read"))

			es := a1.ExpressionStatement.(*ast.ExpressionStatement)

			Expect(es.Expression.(*ast.InfixExpression).Left.(*ast.InfixExpression).Left.(*ast.Identifier).String()).Should(Equal("owner"))
			Expect(es.Expression.(*ast.InfixExpression).Left.(*ast.InfixExpression).Right.(*ast.Identifier).String()).Should(Equal("parent.admin"))
			Expect(es.Expression.(*ast.InfixExpression).Right.(*ast.Identifier).String()).Should(Equal("parent.member"))
		})

		It("Case 3", func() {
			pr := NewParser("entity organization {\n\nrelation owner @user\n\naction delete = owner\n\n\n} `table:organization|identifier:id`\n\n")
			schema := pr.Parse()
			st := schema.Statements[0].(*ast.EntityStatement)

			Expect(st.Name.Literal).Should(Equal("organization"))
			Expect(st.Option.Literal).Should(Equal(`table:organization|identifier:id`))

			r1 := st.RelationStatements[0].(*ast.RelationStatement)
			Expect(r1.Name.Literal).Should(Equal("owner"))

			for _, a := range r1.RelationTypes {
				Expect(a.TokenLiteral()).Should(Equal("user"))
			}

			Expect(r1.Option.Literal).Should(Equal(""))

			a1 := st.ActionStatements[0].(*ast.ActionStatement)
			Expect(a1.Name.Literal).Should(Equal("delete"))

			es := a1.ExpressionStatement.(*ast.ExpressionStatement)

			Expect(es.Expression.(*ast.Identifier).Value).Should(Equal("owner"))
		})

		It("Case 4", func() {
			pr := NewParser("entity organization {\n\nrelation owner @user\n\naction delete = not owner\n\n\n}\n\n")
			schema := pr.Parse()
			st := schema.Statements[0].(*ast.EntityStatement)

			Expect(st.Name.Literal).Should(Equal("organization"))
			Expect(st.Option.Literal).Should(Equal(""))

			r1 := st.RelationStatements[0].(*ast.RelationStatement)
			Expect(r1.Name.Literal).Should(Equal("owner"))

			for _, a := range r1.RelationTypes {
				Expect(a.TokenLiteral()).Should(Equal("user"))
			}

			Expect(r1.Option.Literal).Should(Equal(""))

			a1 := st.ActionStatements[0].(*ast.ActionStatement)
			Expect(a1.Name.Literal).Should(Equal("delete"))

			es := a1.ExpressionStatement.(*ast.ExpressionStatement)

			Expect(es.Expression.(*ast.PrefixExpression).Value).Should(Equal("owner"))
		})
	})
})
