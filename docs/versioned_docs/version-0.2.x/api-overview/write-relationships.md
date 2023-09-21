import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

# Write Relationships

In Permify, relations between your entities, objects and users stored as [relational tuples] in [writeDB]. Since relations and authorization data's are live instances these relational tuples can be created with an simple API call in runtime. 

When using Permify, the application client should update writeDB about the changes happening in entities or resources that are related to the authorization structure. If we consider a document system; when some user joins a group that has edit access on some documents, the application side needs to write relational tuples to keep [writeDB] up-to-date. Besides, each relational tuple should be created according to its authorization model, Permify Schema.

Another example: when one a company executive grant admin role to user (lets say with id = 3) on their organization, application side needs to tell that update to Permify in order to reform that as relation tuples and store in [writeDB].

![tuple-creation](https://user-images.githubusercontent.com/34595361/186637488-30838a3b-849a-4859-ae4f-d664137bb6ba.png)

[relational tuples]: ../getting-started/sync-data
[writeDB]: ../getting-started/sync-data#where-relational-tuples-used-

## Example Request

So if user:3 has been granted an admin role in organization:1, relational tuple `organization:1#admin@user:3` must be created by using **/v1/relationships/write** endpoint.

**Path:** POST /v1/relationships/write

| Required | Argument | Type | Default | Description |
|----------|-------------------|--------|---------|-------------|
| [x]   | tuples | array | - | Can contain multiple relation tuple object|
| [x]   | entity | object | - | Type and id of the entity. Example: "organization:1”|
| [x]   | relation | string | - | Custom relation name. Eg. admin, manager, viewer etc.|
| [x]   | subject | string | - | User or user set who wants to take the action. |
| [ ]   | schema_version | string | 8 | Version of the schema |

### Request

```json
{
    "schema_version": "",
    "tuples": [
        {
        "entity": {
            "type": "organization",
            "id": "1"
        },
        "relation": "admin",
        "subject":{
            "type": "user",
            "id": "3",
            "relation": ""
        }
    }
    ]
}
```

### Response

```json
{
    "snap_token": "FxHhb4CrLBc="
}
```

You can store that snap token alongside with the resource in your relational database, then use it used in endpoints to get fresh results from the API's. For example it can be used in access control check with sending via `snap_token` field to ensure getting check result as fresh as previous request.

See more details on what is [Snap Tokens](../reference/snap-tokens) and how its avoiding stale cache.

### Using gRPC Clients

<Tabs>
<TabItem value="go" label="Go">

```go
rr, err: = client.Relationship.Write(context.Background(), & v1.RelationshipWriteRequest {
    Metadata: &v1.RelationshipWriteRequestMetadata {
        SchemaVersion: ""
    },
    Tuples: [] * v1.Tuple {
        {
            Entity: & v1.Entity {
                Type: "organization",
                Id: "1",
            },
            Relation: "admin",
            Subject: & v1.Subject {
                Type: "admin",
                Id: "3",
            },
        }
    },
})
```

</TabItem>

<TabItem value="node" label="Node">

```javascript
client.relationship.write({
    metadata: {
        schemaVersion: ""
    },
    tuples: [{
        entity: {
            type: "organization",
            id: "1"
        },
        relation: "admin",
        subject: {
            type: "user",
            id: "3"
        }
    }]
}).then((response) => {
    // handle response
})
```

</TabItem>
</Tabs>

## Suggested Workflow 

The most of the data that should written in Permify also needs to be write or engage with applications database as well. So where and how to write relationships into both applications database and Permify ?

### Two Phase Commit Approach
In a standard relational based databases, the suggested place to write relationships to Permify is sending the write request in database transaction of the client action: such as storing the owner of the document when an user creates a document.

To give more concurrent example of this action, let's take a look at below createDocument function 

```go
func CreateDocuments(db *gorm.DB) error {

  tx := db.Begin()
  defer func() {
    if r := recover(); r != nil {
      tx.Rollback()
      // if transaction fails, then delete malformed relation tuple 
      permify.DeleteRelationships(...)
    }
  }()

  if err := tx.Error; err != nil {
    return err
  }

  if err := tx.Create(docs).Error; err != nil {
     tx.Rollback()
     // if transaction fails, then delete malformed relation tuple 
     permify.DeleteRelationships(...)
     return err
  }

  // if transaction successful, write relation tuple to Permify 
  permify.WriteRelationships(...)

  return tx.Commit().Error
}
```
The key point to take way from above approach is if the transaction fails for any reason, the relation will also be deleted from Permify to provide maximum consistency.

### Relationships that not stored in application database

Although ownership generally stored in application databases, there are some relations that not needed to be stored in your actual database. Such as defining organizational roles, group members, project editors etc.

For example, you can model a simple project management authorization in Permify as follows, 

```perm
entity user {}

entity team {

    relation owner @user
    relation member @user
}

entity project {

    relation team @team
    relation owner @user

    action view = team.member or team.owner or project.owner
    action edit = project.owner or team.owner
    action delete = project.owner or team.owner

}
```

This **team member** relation won't need to be stored in the application database. Storing it only in Permify - WriteDB - is enough. In that situation, `WriteRelationships` can be performed in any logical place in your stack.

## Need any help ?

Our team is happy to help you get started with Permify. If you'd like to learn more about using Permify in your app or have any questions about this example, [schedule a call with one of our Permify engineer](https://meetings-eu1.hubspot.com/ege-aytin/call-with-an-expert).