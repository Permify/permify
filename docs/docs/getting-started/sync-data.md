---
sidebar_position: 2
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

# Managing Authorization Data

Permify unifies your authorization data in a database you prefer. We named that database as Write Database, shortly **WriteDB**.

Permify API provides various functionalities - checking access, reasoning permissions, etc - to maintain separate access control mechanisms for individual applications. And **WriteDB** stands as a source of truth for these authorization functionalities.

## Access Control as Relations - Relational Tuples

In Permify, relationship between your entities, objects, and users builds up a collection of access control lists (ACLs). 

These ACLs called relational tuples: the underlying data form that represents object-to-object and object-to-subject relations. Each relational tuple represents an action that a specific user or user set can do on a resource and takes form of `user U has relation R to object O`, where user U could be a simple user or a user set such as team X members.

In Permify, the simplest form of relational tuple structured as: `entity # relation @ user`. Here are some relational tuples with semantics,

![relational-tuples](https://user-images.githubusercontent.com/34595361/183959294-149fcbb9-7f10-4c1e-8d66-20a839893909.png)

## Where Relational Tuples Used ?

In Permify, these relational tuples represents your authorization data. 

Permify stores your relational tuples (authorization data) in a database you prefer. You can configure the database when running Permify Service with using both [configuration flags](../installation/brew#configuration-flags) or [configuration YAML file](https://github.com/Permify/permify/blob/master/example.config.yaml).

Stored relational tuples are queried and utilized in Permify APIs, including the check API, which is an access control check request used to determine whether a user's action is authorized.

As an example; to decide whether a user could view a protected resource, Permify looks up the relations between that specific user and the protected resource. These relation types could be ownership, parent-child relation, or even a role such as an admin or manager.
[WriteDB]: #write-database

## Creating Relational Tuples 

Relational tuples can be created with an simple API call in runtime, since relations and authorization data's are live instances. Each relational tuple should be created according to its authorization model, [Permify Schema]. 

[Permify Schema]: ../getting-started/modeling

![tuple-creation](https://user-images.githubusercontent.com/34595361/186637488-30838a3b-849a-4859-ae4f-d664137bb6ba.png)

Let's follow a simple document management system example with the following Permify Schema to see how to create relation tuples. 

```perm
entity user {} 

entity organization {

    relation admin  @user
    relation member @user

} 

entity document {
    
    relation  owner  @user   
    relation  parent    @organization   
    relation  maintainer  @user @organization#member      

    action view   = owner or parent.member or maintainer or parent.admin
    action edit   = owner or maintainer or parent.admin
    action delete = owner or parent.admin
} 
```

According to the schema above; when a user creates a document in an organization, more specifically let's say, when user:1 create a document:2 we need to create the following relational tuple,

- `document:2#owner@user:1`

[WriteDB]: #write-database

### Write Data API

You can create relational tuples by using `Write Data API`. 

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
                Type: "document",
                Id: "2",
            },
            Relation: "owner",
            Subject: & v1.Subject {
                Type: "user",
                Id: "1",
            },
        }
    },
})
```

</TabItem>

<TabItem value="node" label="Node">

```javascript
client.data.write({
    tenantId: "t1",
    metadata: {
        schemaVersion: ""
    },
    tuples: [{
        entity: {
            type: "document",
            id: "2"
        },
        relation: "owner",
        subject: {
            type: "user",
            id: "1"
        }
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
    "metadata": {
        "schema_version": ""
    },
    "tuples": [
        {
        "entity": {
            "type": "document",
            "id": "2s"
        },
        "relation": "owner",
        "subject":{
            "type": "user",
            "id": "1",
            "relation": ""
        }
    }
    ]
}'
```
</TabItem>
</Tabs>

### Snap Tokens

In Write Data API response you'll get a snap token of the operation. 

```json
{
    "snap_token": "FxHhb4CrLBc="
}
```

This token consists of an encoded timestamp, which is used to ensure fresh results in access control checks. We're suggesting to use snap tokens in production to prevent data inconsistency and optimize the performance. See more on [Snap Tokens](../reference/snap-tokens.md)

## More Examples

Let's create more example data according to the schema we defined above.

### Organization Admin

**relational tuple:** organization:1#admin@user:3

**Semantics:** User 3 is administrator in organization 1.

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
                Type: "user",
                Id: "3",
            },
        }
    },
})
```

</TabItem>

<TabItem value="node" label="Node">

```javascript
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
            id: "3"
        }
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

### Parent Organization

**Relational Tuple:** document:1#parent@organization:1#…

**Semantics:** Organization 1 is parent of document 1.

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
                Type: "document",
                Id: "1",
            },
            Relation: "parent",
            Subject: & v1.Subject {
                Type: "organization",
                Id: "1",
                Relation: "..."
            },
        }
    },
})
```

</TabItem>

<TabItem value="node" label="Node">

```javascript
client.data.write({
    tenantId: "t1",
    metadata: {
        schemaVersion: ""
    },
    tuples: [{
        entity: {
            type: "document",
            id: "1"
        },
        relation: "parent",
        subject: {
            type: "organization",
            id: "1",
            relation: "..."
        }
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
    "metadata": {
        "schema_version": ""
    },
    "tuples": [
        {
        "entity": {
            "type": "document",
            "id": "1"
        },
        "relation": "parent",
        "subject":{
            "type": "organization",
            "id": "1",
            "relation": "..."
        }
    }
    ]
}'
```
</TabItem>
</Tabs>

:::info
Note: `relation: “...”` used when subject type is different from **user** entity. **#…** represents a relation that does not affect the semantics of the tuple.

Simply, the usage of ... is straightforward: if you're use user entity as an subject, you should not be using the `...` If you're using another subject rather than user entity then you need to use the `...` 
:::

### Organization Members Are Maintainers in specific Doc

**Created relational tuple:** document:1#maintainer@organization:2#member

**Definition:** Members of organization 2 are maintainers in document 1.

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
                Type: "document",
                Id: "1",
            },
            Relation: "maintainer",
            Subject: & v1.Subject {
                Type: "organization",
                Id: "2",
                Relation: "member"
            },
        }
    },
})
```

</TabItem>

<TabItem value="node" label="Node">

```javascript
client.data.write({
    tenantId: "t1",
    metadata: {
        schemaVersion: ""
    },
    tuples: [{
        entity: {
            type: "document",
            id: "1"
        },
        relation: "maintainer",
        subject: {
            type: "organization",
            id: "2",
            relation: "member"
        }
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
    "metadata": {
        "schema_version": ""
    },
    "tuples": [
        {
        "entity": {
            "type": "document",
            "id": "1"
        },
        "relation": "maintainer",
        "subject":{
            "type": "organization",
            "id": "2",
            "relation": "member"
        }
    }
    ]
}'
```
</TabItem>
</Tabs>

#### Test this Example on [Playground](https://play.permify.co/?s=bCDvst-22ISFR6DV90y8_)

## Audit Logs For Permission Changes

Permify does support audit logs for permission changes. Leveraging the [MVCC (Multi-Version Concurrency Control)](http://mbukowicz.github.io/databases/2020/05/01/snapshot-isolation-in-postgresql.html) pattern, we maintain a history of all permission data changes. This essentially provides an audit trail, allowing users to track alterations and when they occurred.

In cloud version, our system supports change history auditing. It automatically generates and securely stores logs for all significant actions. These logs detail who made the change, what was changed, and when the change occurred. Furthermore, your system allows for easy searching and analysis of these logs, supporting automated alerting for suspicious activities. This comprehensive approach ensures thorough and effective auditing of all changes

## Permission Baselining (Reviewing)

We have a strong foundation for permission baselining and review, thanks to MVCC.

**Historical Review:** You can review the history of permissions changes as each version is stored. This enables retrospective audits and analysis.

**Current State Review:** You can review the current state of permissions by examining the latest versions of each permission setting.

**Cleanup:** Your system incorporates a garbage collector for managing old data, which helps keep your permissions structure clean and optimized.

## Next 

Let's now head over to the **Access Control Check** section and learn how to perform access control in Permify to ensure that only authorized users have the right level of access to our resources.
