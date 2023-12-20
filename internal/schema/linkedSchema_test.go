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

			ent, err := g.RelationshipLinkedEntrances(&base.RelationReference{
				Type:     "document",
				Relation: "viewer",
			}, &base.RelationReference{
				Type:     "user",
				Relation: "",
			})

			Expect(err).ShouldNot(HaveOccurred())
			Expect(ent).Should(Equal([]*LinkedEntrance{
				{
					Kind: RelationLinkedEntrance,
					TargetEntrance: &base.RelationReference{
						Type:     "document",
						Relation: "viewer",
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

			ent, err := g.RelationshipLinkedEntrances(&base.RelationReference{
				Type:     "document",
				Relation: "view",
			}, &base.RelationReference{
				Type:     "user",
				Relation: "",
			})

			Expect(err).ShouldNot(HaveOccurred())
			Expect(ent).Should(Equal([]*LinkedEntrance{
				{
					Kind: RelationLinkedEntrance,
					TargetEntrance: &base.RelationReference{
						Type:     "document",
						Relation: "viewer",
					},
					TupleSetRelation: "",
				},
				{
					Kind: RelationLinkedEntrance,
					TargetEntrance: &base.RelationReference{
						Type:     "document",
						Relation: "editor",
					},
					TupleSetRelation: "",
				},
				{
					Kind: RelationLinkedEntrance,
					TargetEntrance: &base.RelationReference{
						Type:     "document",
						Relation: "owner",
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

			ent, err := g.RelationshipLinkedEntrances(&base.RelationReference{
				Type:     "document",
				Relation: "view",
			}, &base.RelationReference{
				Type:     "user",
				Relation: "",
			})

			Expect(err).ShouldNot(HaveOccurred())

			Expect(ent).Should(Equal([]*LinkedEntrance{
				{
					Kind: RelationLinkedEntrance,
					TargetEntrance: &base.RelationReference{
						Type:     "document",
						Relation: "viewer",
					},
					TupleSetRelation: "",
				},
				{
					Kind: RelationLinkedEntrance,
					TargetEntrance: &base.RelationReference{
						Type:     "document",
						Relation: "editor",
					},
					TupleSetRelation: "",
				},
				{
					Kind: RelationLinkedEntrance,
					TargetEntrance: &base.RelationReference{
						Type:     "document",
						Relation: "owner",
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

			ent, err := g.RelationshipLinkedEntrances(&base.RelationReference{
				Type:     "document",
				Relation: "view",
			}, &base.RelationReference{
				Type:     "user",
				Relation: "",
			})

			Expect(err).ShouldNot(HaveOccurred())

			Expect(ent).Should(Equal([]*LinkedEntrance{
				{
					Kind: RelationLinkedEntrance,
					TargetEntrance: &base.RelationReference{
						Type:     "document",
						Relation: "viewer",
					},
					TupleSetRelation: "",
				},
				{
					Kind: RelationLinkedEntrance,
					TargetEntrance: &base.RelationReference{
						Type:     "document",
						Relation: "owner",
					},
					TupleSetRelation: "",
				},
				{
					Kind: RelationLinkedEntrance,
					TargetEntrance: &base.RelationReference{
						Type:     "organization",
						Relation: "admin",
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

			ent, err := g.RelationshipLinkedEntrances(&base.RelationReference{
				Type:     "document",
				Relation: "view",
			}, &base.RelationReference{
				Type:     "user",
				Relation: "",
			})

			Expect(err).ShouldNot(HaveOccurred())

			Expect(ent).Should(Equal([]*LinkedEntrance{
				{
					Kind: RelationLinkedEntrance,
					TargetEntrance: &base.RelationReference{
						Type:     "document",
						Relation: "viewer",
					},
					TupleSetRelation: "",
				},
				{
					Kind: RelationLinkedEntrance,
					TargetEntrance: &base.RelationReference{
						Type:     "document",
						Relation: "owner",
					},
					TupleSetRelation: "",
				},
				{
					Kind: RelationLinkedEntrance,
					TargetEntrance: &base.RelationReference{
						Type:     "organization",
						Relation: "admin",
					},
					TupleSetRelation: "",
				},
				{
					Kind: RelationLinkedEntrance,
					TargetEntrance: &base.RelationReference{
						Type:     "container",
						Relation: "container_admin",
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

			ent, err := g.RelationshipLinkedEntrances(&base.RelationReference{
				Type:     "document",
				Relation: "view",
			}, &base.RelationReference{
				Type:     "user",
				Relation: "",
			})

			Expect(err).ShouldNot(HaveOccurred())

			Expect(ent).Should(Equal([]*LinkedEntrance{
				{
					Kind: RelationLinkedEntrance,
					TargetEntrance: &base.RelationReference{
						Type:     "document",
						Relation: "viewer",
					},
					TupleSetRelation: "",
				},
				{
					Kind: RelationLinkedEntrance,
					TargetEntrance: &base.RelationReference{
						Type:     "group",
						Relation: "manager",
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

			ent, err := g.RelationshipLinkedEntrances(&base.RelationReference{
				Type:     "document",
				Relation: "view",
			}, &base.RelationReference{
				Type:     "user",
				Relation: "",
			})

			Expect(err).ShouldNot(HaveOccurred())

			Expect(ent).Should(Equal([]*LinkedEntrance{
				{
					Kind: RelationLinkedEntrance,
					TargetEntrance: &base.RelationReference{
						Type:     "document",
						Relation: "viewer",
					},
					TupleSetRelation: "",
				},
				{
					Kind: RelationLinkedEntrance,
					TargetEntrance: &base.RelationReference{
						Type:     "organization",
						Relation: "banned",
					},
					TupleSetRelation: "",
				},
				{
					Kind: RelationLinkedEntrance,
					TargetEntrance: &base.RelationReference{
						Type:     "organization",
						Relation: "admin",
					},
					TupleSetRelation: "",
				},
				{
					Kind: RelationLinkedEntrance,
					TargetEntrance: &base.RelationReference{
						Type:     "document",
						Relation: "fist_rel",
					},
					TupleSetRelation: "",
				},
				{
					Kind: RelationLinkedEntrance,
					TargetEntrance: &base.RelationReference{
						Type:     "document",
						Relation: "second_rel",
					},
					TupleSetRelation: "",
				},
				{
					Kind: RelationLinkedEntrance,
					TargetEntrance: &base.RelationReference{
						Type:     "document",
						Relation: "third_rel",
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

			ent, err := g.RelationshipLinkedEntrances(&base.RelationReference{
				Type:     "document",
				Relation: "view",
			}, &base.RelationReference{
				Type:     "document",
				Relation: "viewer",
			})

			Expect(err).ShouldNot(HaveOccurred())

			Expect(ent).Should(Equal([]*LinkedEntrance{
				{
					Kind: ComputedUserSetLinkedEntrance,
					TargetEntrance: &base.RelationReference{
						Type:     "document",
						Relation: "view",
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

			ent, err := g.RelationshipLinkedEntrances(&base.RelationReference{
				Type:     "document",
				Relation: "view",
			}, &base.RelationReference{
				Type:     "organization",
				Relation: "admin",
			})

			Expect(err).ShouldNot(HaveOccurred())
			Expect(ent).Should(Equal([]*LinkedEntrance{
				{
					Kind: TupleToUserSetLinkedEntrance,
					TargetEntrance: &base.RelationReference{
						Type:     "document",
						Relation: "view",
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

			ent, err := g.RelationshipLinkedEntrances(&base.RelationReference{
				Type:     "document",
				Relation: "view",
			}, &base.RelationReference{
				Type:     "container",
				Relation: "local_admin",
			})

			Expect(err).ShouldNot(HaveOccurred())

			Expect(ent).Should(Equal([]*LinkedEntrance{
				{
					Kind: ComputedUserSetLinkedEntrance,
					TargetEntrance: &base.RelationReference{
						Type:     "container",
						Relation: "admin",
					},
					TupleSetRelation: "",
				},
				{
					Kind: ComputedUserSetLinkedEntrance,
					TargetEntrance: &base.RelationReference{
						Type:     "container",
						Relation: "test",
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

			ent, err := g.RelationshipLinkedEntrances(&base.RelationReference{
				Type:     "document",
				Relation: "view",
			}, &base.RelationReference{
				Type:     "container",
				Relation: "local_admin",
			})

			Expect(err).ShouldNot(HaveOccurred())

			Expect(ent).Should(Equal([]*LinkedEntrance{
				{
					Kind: ComputedUserSetLinkedEntrance,
					TargetEntrance: &base.RelationReference{
						Type:     "container",
						Relation: "admin",
					},
					TupleSetRelation: "",
				},
				{
					Kind: ComputedUserSetLinkedEntrance,
					TargetEntrance: &base.RelationReference{
						Type:     "container",
						Relation: "test",
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

			ent, err := g.RelationshipLinkedEntrances(&base.RelationReference{
				Type:     "document",
				Relation: "view",
			}, &base.RelationReference{
				Type:     "document",
				Relation: "viewer",
			})

			Expect(err).ShouldNot(HaveOccurred())

			Expect(ent).Should(Equal([]*LinkedEntrance{
				{
					Kind: ComputedUserSetLinkedEntrance,
					TargetEntrance: &base.RelationReference{
						Type:     "document",
						Relation: "view",
					},
					TupleSetRelation: "",
				},
				{
					Kind: ComputedUserSetLinkedEntrance,
					TargetEntrance: &base.RelationReference{
						Type:     "document",
						Relation: "view",
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

			ent, err := g.RelationshipLinkedEntrances(&base.RelationReference{
				Type:     "document",
				Relation: "view",
			}, &base.RelationReference{
				Type:     "user",
				Relation: "",
			})

			Expect(err).ShouldNot(HaveOccurred())

			Expect(ent).Should(Equal([]*LinkedEntrance{
				{
					Kind: RelationLinkedEntrance,
					TargetEntrance: &base.RelationReference{
						Type:     "organization",
						Relation: "viewer",
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

			ent, err := g.RelationshipLinkedEntrances(&base.RelationReference{
				Type:     "document",
				Relation: "view",
			}, &base.RelationReference{
				Type:     "user",
				Relation: "",
			})

			Expect(err).ShouldNot(HaveOccurred())

			Expect(ent).Should(Equal([]*LinkedEntrance{
				{
					Kind: RelationLinkedEntrance,
					TargetEntrance: &base.RelationReference{
						Type:     "document",
						Relation: "viewer",
					},
					TupleSetRelation: "",
				},
				{
					Kind: RelationLinkedEntrance,
					TargetEntrance: &base.RelationReference{
						Type:     "document",
						Relation: "editor",
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

			ent, err := g.RelationshipLinkedEntrances(&base.RelationReference{
				Type:     "account",
				Relation: "add_member",
			}, &base.RelationReference{
				Type:     "user",
				Relation: "",
			})

			Expect(err).ShouldNot(HaveOccurred())
			Expect(ent).Should(Equal([]*LinkedEntrance{
				{
					Kind: RelationLinkedEntrance,
					TargetEntrance: &base.RelationReference{
						Type:     "account",
						Relation: "admin",
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

			ent, err := g.RelationshipLinkedEntrances(&base.RelationReference{
				Type:     "account",
				Relation: "add_member",
			}, &base.RelationReference{
				Type:     "account",
				Relation: "admin",
			})

			Expect(err).ShouldNot(HaveOccurred())
			Expect(ent).Should(Equal([]*LinkedEntrance{
				{
					Kind: ComputedUserSetLinkedEntrance,
					TargetEntrance: &base.RelationReference{
						Type:     "account",
						Relation: "add_member",
					},
					TupleSetRelation: "",
				},
				{
					Kind: RelationLinkedEntrance,
					TargetEntrance: &base.RelationReference{
						Type:     "account",
						Relation: "admin",
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

			ent, err := g.RelationshipLinkedEntrances(&base.RelationReference{
				Type:     "account",
				Relation: "admin",
			}, &base.RelationReference{
				Type:     "account",
				Relation: "admin",
			})

			Expect(err).ShouldNot(HaveOccurred())
			Expect(ent).Should(Equal([]*LinkedEntrance{
				{
					Kind: RelationLinkedEntrance,
					TargetEntrance: &base.RelationReference{
						Type:     "account",
						Relation: "admin",
					},
					TupleSetRelation: "",
				},
			}))
		})
	})
})
