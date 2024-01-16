import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

# Read Attributes

Read API allows for directly querying the stored graph data to display and filter stored attributes.

## Request
```javascript
POST /v1/tenants/{tenant_id}/data/attributes/read
```

[![View in Swagger](http://jessemillar.github.io/view-in-swagger-button/button.svg)](https://permify.github.io/permify-swagger/#/Data/data.attributes.read)

| Required | Argument | Type | Description |
|----------|----------|---------|---------|-------------------------------------------------------------------------------------------|
| [x]   | tenant_id | string | identifier of the tenant, if you are not using multi-tenancy (have only one tenant) use pre-inserted tenant `t1` for this field.
| [ ]   | snap_token | string |  the snap token to avoid stale cache, see more details on [Snap Tokens](../../reference/snap-tokens) |
| [x]   | entity | object |  contains entity type and id of the entity. Example: repository:1”.
| [x]   | attributes | string array |  attributes of the given entity |


<Tabs>
<TabItem value="go" label="Go">

```go
rr, err: = client.Data.ReadAttributes(context.Background(), & v1.Data.AttributeReadRequest {
    TenantId: "t1",
    Metadata: &v1.Data.AttributeReadRequestMetadata {
        SnapToken: ""
    },
    Filter: &v1.AttributeFilter {
        Entity: &v1.EntityFilter {
        Type: "organization",
        Ids: []string {"1"} ,
    },
    Attributes: []string {"private"},
})
```

</TabItem>

<TabItem value="node" label="Node">

```javascript
client.data.readAttributes({
  tenantId: "t1",
  metadata: {
     snap_token: "",
  },
  filter: {
    entity: {
      type: "organization",
      ids: [
        "1"
      ]
    },
    attributes: [
        "private"
    ],
  }
}).then((response) => {
    // handle response
})
```

</TabItem>
<TabItem value="curl" label="cURL">

```curl
curl --location --request POST 'localhost:3476/v1/tenants/{tenant_id}/data/attributes/read' \
--header 'Content-Type: application/json' \
--data-raw '{
  metadata: {
     snap_token: "",
  },
  filter: {
    entity: {
      type: "organization",
      ids: [
        "1"
      ]
    },
    attributes: [
        "private"
      ],
  }
}'
```
</TabItem>
</Tabs>

## Need any help ?

Our team is happy to help you get started with Permify. If you'd like to learn more about using Permify in your app or have any questions about this example, [schedule a call with one of our Permify engineer](https://meetings-eu1.hubspot.com/ege-aytin/call-with-an-expert).
