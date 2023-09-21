import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

# Schema Lookup

You can use schema lookup API endpoint to retrieve all permissions associated with a resource relation. Basically, you can perform enforcement without checking stored authorization data. For example in given a Permify Schema like:

```
entity user {}

entity document { 

 relation assignee @user  
 relation manager @user     
 
 action view = assignee or manager
 action edit = manager
 
}

```

Let's say you have a user X with a manager role. If you want to check what user X can do on a documents ? You can use the schema lookup endpoint as follows,

**Path:** POST /v1/permissions/lookup-schema

| Required | Argument | Type | Default | Description |
|----------|----------|---------|---------|-------------------------------------------------------------------------------------------|
| [ ]   | schema_version | string | 8 | Version of the schema |
| [x]   | entity_type | string | - | type of the entity. 
| [x]   | relation_names | string[] | - | string array that holds entity relations |

#### Request

```json
{
  "metadata": {
    "schema_version": ""
  },
  "entity_type": "document",
  "relation_names": [ "manager" ]
}
```

#### Response

```json
{
  "data": {
    "action_names": [ 
        "view",
        "edit"
     ]
   }
}
```


### Using gRPC Clients

<Tabs>
<TabItem value="go" label="Go">

```go
cr, err: = client.Permission.LookupSchema(context.Background(), & v1.PermissionLookupSchemaRequest {
    Metadata: & v1.PermissionLookupSchemaRequestMetadata {
        SchemaVersion: ""
    },
    EntityType: "document",
    RelationNames: []string {"manager"},
})
```

</TabItem>
<TabItem value="node" label="Node">

```javascript
client.permission.LookupSchema({
     metadata: {
      schema_version: ""
    },
    entity_type: "document",
    relation_names: [ "manager" ]
})
```

</TabItem>
</Tabs>

The response will return all the possible actions that manager can perform on documents. Also you can extend relation lookup as much as you want by adding relations to the **"relation_names"** array.