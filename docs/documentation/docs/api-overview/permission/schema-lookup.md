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

## Request

**Path:** POST /v1/permissions/lookup-schema

| Required | Argument | Type | Default | Description |
|----------|----------|---------|---------|-------------------------------------------------------------------------------------------|
| [x]   | tenant_id | string | - | identifier of the tenant, if you are not using multi-tenancy (have only one tenant) use pre-inserted tenant `t1` for this field.
| [ ]   | schema_version | string | 8 | Version of the schema |
| [x]   | entity_type | string | - | type of the entity. 
| [x]   | relation_names | string[] | - | string array that holds entity relations |

<Tabs>
<TabItem value="go" label="Go">

```go
cr, err: = client.Permission.LookupSchema(context.Background(), & v1.PermissionLookupSchemaRequest {
    TenantId: "t1",
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
     tenantId: "t1",
     metadata: {
      schema_version: ""
    },
    entity_type: "document",
    relation_names: [ "manager" ]
})
```

</TabItem>
<TabItem value="curl" label="cURL">

```curl
curl --location --request POST 'localhost:3476/v1/tenants/{tenant_id}/permissions/lookup-schema' \
--header 'Content-Type: application/json' \
--data-raw '{
  "metadata": {
    "schema_version": ""
  },
  "entity_type": "document",
  "relation_names": [ "manager" ]
}'
```
</TabItem>
</Tabs>

## Response

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


The response will return all the possible actions that manager can perform on documents. Also you can extend relation lookup as much as you want by adding relations to the **"relation_names"** array.