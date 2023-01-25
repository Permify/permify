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
						slug: "/api-overview//permission",
					},
					items: [
						"api-overview/permission/check-api",
						"api-overview/permission/lookup-entity",
						"api-overview/permission/expand-api",
						"api-overview/permission/schema-lookup"
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
		/* {
			type: "doc",
			id: "comparison",
			label: "Comparision",
		}, */
		{
			type: "category",
			label: "Common Use Cases",
			link: {
				type: "generated-index",
				title: "Common Use Cases",
				slug: "/example-use-cases"
			},
			items: [
				"example-use-cases/simple-rbac",
				"example-use-cases/organizational",
				"example-use-cases/ownership",
				"example-use-cases/parent-child",
				"example-use-cases/sharing",
				"example-use-cases/user-groups"
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
				"reference/snap-tokens",
				"reference/tracing"
			],
			collapsed: true
		},
  ],
};
