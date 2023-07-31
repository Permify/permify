package schema

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Permify/permify/pkg/dsl/compiler"
	"github.com/Permify/permify/pkg/dsl/parser"
)

var _ = Describe("walker", func() {
	Context("walker", func() {
		It("Case 1", func() {
			sch, err := parser.NewParser(`
			entity user {}
			entity organization {
				relation admin @user

				attribute is_public boolean

				permission edit = admin or is_public
			}
			entity container {
				relation parent @organization
				relation container_admin @user

				permission admin = parent.admin or container_admin

				permission edit = container_admin or parent.edit
			}
			`).Parse()

			Expect(err).ShouldNot(HaveOccurred())

			c := compiler.NewCompiler(true, sch)
			e, r, err := c.Compile()

			Expect(err).ShouldNot(HaveOccurred())

			w := NewWalker(NewSchemaFromEntityAndRuleDefinitions(e, r))

			err = w.Walk("container", "edit")

			Expect(err).Should(Equal(ErrUnimplemented))

			err = w.Walk("container", "admin")

			Expect(err).ShouldNot(HaveOccurred())
		})

		It("Case 2", func() {
			sch, err := parser.NewParser(`
		entity user {}
		
		entity organization {
		
			relation member @user
		
			attribute balance integer

			permission view = check_balance(balance) and member
		}
		
		entity repository {
		
			relation organization  @organization
			
			attribute is_public boolean

			permission view = is_public
			permission edit = organization.view
			permission delete = is_weekday(request.day_of_week)
			permission create = organization.member
		}
		
		rule check_balance(balance integer) {
			balance > 5000
		}

		rule is_weekday(day_of_week string) {
			  day_of_week != 'saturday' && day_of_week != 'sunday'
		}
			`).Parse()

			Expect(err).ShouldNot(HaveOccurred())

			c := compiler.NewCompiler(true, sch)
			e, r, err := c.Compile()

			Expect(err).ShouldNot(HaveOccurred())

			w := NewWalker(NewSchemaFromEntityAndRuleDefinitions(e, r))

			err = w.Walk("organization", "member")

			Expect(err).ShouldNot(HaveOccurred())

			err = w.Walk("repository", "view")

			Expect(err).Should(Equal(ErrUnimplemented))

			err = w.Walk("repository", "delete")

			Expect(err).Should(Equal(ErrUnimplemented))

			err = w.Walk("repository", "edit")

			Expect(err).Should(Equal(ErrUnimplemented))
		})
	})
})
