/** @type {import('@docusaurus/plugin-content-docs/src/sidebars/types').Sidebars} */
module.exports = {
  someSidebar: [
		{
			type: "category",
			label: "First Glance",
			link: {
					type: "generated-index",
					title: "First Glance",
					slug: "/permify-overview",
			},
			items: [
				"permify-overview/intro",
				"permify-overview/authorization-service",
				"permify-overview/infrastructure"
			],
			collapsed: false,
		},
		{
			type: "category",
			label: "Getting Started",
			link: {
				type: "generated-index",
				title: "Getting Started",
				slug: "/getting-started",
			},
			items: [
				"getting-started/modeling",
				"getting-started/sync-data",
				"getting-started/enforcement",
				"getting-started/testing",
				{
					type: 'category',
					label: 'Real World Examples',
					link: {
						type: "generated-index",
						title: "Real World Examples",
						slug: "/getting-started/examples",
					},
					items: [
						"getting-started/examples/google-docs",
						"getting-started/examples/facebook-groups",
						"getting-started/examples/notion"
					],
				},
			],
			collapsed: false,
		},
		{
			type: "category",
			label: "Set Up Permify",
			link: {
				type: "doc",
				id: "installation",
			},
			items: [
				"installation/overview",
				"installation/brew",
				"installation/container",
				"installation/aws",
				"installation/azure",
				"installation/google",
				"installation/kubernetes",
			],
			collapsed: true,
		},
		{
			type: "category",
			label: "Using the API",
			link: {
				type: "doc",
                id: "api-overview",
			},
			items: [
				{
					type: 'category',
					label: 'Schema Service',
					link: {
						type: "generated-index",
						title: "Schema Service",
						slug: "/api-overview/schema",
					},
					items: [
						"api-overview/schema/write-schema"
					],
				  },
				  {
					type: 'category',
					label: 'Relationship Service',
					link: {
						type: "generated-index",
						title: "Relationship Service",
						slug: "/api-overview/relationship",
					},
					items: [
						"api-overview/relationship/write-relationships",
						"api-overview/relationship/read-api", 
						"api-overview/relationship/delete-relationships"
					],
				  },
				  {
					type: 'category',
					label: 'Permission Service',
					link: {
						type: "generated-index",
						title: "Permission Service",
						slug: "/api-overview/permission",
					},
					items: [
						"api-overview/permission/check-api",
						"api-overview/permission/lookup-entity",
						"api-overview/permission/lookup-subject",
						"api-overview/permission/expand-api",
						"api-overview/permission/subject-permission"
					],
				  },
				  {
					type: 'category',
					label: 'Tenancy Service',
					link: {
						type: "generated-index",
						title: "Tenancy Service",
						slug: "/api-overview/tenancy",
					},
					items: [
						"api-overview/tenancy/create-tenant",
						"api-overview/tenancy/delete-tenant",
					],
				  },
				{
					type: 'category',
					label: 'Watch Service',
					link: {
						type: "generated-index",
						title: "Watch Service",
						slug: "/api-overview/watch",
					},
					items: [
						"api-overview/watch/watch-changes",
					],
				},
			],
			collapsed: true
		},
		{
			type: "doc",
			id: "playground",
			label: "Permify Playground",
		},
		{
			type: "doc",
			id: "multi-tenancy",
			label: "Multi Tenancy",
		},
		{
			type: "category",
			label: "Common Use Cases",
			link: {
				type: "doc",
                id: "use-cases",
			},
			items: [
				"use-cases/simple-rbac",
				"use-cases/abac",
				"use-cases/organizational",
				"use-cases/ownership",
				"use-cases/custom-roles",
				"use-cases/nested-hierarchies",
				"use-cases/user-groups",
				"use-cases/sharing",
			],
			collapsed: true,
		},
		{
			type: "category",
			label: "Reference",
			link: {
				type: "generated-index",
				title: "Reference",
				slug: "/reference"
			},
			items: [
				"reference/glossary",
				"reference/configuration",
				"reference/contextual-tuples",
				"reference/snap-tokens",
				"reference/cache",
				"reference/tracing"
			],
			collapsed: true
		},
  ],
};
