package compiler

import (
	"errors"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Permify/permify/pkg/dsl/parser"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// TestCompiler -
func TestCompiler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "compiler-suite")
}

var _ = Describe("compiler", func() {
	Context("NewCompiler", func() {
		It("Case 1", func() {
			sch, err := parser.NewParser(`
			entity user {}`).Parse()

			Expect(err).ShouldNot(HaveOccurred())

			c := NewCompiler(true, sch)

			var is []*base.EntityDefinition
			is, err = c.Compile()
			Expect(err).ShouldNot(HaveOccurred())

			Expect(is).Should(Equal([]*base.EntityDefinition{
				{
					Name:        "user",
					Relations:   map[string]*base.RelationDefinition{},
					Permissions: map[string]*base.PermissionDefinition{},
					References:  map[string]base.EntityDefinition_RelationalReference{},
				},
			}))
		})

		It("Case 2", func() {
			sch, err := parser.NewParser(`
			entity user {}
		
			entity organization {
		
				relation owner @user
				relation admin @user
		
				permission update = owner or admin
			}
			`).Parse()

			Expect(err).ShouldNot(HaveOccurred())

			c := NewCompiler(true, sch)

			var is []*base.EntityDefinition
			is, err = c.Compile()
			Expect(err).ShouldNot(HaveOccurred())

			i := []*base.EntityDefinition{
				{
					Name:        "user",
					Relations:   map[string]*base.RelationDefinition{},
					Permissions: map[string]*base.PermissionDefinition{},
					References:  map[string]base.EntityDefinition_RelationalReference{},
				},
				{
					Name: "organization",
					Permissions: map[string]*base.PermissionDefinition{
						"update": {
							Name: "update",
							Child: &base.Child{
								Type: &base.Child_Rewrite{
									Rewrite: &base.Rewrite{
										RewriteOperation: base.Rewrite_OPERATION_UNION,
										Children: []*base.Child{
											{
												Type: &base.Child_Leaf{
													Leaf: &base.Leaf{
														Type: &base.Leaf_ComputedUserSet{
															ComputedUserSet: &base.ComputedUserSet{
																Relation: "owner",
															},
														},
													},
												},
											},
											{
												Type: &base.Child_Leaf{
													Leaf: &base.Leaf{
														Type: &base.Leaf_ComputedUserSet{
															ComputedUserSet: &base.ComputedUserSet{
																Relation: "admin",
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
					Relations: map[string]*base.RelationDefinition{
						"owner": {
							Name: "owner",
							RelationReferences: []*base.RelationReference{
								{
									Type:     "user",
									Relation: "",
								},
							},
						},
						"admin": {
							Name: "admin",
							RelationReferences: []*base.RelationReference{
								{
									Type:     "user",
									Relation: "",
								},
							},
						},
					},
					References: map[string]base.EntityDefinition_RelationalReference{
						"owner":  base.EntityDefinition_RELATIONAL_REFERENCE_RELATION,
						"admin":  base.EntityDefinition_RELATIONAL_REFERENCE_RELATION,
						"update": base.EntityDefinition_RELATIONAL_REFERENCE_PERMISSION,
					},
				},
			}

			Expect(is).Should(Equal(i))
		})

		It("Case 3", func() {
			sch, err := parser.NewParser(`
			entity user {}
		
			entity organization {
		
				relation owner @user
				relation admin @user
		
				permission update = owner or (admin and owner)
			}
			`).Parse()

			Expect(err).ShouldNot(HaveOccurred())

			c := NewCompiler(true, sch)

			var is []*base.EntityDefinition
			is, err = c.Compile()
			Expect(err).ShouldNot(HaveOccurred())

			i := []*base.EntityDefinition{
				{
					Name:        "user",
					Relations:   map[string]*base.RelationDefinition{},
					Permissions: map[string]*base.PermissionDefinition{},
					References:  map[string]base.EntityDefinition_RelationalReference{},
				},
				{
					Name: "organization",
					Permissions: map[string]*base.PermissionDefinition{
						"update": {
							Name: "update",
							Child: &base.Child{
								Type: &base.Child_Rewrite{
									Rewrite: &base.Rewrite{
										RewriteOperation: base.Rewrite_OPERATION_UNION,
										Children: []*base.Child{
											{
												Type: &base.Child_Leaf{
													Leaf: &base.Leaf{
														Type: &base.Leaf_ComputedUserSet{
															ComputedUserSet: &base.ComputedUserSet{
																Relation: "owner",
															},
														},
													},
												},
											},
											{
												Type: &base.Child_Rewrite{
													Rewrite: &base.Rewrite{
														RewriteOperation: base.Rewrite_OPERATION_INTERSECTION,
														Children: []*base.Child{
															{
																Type: &base.Child_Leaf{
																	Leaf: &base.Leaf{
																		Type: &base.Leaf_ComputedUserSet{
																			ComputedUserSet: &base.ComputedUserSet{
																				Relation: "admin",
																			},
																		},
																	},
																},
															},
															{
																Type: &base.Child_Leaf{
																	Leaf: &base.Leaf{
																		Type: &base.Leaf_ComputedUserSet{
																			ComputedUserSet: &base.ComputedUserSet{
																				Relation: "owner",
																			},
																		},
																	},
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
					Relations: map[string]*base.RelationDefinition{
						"owner": {
							Name: "owner",
							RelationReferences: []*base.RelationReference{
								{
									Type:     "user",
									Relation: "",
								},
							},
						},
						"admin": {
							Name: "admin",
							RelationReferences: []*base.RelationReference{
								{
									Type:     "user",
									Relation: "",
								},
							},
						},
					},
					References: map[string]base.EntityDefinition_RelationalReference{
						"owner":  base.EntityDefinition_RELATIONAL_REFERENCE_RELATION,
						"admin":  base.EntityDefinition_RELATIONAL_REFERENCE_RELATION,
						"update": base.EntityDefinition_RELATIONAL_REFERENCE_PERMISSION,
					},
				},
			}

			Expect(is).Should(Equal(i))
		})

		It("Case 4", func() {
			sch, err := parser.NewParser(`
			entity user {}
		
			entity organization {
		
				relation owner @user
				relation admin @user
		
				permission update = owner
			}
			`).Parse()

			Expect(err).ShouldNot(HaveOccurred())

			c := NewCompiler(true, sch)

			var is []*base.EntityDefinition
			is, err = c.Compile()
			Expect(err).ShouldNot(HaveOccurred())

			i := []*base.EntityDefinition{
				{
					Name:        "user",
					Relations:   map[string]*base.RelationDefinition{},
					Permissions: map[string]*base.PermissionDefinition{},
					References:  map[string]base.EntityDefinition_RelationalReference{},
				},
				{
					Name: "organization",
					Permissions: map[string]*base.PermissionDefinition{
						"update": {
							Name: "update",
							Child: &base.Child{
								Type: &base.Child_Leaf{
									Leaf: &base.Leaf{
										Type: &base.Leaf_ComputedUserSet{
											ComputedUserSet: &base.ComputedUserSet{
												Relation: "owner",
											},
										},
									},
								},
							},
						},
					},
					Relations: map[string]*base.RelationDefinition{
						"owner": {
							Name: "owner",
							RelationReferences: []*base.RelationReference{
								{
									Type:     "user",
									Relation: "",
								},
							},
						},
						"admin": {
							Name: "admin",
							RelationReferences: []*base.RelationReference{
								{
									Type:     "user",
									Relation: "",
								},
							},
						},
					},
					References: map[string]base.EntityDefinition_RelationalReference{
						"owner":  base.EntityDefinition_RELATIONAL_REFERENCE_RELATION,
						"admin":  base.EntityDefinition_RELATIONAL_REFERENCE_RELATION,
						"update": base.EntityDefinition_RELATIONAL_REFERENCE_PERMISSION,
					},
				},
			}

			Expect(is).Should(Equal(i))
		})

		It("Case 5", func() {
			sch, err := parser.NewParser(`
			entity user {}
		
			entity organization {
		
				relation owner @user
				relation admin @user
		
				permission update = maintainer or admin
			}
			`).Parse()

			Expect(err).ShouldNot(HaveOccurred())

			c := NewCompiler(true, sch)

			_, err = c.Compile()
			Expect(err).Should(Equal(errors.New("9:26: undefined relation reference")))
		})

		It("Case 6", func() {
			sch, err := parser.NewParser(`
			entity user {}
		
			entity parent {
		
				relation admin @user
			}
		
			entity organization {
		
				relation parent @parent
				relation admin @user
			}
		
			entity repository {
		
				relation parent @organization
				permission update = parent.parent.admin or admin
			}
			`).Parse()

			Expect(err).ShouldNot(HaveOccurred())

			c := NewCompiler(true, sch)

			_, err = c.Compile()
			Expect(err).Should(Equal(errors.New("18:40: not supported relation walk")))
		})

		It("Case 7", func() {
			sch, err := parser.NewParser(`
			entity user {}
		
			entity organization {
		
				relation owner @user
				relation admin @user
		
				permission update = owner or admin
			}
		
			entity repository {
		
				relation parent @organization
				relation owner @user
		
				permission delete = owner or (parent.update not parent.owner)
			}
		
			`).Parse()

			Expect(err).ShouldNot(HaveOccurred())

			c := NewCompiler(true, sch)

			var is []*base.EntityDefinition
			is, err = c.Compile()
			Expect(err).ShouldNot(HaveOccurred())

			i := []*base.EntityDefinition{
				{
					Name:        "user",
					Relations:   map[string]*base.RelationDefinition{},
					Permissions: map[string]*base.PermissionDefinition{},
					References:  map[string]base.EntityDefinition_RelationalReference{},
				},
				{
					Name: "organization",
					Permissions: map[string]*base.PermissionDefinition{
						"update": {
							Name: "update",
							Child: &base.Child{
								Type: &base.Child_Rewrite{
									Rewrite: &base.Rewrite{
										RewriteOperation: base.Rewrite_OPERATION_UNION,
										Children: []*base.Child{
											{
												Type: &base.Child_Leaf{
													Leaf: &base.Leaf{
														Type: &base.Leaf_ComputedUserSet{
															ComputedUserSet: &base.ComputedUserSet{
																Relation: "owner",
															},
														},
													},
												},
											},
											{
												Type: &base.Child_Leaf{
													Leaf: &base.Leaf{
														Type: &base.Leaf_ComputedUserSet{
															ComputedUserSet: &base.ComputedUserSet{
																Relation: "admin",
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
					Relations: map[string]*base.RelationDefinition{
						"owner": {
							Name: "owner",
							RelationReferences: []*base.RelationReference{
								{
									Type:     "user",
									Relation: "",
								},
							},
						},
						"admin": {
							Name: "admin",
							RelationReferences: []*base.RelationReference{
								{
									Type:     "user",
									Relation: "",
								},
							},
						},
					},
					References: map[string]base.EntityDefinition_RelationalReference{
						"owner":  base.EntityDefinition_RELATIONAL_REFERENCE_RELATION,
						"admin":  base.EntityDefinition_RELATIONAL_REFERENCE_RELATION,
						"update": base.EntityDefinition_RELATIONAL_REFERENCE_PERMISSION,
					},
				},
				{
					Name: "repository",
					Permissions: map[string]*base.PermissionDefinition{
						"delete": {
							Name: "delete",
							Child: &base.Child{
								Type: &base.Child_Rewrite{
									Rewrite: &base.Rewrite{
										RewriteOperation: base.Rewrite_OPERATION_UNION,
										Children: []*base.Child{
											{
												Type: &base.Child_Leaf{
													Leaf: &base.Leaf{
														Type: &base.Leaf_ComputedUserSet{
															ComputedUserSet: &base.ComputedUserSet{
																Relation: "owner",
															},
														},
													},
												},
											},
											{
												Type: &base.Child_Rewrite{
													Rewrite: &base.Rewrite{
														RewriteOperation: base.Rewrite_OPERATION_EXCLUSION,
														Children: []*base.Child{
															{
																Type: &base.Child_Leaf{
																	Leaf: &base.Leaf{
																		Type: &base.Leaf_TupleToUserSet{
																			TupleToUserSet: &base.TupleToUserSet{
																				TupleSet: &base.TupleSet{
																					Relation: "parent",
																				},
																				Computed: &base.ComputedUserSet{
																					Relation: "update",
																				},
																			},
																		},
																	},
																},
															},
															{
																Type: &base.Child_Leaf{
																	Leaf: &base.Leaf{
																		Type: &base.Leaf_TupleToUserSet{
																			TupleToUserSet: &base.TupleToUserSet{
																				TupleSet: &base.TupleSet{
																					Relation: "parent",
																				},
																				Computed: &base.ComputedUserSet{
																					Relation: "owner",
																				},
																			},
																		},
																	},
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
					Relations: map[string]*base.RelationDefinition{
						"parent": {
							Name: "parent",
							RelationReferences: []*base.RelationReference{
								{
									Type:     "organization",
									Relation: "",
								},
							},
						},
						"owner": {
							Name: "owner",
							RelationReferences: []*base.RelationReference{
								{
									Type:     "user",
									Relation: "",
								},
							},
						},
					},
					References: map[string]base.EntityDefinition_RelationalReference{
						"parent": base.EntityDefinition_RELATIONAL_REFERENCE_RELATION,
						"owner":  base.EntityDefinition_RELATIONAL_REFERENCE_RELATION,
						"delete": base.EntityDefinition_RELATIONAL_REFERENCE_PERMISSION,
					},
				},
			}

			Expect(is).Should(Equal(i))
		})

		It("Case 8", func() {
			sch, err := parser.NewParser(`
			entity user {}
		
			entity organization {
		
				relation owner @user
				relation admin @user
		
				permission update = owner or admin
			}
		
			entity repository {
		
				relation parent @organization
				relation owner @user @organization#admin @organization#owner
		
				permission delete = owner or (parent.update not parent.owner)
			}
		
			`).Parse()

			Expect(err).ShouldNot(HaveOccurred())

			c := NewCompiler(true, sch)

			var is []*base.EntityDefinition
			is, err = c.Compile()
			Expect(err).ShouldNot(HaveOccurred())

			i := []*base.EntityDefinition{
				{
					Name:        "user",
					Relations:   map[string]*base.RelationDefinition{},
					Permissions: map[string]*base.PermissionDefinition{},
					References:  map[string]base.EntityDefinition_RelationalReference{},
				},
				{
					Name: "organization",
					Permissions: map[string]*base.PermissionDefinition{
						"update": {
							Name: "update",
							Child: &base.Child{
								Type: &base.Child_Rewrite{
									Rewrite: &base.Rewrite{
										RewriteOperation: base.Rewrite_OPERATION_UNION,
										Children: []*base.Child{
											{
												Type: &base.Child_Leaf{
													Leaf: &base.Leaf{
														Type: &base.Leaf_ComputedUserSet{
															ComputedUserSet: &base.ComputedUserSet{
																Relation: "owner",
															},
														},
													},
												},
											},
											{
												Type: &base.Child_Leaf{
													Leaf: &base.Leaf{
														Type: &base.Leaf_ComputedUserSet{
															ComputedUserSet: &base.ComputedUserSet{
																Relation: "admin",
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
					Relations: map[string]*base.RelationDefinition{
						"owner": {
							Name: "owner",
							RelationReferences: []*base.RelationReference{
								{
									Type:     "user",
									Relation: "",
								},
							},
						},
						"admin": {
							Name: "admin",
							RelationReferences: []*base.RelationReference{
								{
									Type:     "user",
									Relation: "",
								},
							},
						},
					},
					References: map[string]base.EntityDefinition_RelationalReference{
						"owner":  base.EntityDefinition_RELATIONAL_REFERENCE_RELATION,
						"admin":  base.EntityDefinition_RELATIONAL_REFERENCE_RELATION,
						"update": base.EntityDefinition_RELATIONAL_REFERENCE_PERMISSION,
					},
				},
				{
					Name: "repository",
					Permissions: map[string]*base.PermissionDefinition{
						"delete": {
							Name: "delete",
							Child: &base.Child{
								Type: &base.Child_Rewrite{
									Rewrite: &base.Rewrite{
										RewriteOperation: base.Rewrite_OPERATION_UNION,
										Children: []*base.Child{
											{
												Type: &base.Child_Leaf{
													Leaf: &base.Leaf{
														Type: &base.Leaf_ComputedUserSet{
															ComputedUserSet: &base.ComputedUserSet{
																Relation: "owner",
															},
														},
													},
												},
											},
											{
												Type: &base.Child_Rewrite{
													Rewrite: &base.Rewrite{
														RewriteOperation: base.Rewrite_OPERATION_EXCLUSION,
														Children: []*base.Child{
															{
																Type: &base.Child_Leaf{
																	Leaf: &base.Leaf{
																		Type: &base.Leaf_TupleToUserSet{
																			TupleToUserSet: &base.TupleToUserSet{
																				TupleSet: &base.TupleSet{
																					Relation: "parent",
																				},
																				Computed: &base.ComputedUserSet{
																					Relation: "update",
																				},
																			},
																		},
																	},
																},
															},
															{
																Type: &base.Child_Leaf{
																	Leaf: &base.Leaf{
																		Type: &base.Leaf_TupleToUserSet{
																			TupleToUserSet: &base.TupleToUserSet{
																				TupleSet: &base.TupleSet{
																					Relation: "parent",
																				},
																				Computed: &base.ComputedUserSet{
																					Relation: "owner",
																				},
																			},
																		},
																	},
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
					Relations: map[string]*base.RelationDefinition{
						"parent": {
							Name: "parent",
							RelationReferences: []*base.RelationReference{
								{
									Type:     "organization",
									Relation: "",
								},
							},
						},
						"owner": {
							Name: "owner",
							RelationReferences: []*base.RelationReference{
								{
									Type:     "user",
									Relation: "",
								},
								{
									Type:     "organization",
									Relation: "admin",
								},
								{
									Type:     "organization",
									Relation: "owner",
								},
							},
						},
					},
					References: map[string]base.EntityDefinition_RelationalReference{
						"parent": base.EntityDefinition_RELATIONAL_REFERENCE_RELATION,
						"owner":  base.EntityDefinition_RELATIONAL_REFERENCE_RELATION,
						"delete": base.EntityDefinition_RELATIONAL_REFERENCE_PERMISSION,
					},
				},
			}

			Expect(is).Should(Equal(i))
		})

		It("Case 9", func() {
			sch, err := parser.NewParser(`
			entity user {}
		
			entity organization {
		
				relation owner @user
				relation admin @user
		
				permission update = owner or admin
			}
		
			entity repository {
		
				relation parent @organization
				relation owner @user @organization#update
		
				permission delete = owner or (parent.update not parent.owner)
			}
		
			`).Parse()

			Expect(err).ShouldNot(HaveOccurred())

			c := NewCompiler(true, sch)

			_, err = c.Compile()
			Expect(err.Error()).Should(Equal("15:28: relation reference not found in entity references"))
		})

		It("Case 10", func() {
			sch, err := parser.NewParser(`
			entity user {
				relation org @organization
		
				permission read = org.admin
				permission write = org.admin
			}
		
			entity organization {
				relation admin @user
			}
		
			entity division {
				relation manager @user @organization#admin
		
				permission read = manager
				permission write = manager
			}
			`).Parse()

			Expect(err).ShouldNot(HaveOccurred())

			c := NewCompiler(true, sch)

			var is []*base.EntityDefinition
			is, err = c.Compile()

			Expect(err).ShouldNot(HaveOccurred())

			i := []*base.EntityDefinition{
				{
					Name: "user",
					Relations: map[string]*base.RelationDefinition{
						"org": {
							Name: "org",
							RelationReferences: []*base.RelationReference{
								{
									Type:     "organization",
									Relation: "",
								},
							},
						},
					},
					Permissions: map[string]*base.PermissionDefinition{
						"read": {
							Name: "read",
							Child: &base.Child{
								Type: &base.Child_Leaf{
									Leaf: &base.Leaf{
										Type: &base.Leaf_TupleToUserSet{
											TupleToUserSet: &base.TupleToUserSet{
												TupleSet: &base.TupleSet{
													Relation: "org",
												},
												Computed: &base.ComputedUserSet{
													Relation: "admin",
												},
											},
										},
									},
								},
							},
						},
						"write": {
							Name: "write",
							Child: &base.Child{
								Type: &base.Child_Leaf{
									Leaf: &base.Leaf{
										Type: &base.Leaf_TupleToUserSet{
											TupleToUserSet: &base.TupleToUserSet{
												TupleSet: &base.TupleSet{
													Relation: "org",
												},
												Computed: &base.ComputedUserSet{
													Relation: "admin",
												},
											},
										},
									},
								},
							},
						},
					},
					References: map[string]base.EntityDefinition_RelationalReference{
						"org":   base.EntityDefinition_RELATIONAL_REFERENCE_RELATION,
						"read":  base.EntityDefinition_RELATIONAL_REFERENCE_PERMISSION,
						"write": base.EntityDefinition_RELATIONAL_REFERENCE_PERMISSION,
					},
				},
				{
					Name: "organization",
					Relations: map[string]*base.RelationDefinition{
						"admin": {
							Name: "admin",
							RelationReferences: []*base.RelationReference{
								{
									Type:     "user",
									Relation: "",
								},
							},
						},
					},
					Permissions: map[string]*base.PermissionDefinition{},
					References: map[string]base.EntityDefinition_RelationalReference{
						"admin": base.EntityDefinition_RELATIONAL_REFERENCE_RELATION,
					},
				},
				{
					Name: "division",
					Relations: map[string]*base.RelationDefinition{
						"manager": {
							Name: "manager",
							RelationReferences: []*base.RelationReference{
								{
									Type:     "user",
									Relation: "",
								},
								{
									Type:     "organization",
									Relation: "admin",
								},
							},
						},
					},
					Permissions: map[string]*base.PermissionDefinition{
						"read": {
							Name: "read",
							Child: &base.Child{
								Type: &base.Child_Leaf{
									Leaf: &base.Leaf{
										Type: &base.Leaf_ComputedUserSet{
											ComputedUserSet: &base.ComputedUserSet{
												Relation: "manager",
											},
										},
									},
								},
							},
						},
						"write": {
							Name: "write",
							Child: &base.Child{
								Type: &base.Child_Leaf{
									Leaf: &base.Leaf{
										Type: &base.Leaf_ComputedUserSet{
											ComputedUserSet: &base.ComputedUserSet{
												Relation: "manager",
											},
										},
									},
								},
							},
						},
					},
					References: map[string]base.EntityDefinition_RelationalReference{
						"manager": base.EntityDefinition_RELATIONAL_REFERENCE_RELATION,
						"read":    base.EntityDefinition_RELATIONAL_REFERENCE_PERMISSION,
						"write":   base.EntityDefinition_RELATIONAL_REFERENCE_PERMISSION,
					},
				},
			}

			Expect(is).Should(Equal(i))
		})

		It("Case 11", func() {
			sch, err := parser.NewParser(`
			entity user {
				relation org @organization
		
				permission read = org.admin
				permission write = org.admin
			}
		
			entity organization {
				relation admin @user
			}
		
			entity division {
				relation manager @user @organization#admin
		
				permission read = manager
				permission write = manager
			}
			`).Parse()

			Expect(err).ShouldNot(HaveOccurred())

			c := NewCompiler(true, sch)

			var is []*base.EntityDefinition
			is, err = c.Compile()

			Expect(err).ShouldNot(HaveOccurred())

			i := []*base.EntityDefinition{
				{
					Name: "user",
					Relations: map[string]*base.RelationDefinition{
						"org": {
							Name: "org",
							RelationReferences: []*base.RelationReference{
								{
									Type:     "organization",
									Relation: "",
								},
							},
						},
					},
					Permissions: map[string]*base.PermissionDefinition{
						"read": {
							Name: "read",
							Child: &base.Child{
								Type: &base.Child_Leaf{
									Leaf: &base.Leaf{
										Type: &base.Leaf_TupleToUserSet{
											TupleToUserSet: &base.TupleToUserSet{
												TupleSet: &base.TupleSet{
													Relation: "org",
												},
												Computed: &base.ComputedUserSet{
													Relation: "admin",
												},
											},
										},
									},
								},
							},
						},
						"write": {
							Name: "write",
							Child: &base.Child{
								Type: &base.Child_Leaf{
									Leaf: &base.Leaf{
										Type: &base.Leaf_TupleToUserSet{
											TupleToUserSet: &base.TupleToUserSet{
												TupleSet: &base.TupleSet{
													Relation: "org",
												},
												Computed: &base.ComputedUserSet{
													Relation: "admin",
												},
											},
										},
									},
								},
							},
						},
					},
					References: map[string]base.EntityDefinition_RelationalReference{
						"org":   base.EntityDefinition_RELATIONAL_REFERENCE_RELATION,
						"read":  base.EntityDefinition_RELATIONAL_REFERENCE_PERMISSION,
						"write": base.EntityDefinition_RELATIONAL_REFERENCE_PERMISSION,
					},
				},
				{
					Name: "organization",
					Relations: map[string]*base.RelationDefinition{
						"admin": {
							Name: "admin",
							RelationReferences: []*base.RelationReference{
								{
									Type:     "user",
									Relation: "",
								},
							},
						},
					},
					Permissions: map[string]*base.PermissionDefinition{},
					References: map[string]base.EntityDefinition_RelationalReference{
						"admin": base.EntityDefinition_RELATIONAL_REFERENCE_RELATION,
					},
				},
				{
					Name: "division",
					Relations: map[string]*base.RelationDefinition{
						"manager": {
							Name: "manager",
							RelationReferences: []*base.RelationReference{
								{
									Type:     "user",
									Relation: "",
								},
								{
									Type:     "organization",
									Relation: "admin",
								},
							},
						},
					},
					Permissions: map[string]*base.PermissionDefinition{
						"read": {
							Name: "read",
							Child: &base.Child{
								Type: &base.Child_Leaf{
									Leaf: &base.Leaf{
										Type: &base.Leaf_ComputedUserSet{
											ComputedUserSet: &base.ComputedUserSet{
												Relation: "manager",
											},
										},
									},
								},
							},
						},
						"write": {
							Name: "write",
							Child: &base.Child{
								Type: &base.Child_Leaf{
									Leaf: &base.Leaf{
										Type: &base.Leaf_ComputedUserSet{
											ComputedUserSet: &base.ComputedUserSet{
												Relation: "manager",
											},
										},
									},
								},
							},
						},
					},
					References: map[string]base.EntityDefinition_RelationalReference{
						"manager": base.EntityDefinition_RELATIONAL_REFERENCE_RELATION,
						"read":    base.EntityDefinition_RELATIONAL_REFERENCE_PERMISSION,
						"write":   base.EntityDefinition_RELATIONAL_REFERENCE_PERMISSION,
					},
				},
			}

			Expect(is).Should(Equal(i))
		})

		It("Case 12", func() {
			sch, err := parser.NewParser(`
		
			entity usertype {}
		
			entity company {
				relation admin @usertype
			}
		
			entity organization {
				relation admin @usertype
			}
		
			entity department {
				relation parent @company @organization
		
				permission read = parent.admin
		
			}
			`).Parse()

			Expect(err).ShouldNot(HaveOccurred())

			c := NewCompiler(true, sch)

			var is []*base.EntityDefinition
			is, err = c.Compile()

			Expect(err).ShouldNot(HaveOccurred())

			i := []*base.EntityDefinition{
				{
					Name:        "usertype",
					Relations:   map[string]*base.RelationDefinition{},
					Permissions: map[string]*base.PermissionDefinition{},
					References:  map[string]base.EntityDefinition_RelationalReference{},
				},
				{
					Name: "company",
					Relations: map[string]*base.RelationDefinition{
						"admin": {
							Name: "admin",
							RelationReferences: []*base.RelationReference{
								{
									Type:     "usertype",
									Relation: "",
								},
							},
						},
					},
					Permissions: map[string]*base.PermissionDefinition{},
					References: map[string]base.EntityDefinition_RelationalReference{
						"admin": base.EntityDefinition_RELATIONAL_REFERENCE_RELATION,
					},
				},
				{
					Name: "organization",
					Relations: map[string]*base.RelationDefinition{
						"admin": {
							Name: "admin",
							RelationReferences: []*base.RelationReference{
								{
									Type:     "usertype",
									Relation: "",
								},
							},
						},
					},
					Permissions: map[string]*base.PermissionDefinition{},
					References: map[string]base.EntityDefinition_RelationalReference{
						"admin": base.EntityDefinition_RELATIONAL_REFERENCE_RELATION,
					},
				},
				{
					Name: "department",
					Relations: map[string]*base.RelationDefinition{
						"parent": {
							Name: "parent",
							RelationReferences: []*base.RelationReference{
								{
									Type:     "company",
									Relation: "",
								},
								{
									Type:     "organization",
									Relation: "",
								},
							},
						},
					},
					Permissions: map[string]*base.PermissionDefinition{
						"read": {
							Name: "read",
							Child: &base.Child{
								Type: &base.Child_Leaf{
									Leaf: &base.Leaf{
										Type: &base.Leaf_TupleToUserSet{
											TupleToUserSet: &base.TupleToUserSet{
												TupleSet: &base.TupleSet{
													Relation: "parent",
												},
												Computed: &base.ComputedUserSet{
													Relation: "admin",
												},
											},
										},
									},
								},
							},
						},
					},
					References: map[string]base.EntityDefinition_RelationalReference{
						"parent": base.EntityDefinition_RELATIONAL_REFERENCE_RELATION,
						"read":   base.EntityDefinition_RELATIONAL_REFERENCE_PERMISSION,
					},
				},
			}

			Expect(is).Should(Equal(i))
		})

		It("Case 13", func() {
			sch, err := parser.NewParser(`
			entity usertype {}

			entity company {
    			relation owner @usertype
			}

			entity organization {
    			relation parent @company @organization

				relation owner @usertype
			}

			entity repository {

    			relation parent @organization#parent
    			relation owner  @usertype

    			permission edit  = parent.owner or owner
    			permission delete  = edit
			} 
			`).Parse()

			Expect(err).ShouldNot(HaveOccurred())

			c := NewCompiler(true, sch)

			var is []*base.EntityDefinition
			is, err = c.Compile()

			Expect(err).ShouldNot(HaveOccurred())

			i := []*base.EntityDefinition{
				{
					Name:        "usertype",
					Relations:   map[string]*base.RelationDefinition{},
					Permissions: map[string]*base.PermissionDefinition{},
					References:  map[string]base.EntityDefinition_RelationalReference{},
				},
				{
					Name: "company",
					Relations: map[string]*base.RelationDefinition{
						"owner": {
							Name: "owner",
							RelationReferences: []*base.RelationReference{
								{
									Type:     "usertype",
									Relation: "",
								},
							},
						},
					},
					Permissions: map[string]*base.PermissionDefinition{},
					References: map[string]base.EntityDefinition_RelationalReference{
						"owner": base.EntityDefinition_RELATIONAL_REFERENCE_RELATION,
					},
				},
				{
					Name: "organization",
					Relations: map[string]*base.RelationDefinition{
						"parent": {
							Name: "parent",
							RelationReferences: []*base.RelationReference{
								{
									Type:     "company",
									Relation: "",
								},
								{
									Type:     "organization",
									Relation: "",
								},
							},
						},
						"owner": {
							Name: "owner",
							RelationReferences: []*base.RelationReference{
								{
									Type:     "usertype",
									Relation: "",
								},
							},
						},
					},
					Permissions: map[string]*base.PermissionDefinition{},
					References: map[string]base.EntityDefinition_RelationalReference{
						"parent": base.EntityDefinition_RELATIONAL_REFERENCE_RELATION,
						"owner":  base.EntityDefinition_RELATIONAL_REFERENCE_RELATION,
					},
				},
				{
					Name: "repository",
					Relations: map[string]*base.RelationDefinition{
						"parent": {
							Name: "parent",
							RelationReferences: []*base.RelationReference{
								{
									Type:     "organization",
									Relation: "parent",
								},
							},
						},
						"owner": {
							Name: "owner",
							RelationReferences: []*base.RelationReference{
								{
									Type:     "usertype",
									Relation: "",
								},
							},
						},
					},
					Permissions: map[string]*base.PermissionDefinition{
						"edit": {
							Name: "edit",
							Child: &base.Child{
								Type: &base.Child_Rewrite{
									Rewrite: &base.Rewrite{
										RewriteOperation: base.Rewrite_OPERATION_UNION,
										Children: []*base.Child{
											{
												Type: &base.Child_Leaf{
													Leaf: &base.Leaf{
														Type: &base.Leaf_TupleToUserSet{
															TupleToUserSet: &base.TupleToUserSet{
																TupleSet: &base.TupleSet{
																	Relation: "parent",
																},
																Computed: &base.ComputedUserSet{
																	Relation: "owner",
																},
															},
														},
													},
												},
											},
											{
												Type: &base.Child_Leaf{
													Leaf: &base.Leaf{
														Type: &base.Leaf_ComputedUserSet{
															ComputedUserSet: &base.ComputedUserSet{
																Relation: "owner",
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
						"delete": {
							Name: "delete",
							Child: &base.Child{
								Type: &base.Child_Leaf{
									Leaf: &base.Leaf{
										Type: &base.Leaf_ComputedUserSet{
											ComputedUserSet: &base.ComputedUserSet{
												Relation: "edit",
											},
										},
									},
								},
							},
						},
					},
					References: map[string]base.EntityDefinition_RelationalReference{
						"parent": base.EntityDefinition_RELATIONAL_REFERENCE_RELATION,
						"owner":  base.EntityDefinition_RELATIONAL_REFERENCE_RELATION,
						"edit":   base.EntityDefinition_RELATIONAL_REFERENCE_PERMISSION,
						"delete": base.EntityDefinition_RELATIONAL_REFERENCE_PERMISSION,
					},
				},
			}

			Expect(is).Should(Equal(i))
		})

		It("Case 14", func() {
			sch, err := parser.NewParser(`
			entity usertype {}

			entity company {}

			entity organization {
    			relation parent @company
			}

			entity repository {

    			relation parent @organization#parent
    			relation owner  @usertype

    			permission edit   = parent.owner or owner
			} 
			`).Parse()

			Expect(err).ShouldNot(HaveOccurred())

			c := NewCompiler(true, sch)

			_, err = c.Compile()
			Expect(err.Error()).Should(Equal("15:36: undefined relation reference"))
		})
	})
})
