import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

# Check Access Control

In Permify, you can perform two different types access checks,

- **resource based** authorization checks, in form of `Can user U perform action Y in resource Z ?`
- **subject based** authorization checks, in form of `Which resources can user U edit ?`

In this section we'll look at the resource based check request of Permify. You can find subject based access checks in [Entity (Data) Filtering] section.

[Entity (Data) Filtering]: ../lookup-entity

## Request

**Path:** POST /v1/permissions/check

| Required | Argument          | Type    | Default | Description                                                                                                                                                                  |
|----------|-------------------|---------|---------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| [x]      | tenant_id         | string  | -       | identifier of the tenant, if you are not using multi-tenancy (have only one tenant) use pre-inserted tenant `t1` for this field.                                             |
| [ ]      | schema_version    | string  | 8       | Version of the schema                                                                                                                                                        |
| [ ]      | snap_token        | string  | -       | the snap token to avoid stale cache, see more details on [Snap Tokens](../../../reference/snap-tokens).                                                                         |
| [x]      | entity            | object  | -       | contains entity type and id of the entity. Example: repository:1.                                                                                                            |
| [x]      | permission        | string  | -       | the action the user wants to perform on the resource                                                                                                                         |
| [x]      | subject           | object  | -       | the user or user set who wants to take the action. It contains type and id of the subject.                                                                                   |
| [x]      | depth             | integer | 8       | Timeout limit when if recursive database queries got in loop                                                                                                                 |
| [ ]      | contextual_tuples | object  | -       | Contextual tuples are relations that can be dynamically added to permission request operations. , see more details on [Contextual Tuples](../../../reference/contextual-tuples) |

<Tabs>
<TabItem value="go" label="Go">

```go
cr, err: = client.Permission.Check(context.Background(), &v1.PermissionCheckRequest {
    TenantId: "t1",
    Metadata: &v1.PermissionCheckRequestMetadata {
        SnapToken: "",
        SchemaVersion: "",
        Depth: 20,
    },
    Entity: &v1.Entity {
        Type: "repository",
        Id: "1",
    },
    Permission: "edit",
    Subject: &v1.Subject {
        Type: "user",
        Id: "1",
    },

    if (cr.can === PermissionCheckResponse_Result.RESULT_ALLOWED) {
        // RESULT_ALLOWED
    } else {
        // RESULT_DENIED
    }
})
```

</TabItem>
<TabItem value="node" label="Node">

```javascript
client.permission.check({
    tenantId: "t1", 
    metadata: {
        snapToken: "",
        schemaVersion: "",
        depth: 20
    },
    entity: {
        type: "repository",
        id: "1"
    },
    permission: "edit",
    subject: {
        type: "user",
        id: "1"
    }
}).then((response) => {
    if (response.can === PermissionCheckResponse_Result.RESULT_ALLOWED) {
        console.log("RESULT_ALLOWED")
    } else {
        console.log("RESULT_DENIED")
    }
})
```

</TabItem>
<TabItem value="curl" label="cURL">

```curl
curl --location --request POST 'localhost:3476/v1/tenants/{tenant_id}/permissions/check' \
--header 'Content-Type: application/json' \
--data-raw '{
  "metadata":{
    "snap_token": "",
    "schema_version": "",
    "depth": 20
  },
  "entity": {
    "type": "repository",
    "id": "1"
  },
  "permission": "edit",
  "subject": {
    "type": "user",
    "id": "1",
    "relation": ""
  },
}'
```
</TabItem>
</Tabs>

## Response

```json
{
  "can": "RESULT_ALLOWED",
  "remaining_depth": 0
}
```

Answering access checks is accomplished within Permify using a basic graph walking mechanism. 

## How Access Decisions Evaluated?

Access decisions are evaluated by stored [relational tuples] and your authorization model, [Permify Schema]. 

In high level, access of an subject related with the relationships created between the subject and the resource. You can define this relationships in Permify Schema then create and store them as relational tuples, which is basically your authorization data. 

Permify Engine to compute access decision in 2 steps, 
1. Looking up authorization model for finding the given action's ( **edit**, **push**, **delete** etc.) relations.
2. Walk over a graph of each relation to find whether given subject ( user or user set ) is related with the action. 

Let's turn back to above authorization question ( ***"Can the user 3 edit document 12 ?"*** ) to better understand how decision evaluation works. 

[relational tuples]: ../../getting-started/sync-data.md
[Permify Schema]:  ../../getting-started/modeling.md

When Permify Engine receives this question it directly looks up to authorization model to find document `‍edit` action. Let's say we have a model as follows

```perm
entity user {}
        
entity organization {

    // organizational roles
    relation admin @user
    relation member @user
}

entity document {

    // represents documents parent organization
    relation parent @organization
    
    // represents owner of this document
    relation owner  @user
    
    // permissions
    action edit   = parent.admin or owner
    action delete = owner
} 
```

Which has a directed graph as follows:

![relational-tuples](https://github.com/Permify/permify/assets/39353278/cec9936c-f907-42c0-a419-032ebb45454e)

As we can see above: only users with an admin role in an organization, which `document:12` belongs, and owners of the `document:12` can edit. Permify runs two concurrent queries for **parent.admin** and **owner**:

**Q1:** Get the owners of the `document:12`.

**Q2:** Get admins of the organization where `document:12` belongs to.

Since edit action consist **or** between owner and parent.admin, if Permify Engine found user:3 in results of one of these queries then it terminates the other ongoing queries and returns authorized true to the client.

Rather than **or**, if we had an **and** relation then Permify Engine waits the results of these queries to returning a decision. 

## Latency & Performance

With the right architecture we expect **7-12 ms** latency. Depending on your load, cache usage and architecture you can get up to **30ms**.

Permify implements several cache mechanisms in order to achieve low latency in scaled distributed systems. See more on the section [Cache Mechanisims](../../reference/cache.md) 

## Need any help ?

:::info
Bulk permission check or with other name data filtering is a common use case we have seen so far. If you have a similar use case we would love to hear from you. Join our [discord](https://discord.gg/n6KfzYxhPp) to discuss or [schedule a call with one of our Permify engineers](https://meetings-eu1.hubspot.com/ege-aytin/call-with-an-expert).
:::

