package coverage

import (
	"testing"

	"github.com/Permify/permify/pkg/development/file"
)

func TestCoverage(t *testing.T) {
	shp := file.Shape{
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
	}

	Run(shp)
}
