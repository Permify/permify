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
| [x]   | action | string | - | the action the user wants to perform on the resource |
| [x]   | subject | object | - | the user or user set who wants to take the action. It containes type and id of the subject.  |
| [ ]   | depth | integer | 8 | Timeout limit when if recursive database queries got in loop|

#### Request

```json
{
  "schema_version": "",
  "snap_token": "",
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
  "depth": 0
}
```

#### Response

```json
{
  "can": "RESULT_ALLOW",
  "remaining_depth": 0
}
```

Answering access checks is accomplished within Permify using a basic graph walking mechanism. See how [access decisions evaluated] in Permify. 

[access decisions evaluated]: ../../docs/getting-started/enforcement#how-access-decisions-are-evaluated

## Need any help ?

Our team is happy to help you get started with Permify. If you'd like to learn more about using Permify in your app or have any questions about this example, [schedule a call with one of our Permify engineer](https://meetings-eu1.hubspot.com/ege-aytin/call-with-an-expert).