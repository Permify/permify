# Expand API
We developed a expand API to see Permify Schema actions in a tree structure to improve observability and reasonability of access permissions.

Expand API is represented by a user set tree whose leaf nodes are user IDs or user sets pointing to other ⟨object#relation⟩ pairs, and intermediate nodes represent union, intersection, or exclusion operators.

Expand is crucial for our users to reason about the complete set of users and groups that have access to their objects, which allows them to build efficient search indices for access-controlled content. Unlike the Read API, Expand follows indirect references expressed through user set rewrite rules.

## Usage

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

Above schema - modeled with Permify's DSL - represents a simplified version of GitHub access control. When we look at the repository entity, we can see two actions and corresponding accesses:

 - Only owners can push to a private repository.
 - To read a private repository, the user should be one of the owners of that repository and need to belong to the parent organization of that repository ( user can either be admin or member on that organization).

According to above authorization model, let's create 3 example relation tuples for testing expand API,

`organization:1#admin@user:1`  --> User 1 is admin in organization 1‍

`repository:1#owner@user:1`  --> User 1 is owner of repository 1  

`repository:1#parent@organization:1#...`  --> repository 1 belongs to organization 1

We can use expand API to reason the access actions. If we want to reason access structure for actions of repository entity, we can use expand API with ***POST "/v1/permissions/expand"***. 

**Path:** POST /v1/permissions/expand

| Required | Argument | Type | Default | Description |
|----------|----------|---------|---------|-------------------------------------------------------------------------------------------|
| [ ]   | schema_version | string | 8 | Version of the schema |
| [ ]   | snap_token | string | - | the snap token to avoid stale cache, see more details on [Snap Tokens](/docs/reference/snap-tokens) |
| [x]   | entity | string | - | Name and id of the entity. Example: repository:1”.
| [x]   | action | string | - | The action the user wants to perform on the resource |

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
            "exclusion": false,
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
    "entity": {
        "type": "repository",
        "id": "1"
    },
    "action": "read"
}
```

</p>
</details>

<details><summary>Response</summary>
<p>

```json
{
    "tree": {
        "target": null,
        "expand": {
            "operation": "INTERSECTION",
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
                        "exclusion": false,
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
                    "target": null,
                    "expand": {
                        "operation": "UNION",
                        "children": [
                            {
                                "target": null,
                                "expand": {
                                    "operation": "UNION",
                                    "children": [
                                        {
                                            "target": {
                                                "entity": {
                                                    "type": "repository",
                                                    "id": "1"
                                                },
                                                "relation": "parent.admin"
                                            },
                                            "leaf": {
                                                "exclusion": false,
                                                "subjects": [
                                                    {
                                                        "type": "organization",
                                                        "id": "1",
                                                        "relation": "admin"
                                                    }
                                                ]
                                            }
                                        },
                                        {
                                            "target": {
                                                "entity": {
                                                    "type": "organization",
                                                    "id": "1"
                                                },
                                                "relation": "admin"
                                            },
                                            "leaf": {
                                                "exclusion": false,
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
                                "target": null,
                                "expand": {
                                    "operation": "UNION",
                                    "children": [
                                        {
                                            "target": {
                                                "entity": {
                                                    "type": "repository",
                                                    "id": "1"
                                                },
                                                "relation": "parent.member"
                                            },
                                            "leaf": {
                                                "exclusion": false,
                                                "subjects": [
                                                    {
                                                        "type": "organization",
                                                        "id": "1",
                                                        "relation": "member"
                                                    }
                                                ]
                                            }
                                        },
                                        {
                                            "target": {
                                                "entity": {
                                                    "type": "organization",
                                                    "id": "1"
                                                },
                                                "relation": "member"
                                            },
                                            "leaf": {
                                                "exclusion": false,
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

#### **Graph Representation of Expanding Read Action**

![graph-of-relations](https://user-images.githubusercontent.com/34595361/186653899-7090feb5-8ef4-4a8c-991f-ed9475a5e1f7.png)