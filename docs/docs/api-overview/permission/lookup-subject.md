---
title: Subject Filtering
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

# Subject Filtering

Lookup Subject endpoint lets you ask questions in form of **“Which subjects can do action Y on entity:X?”**. As a response of this you’ll get a subject results in a format of string array.

So, we provide 1 endpoint for subject filtering request,

- [/v1/permissions/lookup-subject](#lookup-subject)

## Lookup Subject

In this endpoint you'll get directly the IDs' of the subjects that are authorized in an array.

**POST** /v1/permissions/lookup-subject

[![View in Swagger](http://jessemillar.github.io/view-in-swagger-button/button.svg)](https://permify.github.io/permify-swagger/#/Permission/permissions.lookupSubject)

| Required | Argument            | Type     | Default | Description                                                                                                                                                                |
|----------|---------------------|----------|---------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| [x]      | tenant_id           | string   | -       | identifier of the tenant, if you are not using multi-tenancy (have only one tenant) use pre-inserted tenant `t1` for this field.                                           |
| [ ]      | schema_version      | string   | -       | Version of the schema                                                                                                                                                      |
| [x]      | depth             | integer | 8       | Timeout limit when if recursive database queries got in loop                                                                                                                 |
| [ ]      | snap_token          | string   | -       | the snap token to avoid stale cache, see more details on [Snap Tokens](../../reference/snap-tokens.md).                                                                       |
| [x]      | entity              | object   | -       | contains entity type and id of the entity. Example: repository:1                                                                                                           |
| [x]      | permission          | string   | -       | the action the user wants to perform on the resource                                                                                                                       |
| [x]      | subject_reference   | object   | -       | the subject or subject reference who wants to take the action. It contains type and relation of the subject.                                                               |
| [ ]      | context   | object   | -       | Contextual data that can be dynamically added to permission check requests. See details on [Contextual Data](../../reference/contextual-tuples.md) |

<Tabs>
<TabItem value="go" label="Go">

```go
cr, err: = client.Permission.LookupSubject(context.Background(), &v1.PermissionLookupSubjectRequest {
    TenantId: "t1",
    Metadata: &v1.PermissionLookupSubjectRequestMetadata{
        SnapToken: "",
        SchemaVersion: "",
        Depth: 20,
    },
    Entity: &v1.Entity{
        Type: "document",
        Id: "1",
    },
    Permission: "edit",
    SubjectReference: &v1.RelationReference{
        Type: "user",
        Relation: "",
    }
})
```

</TabItem>
<TabItem value="node" label="Node">

```javascript
client.permission.lookupSubject({
    tenantId: "t1",
    metadata: {
        snapToken: "",
        schemaVersion: ""
        depth: 20,
    },
    Entity: {
        Type: "document",
        Id: "1",
    },
    permission: "edit",
    subject_reference: {
        type: "user",
        relation: ""
    }
}).then((response) => {
    console.log(response.subject_ids)
})
```

</TabItem>
<TabItem value="curl" label="cURL">

```curl
curl --location --request POST 'localhost:3476/v1/tenants/{tenant_id}/permissions/lookup-subject' \
--header 'Content-Type: application/json' \
--data-raw '{
  "metadata":{
    "snap_token": "",
    "schema_version": ""
    "depth": 20,
  },
  "entity": {
    type: "document",
    id: "1'
  },
  "permission": "edit",
  "subject_reference": {
    "type": "user",
    "relation": ""
  }
}'
```

</TabItem>
</Tabs>


## Need any help ?

Our team is happy to help you get started with Permify. If you'd like to learn more about using Permify in your app or have any questions about this example, [schedule a call with one of our Permify engineer](https://meetings-eu1.hubspot.com/ege-aytin/call-with-an-expert).
