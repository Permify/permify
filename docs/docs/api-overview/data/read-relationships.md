import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

# Read Relational Tuples

Read API allows for directly querying the stored graph data to display and filter stored relational tuples.

## Request

**Path:** POST /v1/tenants/{tenant_id/data/relationships/read

| Required | Argument | Type | Default | Description |
|----------|----------|---------|---------|-------------------------------------------------------------------------------------------|
| [x]   | tenant_id | string | - | identifier of the tenant, if you are not using multi-tenancy (have only one tenant) use pre-inserted tenant `t1` for this field.
| [ ]   | snap_token | string | - | the snap token to avoid stale cache, see more details on [Snap Tokens](../../reference/snap-tokens) |
| [x]   | entity | object | - | contains entity type and id of the entity. Example: repository:1”.
| [x]   | relation | string | - | relation of the given entity |
| [ ]   | subject | object | - | the user or user set. It containes type and id of the subject.  ||

<Tabs>
<TabItem value="go" label="Go">

```go
rr, err: = client.Data.Relationship.Read(context.Background(), & v1.Data.RelationshipReadRequest {
    TenantId: "t1",
    Metadata: &v1.Data.RelationshipReadRequestMetadata {
        SnapToken: ""
    },
    Filter: &v1.TupleFilter {
        Entity: &v1.EntityFilter {
        Type: "organization",
        Ids: []string {"1"} ,
    },
    Relation: "member",
    Subject: &v1.SubjectFilter {
        Type: "",
        Id: []string {""},
        Relation: ""
    }}
})
```

</TabItem>

<TabItem value="node" label="Node">

```javascript
client.data.relationship.read({
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
    relation: "member",
    subject: {
      type: "",
      ids: [],
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
curl --location --request POST 'localhost:3476/v1/tenants/{tenant_id}/data/relationships/read' \
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
    relation: "member",
    subject: {
      type: "",
      ids: [],
      relation: ""
    }
  }
}'
```
</TabItem>
</Tabs>

## Need any help ?

Our team is happy to help you get started with Permify. If you'd like to learn more about using Permify in your app or have any questions about this example, [schedule a call with one of our Permify engineer](https://meetings-eu1.hubspot.com/ege-aytin/call-with-an-expert).
