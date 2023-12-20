import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

# Write Authorization Data

In Permify, relations between your entities, objects and users stored as [relational tuples] in a [preferred database]. Since relations and authorization data's are live instances these relational tuples can be created with an simple API call in runtime.

When using Permify, the application client should update preferred database about the changes happening in entities or resources that are related to the authorization structure. If we consider a document system; when some user joins a group that has edit access on some documents, the application side needs to write relational tuples to keep [preferred database] up-to-date. Besides, each relational tuple should be created according to its authorization model, Permify Schema.

Another example: when one a company executive grant admin role to user (lets say with id = 3) on their organization, application side needs to tell that update to Permify in order to reform that as relation tuples and store in [preferred database].

![tuple-creation](https://user-images.githubusercontent.com/34595361/186637488-30838a3b-849a-4859-ae4f-d664137bb6ba.png)

[relational tuples]: ../../../getting-started/sync-data
[preferred database]: ../../../getting-started/sync-data#where-relational-tuples-used

## Write Request

:::info
You can use the **/v1/tenants/{tenant_id}/data/write** endpoint for both creating **relation tuples** and for creating **attribute data**.
:::

**Path:** POST /v1/tenants/{tenant_id}/data/write

[![View in Swagger](http://jessemillar.github.io/view-in-swagger-button/button.svg)](https://permify.github.io/permify-swagger/#/Data/data.write)

#### Glossary for parameters & payload objects:

| Required | Argument       | Type   | Default | Description                                                                                                                                           |
| -------- | -------------- | ------ | ------- | ----------------------------------------------------------------------------------------------------------------------------------------------------- |
| [x]      | tenant_id      | string | -       | identifier of the tenant, if you are not using multi-tenancy (have only one tenant in your system) use pre-inserted tenant **t1** for this parameter. |
| [ ]      | schema_version | string | 8       | Version of the schema.                                                                                                                 |
| [x] | tuples | array | - | Array of objects that are used to define relationships. Each object contains **entity**, **relation**, and **subject** arguments.|
| [x] | attributes | array | - | Array of objects that are used to define relationships. Each object contains **entity**, **attribute**, and **value** arguments. |
| [x] | entity | object | - | Type and id of the entity. Example: "organization:1” |
| [x] | subject | string | - | User or user set who wants to take the action. |
| [x] | relation | string | - | Custom relation name. Eg. admin, manager, viewer etc. |
| [x] | attribute | string | - | Custom attribute name. |
| [x] | value     | object | - | Represents value and type of the attribute data. |


### Creating Relational Tuple

Let's create an example relation tuple. If user:3 has been granted an admin role in organization:1, relational tuple `organization:1#admin@user:3` should be created as follows:

<Tabs>
<TabItem value="go" label="Go">

```go
rr, err: = client.Data.Write(context.Background(), & v1.DataWriteRequest {
    TenantId: "t1",
    Metadata: &v1.DataWriteRequestMetadata {
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
client.data
  .write({
    tenantId: "t1",
    metadata: {
      schemaVersion: "",
    },
    tuples: [
      {
        entity: {
          type: "organization",
          id: "1",
        },
        relation: "admin",
        subject: {
          type: "user",
          id: "3",
        },
      },
    ],
  })
  .then((response) => {
    // handle response
  });
```

</TabItem>
<TabItem value="curl" label="cURL">

```curl
curl --location --request POST 'localhost:3476/v1/tenants/{tenant_id}/data/write' \
--header 'Content-Type: application/json' \
--data-raw '{
    "metadata": {
        "schema_version": ""
    },
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
}'
```

</TabItem>
</Tabs>

### Creating Attribute Data

You can use `attributes` argument to create attribute/attributes, similarly the `tuples`. 

Let's conitnue with an example: Assume user:1 has been granted an admin role in organization:1 and organization:1 is a private (boolean) organization:

:::warning **value** field
**value** field is mandatory on attribute data creation.

Here are the available attribute value types:

- **type.googleapis.com/base.v1.StringValue**
- **type.googleapis.com/base.v1.BooleanValue**
- **type.googleapis.com/base.v1.IntegerValue**
- **type.googleapis.com/base.v1.DoubleValue**
- **type.googleapis.com/base.v1.StringArrayValue**
- **type.googleapis.com/base.v1.BooleanArrayValue**
- **type.googleapis.com/base.v1.IntegerArrayValue**
- **type.googleapis.com/base.v1.DoubleArrayValue**
:::

<Tabs>
<TabItem value="go" label="Go">

```go
// Convert the wrapped attribute value into Any proto message
value, err := anypb.New(&v1.BooleanValue{
    Data: true,
})
if err != nil {
	// Handle error
}

cr, err := client.Data.Write(context.Background(), &v1.DataWriteRequest{
    TenantId: "t1",,
    Metadata: &v1.DataWriteRequestMetadata{
        SchemaVersion: "",
    },
	Tuples: []*v1.Attribute{
        {
            Entity: &v1.Entity{
                Type: "organization",
                Id:   "1",
            },
            Relation: "admin",
            Subject:  &v1.Subject{
		        Type: "user",
		        Id:   "1",
				Relation: "",
			},
        },
    },
    Attributes: []*v1.Attribute{
        {
            Entity: &v1.Entity{
                Type: "account",
                Id:   "1",
            },
            Attribute: "public",
            Value:     value,
        },
    },
})
```

</TabItem>

<TabItem value="node" label="Node">

```javascript
const booleanValue = BooleanValue.fromJSON({ data: true });

const value = Any.fromJSON({
    typeUrl: 'type.googleapis.com/base.v1.BooleanValue',
    value: BooleanValue.encode(booleanValue).finish()
});

client.data.write({
    tenantId: "t1",
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
            id: "1"
        }
    }],
    attributes: [{
        entity: {
            type: "document",
            id: "1"
        },
        attribute: "public",
        value: value,
    }]
}).then((response) => {
    // handle response
})
```

</TabItem>
<TabItem value="curl" label="cURL">

```curl
curl --location --request POST 'localhost:3476/v1/tenants/{tenant_id}/data/write' \
--header 'Content-Type: application/json' \
--data-raw '{
{
    "metadata": {
        "schema_version": ""
    },
    "tuples": [
      {
        "entity": {
          "type": "organization",
          "id": "1"
        },
        "relation": "admin",
        "subject": {
          "type": "user",
          "id": "1"
        }
    }
    ],
    "attributes": [
        {
            "entity": {
                "type": "organization",
                "id": "1"
            },
            "attribute": "private",
            "value": {
                "@type": "type.googleapis.com/base.v1.BooleanValue",
                "data": true
            }
        }
    ]
}
}'
```

</TabItem>
</Tabs>

## Response

```json
{
  "snap_token": "FxHhb4CrLBc="
}
```

You can store that snap token alongside with the resource in your relational database, then use it used in endpoints to get fresh results from the API's. For example it can be used in access control check with sending via `snap_token` field to ensure getting check result as fresh as previous request.

See more details on what is [Snap Tokens](../../../reference/snap-tokens) and how its avoiding stale cache.

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
      permify.DeleteData(...)
    }
  }()

  if err := tx.Error; err != nil {
    return err
  }

  if err := tx.Create(docs).Error; err != nil {
     tx.Rollback()
     // if transaction fails, then delete malformed relation tuple
     permify.DeleteData(...)
     return err
  }

  // if transaction successful, write relation tuple to Permify
  permify.WriteData(...)

  return tx.Commit().Error
}
```

The key point to take way from above approach is if the transaction fails for any reason, the relation will also be deleted from Permify to provide maximum consistency.

### Data that not stored in application database

Although ownership generally stored in application databases, there are some data that not needed to be stored in your actual database. Such as defining organizational roles, group members, project editors etc.

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

This **team member** relation won't need to be stored in the application database. Storing it only in Permify - preferred database - is enough. In that situation, `WriteData` can be performed in any logical place in your stack.

## Need any help ?

Our team is happy to help you get started with Permify. If you'd like to learn more about using Permify in your app or have any questions about this example, [schedule a call with one of our Permify engineer](https://meetings-eu1.hubspot.com/ege-aytin/call-with-an-expert).
