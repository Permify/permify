import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

# Delete Tenant

You can delete a tenant with following API.

## Request
```javascript
DELETE /v1/tenants/{id}
```

[![View in Swagger](http://jessemillar.github.io/view-in-swagger-button/button.svg)](https://permify.github.io/permify-swagger/#/Tenancy/tenants.delete)

<Tabs>
<TabItem value="go" label="Go">

```go
rr, err: = client.Tenancy.Delete(context.Background(), & v1.TenantDeleteRequest {
    Id: ""
})
```

</TabItem>

<TabItem value="node" label="Node">

```javascript
client.tenancy.delete({
   id: "",
}).then((response) => {
    // handle response
})
```

</TabItem>
<TabItem value="curl" label="cURL">

```curl
curl --location --request DELETE 'http://localhost:3476/v1/tenants/t1'
```
</TabItem>
</Tabs>

## Need any help ?

Our team is happy to help you get started with Permify. If you'd like to learn more about using Permify in your app or have any questions about this example, [schedule a call with one of our Permify engineer](https://meetings-eu1.hubspot.com/ege-aytin/call-with-an-expert).