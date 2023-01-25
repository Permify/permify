import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

# Check Access Control

In Permify, you can perform two different types access checks,

- **resource based** authorization checks, in form of `Can user U perform action Y in resource Z ?`
- **subject based** authorization checks, in form of `Which resources can user U edit ?`

In this section we'll look at the resource based check request of Permify. You can find subject based access checks in [Check Entities' Permissions] section.

[Check Entities' Permissions]: ./lookup-entity

## Request

**Path:** POST /v1/permissions/check

| Required | Argument | Type | Default | Description |
|----------|----------|---------|---------|-------------------------------------------------------------------------------------------|
| [x]   | tenant_id | string | - | identifier of the tenant, if you are not using multi-tenancy (have only one tenant) use pre-inserted tenant `t1` for this field.
| [ ]   | schema_version | string | 8 | Version of the schema |
| [ ]   | snap_token | string | - | the snap token to avoid stale cache, see more details on [Snap Tokens](/docs/reference/snap-tokens) |
| [x]   | entity | object | - | contains entity type and id of the entity. Example: repository:1”.
| [x]   | permission | string | - | the action the user wants to perform on the resource |
| [x]   | subject | object | - | the user or user set who wants to take the action. It contains type and id of the subject.  |
| [ ]   | depth | integer | 8 | Timeout limit when if recursive database queries got in loop|


<Tabs>
<TabItem value="go" label="Go">

```go
cr, err: = client.Permission.Check(context.Background(), & v1.PermissionCheckRequest {
    TenantId: "t1",
    Metadata: & v1.PermissionCheckRequestMetadata {
        SnapToken: ""
        SchemaVersion: ""
        Depth: 20,
    },
    Entity: & v1.Entity {
        Type: "repository",
        Id: "1",
    },
    Permission: "edit",
    Subject: & v1.Subject {
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
  "can": "RESULT_ALLOW",
  "remaining_depth": 0
}
```

Answering access checks is accomplished within Permify using a basic graph walking mechanism. See how [access decisions evaluated] in Permify.

[access decisions evaluated]: ../../getting-started/enforcement#how-access-decisions-evaluated

## Need any help ?

Our team is happy to help you get started with Permify. If you'd like to learn more about using Permify in your app or have any questions about this example, [schedule a call with one of our Permify engineer](https://meetings-eu1.hubspot.com/ege-aytin/call-with-an-expert).