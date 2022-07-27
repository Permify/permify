
# Move & Synchronize Authorization Data

Permify aims to unify your authorization data which behave as centralized data source used in access controls check.

You can convert, coordinate & sync authorization data based on a YAML config file.

***Example config file:***

```yaml
app:
  name: ‘permify’
  version: ‘0.0.1’
http:
  port: ‘3476’
logger:
  log_level: ‘debug’
  rollbar_env: ‘permify’
database:
  listen:
    connection: postgres
    slot_name: replication_slot
    output_plugin: pgoutput
    pool_max: 2
    url: 'postgres://postgres:SphU4Uf3QXNnZEtT@permify-projects.ceuo5kqsxyea.us-east-1.rds.amazonaws.com:5432/github?replication=database'
    tables:
        - organization
        - repository
  write:
    connection: postgres
    pool_max: 2
    url: ‘postgres://user:password@host:5432/database_name’
```

- **WriteDB:** Where your want to store relation tuples, audits , and decision logs.

- **ListenDB (optional):** Where your application data is stored. Mandatory when using Change Data Capture approach.

There are 2 approaches to move and syncronize your authorization data in Permify; 
 - [With Change Data Capture](#with-change-data-capture)
 - [Creating Custom Relational Tuples](#creating-custom-relational-tuples)

## With Change Data Capture

Permify applies change data capture (CDC) pattern to coordinate authorization related data in your databases with YAML config file and [Permify Schema](https://github.com/Permify/permify/blob/master/assets/content/MODEL.md) in which you define your authorization relations. 

We publish & subscribe to your Listen DB. And based on a YAML schema file; Any time you create, update, or delete data; we convert, coordinate and sync your authorization data as relation tuples into your database (WriteDB) you point at. Note that you can define multiple listenDB's for data capturing. 

When using CDC pattern additional changes (below) needs to be made in your postgre congif (*postgresql.conf*) file to complete. 

```yaml
    wal_level = logical
    #
    # these parameters only need to set in versions 9.4, 9.5 and 9.6
    # default values are ok in version 10 or later
    #
    max_replication_slots = 10
    max_wal_senders = 10
```

After updating your *postgresql.conf* file you need to reboot your instance.

## Creating Custom Relational Tuples

In case you don't want a Permify listen your databases constanty. You can skip defining listenDB in config YAMl file and can create custom relational tuples manually. You can create custom relational tuples with using "/v1/relationships/write" endpoint.

**Path:** POST /v1/relationships/write
| Required | Argument | Type | Default | Description |
|----------|-------------------|--------|---------|-------------|
| [x]   | entity | string | - | Name of the object or resource type|
| [x]   | object_id | string | - | entity id|
| [x]   | relation | string | - | custom relation name. Eg. admin, manager, viewer etc. |
| [ ]   | userset_entity | string | - | user or resource type, which has relation with entity  |
| [x]   | userset_object_id | string | - | user or resource id, which has relation with entity |
| [ ]   | userset_relation | string | - | user or resource relation of given userset object. |

### Examples 

#### **Organization Admin**

Request

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

Response

```json
{
  "message": "success"
}
```

Created relational tuple: ***organization:1#admin@1***

Definition: user 1 has admin role on organization 1.

#### **Organization Members are Viewer of Repo** 

Request

```json
{
  "entity": "repository",
  "object_id": "1",
  "relation": "viewer",
  "userset_entity": "organization",
  "userset_object_id": "2",
  "userset_relation": "member"
}
```

Response

```json
{
  "message": "success"
}
```

Created relational tuple: ***repository:1#admin@organization:2#member***

Definition: members of organization 2 are viewers of repository 1.

#### **#... case (Parent Organization)**

Request

```json
{
  "entity": "repository",
  "object_id": "1",
  "relation": "parent",
  "userset_entity": "organization",
  "userset_object_id": "1",
  "userset_relation": "..."
}
```

Response

```json
{
  "message": "success"
}
```

Created relational tuple: ***repository:1#parent@organization:1#…***

Definition: organization 1 is parent of repository 1.

Note: “#...” represents a relation that does not affect the semantics of the tuple.
