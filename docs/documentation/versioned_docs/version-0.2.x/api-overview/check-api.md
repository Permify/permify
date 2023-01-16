import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

# Check Access Control

In Permify, you can perform two different types access checks,

- **resource based** authorization checks, in form of `Can user U perform action Y in resource Z ?`
- **data filtering (coming soon)** authorization checks , in form of `Which records can user U edit ?`

In this section we'll investigate proior check request of Permify: **resource based** authorization checks. You can find subject based access checks in [data filtering] section.

**Path:** POST /v1/permissions/check

| Required | Argument | Type | Default | Description |
|----------|----------|---------|---------|-------------------------------------------------------------------------------------------|
| [ ]   | schema_version | string | 8 | Version of the schema |
| [ ]   | snap_token | string | - | the snap token to avoid stale cache, see more details on [Snap Tokens](/docs/reference/snap-tokens) |
| [x]   | entity | object | - | contains entity type and id of the entity. Example: repository:1”.
| [x]   | permission | string | - | the action the user wants to perform on the resource |
| [x]   | subject | object | - | the user or user set who wants to take the action. It containes type and id of the subject.  |
| [ ]   | depth | integer | 8 | Timeout limit when if recursive database queries got in loop|

#### Request

```json
{
  "metadata": {
    "schema_version": "",
    "snap_token": "",
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
}
```

#### Response

```json
{
  "can": "RESULT_ALLOW",
  "remaining_depth": 0
}
```

### Using Clients

<Tabs>
<TabItem value="go" label="Go">

```go
cr, err: = client.Permission.Check(context.Background(), & v1.PermissionCheckRequest {
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
</Tabs>

Answering access checks is accomplished within Permify using a basic graph walking mechanism. See how [access decisions evaluated] in Permify. 

[access decisions evaluated]: ../../../docs/getting-started/enforcement#how-access-decisions-are-evaluated

## Need any help ?

Our team is happy to help you get started with Permify. If you'd like to learn more about using Permify in your app or have any questions about this example, [schedule a call with one of our Permify engineer](https://meetings-eu1.hubspot.com/ege-aytin/call-with-an-expert).