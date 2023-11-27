import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

# Delete Bundle [Beta]

The "Delete Bundle" API is designed for removing specific data bundles within a multi-tenant application environment. This API facilitates the deletion of a bundle, identified by its unique name, from a designated tenant's environment.

## Request

**Path:** POST /v1/tenants/{tenant_id}/bundle/delete

[![View in Swagger](http://jessemillar.github.io/view-in-swagger-button/button.svg)](https://permify.github.io/permify-swagger/#/Bundle/bundle.delete)

| Required | Argument | Type | Description |
|----------|----------|---------|---------|-------------------------------------------------------------------------------------------|
| [x]   | tenant_id | string | identifier of the tenant, if you are not using multi-tenancy (have only one tenant) use pre-inserted tenant `t1` for this field. |
| [x]   | name | string | unique name identifying the bundle. |

<Tabs>
<TabItem value="go" label="Go">

```go
rr, err: = client.Bundle.Delete(context.Background(), &v1.BundleDeleteRequest{
    TenantId: "t1",
    Name:     "organization_created",
})
```

</TabItem>

<TabItem value="node" label="Node">

```javascript
client.bundle.delete({
    tenantId: "t1",
    name: "organization_created",
}).then((response) => {
    // handle response
})
```

</TabItem>
<TabItem value="curl" label="cURL">

```curl
curl --location --request POST 'localhost:3476/v1/tenants/{tenant_id}/bundle/delete' \
--header 'Content-Type: application/json' \
--data-raw '{
    "name": "organization_created",
}'
```

</TabItem>
</Tabs>

## Need any help ?

Our team is happy to help you get started with Permify. If you'd like to learn more about using Permify in your app or have any questions about this example, [schedule a call with one of our Permify engineer](https://meetings-eu1.hubspot.com/ege-aytin/call-with-an-expert).
