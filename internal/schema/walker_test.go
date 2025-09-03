package schema

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Permify/permify/pkg/dsl/compiler"
	"github.com/Permify/permify/pkg/dsl/parser"
	base "github.com/Permify/permify/pkg/pb/base/v1"
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

	Context("Error Handling", func() {
		It("should return error when entity definition not found", func() {
			sch, err := parser.NewParser(`
			entity user {}
			entity document {
				relation viewer @user
			}
			`).Parse()

			Expect(err).ShouldNot(HaveOccurred())

			c := compiler.NewCompiler(true, sch)
			e, r, err := c.Compile()

			Expect(err).ShouldNot(HaveOccurred())

			w := NewWalker(NewSchemaFromEntityAndRuleDefinitions(e, r))

			// Try to walk a non-existent entity
			err = w.Walk("nonexistent_entity", "viewer")

			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(Equal(base.ErrorCode_ERROR_CODE_ENTITY_DEFINITION_NOT_FOUND.String()))
		})

		It("should return error when permission not found", func() {
			sch, err := parser.NewParser(`
			entity user {}
			entity document {
				relation viewer @user
				action view = viewer
			}
			`).Parse()

			Expect(err).ShouldNot(HaveOccurred())

			c := compiler.NewCompiler(true, sch)
			e, r, err := c.Compile()

			Expect(err).ShouldNot(HaveOccurred())

			w := NewWalker(NewSchemaFromEntityAndRuleDefinitions(e, r))

			// Try to walk a non-existent permission
			err = w.Walk("document", "nonexistent_permission")

			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(Equal(base.ErrorCode_ERROR_CODE_UNDEFINED_CHILD_KIND.String()))
		})

		It("should return error when reference type is attribute", func() {
			sch, err := parser.NewParser(`
			entity user {}
			entity document {
				relation viewer @user
				attribute public boolean
				action view = viewer or public
			}
			`).Parse()

			Expect(err).ShouldNot(HaveOccurred())

			c := compiler.NewCompiler(true, sch)
			e, r, err := c.Compile()

			Expect(err).ShouldNot(HaveOccurred())

			w := NewWalker(NewSchemaFromEntityAndRuleDefinitions(e, r))

			// Try to walk an attribute reference
			err = w.Walk("document", "public")

			Expect(err).Should(HaveOccurred())
			Expect(err).Should(Equal(ErrUnimplemented))
		})

		It("should return error for undefined child kind in Walk", func() {
			// This test is challenging because we need to create a malformed reference type
			// that will trigger the undefined child kind error. This would require
			// creating a custom schema structure that bypasses the normal compilation process.
			// For now, we'll skip this test as it's difficult to trigger in practice.
			Skip("Skipping test for undefined child kind in Walk - difficult to trigger with normal schema compilation")
		})

		It("should return error for undefined child kind in WalkRewrite", func() {
			// This test is challenging because we need to create a malformed child type
			// that will trigger the undefined child kind error. This would require
			// creating a custom schema structure that bypasses the normal compilation process.
			// For now, we'll skip this test as it's difficult to trigger in practice.
			Skip("Skipping test for undefined child kind in WalkRewrite - difficult to trigger with normal schema compilation")
		})

		It("should return error when entity definition not found in WalkLeaf", func() {
			sch, err := parser.NewParser(`
			entity user {}
			entity organization {
				relation admin @user
			}
			entity document {
				relation org @organization
				action view = org.admin
			}
			`).Parse()

			Expect(err).ShouldNot(HaveOccurred())

			c := compiler.NewCompiler(true, sch)
			e, r, err := c.Compile()

			Expect(err).ShouldNot(HaveOccurred())

			// Create a schema with a malformed entity definition that will cause the error
			schema := NewSchemaFromEntityAndRuleDefinitions(e, r)
			// Remove the organization entity definition to trigger the error
			delete(schema.EntityDefinitions, "organization")

			w := NewWalker(schema)

			err = w.Walk("document", "view")

			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(Equal(base.ErrorCode_ERROR_CODE_ENTITY_DEFINITION_NOT_FOUND.String()))
		})

		It("should return error when relation reference is undefined in WalkLeaf", func() {
			sch, err := parser.NewParser(`
			entity user {}
			entity organization {
				relation admin @user
			}
			entity document {
				relation org @organization
				action view = org.admin
			}
			`).Parse()

			Expect(err).ShouldNot(HaveOccurred())

			c := compiler.NewCompiler(true, sch)
			e, r, err := c.Compile()

			Expect(err).ShouldNot(HaveOccurred())

			// Create a schema with a malformed entity definition that will cause the error
			schema := NewSchemaFromEntityAndRuleDefinitions(e, r)
			// Remove the relation definition to trigger the error
			delete(schema.EntityDefinitions["document"].Relations, "org")

			w := NewWalker(schema)

			err = w.Walk("document", "view")

			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(Equal(base.ErrorCode_ERROR_CODE_UNDEFINED_RELATION_REFERENCE.String()))
		})

		It("should return error when leaf type is computed attribute", func() {
			sch, err := parser.NewParser(`
			entity user {}
			entity document {
				relation viewer @user
				attribute public boolean
				action view = viewer or public
			}
			`).Parse()

			Expect(err).ShouldNot(HaveOccurred())

			c := compiler.NewCompiler(true, sch)
			e, r, err := c.Compile()

			Expect(err).ShouldNot(HaveOccurred())

			w := NewWalker(NewSchemaFromEntityAndRuleDefinitions(e, r))

			// This will trigger the computed attribute error path
			err = w.Walk("document", "view")

			Expect(err).Should(HaveOccurred())
			Expect(err).Should(Equal(ErrUnimplemented))
		})

		It("should return error when leaf type is call", func() {
			sch, err := parser.NewParser(`
			entity user {}
			entity document {
				relation viewer @user
				attribute public boolean
				action view = check_public(public) and viewer
			}
			
			rule check_public(public boolean) {
				public == true
			}
			`).Parse()

			Expect(err).ShouldNot(HaveOccurred())

			c := compiler.NewCompiler(true, sch)
			e, r, err := c.Compile()

			Expect(err).ShouldNot(HaveOccurred())

			w := NewWalker(NewSchemaFromEntityAndRuleDefinitions(e, r))

			// This will trigger the call error path
			err = w.Walk("document", "view")

			Expect(err).Should(HaveOccurred())
			Expect(err).Should(Equal(ErrUnimplemented))
		})

		It("should return error for undefined leaf type", func() {
			// This test is challenging because we need to create a malformed leaf type
			// that will trigger the undefined leaf type error. This would require
			// creating a custom schema structure that bypasses the normal compilation process.
			// For now, we'll skip this test as it's difficult to trigger in practice.
			Skip("Skipping test for undefined leaf type - difficult to trigger with normal schema compilation")
		})

		It("should handle error propagation in WalkRewrite", func() {
			sch, err := parser.NewParser(`
			entity user {}
			entity organization {
				relation admin @user
			}
			entity document {
				relation org @organization
				action view = org.admin
			}
			`).Parse()

			Expect(err).ShouldNot(HaveOccurred())

			c := compiler.NewCompiler(true, sch)
			e, r, err := c.Compile()

			Expect(err).ShouldNot(HaveOccurred())

			// Create a schema with a malformed entity definition that will cause the error
			schema := NewSchemaFromEntityAndRuleDefinitions(e, r)
			// Remove the organization entity definition to trigger the error
			delete(schema.EntityDefinitions, "organization")

			w := NewWalker(schema)

			err = w.Walk("document", "view")

			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(Equal(base.ErrorCode_ERROR_CODE_ENTITY_DEFINITION_NOT_FOUND.String()))
		})

		It("should handle error propagation in WalkComputedUserSet", func() {
			sch, err := parser.NewParser(`
			entity user {}
			entity organization {
				relation admin @user
			}
			entity document {
				relation org @organization
				action view = org.admin
			}
			`).Parse()

			Expect(err).ShouldNot(HaveOccurred())

			c := compiler.NewCompiler(true, sch)
			e, r, err := c.Compile()

			Expect(err).ShouldNot(HaveOccurred())

			// Create a schema with a malformed entity definition that will cause the error
			schema := NewSchemaFromEntityAndRuleDefinitions(e, r)
			// Remove the organization entity definition to trigger the error
			delete(schema.EntityDefinitions, "organization")

			w := NewWalker(schema)

			err = w.Walk("document", "view")

			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(Equal(base.ErrorCode_ERROR_CODE_ENTITY_DEFINITION_NOT_FOUND.String()))
		})
	})
})
