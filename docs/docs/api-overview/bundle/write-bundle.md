import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

# Write Bundle [Beta]

The "Write Bundle" API is designed for handling data in a multi-tenant application environment. Its primary function is to write and delete data according to predefined structures. This API allows users to define or update data bundles, each distinguished by a unique name.

## Request

**Path:** POST /v1/tenants/{tenant_id}/bundle/write

[![View in Swagger](http://jessemillar.github.io/view-in-swagger-button/button.svg)](https://permify.github.io/permify-swagger/#/Bundle/bundle.write)

| Required | Argument | Type | Description |
|----------|----------|---------|---------|-------------------------------------------------------------------------------------------|
| [x]   | tenant_id | string | identifier of the tenant, if you are not using multi-tenancy (have only one tenant) use pre-inserted tenant `t1` for this field. |
| [x]   | name | string | unique name identifying the bundle. |

<Tabs>
<TabItem value="go" label="Go">

```go
rr, err := client.Bundle.Write(context.Background(), &v1.BundleWriteRequest{
    TenantId: "t1",
    Bundles: []*v1.DataBundle{
        {
            Name: "organization_created",
            Arguments: []string{
                "creatorID",
                "organizationID",
            },
			Operations: []*v1.Operation{
			    {
                    RelationshipsWrite: []string{
                        "organization:{{.organizationID}}#admin@user:{{.creatorID}}",
                        "organization:{{.organizationID}}#manager@user:{{.creatorID}}",
                    },
                    AttributesWrite: []string{
                        "organization:{{.organizationID}}$public|boolean:false",
					},
				},
			},
		},
	},
})
```

</TabItem>

<TabItem value="node" label="Node">

```javascript
client.bundle.write({
    tenantId: "t1",
    bundles: [
        {
            name: "organization_created",
            arguments: [
                "creatorID",
                "organizationID",
            ],
            operations: [
                {
                    relationships_write: [
                        "organization:{{.organizationID}}#admin@user:{{.creatorID}}",
                        "organization:{{.organizationID}}#manager@user:{{.creatorID}}",
                    ],
                    attributes_write: [
                        "organization:{{.organizationID}}$public|boolean:false",
                    ]
                }
            ]
        }
    ]
}).then((response) => {
    // handle response
})
```

</TabItem>
<TabItem value="curl" label="cURL">

```curl
curl --location --request POST 'localhost:3476/v1/tenants/{tenant_id}/bundle/write' \
--header 'Content-Type: application/json' \
--data-raw '{
    "bundles": [
        {
            "name": "organization_created"
            "arguments": [
                "creatorID",
                "organizationID"
            ],
            "operations": [
                {
                    "relationships_write": [
                        "organization:{{.organizationID}}#admin@user:{{.creatorID}}",
                        "organization:{{.organizationID}}#manager@user:{{.creatorID}}",
                    ],
                    "attributes_write": [
                        "organization:{{.organizationID}}$public|boolean:false",
                    ],
                },
            ],
        },
    ],
}'
```

</TabItem>
</Tabs>

## Need any help ?

Our team is happy to help you get started with Permify. If you'd like to learn more about using Permify in your app or have any questions about this example, [schedule a call with one of our Permify engineer](https://meetings-eu1.hubspot.com/ege-aytin/call-with-an-expert).
