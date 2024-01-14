import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

# Create Tenant

Permify Multi Tenancy support you can create custom schemas for tenants and manage them in a single place. You can create a tenant with following API.

:::caution 
We have a pre-inserted tenant - **t1** - by default for the ones that don't use multi-tenancy.  
:::

## Request

```javascript
POST /v1/tenants/create
```

[![View in Swagger](http://jessemillar.github.io/view-in-swagger-button/button.svg)](https://permify.github.io/permify-swagger/#/Tenancy/tenants.create)

<Tabs>
<TabItem value="go" label="Go">

```go
rr, err: = client.Tenancy.Create(context.Background(), & v1.TenantCreateRequest {
    Id: ""
    Name: ""
})
```

</TabItem>

<TabItem value="node" label="Node">

```javascript
client.tenancy.create({
   id: "",
   name: ""
}).then((response) => {
    // handle response
})
```

</TabItem>
<TabItem value="curl" label="cURL">

```curl
curl --location --request POST 'http://localhost:3476/v1/tenants/create' \
--header 'Content-Type: application/json' \
--data-raw '{
    "id": "",
    "name": ""
}'
```
</TabItem>
</Tabs>

## Need any help ?

Our team is happy to help you get started with Permify. If you'd like to learn more about using Permify in your app or have any questions about this example, [schedule a call with one of our Permify engineer](https://meetings-eu1.hubspot.com/ege-aytin/call-with-an-expert).