import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

# Read Bundle

The "Read Bundle" API is a crucial tool for retrieving details of specific data bundles in a multi-tenant application setup. It is designed to access information about a bundle, uniquely identified by its name, within the specified tenant's environment.

## Request

**Path:** POST /v1/tenants/{tenant_id}/bundle/read

[![View in Swagger](http://jessemillar.github.io/view-in-swagger-button/button.svg)](https://permify.github.io/permify-swagger/#/Bundle/bundle.read)

| Required | Argument | Type | Description |
|----------|----------|---------|---------|-------------------------------------------------------------------------------------------|
| [x]   | tenant_id | string | identifier of the tenant, if you are not using multi-tenancy (have only one tenant) use pre-inserted tenant `t1` for this field. |
| [x]   | name | string | unique name identifying the bundle. |

<Tabs>
<TabItem value="go" label="Go">

```go
rr, err: = client.Bundle.Read(context.Background(), &v1.BundleReadRequest{
    TenantId: "t1",
    Name:     "organization_created",
})
```

</TabItem>

<TabItem value="node" label="Node">

```javascript
client.bundle.read({
    tenantId: "t1",
    name: "organization_created",
}).then((response) => {
    // handle response
})
```

</TabItem>
<TabItem value="curl" label="cURL">

```curl
curl --location --request POST 'localhost:3476/v1/tenants/{tenant_id}/bundle/read' \
--header 'Content-Type: application/json' \
--data-raw '{
    "name": "organization_created",
}'
```

</TabItem>
</Tabs>

## Need any help ?

Our team is happy to help you get started with Permify. If you'd like to learn more about using Permify in your app or have any questions about this example, [schedule a call with one of our Permify engineer](https://meetings-eu1.hubspot.com/ege-aytin/call-with-an-expert).
