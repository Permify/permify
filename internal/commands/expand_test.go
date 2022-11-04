package commands

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Permify/permify/internal/repositories/mocks"
	"github.com/Permify/permify/pkg/database"
	"github.com/Permify/permify/pkg/dsl/compiler"
	"github.com/Permify/permify/pkg/dsl/schema"
	"github.com/Permify/permify/pkg/logger"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/tuple"
)

var _ = Describe("expand-command", func() {
	var expandCommand *ExpandCommand
	l := logger.New("debug")

	// DRIVE SAMPLE
	driveSchema := `
entity user {}

entity organization {
	relation admin @user
}

entity folder {
	relation org @organization
	relation creator @user
	relation collaborator @user

	action read = collaborator
	action update = collaborator
	action delete = creator or org.admin
}

entity doc {
	relation org @organization
	relation parent @folder
	relation owner @user
	
	action read = (owner or parent.collaborator) or org.admin
	action update = owner and org.admin
	action delete = owner or org.admin
}
`

	Context("Drive Sample: Expand", func() {
		It("Drive Sample: Case 1", func() {
			relationTupleRepository := new(mocks.RelationTupleRepository)

			getDocOwners := []*base.Tuple{
				{
					Entity: &base.Entity{
						Type: "doc",
						Id:   "1",
					},
					Relation: "owner",
					Subject: &base.Subject{
						Type:     tuple.USER,
						Id:       "2",
						Relation: "",
					},
				},
			}

			getDocParent := []*base.Tuple{
				{
					Entity: &base.Entity{
						Type: "doc",
						Id:   "1",
					},
					Relation: "parent",
					Subject: &base.Subject{
						Type:     "folder",
						Id:       "1",
						Relation: tuple.ELLIPSIS,
					},
				},
			}

			getParentCollaborators := []*base.Tuple{
				{
					Entity: &base.Entity{
						Type: "folder",
						Id:   "1",
					},
					Relation: "collaborator",
					Subject: &base.Subject{
						Type:     tuple.USER,
						Id:       "1",
						Relation: "",
					},
				},
				{
					Entity: &base.Entity{
						Type: "folder",
						Id:   "1",
					},
					Relation: "collaborator",
					Subject: &base.Subject{
						Type:     tuple.USER,
						Id:       "3",
						Relation: "",
					},
				},
			}

			getDocOrg := []*base.Tuple{
				{
					Entity: &base.Entity{
						Type: "doc",
						Id:   "1",
					},
					Relation: "org",
					Subject: &base.Subject{
						Type:     "organization",
						Id:       "1",
						Relation: tuple.ELLIPSIS,
					},
				},
			}

			getOrgAdmins := []*base.Tuple{
				{
					Entity: &base.Entity{
						Type: "organization",
						Id:   "1",
					},
					Relation: "admin",
					Subject: &base.Subject{
						Type:     tuple.USER,
						Id:       "1",
						Relation: "",
					},
				},
			}

			relationTupleRepository.On("QueryTuples", "doc", "1", "owner").Return(database.NewTupleCollection(getDocOwners...).CreateTupleIterator(), nil).Times(1)
			relationTupleRepository.On("QueryTuples", "doc", "1", "parent").Return(database.NewTupleCollection(getDocParent...).CreateTupleIterator(), nil).Times(1)
			relationTupleRepository.On("QueryTuples", "folder", "1", "collaborator").Return(database.NewTupleCollection(getParentCollaborators...).CreateTupleIterator(), nil).Times(1)
			relationTupleRepository.On("QueryTuples", "doc", "1", "org").Return(database.NewTupleCollection(getDocOrg...).CreateTupleIterator(), nil).Times(1)
			relationTupleRepository.On("QueryTuples", "organization", "1", "admin").Return(database.NewTupleCollection(getOrgAdmins...).CreateTupleIterator(), nil).Times(1)

			expandCommand = NewExpandCommand(relationTupleRepository, l)

			re := &ExpandQuery{
				Entity: &base.Entity{Type: "doc", Id: "1"},
			}

			sch, err := compiler.NewSchema(driveSchema)
			Expect(err).ShouldNot(HaveOccurred())

			en, err := schema.GetEntityByName(sch, re.Entity.Type)
			Expect(err).ShouldNot(HaveOccurred())

			ac, err := schema.GetActionByNameInEntityDefinition(en, "read")
			Expect(err).ShouldNot(HaveOccurred())

			actualResult, err := expandCommand.Execute(context.Background(), re, ac.Child)
			Expect(err).ShouldNot(HaveOccurred())

			Expect(&base.Expand{
				Node: &base.Expand_Expand{
					Expand: &base.ExpandTreeNode{
						Operation: base.ExpandTreeNode_UNION,
						Children: []*base.Expand{
							{
								Node: &base.Expand_Expand{
									Expand: &base.ExpandTreeNode{
										Operation: base.ExpandTreeNode_UNION,
										Children: []*base.Expand{
											{
												Target: &base.EntityAndRelation{
													Entity: &base.Entity{
														Type: "doc",
														Id:   "1",
													},
													Relation: "owner",
												}, Node: &base.Expand_Leaf{
													Leaf: &base.Subjects{
														Subjects: []*base.Subject{
															{
																Type: tuple.USER,
																Id:   "2",
															},
														},
													},
												},
											},
											{
												Node: &base.Expand_Expand{
													Expand: &base.ExpandTreeNode{
														Operation: base.ExpandTreeNode_UNION,
														Children: []*base.Expand{
															{
																Target: &base.EntityAndRelation{
																	Entity: &base.Entity{
																		Type: "doc",
																		Id:   "1",
																	},
																	Relation: "parent.collaborator",
																}, Node: &base.Expand_Leaf{
																	Leaf: &base.Subjects{
																		Subjects: []*base.Subject{
																			{
																				Type:     "folder",
																				Id:       "1",
																				Relation: "collaborator",
																			},
																		},
																	},
																},
															},
															{
																Target: &base.EntityAndRelation{
																	Entity: &base.Entity{
																		Type: "folder",
																		Id:   "1",
																	},
																	Relation: "collaborator",
																}, Node: &base.Expand_Leaf{
																	Leaf: &base.Subjects{
																		Subjects: []*base.Subject{
																			{
																				Type: tuple.USER,
																				Id:   "1",
																			},
																			{
																				Type: tuple.USER,
																				Id:   "3",
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
							}, {Node: &base.Expand_Expand{
								Expand: &base.ExpandTreeNode{
									Operation: base.ExpandTreeNode_UNION,
									Children: []*base.Expand{
										{
											Target: &base.EntityAndRelation{
												Entity: &base.Entity{
													Type: "doc",
													Id:   "1",
												},
												Relation: "org.admin",
											}, Node: &base.Expand_Leaf{
												Leaf: &base.Subjects{
													Subjects: []*base.Subject{
														{
															Type:     "organization",
															Id:       "1",
															Relation: "admin",
														},
													},
												},
											},
										},
										{
											Target: &base.EntityAndRelation{
												Entity: &base.Entity{
													Type: "organization",
													Id:   "1",
												},
												Relation: "admin",
											}, Node: &base.Expand_Leaf{
												Leaf: &base.Subjects{
													Subjects: []*base.Subject{
														{
															Type:     tuple.USER,
															Id:       "1",
															Relation: "",
														},
													},
												},
											},
										},
									},
								},
							}},
						},
					},
				},
			}).Should(Equal(actualResult.Tree))
		})
	})
})
