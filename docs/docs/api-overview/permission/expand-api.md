import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

# Expand API 

Retrieve all subjects (users and user sets) that have a relationship or attribute with given entity and permission

Expand API response is represented by a user set tree, whose leaf nodes are user IDs or user sets pointing to other ⟨object#relation⟩ pairs. 

:::caution When To Use ?
Expand is designed for reasoning the complete set of users that have access to their objects, which allows our users to build efficient search indices for access-controlled content. 

It is not designed to use as a check access. Expand request has a high latency which can cause a performance issues when its used as access check.
:::

[![View in Swagger](http://jessemillar.github.io/view-in-swagger-button/button.svg)](https://permify.github.io/permify-swagger/#/Permission/permissions.expand)

<Tabs>
<TabItem value="go" label="Go">

```go
cr, err: = client.Permission.Expand(context.Background(), &v1.PermissionExpandRequest{
    TenantId: "t1",
    Metadata: &v1.PermissionExpandRequestMetadata{
        SnapToken: "",
        SchemaVersion: "",
    },
    Entity: &v1.Entity{
        Type: "repository",
        Id: "1",
    },
    Permission: "push",
})
```

</TabItem>

<TabItem value="node" label="Node">

```javascript
client.permission.expand({
  tenantId: "t1",
  metadata: {
        snapToken: "",
        schemaVersion: ""
    },
    entity: {
        type: "repository",
        id: "1"
    },
    permission: "push",
})
```

</TabItem>
<TabItem value="curl" label="cURL">

```curl
curl --location --request POST 'localhost:3476/v1/tenants/{tenant_id}/permissions/expand' \
--header 'Content-Type: application/json' \
--data-raw '{
  "metadata": {
    "schema_version": "",
    "snap_token": ""
  },
  "entity": {
    "type": "repository",
    "id": "1"
  },
  "permission": "push"
}'
```
</TabItem>
</Tabs>

## Example Usage

To give an example usage for Expand API, let's examine following authorization model.

```perm
entity user {} 

entity organization {

    relation admin @user    
    relation member @user    

    action create_repository = admin or member
    action delete = admin

} 

entity repository {

    relation    parent   @organization 
    relation    owner    @user           

    action push   = owner
    action read   = owner and (parent.admin or parent.member)

} 
```

Above schema - modeled with Permify DSL - represents a simplified version of GitHub access control. When we look at the repository entity, we can see two actions and corresponding accesses:

 - Only owners can push to a private repository.
 - To read a private repository, the user should be one of the owners of that repository and need to belong to the parent organization of that repository ( user can either be admin or member on that organization).

According to above authorization model, let's create 3 example relation tuples for testing expand API,

`organization:1#admin@user:1`  --> User 1 is admin in organization 1‍

`repository:1#owner@user:1`  --> User 1 is owner of repository 1  

`repository:1#parent@organization:1#...`  --> repository 1 belongs to organization 1

We can use expand API to reason the access actions. If we want to reason access structure for actions of repository entity, we can use expand API with ***POST "/v1/permissions/expand"***. 

**Path:** POST /v1/tenants/{tenant_id}/permissions/expand

| Required | Argument          | Type   | Default | Description                                                                                                                                                                |
|----------|-------------------|--------|---------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| [x]      | tenant_id         | string | -       | identifier of the tenant, if you are not using multi-tenancy (have only one tenant) use pre-inserted tenant `t1` for this field.                                           |
| [ ]      | schema_version    | string | -       | Version of the schema                                                                                                                                                      |
| [ ]      | snap_token        | string | -       | the snap token to avoid stale cache, see more details on [Snap Tokens](../../reference/snap-tokens)                                                                        |
| [x]      | entity            | string | -       | Name and id of the entity. Example: repository:1”.                                                                                                                         |
| [x]      | permission        | string | -       | The permission the user wants to perform on the resource                                                                                                                   |
| [ ]      | context | object | -       | Contextual tuples are relations that can be dynamically added to permission request operations. See more details on [Contextual Tuples](../../reference/contextual-tuples) |

### Expand Push Action 

<details><summary>Request</summary>
<p>

```json
{
  "metadata": {
    "schema_version": "",
    "snap_token": ""
  },
  "entity": {
    "type": "repository",
    "id": "1"
  },
  "permission": "push"
}
```

</p>
</details>

<details><summary>Response</summary>
<p>

```json
{
  "tree": {
    "target": {
      "entity": {
        "type": "repository",
        "id": "1"
      },
      "relation": "owner"
    },
    "leaf": {
      "subjects": [
        {
          "type": "user",
          "id": "1",
          "relation": ""
        }
      ]
    }
  }
}
```

</p>
</details>

### Expand Read Action 

<details><summary>Request</summary>
<p>

```json
{
  "metadata": {
    "schema_version": "",
    "snap_token": ""
  },
  "entity": {
    "type": "repository",
    "id": "1"
  },
  "permission": "read"
}
```

</p>
</details>

<details><summary>Response</summary>
<p>

```json
{
  "tree": {
    "target": {
      "entity": {
        "type": "repository",
        "id": "1"
      },
      "relation": "read"
    },
    "expand": {
      "operation": "OPERATION_INTERSECTION",
      "children": [
        {
          "target": {
            "entity": {
              "type": "repository",
              "id": "1"
            },
            "relation": "owner"
          },
          "leaf": {
            "subjects": [
              {
                "type": "user",
                "id": "1",
                "relation": ""
              }
            ]
          }
        },
        {
          "target": {
            "entity": {
              "type": "repository",
              "id": "1"
            },
            "relation": "read"
          },
          "expand": {
            "operation": "OPERATION_UNION",
            "children": [
              {
                "target": {
                  "entity": {
                    "type": "repository",
                    "id": "1"
                  },
                  "relation": "read"
                },
                "expand": {
                  "operation": "OPERATION_UNION",
                  "children": [
                    {
                      "target": {
                        "entity": {
                          "type": "organization",
                          "id": "1"
                        },
                        "relation": "admin"
                      },
                      "leaf": {
                        "subjects": [
                          {
                            "type": "user",
                            "id": "1",
                            "relation": ""
                          }
                        ]
                      }
                    }
                  ]
                }
              },
              {
                "target": {
                  "entity": {
                    "type": "repository",
                    "id": "1"
                  },
                  "relation": "read"
                },
                "expand": {
                  "operation": "OPERATION_UNION",
                  "children": [
                    {
                      "target": {
                        "entity": {
                          "type": "organization",
                          "id": "1"
                        },
                        "relation": "member"
                      },
                      "leaf": {
                        "subjects": []
                      }
                    }
                  ]
                }
              }
            ]
          }
        }
      ]
    }
  }
}
```
</p>
</details>

