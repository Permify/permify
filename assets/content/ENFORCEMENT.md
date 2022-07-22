# Access Control Check

You can check authorization with
single API call. This check request returns a decision about whether user can perform an action on a certain resource.

Access decisions generated according to relational tuples, which stored in your database (writeDB) and [Permify Schema](https://github.com/Permify/permify/blob/master/assets/content/MODEL.md) action conditions.

## Example Check

Lets examine a example access control decision on github: 

***Can the user X push on a repository Y ?***

### Path: 

POST /v1/permissions/check

| Required | Argument | Type | Default | Description |
|----------|----------|---------|---------|-------------------------------------------------------------------------------------------|
| [x]   | user | string | - | the user or user set who wants to take the action.
| [x]   | action | string | - | the action the user wants to perform on the resource |
| [x]   | object | string | - | name and id of the resource. Example: repository:1” |
| [ ]   | depth | integer | 8 | |

### Request

```json
{
  "user": "1",
  "action": "push",
  "object": "repository:1"
}
```

### Response

```json
{
  "can": true,
  "debug": "user 1 is a owner of organization 1"
}
```
