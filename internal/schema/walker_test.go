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
			permission delete = is_workday(is_public)
			permission create = organization.member
		}
		
		rule check_balance(balance integer) {
			balance > 5000
		}

		rule is_workday(is_public boolean) {
			  is_public == true && (context.data.day_of_week != 'saturday' && context.data.day_of_week != 'sunday')
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

		It("Case 3", func() {
			sch, err := parser.NewParser(`
			entity user {}

			entity tag {
				relation assignee @department
				permission view_document = assignee.view_document
			}

			entity document {
				relation owner @department
			
				permission edit = owner.edit_document
				permission view = owner.view_document or owner.peek_document
			}
			
			entity department {
				relation parent @department
				relation admin @user
				relation viewer @user
				relation assigned_tag @tag
				
				permission peek_document = assigned_tag.view_document or parent.peek_document
				permission edit_document = admin or parent.edit_document
				permission view_document = viewer or admin or parent.view_document
			}
			`).Parse()

			Expect(err).ShouldNot(HaveOccurred())

			c := compiler.NewCompiler(true, sch)
			e, r, err := c.Compile()

			Expect(err).ShouldNot(HaveOccurred())

			w := NewWalker(NewSchemaFromEntityAndRuleDefinitions(e, r))

			err = w.Walk("document", "view")

			Expect(err).ShouldNot(HaveOccurred())
		})

		It("Case 4", func() {
			sch, err := parser.NewParser(`
			  entity user {}
			
			  entity organization {
			
				  attribute active boolean
			
				  relation active_user @user
				  relation member @user @organization#member
			
				  permission active_member = member and active_user and active
			
				  action view_one = member and active_user and active
				  action view_two = active_member
			
			  }
			`).Parse()

			Expect(err).ShouldNot(HaveOccurred())

			c := compiler.NewCompiler(true, sch)
			e, r, err := c.Compile()

			Expect(err).ShouldNot(HaveOccurred())

			w := NewWalker(NewSchemaFromEntityAndRuleDefinitions(e, r))

			err = w.Walk("organization", "view_two")

			Expect(err).Should(Equal(ErrUnimplemented))
		})
	})
})
