package coverage

import (
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
		It("Case 1", func() {
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
					"repository:1#parent@organization:1#...",
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
			Expect(sci.EntityCoverageInfo[2].UncoveredRelationships).Should(Equal([]string{
				"repository#owner@user",
				"repository#owner@organization#admin",
			}))
			Expect(sci.EntityCoverageInfo[2].CoverageRelationshipsPercent).Should(Equal(33))
			Expect(sci.EntityCoverageInfo[2].UncoveredAssertions).Should(Equal(map[string][]string{
				"scenario 2": {
					"repository#delete",
				},
			}))
			Expect(sci.EntityCoverageInfo[2].CoverageAssertionsPercent["scenario 2"]).Should(Equal(50))
			Expect(sci.EntityCoverageInfo[2].CoverageAssertionsPercent["scenario 1"]).Should(Equal(100))
		})
	})
})
