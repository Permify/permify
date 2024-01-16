import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

# Run Bundle [Beta]

The "Run Bundle" API provides a straightforward way to execute predefined bundles within your application's tenant
environment. By sending a POST request to this endpoint, you can activate specific functionalities or processes
encapsulated in a bundle.

## Request

```javascript
 POST /v1/tenants/{tenant_id}/data/run-bundle
```

[![View in Swagger](http://jessemillar.github.io/view-in-swagger-button/button.svg)](https://permify.github.io/permify-swagger/#/Data/bundle.run)

| Required | Argument | Type | Description |
|----------|----------|---------|---------|-------------------------------------------------------------------------------------------|
| [x]   | tenant_id | string | identifier of the tenant, if you are not using multi-tenancy (have only one tenant) use pre-inserted tenant `t1` for this field. |
| [x]   | name | string | unique name identifying the bundle. |
| [ ]   | arguments | map | parameters for the bundle in key-value format. |

<Tabs>
<TabItem value="go" label="Go">

```go
rr, err: = client.Data.RunBundle(context.Background(), &v1.BundleRunRequest{
    TenantId: "t1",
    Name:     "organization_created",
    Arguments: map[string]string{
        "creatorID":      "564",
        "organizationID": "789",
    },
})
```

</TabItem>

<TabItem value="node" label="Node">

```javascript
client.data.runBundle({
    tenantId: "t1",
    name: "organization_created",
    arguments: {
        creatorID: "564",
        organizationID: "789",
    }
}).then((response) => {
    // handle response
})
```

</TabItem>
<TabItem value="curl" label="cURL">

```curl
curl --location --request POST 'localhost:3476/v1/tenants/{tenant_id}/data/run-bundle' \
--header 'Content-Type: application/json' \
--data-raw '{
    "name": "organization_created",
    "arguments": {
        "creatorID": "564",
        "organizationID": "789",
    }
}'
```

</TabItem>
</Tabs>

## Need any help ?

Our team is happy to help you get started with Permify. If you'd like to learn more about using Permify in your app or have any questions about this example, [schedule a call with one of our Permify engineer](https://meetings-eu1.hubspot.com/ege-aytin/call-with-an-expert).
