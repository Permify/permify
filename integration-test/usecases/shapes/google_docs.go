package shapes

import (
	"github.com/Permify/permify/pkg/development/file"
)

// GOOGLE DOCS SAMPLE

var InitialGoogleDocsShape = file.Shape{
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
}
    `,
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
			Name:        "Scenario 1",
			Description: "Scenario Description",
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
			EntityFilters:  []file.EntityFilter{},
			SubjectFilters: []file.SubjectFilter{},
		},
	},
}
