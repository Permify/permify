# API Overview

Permify is an open-source authorization service that you can run with docker and works on a Rest API.

## Paths

### Check Access 

Returns a decision about whether user can perform an action on a certain resource. For example, can the user do push on
a repository object?
**Path:** POST /v1/permissions/check
| Required | Argument | Type | Default | Description |
|----------|----------|---------|---------|-------------------------------------------------------------------------------------------|
| [x]   | user | string | - | the user or user set who wants to take the action. Examples: “1”, “organization:1#owners”
| [x]   | action | string | - | the action the user wants to perform on the resource |
| [x]   | object | string | - | name and id of the resource. Example: “organization:1” |
| [ ]   | depth | integer | 8 | |

#### Request

```json
{
  "user": "1",
  "action": "push",
  "object": "repository:1"
}
```

#### Response

```json
{
  "can": true,
  "debug": "user 1 is a owner of organization 1"
}
```

### Create Custom Tuple

We examined how we created the tuple by listening to the tables. Permify allows to create custom tuples.
**Path:** POST /v1/relationships/write
| Required | Argument | Type | Default | Description |
|----------|-------------------|--------|---------|-------------|
| [x]   | entity | string | - | |
| [x]   | object_id | string | - | |
| [x]   | relation | string | - | |
| [ ]   | userset_entity | string | - | |
| [x]   | userset_object_id | string | - | |
| [ ]   | userset_relation | string | - | |

#### Request

```json
{
  "entity": "organization",
  "object_id": "1",
  "relation": "admin",
  "userset_entity": "",
  "userset_object_id": "1",
  "userset_relation": ""
}
```

#### Response

```json
{
  "message": "success"
}
```

### Delete Tuple

Delete relation tuple.
**Path:** POST /v1/relationships/delete
| Required | Argument | Type | Default | Description |
|----------|-------------------|--------|---------|-------------|
| [x]   | entity | string | - | |
| [x]   | object_id | string | - | |
| [x]   | relation | string | - | |
| [ ]   | userset_entity | string | - | |
| [x]   | userset_object_id | string | - | |
| [ ]   | userset_relation | string | - | |

### Request

```json
{
  "entity": "organization",
  "object_id": "1",
  "relation": "admin",
  "userset_entity": "",
  "userset_object_id": "1",
  "userset_relation": ""
}
```

### Response

```json
{
  "message": "success"
}
```

### Status Ping

Delete relation tuple.
**Path:** GET /v1/status/ping

### Status Version

Delete relation tuple.
**Path:** GET /v1/status/version


You can find more on [API docs](https://github.com/Permify/permify/tree/master/docs)


