---
title: Entity Filtering
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

# Entity Filtering

Lookup Entity endpoint lets you ask questions in form of **“Which resources can user:X do action Y?”**. As a response of this you’ll get a entity results in a format of string array or as a streaming response depending on the endpoint you're using.

So, we provide 2 separate endpoints for data filtering check request,

- [/v1/permissions/lookup-entity](#lookup-entity)
- [/v1/permissions/lookup-entity-stream](#lookup-entity-streaming)

## Lookup Entity 

In this endpoint you'll get directly the IDs' of the entities that are authorized in an array.

**POST** /v1/permissions/lookup-entity

| Required | Argument          | Type   | Default | Description                                                                                                                                                                |
|----------|-------------------|--------|---------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| [x]      | tenant_id         | string | -       | identifier of the tenant, if you are not using multi-tenancy (have only one tenant) use pre-inserted tenant `t1` for this field.                                           |
| [ ]      | schema_version    | string | 8       | Version of the schema                                                                                                                                                      |
| [ ]      | snap_token        | string | -       | the snap token to avoid stale cache, see more details on [Snap Tokens](../../reference/snap-tokens)                                                                        |
| [x]      | entity_type       | object | -       | type of the  entity. Example: repository”.                                                                                                                                 |
| [x]      | permission        | string | -       | the action the user wants to perform on the resource                                                                                                                       |
| [x]      | subject           | object | -       | the user or user set who wants to take the action. It contains type and id of the subject.                                                                                 |
| [ ]      | contextual_tuples | object | -       | Contextual tuples are relations that can be dynamically added to permission request operations. See more details on [Contextual Tuples](../../reference/contextual-tuples) |

<Tabs>
<TabItem value="go" label="Go">

```go
cr, err: = client.Permission.LookupEntity(context.Background(), & v1.PermissionLookupEntityRequest {
    TenantId: "t1",
    Metadata: & v1.PermissionLookupEntityRequestMetadata {
        SnapToken: ""
        SchemaVersion: ""
        Depth: 20,
    },
    EntityType: "document",
    Permission: "edit",
    Subject: & v1.Subject {
        Type: "user",
        Id: "1",
    }
})
```

</TabItem>
<TabItem value="node" label="Node">

```javascript
client.permission.lookupEntity({
    tenantId: "t1",
    metadata: {
        snapToken: "",
        schemaVersion: "",
        depth: 20
    },
    entity_type: "document",
    permission: "edit",
    subject: {
        type: "user",
        id: "1"
    }
}).then((response) => {
    console.log(response.entity_ids)
})
```

</TabItem>
<TabItem value="curl" label="cURL">

```curl
curl --location --request POST 'localhost:3476/v1/tenants/{tenant_id}/permissions/lookup-entity' \
--header 'Content-Type: application/json' \
--data-raw '{
  "metadata":{
    "snap_token": "",
    "schema_version": "",
    "depth": 20
  },
  "entity_type": "document",
  "permission": "edit",
  "subject": {
    "type":"user",
    "id":"1"
  }
}'
```
</TabItem>
</Tabs>

### Lookup Entity (Streaming)

The difference between this endpoint from direct Lookup Entity is response of this entity gives the IDs' as stream. This could be useful if you have large data set that getting all of the authorized data can take long with direct lookup entity endpoint.

**POST** /v1/permissions/lookup-entity-stream

| Required | Argument          | Type   | Default | Description                                                                                                                                                                |
|----------|-------------------|--------|---------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| [ ]      | schema_version    | string | 8       | Version of the schema                                                                                                                                                      |
| [ ]      | snap_token        | string | -       | the snap token to avoid stale cache, see more details on [Snap Tokens](../../reference/snap-tokens)                                                                        |
| [x]      | entity_type       | object | -       | type of the  entity. Example: repository”.                                                                                                                                 |
| [x]      | permission        | string | -       | the action the user wants to perform on the resource                                                                                                                       |
| [x]      | subject           | object | -       | the user or user set who wants to take the action. It contains type and id of the subject.                                                                                 |
| [ ]      | contextual_tuples | object | -       | Contextual tuples are relations that can be dynamically added to permission request operations. See more details on [Contextual Tuples](../../reference/contextual-tuples) |

<Tabs>
<TabItem value="go" label="Go">

```go
str, err: = client.Permission.LookupEntityStream(context.Background(), &v1.PermissionLookupEntityRequest {
    Metadata: &v1.PermissionLookupEntityRequestMetadata {
        SnapToken: "", 
        SchemaVersion: "" 
        Depth: 50,
    },
    EntityType: "document",
    Permission: "view",
    Subject: &v1.Subject {
        Type: "user",
        Id: "1",
    },
})

// handle stream response
for {
    res, err: = str.Recv()

    if err == io.EOF {
        break
    }

    // res.EntityId
}
```

</TabItem>
<TabItem value="node" label="Node">

```javascript
const permify = require("@permify/permify-node");
const {PermissionLookupEntityStreamResponse} = require("@permify/permify-node/dist/src/grpc/generated/base/v1/service");

function main() {
    const client = new permify.grpc.newClient({
        endpoint: "localhost:3478",
    })

    let res = client.permission.lookupEntityStream({
        metadata: {
            snapToken: "",
            schemaVersion: "",
            depth: 20
        },
        entityType: "document",
        permission: "view",
        subject: {
            type: "user",
            id: "1"
        }
    })

    handle(res)
}

async function handle(res: AsyncIterable<PermissionLookupEntityStreamResponse>) {
    for await (const response of res) {
        // response.entityId
    }
}
```

</TabItem>
</Tabs>