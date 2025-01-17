package schema

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Permify/permify/pkg/dsl/compiler"
	"github.com/Permify/permify/pkg/dsl/parser"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

var _ = Describe("linked schema", func() {
	Context("linked schema", func() {
		It("Case 1", func() {
			sch, err := parser.NewParser(`
			entity user {}
			entity document {
				relation viewer @user
			}
			`).Parse()

			Expect(err).ShouldNot(HaveOccurred())

			c := compiler.NewCompiler(true, sch)
			a, _, _ := c.Compile()

			g := NewLinkedGraph(NewSchemaFromEntityAndRuleDefinitions(a, nil))

			ent, err := g.LinkedEntrances(&base.Entrance{
				Type:  "document",
				Value: "viewer",
			}, &base.Entrance{
				Type:  "user",
				Value: "",
			})

			Expect(err).ShouldNot(HaveOccurred())
			Expect(ent).Should(Equal([]*LinkedEntrance{
				{
					Kind: RelationLinkedEntrance,
					TargetEntrance: &base.Entrance{
						Type:  "document",
						Value: "viewer",
					},
					TupleSetRelation: "",
				},
			}))
		})

		It("Case 2", func() {
			sch, err := parser.NewParser(`
			entity user {}
			entity document {
				relation viewer @user
				relation editor @user
				relation owner  @user
				action view = viewer or editor or owner
			}
			`).Parse()

			Expect(err).ShouldNot(HaveOccurred())

			c := compiler.NewCompiler(true, sch)
			a, _, _ := c.Compile()

			g := NewLinkedGraph(NewSchemaFromEntityAndRuleDefinitions(a, nil))

			ent, err := g.LinkedEntrances(&base.Entrance{
				Type:  "document",
				Value: "view",
			}, &base.Entrance{
				Type:  "user",
				Value: "",
			})

			Expect(err).ShouldNot(HaveOccurred())
			Expect(ent).Should(Equal([]*LinkedEntrance{
				{
					Kind: RelationLinkedEntrance,
					TargetEntrance: &base.Entrance{
						Type:  "document",
						Value: "viewer",
					},
					TupleSetRelation: "",
				},
				{
					Kind: RelationLinkedEntrance,
					TargetEntrance: &base.Entrance{
						Type:  "document",
						Value: "editor",
					},
					TupleSetRelation: "",
				},
				{
					Kind: RelationLinkedEntrance,
					TargetEntrance: &base.Entrance{
						Type:  "document",
						Value: "owner",
					},
					TupleSetRelation: "",
				},
			}))
		})

		It("Case 3", func() {
			sch, err := parser.NewParser(`
			entity user {}
			entity document {
				relation viewer @user
				relation editor @user
				relation owner @user
				action view = viewer and editor not owner
			}
			`).Parse()

			Expect(err).ShouldNot(HaveOccurred())

			c := compiler.NewCompiler(true, sch)
			a, _, _ := c.Compile()

			g := NewLinkedGraph(NewSchemaFromEntityAndRuleDefinitions(a, nil))

			ent, err := g.LinkedEntrances(&base.Entrance{
				Type:  "document",
				Value: "view",
			}, &base.Entrance{
				Type:  "user",
				Value: "",
			})

			Expect(err).ShouldNot(HaveOccurred())

			Expect(ent).Should(Equal([]*LinkedEntrance{
				{
					Kind: RelationLinkedEntrance,
					TargetEntrance: &base.Entrance{
						Type:  "document",
						Value: "viewer",
					},
					TupleSetRelation: "",
				},
				{
					Kind: RelationLinkedEntrance,
					TargetEntrance: &base.Entrance{
						Type:  "document",
						Value: "editor",
					},
					TupleSetRelation: "",
				},
				{
					Kind: RelationLinkedEntrance,
					TargetEntrance: &base.Entrance{
						Type:  "document",
						Value: "owner",
					},
					TupleSetRelation: "",
				},
			}))
		})

		It("Case 4", func() {
			sch, err := parser.NewParser(`
			entity user {}
			entity organization {
				relation admin @user
			}
			entity document {
				relation org @organization
				relation viewer @user
				relation owner @user
				action view = viewer or owner or org.admin
			}
			`).Parse()

			Expect(err).ShouldNot(HaveOccurred())

			c := compiler.NewCompiler(true, sch)
			a, _, _ := c.Compile()

			g := NewLinkedGraph(NewSchemaFromEntityAndRuleDefinitions(a, nil))

			ent, err := g.LinkedEntrances(&base.Entrance{
				Type:  "document",
				Value: "view",
			}, &base.Entrance{
				Type:  "user",
				Value: "",
			})

			Expect(err).ShouldNot(HaveOccurred())

			Expect(ent).Should(Equal([]*LinkedEntrance{
				{
					Kind: RelationLinkedEntrance,
					TargetEntrance: &base.Entrance{
						Type:  "document",
						Value: "viewer",
					},
					TupleSetRelation: "",
				},
				{
					Kind: RelationLinkedEntrance,
					TargetEntrance: &base.Entrance{
						Type:  "document",
						Value: "owner",
					},
					TupleSetRelation: "",
				},
				{
					Kind: RelationLinkedEntrance,
					TargetEntrance: &base.Entrance{
						Type:  "organization",
						Value: "admin",
					},
					TupleSetRelation: "",
				},
			}))
		})

		It("Case 5", func() {
			sch, err := parser.NewParser(`
			entity user {}
			entity organization {
				relation admin @user
			}
			entity container {
				relation parent @organization
				relation container_admin @user
				action admin = parent.admin or container_admin
			}
			entity document {
				relation container @container
				relation viewer @user
				relation owner @user
				action view = viewer or owner or container.admin
			}
			`).Parse()

			Expect(err).ShouldNot(HaveOccurred())

			c := compiler.NewCompiler(true, sch)
			a, _, _ := c.Compile()

			g := NewLinkedGraph(NewSchemaFromEntityAndRuleDefinitions(a, nil))

			ent, err := g.LinkedEntrances(&base.Entrance{
				Type:  "document",
				Value: "view",
			}, &base.Entrance{
				Type:  "user",
				Value: "",
			})

			Expect(err).ShouldNot(HaveOccurred())

			Expect(ent).Should(Equal([]*LinkedEntrance{
				{
					Kind: RelationLinkedEntrance,
					TargetEntrance: &base.Entrance{
						Type:  "document",
						Value: "viewer",
					},
					TupleSetRelation: "",
				},
				{
					Kind: RelationLinkedEntrance,
					TargetEntrance: &base.Entrance{
						Type:  "document",
						Value: "owner",
					},
					TupleSetRelation: "",
				},
				{
					Kind: RelationLinkedEntrance,
					TargetEntrance: &base.Entrance{
						Type:  "organization",
						Value: "admin",
					},
					TupleSetRelation: "",
				},
				{
					Kind: RelationLinkedEntrance,
					TargetEntrance: &base.Entrance{
						Type:  "container",
						Value: "container_admin",
					},
					TupleSetRelation: "",
				},
			}))
		})

		It("Case 6", func() {
			sch, err := parser.NewParser(`
			entity user {}
			entity group {
				relation direct_member @user
				relation manager @user
				action member = direct_member or manager
			}
			entity document {
				relation viewer @user @group#manager
				action view = viewer
			}
			`).Parse()

			Expect(err).ShouldNot(HaveOccurred())

			c := compiler.NewCompiler(true, sch)
			a, _, err := c.Compile()
			Expect(err).ShouldNot(HaveOccurred())

			g := NewLinkedGraph(NewSchemaFromEntityAndRuleDefinitions(a, nil))

			ent, err := g.LinkedEntrances(&base.Entrance{
				Type:  "document",
				Value: "view",
			}, &base.Entrance{
				Type:  "user",
				Value: "",
			})

			Expect(err).ShouldNot(HaveOccurred())

			Expect(ent).Should(Equal([]*LinkedEntrance{
				{
					Kind: RelationLinkedEntrance,
					TargetEntrance: &base.Entrance{
						Type:  "document",
						Value: "viewer",
					},
					TupleSetRelation: "",
				},
				{
					Kind: RelationLinkedEntrance,
					TargetEntrance: &base.Entrance{
						Type:  "group",
						Value: "manager",
					},
					TupleSetRelation: "",
				},
			}))
		})

		It("Case 7", func() {
			sch, err := parser.NewParser(`
			entity user {}
			entity organization {
				relation admin @user
				relation banned @user
			}
			entity document {
				relation org @organization
				relation viewer @user
				relation fist_rel @user
				relation second_rel @user
				relation third_rel @user
				action view = ((((viewer and org.banned) and org.admin) or fist_rel) and second_rel) or third_rel
			}
			`).Parse()

			Expect(err).ShouldNot(HaveOccurred())

			c := compiler.NewCompiler(true, sch)
			a, _, err := c.Compile()
			Expect(err).ShouldNot(HaveOccurred())

			g := NewLinkedGraph(NewSchemaFromEntityAndRuleDefinitions(a, nil))

			ent, err := g.LinkedEntrances(&base.Entrance{
				Type:  "document",
				Value: "view",
			}, &base.Entrance{
				Type:  "user",
				Value: "",
			})

			Expect(err).ShouldNot(HaveOccurred())

			Expect(ent).Should(Equal([]*LinkedEntrance{
				{
					Kind: RelationLinkedEntrance,
					TargetEntrance: &base.Entrance{
						Type:  "document",
						Value: "viewer",
					},
					TupleSetRelation: "",
				},
				{
					Kind: RelationLinkedEntrance,
					TargetEntrance: &base.Entrance{
						Type:  "organization",
						Value: "banned",
					},
					TupleSetRelation: "",
				},
				{
					Kind: RelationLinkedEntrance,
					TargetEntrance: &base.Entrance{
						Type:  "organization",
						Value: "admin",
					},
					TupleSetRelation: "",
				},
				{
					Kind: RelationLinkedEntrance,
					TargetEntrance: &base.Entrance{
						Type:  "document",
						Value: "fist_rel",
					},
					TupleSetRelation: "",
				},
				{
					Kind: RelationLinkedEntrance,
					TargetEntrance: &base.Entrance{
						Type:  "document",
						Value: "second_rel",
					},
					TupleSetRelation: "",
				},
				{
					Kind: RelationLinkedEntrance,
					TargetEntrance: &base.Entrance{
						Type:  "document",
						Value: "third_rel",
					},
					TupleSetRelation: "",
				},
			}))
		})

		It("Case 8", func() {
			sch, err := parser.NewParser(`
		entity user {}
			entity organization {
				relation admin @user
				relation banned @user
			}
			entity document {
				relation org @organization
				relation viewer @user
				relation fist_rel @user
				relation second_rel @user
				relation third_rel @user
				action view = ((((viewer and org.banned) and org.admin) or fist_rel) and second_rel) or third_rel
			}
			`).Parse()

			Expect(err).ShouldNot(HaveOccurred())

			c := compiler.NewCompiler(true, sch)
			a, _, err := c.Compile()
			Expect(err).ShouldNot(HaveOccurred())

			g := NewLinkedGraph(NewSchemaFromEntityAndRuleDefinitions(a, nil))

			ent, err := g.LinkedEntrances(&base.Entrance{
				Type:  "document",
				Value: "view",
			}, &base.Entrance{
				Type:  "document",
				Value: "viewer",
			})

			Expect(err).ShouldNot(HaveOccurred())

			Expect(ent).Should(Equal([]*LinkedEntrance{
				{
					Kind: ComputedUserSetLinkedEntrance,
					TargetEntrance: &base.Entrance{
						Type:  "document",
						Value: "view",
					},
					TupleSetRelation: "",
				},
			}))
		})

		It("Case 9", func() {
			sch, err := parser.NewParser(`
			entity user {}
			entity organization {
				relation admin @user
				relation banned @user
			}
			entity document {
				relation org @organization
				relation viewer @user
				relation fist_rel @user
				relation second_rel @user
				relation third_rel @user
				action view = ((((viewer and org.banned) and org.admin) or fist_rel) and second_rel) or third_rel
			}
			`).Parse()

			Expect(err).ShouldNot(HaveOccurred())

			c := compiler.NewCompiler(true, sch)
			a, _, err := c.Compile()
			Expect(err).ShouldNot(HaveOccurred())

			g := NewLinkedGraph(NewSchemaFromEntityAndRuleDefinitions(a, nil))

			ent, err := g.LinkedEntrances(&base.Entrance{
				Type:  "document",
				Value: "view",
			}, &base.Entrance{
				Type:  "organization",
				Value: "admin",
			})

			Expect(err).ShouldNot(HaveOccurred())
			Expect(ent).Should(Equal([]*LinkedEntrance{
				{
					Kind: TupleToUserSetLinkedEntrance,
					TargetEntrance: &base.Entrance{
						Type:  "document",
						Value: "view",
					},
					TupleSetRelation: "org",
				},
			}))
		})

		It("Case 10", func() {
			sch, err := parser.NewParser(`
			entity user {}
			entity organization {
				relation admin @user
			}
			entity container {
				relation parent @organization
				relation local_admin @user
				action admin = parent.admin or local_admin
				action test = local_admin
			}
			entity document {
				relation container @container
				relation viewer @user
				relation owner @user
				action view = viewer or owner or container.admin or container.test
			}
			`).Parse()

			Expect(err).ShouldNot(HaveOccurred())

			c := compiler.NewCompiler(true, sch)
			a, _, err := c.Compile()
			Expect(err).ShouldNot(HaveOccurred())

			g := NewLinkedGraph(NewSchemaFromEntityAndRuleDefinitions(a, nil))

			ent, err := g.LinkedEntrances(&base.Entrance{
				Type:  "document",
				Value: "view",
			}, &base.Entrance{
				Type:  "container",
				Value: "local_admin",
			})

			Expect(err).ShouldNot(HaveOccurred())

			Expect(ent).Should(Equal([]*LinkedEntrance{
				{
					Kind: ComputedUserSetLinkedEntrance,
					TargetEntrance: &base.Entrance{
						Type:  "container",
						Value: "admin",
					},
					TupleSetRelation: "",
				},
				{
					Kind: ComputedUserSetLinkedEntrance,
					TargetEntrance: &base.Entrance{
						Type:  "container",
						Value: "test",
					},
					TupleSetRelation: "",
				},
			}))
		})

		It("Case 11", func() {
			sch, err := parser.NewParser(`
			entity user {}
			entity organization {
				relation admin @user
			}
			entity container {
				relation parent @organization
				relation local_admin @user
				relation another @user
				action admin = parent.admin or local_admin
				action test = local_admin and another
			}
			entity document {
				relation container @container
				action view = container.admin or container.test
			}
			`).Parse()

			Expect(err).ShouldNot(HaveOccurred())

			c := compiler.NewCompiler(true, sch)
			a, _, err := c.Compile()
			Expect(err).ShouldNot(HaveOccurred())

			g := NewLinkedGraph(NewSchemaFromEntityAndRuleDefinitions(a, nil))

			ent, err := g.LinkedEntrances(&base.Entrance{
				Type:  "document",
				Value: "view",
			}, &base.Entrance{
				Type:  "container",
				Value: "local_admin",
			})

			Expect(err).ShouldNot(HaveOccurred())

			Expect(ent).Should(Equal([]*LinkedEntrance{
				{
					Kind: ComputedUserSetLinkedEntrance,
					TargetEntrance: &base.Entrance{
						Type:  "container",
						Value: "admin",
					},
					TupleSetRelation: "",
				},
				{
					Kind: ComputedUserSetLinkedEntrance,
					TargetEntrance: &base.Entrance{
						Type:  "container",
						Value: "test",
					},
					TupleSetRelation: "",
				},
			}))
		})

		It("Case 12", func() {
			sch, err := parser.NewParser(`
			entity user {}
			entity document {
				relation viewer @user
				relation another @user
				action view = viewer or viewer
			}
			`).Parse()

			Expect(err).ShouldNot(HaveOccurred())

			c := compiler.NewCompiler(true, sch)
			a, _, err := c.Compile()
			Expect(err).ShouldNot(HaveOccurred())

			g := NewLinkedGraph(NewSchemaFromEntityAndRuleDefinitions(a, nil))

			ent, err := g.LinkedEntrances(&base.Entrance{
				Type:  "document",
				Value: "view",
			}, &base.Entrance{
				Type:  "document",
				Value: "viewer",
			})

			Expect(err).ShouldNot(HaveOccurred())

			Expect(ent).Should(Equal([]*LinkedEntrance{
				{
					Kind: ComputedUserSetLinkedEntrance,
					TargetEntrance: &base.Entrance{
						Type:  "document",
						Value: "view",
					},
					TupleSetRelation: "",
				},
				{
					Kind: ComputedUserSetLinkedEntrance,
					TargetEntrance: &base.Entrance{
						Type:  "document",
						Value: "view",
					},
					TupleSetRelation: "",
				},
			}))
		})

		It("Case 13", func() {
			sch, err := parser.NewParser(`
			entity user {}
			entity organization {
				relation viewer @user
			}
			entity document {
				relation parent @organization
				action view = parent.viewer
			}
			`).Parse()

			Expect(err).ShouldNot(HaveOccurred())

			c := compiler.NewCompiler(true, sch)
			a, _, err := c.Compile()
			Expect(err).ShouldNot(HaveOccurred())

			g := NewLinkedGraph(NewSchemaFromEntityAndRuleDefinitions(a, nil))

			ent, err := g.LinkedEntrances(&base.Entrance{
				Type:  "document",
				Value: "view",
			}, &base.Entrance{
				Type:  "user",
				Value: "",
			})

			Expect(err).ShouldNot(HaveOccurred())

			Expect(ent).Should(Equal([]*LinkedEntrance{
				{
					Kind: RelationLinkedEntrance,
					TargetEntrance: &base.Entrance{
						Type:  "organization",
						Value: "viewer",
					},
					TupleSetRelation: "",
				},
			}))
		})

		It("Case 14", func() {
			sch, err := parser.NewParser(`
			entity user {}
			entity document {
				relation viewer @user
				relation editor @user
				action view = viewer or editor
			}
			`).Parse()

			Expect(err).ShouldNot(HaveOccurred())

			c := compiler.NewCompiler(true, sch)
			a, _, err := c.Compile()
			Expect(err).ShouldNot(HaveOccurred())

			g := NewLinkedGraph(NewSchemaFromEntityAndRuleDefinitions(a, nil))

			ent, err := g.LinkedEntrances(&base.Entrance{
				Type:  "document",
				Value: "view",
			}, &base.Entrance{
				Type:  "user",
				Value: "",
			})

			Expect(err).ShouldNot(HaveOccurred())

			Expect(ent).Should(Equal([]*LinkedEntrance{
				{
					Kind: RelationLinkedEntrance,
					TargetEntrance: &base.Entrance{
						Type:  "document",
						Value: "viewer",
					},
					TupleSetRelation: "",
				},
				{
					Kind: RelationLinkedEntrance,
					TargetEntrance: &base.Entrance{
						Type:  "document",
						Value: "editor",
					},
					TupleSetRelation: "",
				},
			}))
		})

		It("Case 15", func() {
			sch, err := parser.NewParser(`
			entity user {}
		
			entity account {
				relation super_admin @user
				relation admin @user @account#admin
				relation member @user @account#member
		
				action add_member = admin
				action delete_member = super_admin
			}
			`).Parse()

			Expect(err).ShouldNot(HaveOccurred())

			c := compiler.NewCompiler(true, sch)
			a, _, err := c.Compile()
			Expect(err).ShouldNot(HaveOccurred())

			g := NewLinkedGraph(NewSchemaFromEntityAndRuleDefinitions(a, nil))

			ent, err := g.LinkedEntrances(&base.Entrance{
				Type:  "account",
				Value: "add_member",
			}, &base.Entrance{
				Type:  "user",
				Value: "",
			})

			Expect(err).ShouldNot(HaveOccurred())
			Expect(ent).Should(Equal([]*LinkedEntrance{
				{
					Kind: RelationLinkedEntrance,
					TargetEntrance: &base.Entrance{
						Type:  "account",
						Value: "admin",
					},
					TupleSetRelation: "",
				},
			}))
		})

		It("Case 16", func() {
			sch, err := parser.NewParser(`
			entity user {}
		
			entity account {
				relation super_admin @user
				relation admin @user @account#admin
				relation member @user @account#member
		
				action add_member = admin
				action delete_member = super_admin
			}
			`).Parse()

			Expect(err).ShouldNot(HaveOccurred())

			c := compiler.NewCompiler(true, sch)
			a, _, err := c.Compile()
			Expect(err).ShouldNot(HaveOccurred())

			g := NewLinkedGraph(NewSchemaFromEntityAndRuleDefinitions(a, nil))

			ent, err := g.LinkedEntrances(&base.Entrance{
				Type:  "account",
				Value: "add_member",
			}, &base.Entrance{
				Type:  "account",
				Value: "admin",
			})

			Expect(err).ShouldNot(HaveOccurred())
			Expect(ent).Should(Equal([]*LinkedEntrance{
				{
					Kind: ComputedUserSetLinkedEntrance,
					TargetEntrance: &base.Entrance{
						Type:  "account",
						Value: "add_member",
					},
					TupleSetRelation: "",
				},
				{
					Kind: RelationLinkedEntrance,
					TargetEntrance: &base.Entrance{
						Type:  "account",
						Value: "admin",
					},
					TupleSetRelation: "",
				},
			}))
		})

		It("Case 17", func() {
			sch, err := parser.NewParser(`
			entity user {}
		
			entity account {
				relation super_admin @user
				relation admin @user @account#admin
				relation member @user @account#member
		
				action add_member = admin or member
				action delete_member = super_admin
			}
			`).Parse()

			Expect(err).ShouldNot(HaveOccurred())

			c := compiler.NewCompiler(true, sch)
			a, _, err := c.Compile()
			Expect(err).ShouldNot(HaveOccurred())

			g := NewLinkedGraph(NewSchemaFromEntityAndRuleDefinitions(a, nil))

			ent, err := g.LinkedEntrances(&base.Entrance{
				Type:  "account",
				Value: "admin",
			}, &base.Entrance{
				Type:  "account",
				Value: "admin",
			})

			Expect(err).ShouldNot(HaveOccurred())
			Expect(ent).Should(Equal([]*LinkedEntrance{
				{
					Kind: RelationLinkedEntrance,
					TargetEntrance: &base.Entrance{
						Type:  "account",
						Value: "admin",
					},
					TupleSetRelation: "",
				},
			}))
		})

		It("Case 18", func() {
			Expect(LinkedEntrance{
				Kind: RelationLinkedEntrance,
				TargetEntrance: &base.Entrance{
					Type:  "account",
					Value: "admin",
				},
				TupleSetRelation: "",
			}.LinkedEntranceKind()).Should(Equal(RelationLinkedEntrance))
		})

		It("Case 19", func() {
			sch, err := parser.NewParser(`
			entity user {}
			entity document {
				relation viewer @user
				attribute public boolean

				permission view = viewer or public
			}
			`).Parse()

			Expect(err).ShouldNot(HaveOccurred())

			c := compiler.NewCompiler(true, sch)
			a, _, _ := c.Compile()

			g := NewLinkedGraph(NewSchemaFromEntityAndRuleDefinitions(a, nil))

			ent, err := g.LinkedEntrances(&base.Entrance{
				Type:  "document",
				Value: "view",
			}, &base.Entrance{
				Type:  "user",
				Value: "",
			})

			Expect(err).ShouldNot(HaveOccurred())
			Expect(ent).Should(Equal([]*LinkedEntrance{
				{
					Kind: RelationLinkedEntrance,
					TargetEntrance: &base.Entrance{
						Type:  "document",
						Value: "viewer",
					},
					TupleSetRelation: "",
				},
				{
					Kind: AttributeLinkedEntrance,
					TargetEntrance: &base.Entrance{
						Type:  "document",
						Value: "public",
					},
					TupleSetRelation: "",
				},
			}))
		})

		It("Case 20", func() {
			sch, err := parser.NewParser(`
			entity user {}
			
			entity account {
				relation owner @user
				
				attribute balance integer
			
				permission withdraw = check_balance(balance) and owner
			}
			
			rule check_balance(balance integer) {
				(balance >= 5000)
			}
			`).Parse()

			Expect(err).ShouldNot(HaveOccurred())

			c := compiler.NewCompiler(true, sch)
			a, _, _ := c.Compile()

			g := NewLinkedGraph(NewSchemaFromEntityAndRuleDefinitions(a, nil))

			ent, err := g.LinkedEntrances(&base.Entrance{
				Type:  "account",
				Value: "withdraw",
			}, &base.Entrance{
				Type:  "user",
				Value: "",
			})

			Expect(err).ShouldNot(HaveOccurred())
			Expect(ent).Should(Equal([]*LinkedEntrance{
				{
					Kind: AttributeLinkedEntrance,
					TargetEntrance: &base.Entrance{
						Type:  "account",
						Value: "balance",
					},
					TupleSetRelation: "",
				},
				{
					Kind: RelationLinkedEntrance,
					TargetEntrance: &base.Entrance{
						Type:  "account",
						Value: "owner",
					},
					TupleSetRelation: "",
				},
			}))
		})

		It("Case 21", func() {
			sch, err := parser.NewParser(`
			entity user {}

			entity organization {
				relation owner @user

				attribute balance integer

				permission withdraw = check_balance(balance) or owner
			}

			entity account {

				relation parent @organization

				relation owner @user
				
				attribute balance integer
			
				permission withdraw = check_balance(balance) and owner or parent.withdraw
			}

			rule check_balance(balance integer) {
				(balance >= 5000)
			}
			`).Parse()

			Expect(err).ShouldNot(HaveOccurred())

			c := compiler.NewCompiler(true, sch)
			a, _, _ := c.Compile()

			g := NewLinkedGraph(NewSchemaFromEntityAndRuleDefinitions(a, nil))

			ent, err := g.LinkedEntrances(&base.Entrance{
				Type:  "account",
				Value: "withdraw",
			}, &base.Entrance{
				Type:  "user",
				Value: "",
			})

			Expect(err).ShouldNot(HaveOccurred())
			Expect(ent).Should(Equal([]*LinkedEntrance{
				{
					Kind: AttributeLinkedEntrance,
					TargetEntrance: &base.Entrance{
						Type:  "account",
						Value: "balance",
					},
					TupleSetRelation: "",
				},
				{
					Kind: RelationLinkedEntrance,
					TargetEntrance: &base.Entrance{
						Type:  "account",
						Value: "owner",
					},
					TupleSetRelation: "",
				},
				{
					Kind: AttributeLinkedEntrance,
					TargetEntrance: &base.Entrance{
						Type:  "organization",
						Value: "balance",
					},
					TupleSetRelation: "",
				},
				{
					Kind: RelationLinkedEntrance,
					TargetEntrance: &base.Entrance{
						Type:  "organization",
						Value: "owner",
					},
					TupleSetRelation: "",
				},
			}))
		})

		It("Case 22", func() {
			sch, err := parser.NewParser(`
			entity user {}

			entity organization {
				relation owner @user

				attribute balance integer

				permission withdraw = check_balance(balance) or owner
			}

			entity account {

				relation parent @organization

				relation owner @user @account#owner
				
				attribute balance integer
			
				permission withdraw = check_balance(balance) and owner or parent.withdraw
			}

			rule check_balance(balance integer) {
				(balance >= 5000)
			}
			`).Parse()

			Expect(err).ShouldNot(HaveOccurred())

			c := compiler.NewCompiler(true, sch)
			a, _, _ := c.Compile()

			g := NewLinkedGraph(NewSchemaFromEntityAndRuleDefinitions(a, nil))

			ent, err := g.LinkedEntrances(&base.Entrance{
				Type:  "account",
				Value: "withdraw",
			}, &base.Entrance{
				Type:  "account",
				Value: "owner",
			})

			Expect(err).ShouldNot(HaveOccurred())
			Expect(ent).Should(Equal([]*LinkedEntrance{
				{
					Kind: AttributeLinkedEntrance,
					TargetEntrance: &base.Entrance{
						Type:  "account",
						Value: "balance",
					},
					TupleSetRelation: "",
				},
				{
					Kind: ComputedUserSetLinkedEntrance,
					TargetEntrance: &base.Entrance{
						Type:  "account",
						Value: "withdraw",
					},
					TupleSetRelation: "",
				},
				{
					Kind: RelationLinkedEntrance,
					TargetEntrance: &base.Entrance{
						Type:  "account",
						Value: "owner",
					},
					TupleSetRelation: "",
				},
				{
					Kind: AttributeLinkedEntrance,
					TargetEntrance: &base.Entrance{
						Type:  "organization",
						Value: "balance",
					},
					TupleSetRelation: "",
				},
			}))
		})

		It("Case 23", func() {
			sch, err := parser.NewParser(`
				entity user {}
				
				entity aaa {
					relation role__admin @user
					permission ccc__read = role__admin
				}
				
				entity bbb {
					relation resource__aaa @aaa
					relation role__admin @user
					attribute attr__is_public boolean
					permission ccc__read = role__admin or attr__is_public
				
				}
				
				entity ccc {
					relation resource__aaa @aaa
					relation resource__bbb @bbb
					permission ccc__read = resource__aaa.ccc__read or resource__bbb.ccc__read
				}
			`).Parse()

			Expect(err).ShouldNot(HaveOccurred())

			c := compiler.NewCompiler(true, sch)
			a, _, _ := c.Compile()

			g := NewLinkedGraph(NewSchemaFromEntityAndRuleDefinitions(a, nil))

			ent, err := g.LinkedEntrances(&base.Entrance{
				Type:  "ccc",
				Value: "ccc__read",
			}, &base.Entrance{
				Type:  "user",
				Value: "",
			})

			Expect(err).ShouldNot(HaveOccurred())
			Expect(ent).Should(Equal([]*LinkedEntrance{
				{
					Kind: RelationLinkedEntrance,
					TargetEntrance: &base.Entrance{
						Type:  "aaa",
						Value: "role__admin",
					},
					TupleSetRelation: "",
				},
				{
					Kind: RelationLinkedEntrance,
					TargetEntrance: &base.Entrance{
						Type:  "bbb",
						Value: "role__admin",
					},
					TupleSetRelation: "",
				},
				{
					Kind: AttributeLinkedEntrance,
					TargetEntrance: &base.Entrance{
						Type:  "bbb",
						Value: "attr__is_public",
					},
					TupleSetRelation: "",
				},
			}))
		})
	})
})
