---
title: Subject Permission List
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

# Subject Permission List

The Subject Permission List endpoint allows you to inquire in the form of **“Which permissions user:x can perform on entity:y?”**. In response, you'll receive a list of permissions specific to the user for the given entity, returned in the format of a map.

So, we provide 1 endpoint for subject permission list,

- [/v1/permissions/subject-permission](#subject-permission)

In this endpoint, you'll receive a map of permissions and their statuses directly. The structure is map[string]CheckResult, such as "sample-permission" -> "ALLOWED". This represents the permissions and their associated states in a key-value pair format.

## Request

**Path:** 
```javascript
POST /v1/permissions/subject-permission
```

[![View in Swagger](http://jessemillar.github.io/view-in-swagger-button/button.svg)](https://permify.github.io/permify-swagger/#/Permission/permissions.subjectPermission)

| Required | Argument          | Type    | Default | Description                                                                                                                                                                                         |
|----------|-------------------|---------|---------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| [x]      | tenant_id         | string  | -       | identifier of the tenant, if you are not using multi-tenancy (have only one tenant) use pre-inserted tenant `t1` for this field.                                                                    |
| [ ]      | schema_version    | string  | 8       | Version of the schema                                                                                                                                                                               |
| [ ]      | snap_token        | string  | -       | the snap token to avoid stale cache, see more details on [Snap Tokens](../../reference/snap-tokens.md).                                                                                                |
| [x]      | entity            | object  | -       | contains entity type and id of the entity. Example: repository:1.                                                                                                                                   |
| [x]      | subject           | object  | -       | the user or user set who wants to take the action. It contains type and id of the subject.                                                                                                          |
| [x]      | depth             | integer | 8       | Timeout limit when if recursive database queries got in loop                                                                                                                                        |
| [ ]      | only_permission   | bool    | false   | By default, the endpoint returns both permissions and relations associated with the user and entity. However, when the "only_permission" parameter is set to true, it returns only the permissions. |                                                                                                               |
| [ ]      | context | object  | -       | Contextual data that can be dynamically added to permission check requests. See details on [Contextual Data](../../reference/contextual-tuples.md)                        |

<Tabs>
<TabItem value="go" label="Go">

```go
cr, err: = client.Permission.SubjectPermission(context.Background(), &v1.PermissionSubjectPermissionRequest {
    TenantId: "t1",
    Metadata: &v1.PermissionSubjectPermissionRequestMetadata {
        SnapToken: "",
        SchemaVersion: "",
		OnlyPermission: false,
        Depth: 20,
    },
    Entity: &v1.Entity {
        Type: "repository",
        Id: "1",
    },
    Subject: &v1.Subject {
        Type: "user",
        Id: "1",
    },
})
```

</TabItem>
<TabItem value="node" label="Node">

```javascript
client.permission.subjectPermission({
    tenantId: "t1", 
    metadata: {
        snapToken: "",
        schemaVersion: "",
        onlyPermission: true,
        depth: 20
    },
    entity: {
        type: "repository",
        id: "1"
    },
    subject: {
        type: "user",
        id: "1"
    }
}).then((response) => {
    console.log(response);
})
```

</TabItem>
<TabItem value="curl" label="cURL">

```curl
curl --location --request POST 'localhost:3476/v1/tenants/{tenant_id}/permissions/subject-permission' \
--header 'Content-Type: application/json' \
--data-raw '{
  "metadata":{
    "snap_token": "",
    "schema_version": "",
    "only_permission": true,
    "depth": 20
  },
  "entity": {
    "type": "repository",
    "id": "1"
  },
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
  "results": [
    {
      "key": "delete",
      "value": "RESULT_ALLOWED"
    },
    {
      "key": "edit",
      "value": "RESULT_ALLOWED"
    }
  ]
}
```

## Need any help ?

Our team is happy to help you get started with Permify. If you'd like to learn more about using Permify in your app or have any questions about this example, [schedule a call with one of our Permify engineer](https://meetings-eu1.hubspot.com/ege-aytin/call-with-an-expert).

