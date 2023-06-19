package coverage

import (
	"sort"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Permify/permify/pkg/development/file"
)

// TestCoverage -
func TestCoverage(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "coverage-suite")
}

var _ = Describe("coverage", func() {
	Context("Run", func() {
		It("Case 1: Github Simplified", func() {
			sci := Run(file.Shape{
				Schema: `
		entity user {}
		
		entity organization {
		   // organizational roles
		   relation admin @user
		   relation member @user
		}
		
		entity repository {
		   // represents repositories parent organization
		   relation parent @organization
		
		   // represents owner of this repository
		   relation owner  @user @organization#admin
		
		   // permissions
		   permission edit   = parent.admin or owner
		   permission delete = owner
		}`,
				Relationships: []string{
					"organization:1#admin@user:1",
					"repository:1#parent@organization:1",
				},
				Scenarios: []file.Scenario{
					{
						Name:        "scenario 1",
						Description: "test description",
						Checks: []file.Check{
							{
								Entity:  "repository:1",
								Subject: "user:1",
								Assertions: map[string]bool{
									"edit":   true,
									"delete": false,
								},
							},
						},
						EntityFilters: []file.EntityFilter{
							{
								EntityType: "repository",
								Subject:    "user:1",
								Assertions: map[string][]string{
									"edit": {"1", "5"},
								},
							},
						},
					},
					{
						Name:        "scenario 2",
						Description: "test description",
						Checks: []file.Check{
							{
								Entity:  "repository:1",
								Subject: "user:1",
								Assertions: map[string]bool{
									"edit": true,
								},
							},
						},
						EntityFilters: []file.EntityFilter{},
					},
				},
			})

			Expect(sci.EntityCoverageInfo[0].EntityName).Should(Equal("user"))
			Expect(sci.EntityCoverageInfo[0].UncoveredRelationships).Should(Equal([]string{}))
			Expect(sci.EntityCoverageInfo[0].CoverageRelationshipsPercent).Should(Equal(100))
			Expect(len(sci.EntityCoverageInfo[0].UncoveredAssertions)).Should(Equal(0))
			Expect(sci.EntityCoverageInfo[0].CoverageAssertionsPercent["scenario 2"]).Should(Equal(100))
			Expect(sci.EntityCoverageInfo[0].CoverageAssertionsPercent["scenario 1"]).Should(Equal(100))

			Expect(sci.EntityCoverageInfo[1].EntityName).Should(Equal("organization"))
			Expect(sci.EntityCoverageInfo[1].UncoveredRelationships).Should(Equal([]string{
				"organization#member@user",
			}))
			Expect(sci.EntityCoverageInfo[1].CoverageRelationshipsPercent).Should(Equal(50))
			Expect(len(sci.EntityCoverageInfo[1].UncoveredAssertions)).Should(Equal(0))
			Expect(sci.EntityCoverageInfo[1].CoverageAssertionsPercent["scenario 2"]).Should(Equal(100))
			Expect(sci.EntityCoverageInfo[1].CoverageAssertionsPercent["scenario 1"]).Should(Equal(100))

			Expect(sci.EntityCoverageInfo[2].EntityName).Should(Equal("repository"))
			Expect(isSameArray(sci.EntityCoverageInfo[2].UncoveredRelationships, []string{
				"repository#owner@user",
				"repository#owner@organization#admin",
			})).Should(Equal(true))
			Expect(sci.EntityCoverageInfo[2].CoverageRelationshipsPercent).Should(Equal(33))
			Expect(sci.EntityCoverageInfo[2].UncoveredAssertions).Should(Equal(map[string][]string{
				"scenario 2": {
					"repository#delete",
				},
			}))
			Expect(sci.EntityCoverageInfo[2].CoverageAssertionsPercent["scenario 2"]).Should(Equal(50))
			Expect(sci.EntityCoverageInfo[2].CoverageAssertionsPercent["scenario 1"]).Should(Equal(100))
		})

		It("Case 2: Google Docs Simplified", func() {
			sci := Run(file.Shape{
				Schema: `
		entity user {}
		
		entity resource {
		   relation viewer  @user  @group#member @group#manager
		   relation manager @user @group#member @group#manager
		
		   action edit = manager
		   action view = viewer or manager
		}
		
		entity group {
		   relation manager @user @group#member @group#manager
		   relation member @user @group#member @group#manager
		}
		
		entity organization {
		   relation group @group
		   relation resource @resource
		
		   relation administrator @user @group#member @group#manager
		   relation direct_member @user
		
		   permission admin = administrator
		   permission member = direct_member or administrator or group.member
		}`,
				Relationships: []string{
					"group:tech#manager@user:ashley",
					"group:tech#member@user:david",
					"group:marketing#manager@user:john",
					"group:marketing#member@user:jenny",
					"group:hr#manager@user:josh",
					"group:hr#member@user:joe",
					"group:tech#member@group:marketing#member",
					"group:tech#member@group:hr#member",
					"organization:acme#group@group:tech",
					"organization:acme#group@group:marketing",
					"organization:acme#group@group:hr",
					"organization:acme#resource@resource:product_database",
					"organization:acme#resource@resource:marketing_materials",
					"organization:acme#resource@resource:hr_documents",
					"organization:acme#administrator@group:tech#manager",
					"organization:acme#administrator@user:jenny",
					"resource:product_database#manager@group:tech#manager",
					"resource:product_database#viewer@group:tech#member",
					"resource:marketing_materials#viewer@group:marketing#member",
					"resource:hr_documents#manager@group:hr#manager",
					"resource:hr_documents#viewer@group:hr#member",
				},
				Scenarios: []file.Scenario{
					{
						Name:        "scenario 1",
						Description: "test description",
						Checks: []file.Check{
							{
								Entity:  "resource:product_database",
								Subject: "user:ashley",
								Assertions: map[string]bool{
									"edit": true,
								},
							},
							{
								Entity:  "resource:hr_documents",
								Subject: "user:joe",
								Assertions: map[string]bool{
									"view": true,
								},
							},
							{
								Entity:  "resource:marketing_materials",
								Subject: "user:david",
								Assertions: map[string]bool{
									"view": false,
								},
							},
						},
						EntityFilters: []file.EntityFilter{
							{
								EntityType: "resource",
								Subject:    "user:ashley",
								Assertions: map[string][]string{
									"edit": {"product_database"},
								},
							},
						},
					},
				},
			})

			Expect(sci.EntityCoverageInfo[0].EntityName).Should(Equal("user"))
			Expect(sci.EntityCoverageInfo[0].UncoveredRelationships).Should(Equal([]string{}))
			Expect(sci.EntityCoverageInfo[0].CoverageRelationshipsPercent).Should(Equal(100))
			Expect(len(sci.EntityCoverageInfo[0].UncoveredAssertions)).Should(Equal(0))
			Expect(sci.EntityCoverageInfo[0].CoverageAssertionsPercent["scenario 1"]).Should(Equal(100))

			Expect(sci.EntityCoverageInfo[1].EntityName).Should(Equal("resource"))
			Expect(isSameArray(sci.EntityCoverageInfo[1].UncoveredRelationships, []string{
				"resource#manager@user",
				"resource#manager@group#member",
				"resource#viewer@user",
				"resource#viewer@group#manager",
			})).Should(Equal(true))
			Expect(sci.EntityCoverageInfo[1].CoverageRelationshipsPercent).Should(Equal(33))
			Expect(len(sci.EntityCoverageInfo[1].UncoveredAssertions)).Should(Equal(0))
			Expect(sci.EntityCoverageInfo[1].CoverageAssertionsPercent["scenario 1"]).Should(Equal(100))

			Expect(sci.EntityCoverageInfo[2].EntityName).Should(Equal("group"))
			Expect(isSameArray(sci.EntityCoverageInfo[2].UncoveredRelationships, []string{
				"group#manager@group#member",
				"group#manager@group#manager",
				"group#member@group#manager",
			})).Should(Equal(true))
			Expect(sci.EntityCoverageInfo[2].CoverageRelationshipsPercent).Should(Equal(50))
			Expect(len(sci.EntityCoverageInfo[2].UncoveredAssertions)).Should(Equal(0))
			Expect(sci.EntityCoverageInfo[2].CoverageAssertionsPercent["scenario 1"]).Should(Equal(100))

			Expect(sci.EntityCoverageInfo[3].EntityName).Should(Equal("organization"))
			Expect(isSameArray(sci.EntityCoverageInfo[3].UncoveredRelationships, []string{
				"organization#administrator@group#member",
				"organization#direct_member@user",
			})).Should(Equal(true))
			Expect(sci.EntityCoverageInfo[3].CoverageRelationshipsPercent).Should(Equal(66))
			Expect(isSameArray(sci.EntityCoverageInfo[3].UncoveredAssertions["scenario 1"], []string{
				"organization#admin",
				"organization#member",
			})).Should(Equal(true))
			Expect(sci.EntityCoverageInfo[3].CoverageAssertionsPercent["scenario 1"]).Should(Equal(0))
		})

		It("Case 3: Facebook Groups", func() {
			sci := Run(file.Shape{
				Schema: `
    entity user {}

    entity group {

        // Relation to represent the members of the group
        relation member @user
        // Relation to represent the admins of the group
        relation admin @user
        // Relation to represent the moderators of the group
        relation moderator @user

        // Permissions for the group entity
        action create = member
        action join = member
        action leave = member
        action invite_to_group = admin
        action remove_from_group = admin or moderator
        action edit_settings = admin or moderator
        action post_to_group = member
        action comment_on_post = member
        action view_group_insights = admin or moderator
    }

    entity post {

        // Relation to represent the owner of the post
        relation owner @user
        // Relation to represent the group that the post belongs to
        relation group @group

        // Permissions for the post entity
        action view_post = owner or group.member
        action edit_post = owner or group.admin
        action delete_post = owner or group.admin

        permission group_member = group.member
    }

    entity comment {

        // Relation to represent the owner of the comment
        relation owner @user

        // Relation to represent the post that the comment belongs to
        relation post @post

        // Permissions for the comment entity
        action view_comment = owner or post.group_member
        action edit_comment = owner
        action delete_comment = owner
    }

    entity like {

        // Relation to represent the owner of the like
        relation owner @user

        // Relation to represent the post that the like belongs to
        relation post @post

        // Permissions for the like entity
        action like_post = owner or post.group_member
        action unlike_post = owner or post.group_member
    }

    entity poll {

        // Relation to represent the owner of the poll
        relation owner @user

        // Relation to represent the group that the poll belongs to
        relation group @group

        // Permissions for the poll entity
        action create_poll = owner or group.admin
        action view_poll = owner or group.member
        action edit_poll = owner or group.admin
        action delete_poll = owner or group.admin
    }

    entity file {

        // Relation to represent the owner of the file
        relation owner @user

        // Relation to represent the group that the file belongs to
        relation group @group

        // Permissions for the file entity
        action upload_file = owner or group.member
        action view_file = owner or group.member
        action delete_file = owner or group.admin
    }

    entity event {

        // Relation to represent the owner of the event
        relation owner @user @tuser
        // Relation to represent the group that the event belongs to
        relation group @group

        // Permissions for the event entity
        action create_event = owner or group.admin
        action view_event = owner or group.member
        action edit_event = owner or group.admin
        action delete_event = owner or group.admin
        action RSVP_to_event = owner or group.member
    }

	entity tuser {}

	`,
				Relationships: []string{
					"group:1#member@user:1",
					"group:1#admin@user:2",
					"group:2#moderator@user:3",
					"group:2#member@user:4",
					"group:1#member@user:5",
					"post:1#owner@user:1",
					"post:1#group@group:1",
					"post:2#owner@user:4",
					"post:2#group@group:1",
					"comment:1#owner@user:2",
					"comment:1#post@post:1",
					"comment:2#owner@user:5",
					"comment:2#post@post:2",
					"like:1#owner@user:3",
					"like:1#post@post:1",
					"like:2#owner@user:4",
					"like:2#post@post:2",
					"poll:1#owner@user:2",
					"poll:1#group@group:1",
					"poll:2#owner@user:5",
					"poll:2#group@group:1",
					"file:1#owner@user:1",
					"file:1#group@group:1",
					"event:1#group@group:1",
				},
				Scenarios: []file.Scenario{
					{
						Name:        "scenario 1",
						Description: "test description",
						Checks: []file.Check{
							{
								Entity:  "event:1",
								Subject: "user:4",
								Assertions: map[string]bool{
									"RSVP_to_event": false,
								},
							},
							{
								Entity:  "comment:1",
								Subject: "user:5",
								Assertions: map[string]bool{
									"view_comment": true,
								},
							},
						},
						EntityFilters: []file.EntityFilter{},
					},
				},
			})

			Expect(sci.EntityCoverageInfo[0].EntityName).Should(Equal("user"))
			Expect(sci.EntityCoverageInfo[0].UncoveredRelationships).Should(Equal([]string{}))
			Expect(sci.EntityCoverageInfo[0].CoverageRelationshipsPercent).Should(Equal(100))
			Expect(len(sci.EntityCoverageInfo[0].UncoveredAssertions)).Should(Equal(0))
			Expect(sci.EntityCoverageInfo[0].CoverageAssertionsPercent["scenario 1"]).Should(Equal(100))

			Expect(sci.EntityCoverageInfo[1].EntityName).Should(Equal("group"))
			Expect(sci.EntityCoverageInfo[1].UncoveredRelationships).Should(Equal([]string{}))
			Expect(sci.EntityCoverageInfo[1].CoverageRelationshipsPercent).Should(Equal(100))
			Expect(isSameArray(sci.EntityCoverageInfo[1].UncoveredAssertions["scenario 1"], []string{
				"group#remove_from_group",
				"group#invite_to_group",
				"group#edit_settings",
				"group#post_to_group",
				"group#comment_on_post",
				"group#view_group_insights",
				"group#create",
				"group#join",
				"group#leave",
			})).Should(Equal(true))
			Expect(sci.EntityCoverageInfo[1].CoverageAssertionsPercent["scenario 1"]).Should(Equal(0))

			Expect(sci.EntityCoverageInfo[2].EntityName).Should(Equal("post"))
			Expect(sci.EntityCoverageInfo[2].UncoveredRelationships).Should(Equal([]string{}))
			Expect(sci.EntityCoverageInfo[2].CoverageRelationshipsPercent).Should(Equal(100))
			Expect(isSameArray(sci.EntityCoverageInfo[2].UncoveredAssertions["scenario 1"], []string{
				"post#view_post",
				"post#edit_post",
				"post#delete_post",
				"post#group_member",
			})).Should(Equal(true))
			Expect(sci.EntityCoverageInfo[2].CoverageAssertionsPercent["scenario 1"]).Should(Equal(0))

			Expect(sci.EntityCoverageInfo[3].EntityName).Should(Equal("comment"))
			Expect(sci.EntityCoverageInfo[3].UncoveredRelationships).Should(Equal([]string{}))
			Expect(sci.EntityCoverageInfo[3].CoverageRelationshipsPercent).Should(Equal(100))
			Expect(isSameArray(sci.EntityCoverageInfo[3].UncoveredAssertions["scenario 1"], []string{
				"comment#edit_comment",
				"comment#delete_comment",
			})).Should(Equal(true))
			Expect(sci.EntityCoverageInfo[3].CoverageAssertionsPercent["scenario 1"]).Should(Equal(33))

			Expect(sci.EntityCoverageInfo[4].EntityName).Should(Equal("like"))
			Expect(sci.EntityCoverageInfo[4].UncoveredRelationships).Should(Equal([]string{}))
			Expect(sci.EntityCoverageInfo[4].CoverageRelationshipsPercent).Should(Equal(100))
			Expect(isSameArray(sci.EntityCoverageInfo[4].UncoveredAssertions["scenario 1"], []string{
				"like#like_post",
				"like#unlike_post",
			})).Should(Equal(true))
			Expect(sci.EntityCoverageInfo[4].CoverageAssertionsPercent["scenario 1"]).Should(Equal(0))

			Expect(sci.EntityCoverageInfo[5].EntityName).Should(Equal("poll"))
			Expect(sci.EntityCoverageInfo[5].UncoveredRelationships).Should(Equal([]string{}))
			Expect(sci.EntityCoverageInfo[5].CoverageRelationshipsPercent).Should(Equal(100))
			Expect(isSameArray(sci.EntityCoverageInfo[5].UncoveredAssertions["scenario 1"], []string{
				"poll#delete_poll",
				"poll#create_poll",
				"poll#view_poll",
				"poll#edit_poll",
			})).Should(Equal(true))
			Expect(sci.EntityCoverageInfo[5].CoverageAssertionsPercent["scenario 1"]).Should(Equal(0))

			Expect(sci.EntityCoverageInfo[6].EntityName).Should(Equal("file"))
			Expect(sci.EntityCoverageInfo[6].UncoveredRelationships).Should(Equal([]string{}))
			Expect(sci.EntityCoverageInfo[6].CoverageRelationshipsPercent).Should(Equal(100))
			Expect(isSameArray(sci.EntityCoverageInfo[6].UncoveredAssertions["scenario 1"], []string{
				"file#upload_file",
				"file#view_file",
				"file#delete_file",
			})).Should(Equal(true))
			Expect(sci.EntityCoverageInfo[6].CoverageAssertionsPercent["scenario 1"]).Should(Equal(0))

			Expect(sci.EntityCoverageInfo[7].EntityName).Should(Equal("event"))
			Expect(sci.EntityCoverageInfo[7].UncoveredRelationships).Should(Equal([]string{
				"event#owner@user",
				"event#owner@tuser",
			}))
			Expect(sci.EntityCoverageInfo[7].CoverageRelationshipsPercent).Should(Equal(33))
			Expect(isSameArray(sci.EntityCoverageInfo[7].UncoveredAssertions["scenario 1"], []string{
				"event#create_event",
				"event#view_event",
				"event#edit_event",
				"event#delete_event",
			})).Should(Equal(true))
			Expect(sci.EntityCoverageInfo[7].CoverageAssertionsPercent["scenario 1"]).Should(Equal(20))

			Expect(sci.EntityCoverageInfo[8].EntityName).Should(Equal("tuser"))
			Expect(sci.EntityCoverageInfo[8].UncoveredRelationships).Should(Equal([]string{}))
			Expect(sci.EntityCoverageInfo[8].CoverageRelationshipsPercent).Should(Equal(100))
			Expect(isSameArray(sci.EntityCoverageInfo[8].UncoveredAssertions["scenario 1"], []string{})).Should(Equal(true))
			Expect(sci.EntityCoverageInfo[8].CoverageAssertionsPercent["scenario 1"]).Should(Equal(100))
		})
	})
})

// isSameArray - check if two arrays are the same
func isSameArray(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	sortedA := make([]string, len(a))
	copy(sortedA, a)
	sort.Strings(sortedA)

	sortedB := make([]string, len(b))
	copy(sortedB, b)
	sort.Strings(sortedB)

	for i := range sortedA {
		if sortedA[i] != sortedB[i] {
			return false
		}
	}

	return true
}
