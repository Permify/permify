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
					Kind: PathChainLinkedEntrance,
					TargetEntrance: &base.Entrance{
						Type:  "organization",
						Value: "balance",
					},
					TupleSetRelation: "",
					PathChain: []*base.RelationReference{
						{
							Type:     "account",
							Relation: "parent",
						},
					},
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
					Kind: PathChainLinkedEntrance,
					TargetEntrance: &base.Entrance{
						Type:  "organization",
						Value: "balance",
					},
					TupleSetRelation: "",
					PathChain: []*base.RelationReference{
						{
							Type:     "account",
							Relation: "parent",
						},
					},
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
					Kind: PathChainLinkedEntrance,
					TargetEntrance: &base.Entrance{
						Type:  "bbb",
						Value: "attr__is_public",
					},
					TupleSetRelation: "",
					PathChain: []*base.RelationReference{
						{
							Type:     "ccc",
							Relation: "resource__bbb",
						},
					},
				},
				{
					Kind: RelationLinkedEntrance,
					TargetEntrance: &base.Entrance{
						Type:  "bbb",
						Value: "role__admin",
					},
					TupleSetRelation: "",
				},
			}))
		})
	})

	Context("Error Handling", func() {
		It("should return error when permission not found", func() {
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

			// Try to access a non-existent permission
			ent, err := g.LinkedEntrances(&base.Entrance{
				Type:  "document",
				Value: "nonexistent_permission",
			}, &base.Entrance{
				Type:  "user",
				Value: "",
			})

			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(Equal("unimplemented"))
			Expect(ent).Should(BeNil())
		})

		It("should return error when attribute not found", func() {
			sch, err := parser.NewParser(`
			entity user {}
			entity document {
				relation viewer @user
				attribute public boolean
			}
			`).Parse()

			Expect(err).ShouldNot(HaveOccurred())

			c := compiler.NewCompiler(true, sch)
			a, _, _ := c.Compile()

			g := NewLinkedGraph(NewSchemaFromEntityAndRuleDefinitions(a, nil))

			// Try to access a non-existent attribute
			ent, err := g.LinkedEntrances(&base.Entrance{
				Type:  "document",
				Value: "nonexistent_attribute",
			}, &base.Entrance{
				Type:  "user",
				Value: "",
			})

			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(Equal("unimplemented"))
			Expect(ent).Should(BeNil())
		})

		It("should return error when entity definition not found", func() {
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

			// Try to access a non-existent entity
			ent, err := g.LinkedEntrances(&base.Entrance{
				Type:  "nonexistent_entity",
				Value: "viewer",
			}, &base.Entrance{
				Type:  "user",
				Value: "",
			})

			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(Equal("entity definition not found"))
			Expect(ent).Should(BeNil())
		})

		It("should return error when relation definition not found", func() {
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

			// Try to access a non-existent relation
			ent, err := g.LinkedEntrances(&base.Entrance{
				Type:  "document",
				Value: "nonexistent_relation",
			}, &base.Entrance{
				Type:  "user",
				Value: "",
			})

			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(Equal("unimplemented"))
			Expect(ent).Should(BeNil())
		})

		It("should return error when entity definition not found in findEntranceLeaf", func() {
			sch, err := parser.NewParser(`
			entity user {}
			entity document {
				relation viewer @user
				action view = viewer
			}
			`).Parse()

			Expect(err).ShouldNot(HaveOccurred())

			c := compiler.NewCompiler(true, sch)
			a, _, _ := c.Compile()

			// Create a schema with a malformed entity definition that will cause the error
			schema := NewSchemaFromEntityAndRuleDefinitions(a, nil)
			// Remove the entity definition to trigger the error
			delete(schema.EntityDefinitions, "document")

			g := NewLinkedGraph(schema)

			ent, err := g.LinkedEntrances(&base.Entrance{
				Type:  "document",
				Value: "view",
			}, &base.Entrance{
				Type:  "user",
				Value: "",
			})

			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(Equal("entity definition not found"))
			Expect(ent).Should(BeNil())
		})

		It("should return error when relation definition not found in findEntranceLeaf", func() {
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
			a, _, _ := c.Compile()

			// Create a schema with a malformed entity definition that will cause the error
			schema := NewSchemaFromEntityAndRuleDefinitions(a, nil)
			// Remove the relation definition to trigger the error
			delete(schema.EntityDefinitions["document"].Relations, "org")

			g := NewLinkedGraph(schema)

			ent, err := g.LinkedEntrances(&base.Entrance{
				Type:  "document",
				Value: "view",
			}, &base.Entrance{
				Type:  "user",
				Value: "",
			})

			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(Equal("relation definition not found"))
			Expect(ent).Should(BeNil())
		})

		It("should return error for undefined child type in findEntranceRewrite", func() {
			// This test is challenging because we need to create a malformed rewrite structure
			// that will trigger the "undefined child type" error. This would require
			// creating a custom schema structure that bypasses the normal compilation process.
			// For now, we'll skip this test as it's difficult to trigger in practice.
			Skip("Skipping test for undefined child type - difficult to trigger with normal schema compilation")
		})

		It("should return error for unimplemented reference type", func() {
			// This test is challenging because we need to create a malformed reference type
			// that will trigger the ErrUnimplemented error. This would require
			// creating a custom schema structure that bypasses the normal compilation process.
			// For now, we'll skip this test as it's difficult to trigger in practice.
			Skip("Skipping test for unimplemented reference type - difficult to trigger with normal schema compilation")
		})

		It("should return error for undefined leaf type", func() {
			// This test is challenging because we need to create a malformed leaf type
			// that will trigger the ErrUndefinedLeafType error. This would require
			// creating a custom schema structure that bypasses the normal compilation process.
			// For now, we'll skip this test as it's difficult to trigger in practice.
			Skip("Skipping test for undefined leaf type - difficult to trigger with normal schema compilation")
		})

		It("should handle error propagation in recursive calls", func() {
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
			a, _, _ := c.Compile()

			// Create a schema with a malformed entity definition that will cause the error
			schema := NewSchemaFromEntityAndRuleDefinitions(a, nil)
			// Remove the organization entity definition to trigger the error
			delete(schema.EntityDefinitions, "organization")

			g := NewLinkedGraph(schema)

			ent, err := g.LinkedEntrances(&base.Entrance{
				Type:  "document",
				Value: "view",
			}, &base.Entrance{
				Type:  "user",
				Value: "",
			})

			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(Equal("entity definition not found"))
			Expect(ent).Should(BeNil())
		})

		It("should handle error propagation in findRelationEntrance", func() {
			sch, err := parser.NewParser(`
			entity user {}
			entity organization {
				relation admin @user
			}
			entity document {
				relation org @organization
				relation viewer @user
				action view = org.admin or viewer
			}
			`).Parse()

			Expect(err).ShouldNot(HaveOccurred())

			c := compiler.NewCompiler(true, sch)
			a, _, _ := c.Compile()

			// Create a schema with a malformed entity definition that will cause the error
			schema := NewSchemaFromEntityAndRuleDefinitions(a, nil)
			// Remove the organization entity definition to trigger the error
			delete(schema.EntityDefinitions, "organization")

			g := NewLinkedGraph(schema)

			ent, err := g.LinkedEntrances(&base.Entrance{
				Type:  "document",
				Value: "view",
			}, &base.Entrance{
				Type:  "user",
				Value: "",
			})

			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(Equal("entity definition not found"))
			Expect(ent).Should(BeNil())
		})

		It("should handle error propagation in findEntranceLeaf with tuple to user set", func() {
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
			a, _, _ := c.Compile()

			// Create a schema with a malformed entity definition that will cause the error
			schema := NewSchemaFromEntityAndRuleDefinitions(a, nil)
			// Remove the organization entity definition to trigger the error
			delete(schema.EntityDefinitions, "organization")

			g := NewLinkedGraph(schema)

			ent, err := g.LinkedEntrances(&base.Entrance{
				Type:  "document",
				Value: "view",
			}, &base.Entrance{
				Type:  "user",
				Value: "",
			})

			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(Equal("entity definition not found"))
			Expect(ent).Should(BeNil())
		})

		It("should handle error propagation in findEntranceLeaf with computed user set", func() {
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
			a, _, _ := c.Compile()

			// Create a schema with a malformed entity definition that will cause the error
			schema := NewSchemaFromEntityAndRuleDefinitions(a, nil)
			// Remove the organization entity definition to trigger the error
			delete(schema.EntityDefinitions, "organization")

			g := NewLinkedGraph(schema)

			ent, err := g.LinkedEntrances(&base.Entrance{
				Type:  "document",
				Value: "view",
			}, &base.Entrance{
				Type:  "user",
				Value: "",
			})

			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(Equal("entity definition not found"))
			Expect(ent).Should(BeNil())
		})

		It("Case 30: Complex Multi-Level Nested Relations with Multiple Entities", func() {
			// Test complex scenario with multiple entity types and deep nesting
			// User -> Group -> Project -> Task -> Document with different permission levels
			sch, err := parser.NewParser(`
			entity user {}

			entity group {
				relation member @user
				relation admin @user
				
				permission manage = admin or member
			}

			entity project {
				relation group @group
				relation owner @user
				attribute visibility string
				
				permission view = owner or group.manage
				permission edit = owner or group.admin
				permission check_visibility = check_project_visibility(visibility)
			}

			entity task {
				relation project @project
				relation assignee @user
				attribute priority string
				
				permission view = assignee or project.view
				permission complete = assignee or project.edit
				permission check_priority = check_task_priority(priority)
			}

			entity document {
				relation task @task
				relation creator @user
				attribute doc_type string
				
				permission read = creator or task.view
				permission modify = creator or task.complete
				permission check_doc_type = check_document_type(doc_type)
				permission check_task_priority = task.check_priority
			}

			rule check_project_visibility(visibility string) {
				true
			}

			rule check_task_priority(priority string) {
				true
			}

			rule check_document_type(doc_type string) {
				true
			}
			`).Parse()

			Expect(err).ShouldNot(HaveOccurred())

			c := compiler.NewCompiler(true, sch)
			a, _, _ := c.Compile()

			g := NewLinkedGraph(NewSchemaFromEntityAndRuleDefinitions(a, nil))

			// Test nested attribute permission: check_task_priority (document -> task -> priority)
			ent, err := g.LinkedEntrances(&base.Entrance{
				Type:  "document",
				Value: "check_task_priority",
			}, &base.Entrance{
				Type:  "user",
				Value: "",
			})

			Expect(err).ShouldNot(HaveOccurred())
			Expect(ent).Should(Equal([]*LinkedEntrance{
				{
					Kind: PathChainLinkedEntrance,
					TargetEntrance: &base.Entrance{
						Type:  "task",
						Value: "priority",
					},
					TupleSetRelation: "",
					PathChain: []*base.RelationReference{
						{
							Type:     "document",
							Relation: "task",
						},
					},
				},
			}))
		})

		It("Case 31: Simple 3-Level Nested Attribute Hierarchy", func() {
			// Test simple 3-level nested attribute: Company -> Department -> Employee -> Attribute
			sch, err := parser.NewParser(`
			entity user {}

			entity company {
				relation department @department
				attribute company_id string
				
				permission ACCESS_COMPANY = check_company(company_id)
			}

			entity department {
				relation company @company
				relation employee @employee
				attribute dept_name string
				
				permission ACCESS_DEPT = company.ACCESS_COMPANY and check_department(dept_name)
			}

			entity employee {
				relation department @department
				attribute employee_id string
				
				permission ACCESS_EMPLOYEE = department.ACCESS_DEPT and check_employee(employee_id)
			}

			rule check_company(company_id string) {
				true
			}

			rule check_department(dept_name string) {
				true
			}

			rule check_employee(employee_id string) {
				true
			}
			`).Parse()

			Expect(err).ShouldNot(HaveOccurred())

			c := compiler.NewCompiler(true, sch)
			a, _, _ := c.Compile()

			g := NewLinkedGraph(NewSchemaFromEntityAndRuleDefinitions(a, nil))

			// Test nested attribute permission: ACCESS_EMPLOYEE (employee -> department -> company -> company attributes)
			ent, err := g.LinkedEntrances(&base.Entrance{
				Type:  "employee",
				Value: "ACCESS_EMPLOYEE",
			}, &base.Entrance{
				Type:  "user",
				Value: "",
			})

			Expect(err).ShouldNot(HaveOccurred())

			Expect(ent).Should(Equal([]*LinkedEntrance{
				{
					Kind: PathChainLinkedEntrance,
					TargetEntrance: &base.Entrance{
						Type:  "company",
						Value: "company_id",
					},
					TupleSetRelation: "",
					PathChain: []*base.RelationReference{
						{
							Type:     "employee",
							Relation: "department",
						},
						{
							Type:     "department",
							Relation: "company",
						},
					},
				},
				{
					Kind: PathChainLinkedEntrance,
					TargetEntrance: &base.Entrance{
						Type:  "department",
						Value: "dept_name",
					},
					TupleSetRelation: "",
					PathChain: []*base.RelationReference{
						{
							Type:     "employee",
							Relation: "department",
						},
					},
				},
				{
					Kind: AttributeLinkedEntrance,
					TargetEntrance: &base.Entrance{
						Type:  "employee",
						Value: "employee_id",
					},
					TupleSetRelation: "",
					PathChain:        nil,
				},
			}))
		})

		It("Case 32: Nested PathChain preserves tuple-set relation", func() {
			sch, err := parser.NewParser(`
			entity user {}

			entity org {
				attribute is_public boolean
				permission view = is_public
			}

			entity folder {
				relation parent @org
				permission view = parent.view
			}

			entity resource {
				relation parent @folder
				relation alt @folder
				permission view = parent.view
			}
			`).Parse()

			Expect(err).ShouldNot(HaveOccurred())

			c := compiler.NewCompiler(true, sch)
			a, _, _ := c.Compile()

			g := NewLinkedGraph(NewSchemaFromEntityAndRuleDefinitions(a, nil))

			ent, err := g.LinkedEntrances(&base.Entrance{
				Type:  "resource",
				Value: "view",
			}, &base.Entrance{
				Type:  "user",
				Value: "",
			})

			Expect(err).ShouldNot(HaveOccurred())
			Expect(ent).Should(Equal([]*LinkedEntrance{
				{
					Kind: PathChainLinkedEntrance,
					TargetEntrance: &base.Entrance{
						Type:  "org",
						Value: "is_public",
					},
					TupleSetRelation: "",
					PathChain: []*base.RelationReference{
						{
							Type:     "resource",
							Relation: "parent",
						},
						{
							Type:     "folder",
							Relation: "parent",
						},
					},
				},
			}))
		})

		It("Case 33: SelfCycleRelationsForPermission returns only same-type relations", func() {
			sch, err := parser.NewParser(`
			entity user {}

			entity resource {
				relation parent @resource
				relation owner @user
				permission view = parent.view or owner
			}
			`).Parse()

			Expect(err).ShouldNot(HaveOccurred())

			c := compiler.NewCompiler(true, sch)
			a, _, _ := c.Compile()

			g := NewLinkedGraph(NewSchemaFromEntityAndRuleDefinitions(a, nil))

			Expect(g.SelfCycleRelationsForPermission("resource", "view")).To(ConsistOf("parent"))
		})

		It("Case 34: SelfCycleRelationsForPermission ignores cross-type relations", func() {
			sch, err := parser.NewParser(`
			entity user {}

			entity org {
				attribute is_public boolean
				permission view = is_public
			}

			entity resource {
				relation parent @org
				permission view = parent.view
			}
			`).Parse()

			Expect(err).ShouldNot(HaveOccurred())

			c := compiler.NewCompiler(true, sch)
			a, _, _ := c.Compile()

			g := NewLinkedGraph(NewSchemaFromEntityAndRuleDefinitions(a, nil))

			Expect(g.SelfCycleRelationsForPermission("resource", "view")).To(BeEmpty())
		})

		It("Case 35: GetSubjectRelationForPathWalk returns nested subject relation", func() {
			sch, err := parser.NewParser(`
			entity user {}

			entity group {
				relation member @user
			}

			entity document {
				relation group @group#member
			}
			`).Parse()

			Expect(err).ShouldNot(HaveOccurred())

			c := compiler.NewCompiler(true, sch)
			a, _, _ := c.Compile()

			g := NewLinkedGraph(NewSchemaFromEntityAndRuleDefinitions(a, nil))

			Expect(g.GetSubjectRelationForPathWalk("document", "group", "group")).To(Equal("member"))
			Expect(g.GetSubjectRelationForPathWalk("document", "group", "user")).To(Equal(""))
		})

		It("Case 36: SelfCycleRelationsForPermission ignores non-self computed relation", func() {
			sch, err := parser.NewParser(`
			entity user {}

			entity resource {
				relation parent @resource
				relation owner @user
				permission edit = owner
				permission view = parent.edit
			}
			`).Parse()

			Expect(err).ShouldNot(HaveOccurred())

			c := compiler.NewCompiler(true, sch)
			a, _, _ := c.Compile()

			g := NewLinkedGraph(NewSchemaFromEntityAndRuleDefinitions(a, nil))

			Expect(g.SelfCycleRelationsForPermission("resource", "view")).To(BeEmpty())
		})
	})

	Context("BuildRelationPathChain", func() {
		It("should find direct relation path", func() {
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

			path, err := g.BuildRelationPathChain("document", "user")

			Expect(err).ShouldNot(HaveOccurred())
			Expect(path).Should(HaveLen(1))
			Expect(path[0].Type).Should(Equal("document"))
			Expect(path[0].Relation).Should(Equal("viewer"))
		})

		It("should find multi-hop path", func() {
			sch, err := parser.NewParser(`
			entity user {}
			entity organization {
				relation admin @user
			}
			entity document {
				relation org @organization
			}
			`).Parse()

			Expect(err).ShouldNot(HaveOccurred())

			c := compiler.NewCompiler(true, sch)
			a, _, _ := c.Compile()

			g := NewLinkedGraph(NewSchemaFromEntityAndRuleDefinitions(a, nil))

			path, err := g.BuildRelationPathChain("document", "user")

			Expect(err).ShouldNot(HaveOccurred())
			Expect(path).Should(HaveLen(2))
			Expect(path[0].Type).Should(Equal("document"))
			Expect(path[0].Relation).Should(Equal("org"))
			Expect(path[1].Type).Should(Equal("organization"))
			Expect(path[1].Relation).Should(Equal("admin"))
		})

		It("should return error when no path found", func() {
			sch, err := parser.NewParser(`
			entity user {}
			entity document {
				relation viewer @user
			}
			entity isolated {
				relation owner @user
			}
			`).Parse()

			Expect(err).ShouldNot(HaveOccurred())

			c := compiler.NewCompiler(true, sch)
			a, _, _ := c.Compile()

			g := NewLinkedGraph(NewSchemaFromEntityAndRuleDefinitions(a, nil))

			path, err := g.BuildRelationPathChain("document", "isolated")

			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(Equal("no path found between entity types"))
			Expect(path).Should(BeNil())
		})

		It("should find path in complex multi-level schema", func() {
			sch, err := parser.NewParser(`
			entity user {}
			entity organization {
				relation admin @user
			}
			entity container {
				relation parent @organization
			}
			entity document {
				relation container @container
			}
			`).Parse()

			Expect(err).ShouldNot(HaveOccurred())

			c := compiler.NewCompiler(true, sch)
			a, _, _ := c.Compile()

			g := NewLinkedGraph(NewSchemaFromEntityAndRuleDefinitions(a, nil))

			path, err := g.BuildRelationPathChain("document", "user")

			Expect(err).ShouldNot(HaveOccurred())
			Expect(path).Should(HaveLen(3))
			Expect(path[0].Type).Should(Equal("document"))
			Expect(path[0].Relation).Should(Equal("container"))
			Expect(path[1].Type).Should(Equal("container"))
			Expect(path[1].Relation).Should(Equal("parent"))
			Expect(path[2].Type).Should(Equal("organization"))
			Expect(path[2].Relation).Should(Equal("admin"))
		})

		It("should handle self-referential relation (same source and target entity type)", func() {
			sch, err := parser.NewParser(`
			entity user {}
			entity account {
				relation admin @user @account#admin
			}
			`).Parse()

			Expect(err).ShouldNot(HaveOccurred())

			c := compiler.NewCompiler(true, sch)
			a, _, _ := c.Compile()

			g := NewLinkedGraph(NewSchemaFromEntityAndRuleDefinitions(a, nil))

			path, err := g.BuildRelationPathChain("account", "account")

			Expect(err).ShouldNot(HaveOccurred())
			Expect(path).Should(HaveLen(1))
			Expect(path[0].Type).Should(Equal("account"))
			Expect(path[0].Relation).Should(Equal("admin"))
		})

		It("should handle non-existent source entity", func() {
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

			path, err := g.BuildRelationPathChain("nonexistent", "user")

			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(Equal("no path found between entity types"))
			Expect(path).Should(BeNil())
		})

		It("should handle non-existent target entity", func() {
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

			path, err := g.BuildRelationPathChain("document", "nonexistent")

			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(Equal("no path found between entity types"))
			Expect(path).Should(BeNil())
		})
	})

	Context("GetSubjectRelationForPathWalk", func() {
		It("should return empty string for simple relation without relation name", func() {
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

			relation := g.GetSubjectRelationForPathWalk("document", "viewer", "user")

			Expect(relation).Should(Equal(""))
		})

		It("should return relation for complex relation with relation name", func() {
			sch, err := parser.NewParser(`
			entity user {}
			entity group {
				relation member @user
			}
			entity document {
				relation viewer @user @group#member
			}
			`).Parse()

			Expect(err).ShouldNot(HaveOccurred())

			c := compiler.NewCompiler(true, sch)
			a, _, _ := c.Compile()

			g := NewLinkedGraph(NewSchemaFromEntityAndRuleDefinitions(a, nil))

			relation := g.GetSubjectRelationForPathWalk("document", "viewer", "group")

			Expect(relation).Should(Equal("member"))
		})

		It("should return empty string for non-existent entity", func() {
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

			relation := g.GetSubjectRelationForPathWalk("nonexistent", "viewer", "user")

			Expect(relation).Should(Equal(""))
		})

		It("should return empty string for non-existent relation", func() {
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

			relation := g.GetSubjectRelationForPathWalk("document", "nonexistent", "user")

			Expect(relation).Should(Equal(""))
		})

		It("should return empty string when target entity type not found in relation references", func() {
			sch, err := parser.NewParser(`
			entity user {}
			entity organization {
				relation admin @user
			}
			entity document {
				relation viewer @user
			}
			`).Parse()

			Expect(err).ShouldNot(HaveOccurred())

			c := compiler.NewCompiler(true, sch)
			a, _, _ := c.Compile()

			g := NewLinkedGraph(NewSchemaFromEntityAndRuleDefinitions(a, nil))

			relation := g.GetSubjectRelationForPathWalk("document", "viewer", "organization")

			Expect(relation).Should(Equal(""))
		})

		It("should return correct relation for multiple relation references", func() {
			sch, err := parser.NewParser(`
			entity user {}
			entity group {
				relation member @user
				relation admin @user
			}
			entity document {
				relation viewer @user @group#member @group#admin
			}
			`).Parse()

			Expect(err).ShouldNot(HaveOccurred())

			c := compiler.NewCompiler(true, sch)
			a, _, _ := c.Compile()

			g := NewLinkedGraph(NewSchemaFromEntityAndRuleDefinitions(a, nil))

			relation := g.GetSubjectRelationForPathWalk("document", "viewer", "group")

			Expect(relation).Should(Equal("member"))
		})
	})

	Context("Attribute Reference", func() {
		It("should return AttributeLinkedEntrance for direct attribute reference", func() {
			sch, err := parser.NewParser(`
			entity user {}
			entity document {
				relation viewer @user
				attribute public boolean
			}
			`).Parse()

			Expect(err).ShouldNot(HaveOccurred())

			c := compiler.NewCompiler(true, sch)
			a, _, _ := c.Compile()

			schema := NewSchemaFromEntityAndRuleDefinitions(a, nil)
			// Manually set attribute as reference to test the attribute reference case
			docDef := schema.EntityDefinitions["document"]
			docDef.References["public"] = base.EntityDefinition_REFERENCE_ATTRIBUTE

			g := NewLinkedGraph(schema)

			ent, err := g.LinkedEntrances(&base.Entrance{
				Type:  "document",
				Value: "public",
			}, &base.Entrance{
				Type:  "user",
				Value: "",
			})

			Expect(err).ShouldNot(HaveOccurred())
			Expect(ent).Should(Equal([]*LinkedEntrance{
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

		It("should return error when attribute not found", func() {
			sch, err := parser.NewParser(`
			entity user {}
			entity document {
				relation viewer @user
				attribute public boolean
			}
			`).Parse()

			Expect(err).ShouldNot(HaveOccurred())

			c := compiler.NewCompiler(true, sch)
			a, _, _ := c.Compile()

			schema := NewSchemaFromEntityAndRuleDefinitions(a, nil)
			// Manually set nonexistent attribute as reference to test the attribute reference case
			// This will trigger the attribute reference case but attribute won't be found
			docDef := schema.EntityDefinitions["document"]
			docDef.References["nonexistent_attribute"] = base.EntityDefinition_REFERENCE_ATTRIBUTE

			g := NewLinkedGraph(schema)

			ent, err := g.LinkedEntrances(&base.Entrance{
				Type:  "document",
				Value: "nonexistent_attribute",
			}, &base.Entrance{
				Type:  "user",
				Value: "",
			})

			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(Equal("attribute not found"))
			Expect(ent).Should(BeNil())
		})

		It("should return AttributeLinkedEntrance for multiple attributes", func() {
			sch, err := parser.NewParser(`
			entity user {}
			entity document {
				relation viewer @user
				attribute public boolean
				attribute published boolean
				attribute archived boolean
			}
			`).Parse()

			Expect(err).ShouldNot(HaveOccurred())

			c := compiler.NewCompiler(true, sch)
			a, _, _ := c.Compile()

			schema := NewSchemaFromEntityAndRuleDefinitions(a, nil)
			// Manually set attribute as reference to test the attribute reference case
			docDef := schema.EntityDefinitions["document"]
			docDef.References["published"] = base.EntityDefinition_REFERENCE_ATTRIBUTE

			g := NewLinkedGraph(schema)

			ent, err := g.LinkedEntrances(&base.Entrance{
				Type:  "document",
				Value: "published",
			}, &base.Entrance{
				Type:  "user",
				Value: "",
			})

			Expect(err).ShouldNot(HaveOccurred())
			Expect(ent).Should(Equal([]*LinkedEntrance{
				{
					Kind: AttributeLinkedEntrance,
					TargetEntrance: &base.Entrance{
						Type:  "document",
						Value: "published",
					},
					TupleSetRelation: "",
				},
			}))
		})

		It("should return AttributeLinkedEntrance for different attribute types", func() {
			sch, err := parser.NewParser(`
			entity user {}
			entity document {
				relation viewer @user
				attribute public boolean
				attribute count integer
				attribute name string
			}
			`).Parse()

			Expect(err).ShouldNot(HaveOccurred())

			c := compiler.NewCompiler(true, sch)
			a, _, _ := c.Compile()

			schema := NewSchemaFromEntityAndRuleDefinitions(a, nil)
			// Manually set attribute as reference to test the attribute reference case
			docDef := schema.EntityDefinitions["document"]
			docDef.References["count"] = base.EntityDefinition_REFERENCE_ATTRIBUTE

			g := NewLinkedGraph(schema)

			ent, err := g.LinkedEntrances(&base.Entrance{
				Type:  "document",
				Value: "count",
			}, &base.Entrance{
				Type:  "user",
				Value: "",
			})

			Expect(err).ShouldNot(HaveOccurred())
			Expect(ent).Should(Equal([]*LinkedEntrance{
				{
					Kind: AttributeLinkedEntrance,
					TargetEntrance: &base.Entrance{
						Type:  "document",
						Value: "count",
					},
					TupleSetRelation: "",
				},
			}))
		})
	})
})
