---
title: "Migrating from 0.2.x to 0.3.x"
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

## Migration overview

This doc guides you through migrating an existing Permify **0.2.x** authorization service to **0.3.x**. With version 0.3.x Permify moved to a tenancy-based infrastructure, which affects almost all of the API operations.

## Multi Tenancy on Permify

Multi-tenancy in Permify refers to an authorization architecture, where a single Permify authorization service serves multiple applications/organizations (tenants).

This allows the ability to customize the authorization for each tenant's specific needs. With Multi-Tenancy support, you can create custom authorization schema and relation tuples accordingly for the different tenants and manage them in a single place - in [WriteDB](./getting-started/sync-data.md).

For the users that don't have/need multi-tenancy in their authorization structure, we created a pre-inserted tenant (id: **t1**) that comes default when you serve a Permify service.

## What have changed ?

Several things changed when we moved to tenant based infrastructure, these are:

* [API endpoints now have Tenant ID field](#api-endpoints-now-have-tenant-id-field)
* [Added Tenancy Service](#added-tenancy-service)
* [WriteDB tables and tenant id column](#writedb-tables-and-tenant-id-column)

### API endpoints now have Tenant ID field 

All API endpoints that cover in 0.2.x now have a `‍tenant_id` mandatory field. Let's examine a check request below,

#### Check API

<Tabs>
<TabItem value="go" label="Go">

```go
cr, err: = client.Permission.Check(context.Background(), & v1.PermissionCheckRequest {
    TenantId: "t1",
    Metadata: & v1.PermissionCheckRequestMetadata {
        SnapToken: ""
        SchemaVersion: ""
        Depth: 20,
    },
    Entity: & v1.Entity {
        Type: "repository",
        Id: "1",
    },
    Permission: "edit",
    Subject: & v1.Subject {
        Type: "user",
        Id: "1",
    },

    if (cr.can === PermissionCheckResponse_Result.RESULT_ALLOWED) {
        // RESULT_ALLOWED
    } else {
        // RESULT_DENIED
    }
})
```

</TabItem>
<TabItem value="node" label="Node">

```javascript
client.permission.check({
    tenantId: "t1", 
    metadata: {
        snapToken: "",
        schemaVersion: "",
        depth: 20
    },
    entity: {
        type: "repository",
        id: "1"
    },
    permission: "edit",
    subject: {
        type: "user",
        id: "1"
    }
}).then((response) => {
    if (response.can === PermissionCheckResponse_Result.RESULT_ALLOWED) {
        console.log("RESULT_ALLOWED")
    } else {
        console.log("RESULT_DENIED")
    }
})
```

</TabItem>
<TabItem value="curl" label="cURL">

```curl
curl --location --request POST 'localhost:3476/v1/tenants/{tenant_id}/permissions/check' \
--header 'Content-Type: application/json' \
--data-raw '{
  "metadata":{
    "snap_token": "",
    "schema_version": "",
    "depth": 20
  },
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
}'
```
</TabItem>
</Tabs>

Users that come from version 0.2.x and users that have a single tenant can enter **t1** as tenant id. See changes on the other endpoints from [API Overview Section](./api-overview/).

### Added Tenancy Service

To manage tenants we have added a Tenancy service; you can create, delete and list tenants accordingly. See the [Tenancy Service](./api-overview/tenancy/) on Using The API section.

### WriteDB tenancy table and tenant id column

#### Tenant Table 

Tenants table have added the Write DB to store tenant's details. The new WriteDB folder structure changed as follows:
```
tables
├── migrations       
├── relation_tuples   
├── schema_definitions   
├── tenants   
├── transactions   
```

#### Tenant ID Column

Relation tuples and schema definition tables now have a tenant_id column, which stores the id of the tenant that data belongs.

Let's take a look at a snapshot of the demo table on an example WriteDB.

Example Relation Tuples data table:
![tenant-id-tuples](https://user-images.githubusercontent.com/34595361/214724165-a3775756-0649-4869-b994-d837fadd271d.png)

Example Schema Definitions data table
![tenant-id-schema](https://user-images.githubusercontent.com/34595361/214724727-01eadad3-720c-4c10-a88d-6ee293ecf4a8.png)

## Need any help ?

Our team is happy to help! If you struggle with migration or need help on using the multi-tenancy, [schedule a call with one of our Permify engineers](https://meetings-eu1.hubspot.com/ege-aytin/call-with-an-expert). Alternatively you can join our [discord community](https://discord.com/invite/MJbUjwskdH) to discuss.
