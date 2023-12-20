import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

# Delete Data

You can delete any stored relation tuples or attributes with following API

## Request

**Path:** POST /v1/tenants/{tenant_id}/data/delete

[![View in Swagger](http://jessemillar.github.io/view-in-swagger-button/button.svg)](https://permify.github.io/permify-swagger/#/Data/data.delete)

| Required | Argument | Type | Description |
|----------|----------|---------|---------|-------------------------------------------------------------------------------------------|
| [x]   | tenant_id | string | identifier of the tenant, if you are not using multi-tenancy (have only one tenant) use pre-inserted tenant `t1` for this field.
| [x]   | tuples_filter | object |filter to delete relational tuples. Contains **entity**, **relation** and **subject**.
| [x]   | attribute_filter | object | filter to delete attributes. Contains **entity** and **attributes**.
| [x]   | entity | object | contains entity type and id of the entity. Example: repository:1”.
| [x]   | relation | string | relation of the given entity |
| [x]   | attribute | string array | attributes to be deleted |
| [ ]   | subject | object | the user or user set. It contains type and id of the subject.  ||

<Tabs>
<TabItem value="go" label="Go">

```go
rr, err: = client.Data.Delete(context.Background(), & v1.DataDeleteRequest {
    TenantId: "t1",
    Metadata: &v1.DataDeleteRequestMetadata {
        SnapToken: ""
    },
    TupleFilter: &v1.TupleFilter {
        Entity: &v1.EntityFilter {
        Type: "organization",
        Ids: []string {"1"} ,
    },
    Relation: "admin",
    Subject: &v1.SubjectFilter {
        Type: "user",
        Id: []string {"1"},
        Relation: ""
    }}
})
```

</TabItem>

<TabItem value="node" label="Node">

```javascript
client.data.delete({
  tenantId: "t1",
  metadata: {
     snap_token: "",
  },
  tupleFilter: {
    entity: {
      type: "organization",
      ids: [
        "1"
      ]
    },
    relation: "admin",
    subject: {
      type: "user",
      ids: [
        "1"
      ],
      relation: ""
    }
  }
}).then((response) => {
    // handle response
})
```

</TabItem>
<TabItem value="curl" label="cURL">

```curl
curl --location --request POST 'localhost:3476/v1/tenants/{tenant_id}/data/delete' \
--header 'Content-Type: application/json' \
--data-raw '{
  "tupleFilter": {
    "entity": {
      "type": "organization",
      "ids": [
        "1"
      ]
    },
    "relation": "admin",
    "subject": {
      "type": "user",
      "ids": [
        "1"
      ],
      "relation": ""
    }
  },
}'
```
</TabItem>
</Tabs>

## Need any help ?

Our team is happy to help you get started with Permify. If you'd like to learn more about using Permify in your app or have any questions about this example, [schedule a call with one of our Permify engineer](https://meetings-eu1.hubspot.com/ege-aytin/call-with-an-expert).