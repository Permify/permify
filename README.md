<p align="center">
<img src="https://media.giphy.com/media/BkaGFfogyFmPLwmSXu/giphy.gif" alt="Permify - Open source authorization as a service"  width="600px" />
</p>

Permify is an open-source authorization service that you can run with docker and works on a Rest API. 

We publish & subscribe to your Postgres DB (listen DB). And based on a YAML schema file; we convert, coordinate and sync
your authorization data as relation tuples into your DB (write db) you point at. And you can check authorization with
single request based on those tuples.
Data model is inspired
by [Google Zanzibar White Paper](https://storage.googleapis.com/pub-tools-public-publication-data/pdf/41f08f03da59f5518802898f68730e247e23c331.pdf)
.
## Getting Started
Permify consists of 3 main parts; data sync, authorization model, and enforcement checks.
### Move & Sync Data
You can convert, coordinate & sync authorization data based on a YAML config file and schema in which you define your
authorization relations.
- **ListenDB:** Where your application data is stored.
- **WriteDB:** Where your want to store relation tuples, audits , and decision logs.
Permify creates a function in your Listen DB that has a trigger based on your config file. Any time you create, update,
or delete data; we sync this data as relation tuples into writeDB.

example config file
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
  pool_max: 2
  url: ‘postgres://user:password@host:5432/database_name’
 write:
  connection: postgres
  pool_max: 2
  url: ‘postgres://user:password@host:5432/database_name’
```
### Modeling Authorization Logic
You can create access decisions are based on the relationships a subject has with Permify Schema. For instance, only
allowing Repository owners to push updates.

```perm
entity user {} `table:users|identifier:id`

entity organization {

    relation admin @user     `rel:custom`
    relation member @user    `rel:many-to-many|table:org_members|cols:org_id,user_id`

    action create_repository = admin or member
    action delete = admin

} `table:organizations|identifier:id`

entity repository {

    relation    owner @user          `rel:belongs-to|cols:owner_id`
    relation    org   @organization    `rel:belongs-to|cols:organization_id`

    action push   = owner
    action read   = (owner or org.member) and org.admin
    action delete = org.admin or owner

} `table:repositories|identifier:id`
```
#### Entities
Entities represent your main tables. The table name and the entity name here must be the same. For example, name of the
user entity in our example represents user table in your database.
Entity has 2 different parts. These are;
- **relations**
- **actions**
#### Relations
Relations represent relationships between entities. For example, each repository is related with an organization
While the `belongs_to` is kept within the entity itself, the `has one` is kept in the table where it is related. And
the `many-to-many` relation is kept in the pivot tables. Therefore, we need to specify the relationship types in the
schema.
- **name:** define the realtion
- **table (optional):** the name of the pivot table. (Only for many-to-many relationships.)
- **rel:** type of relationship (many-to-many, belongs-to or custom)
- **entity:** the entity it’s related with (e.g. user, organization, repo…)
- **cols:** the columns you have created in your database.
```
entity repository {

    relation    owner @user          `rel:belongs-to|cols:owner_id`
    relation    org   @organization    `rel:belongs-to|cols:organization_id`

}
```
→ Each repository belongs to an organization. which is defined as organization_id column in the repository table.
#### Actions 
Actions describe what relations, or relation’s relation can do.

For example, only the repository owner can push to
repository.
```
action push   = owner
```

another example, organization admin (user with admin role) and 
```
action read   = (owner or org.member) and org.admin
```
→ "User with a admin role and either owner of the repository, or member of the organization which repository belongs to" can read.

## Installation
### Container (Docker)
#### With terminal
1. Open your terminal.
2. Run following line.

```shell
docker run -d -p 3476:3476 --name permify-container -v {YOUR-CONFIG-PATH}:/config permify/permify:0.0.1
```

3. Test your connection.
   - Create an HTTP GET request ~ localhost:3476/v1/status/ping
#### With docker desktop
Setup docker desktop, and run service with the following steps;
1. Open your docker account.
2. Open terminal and run following line
```shell
docker pull permify/permify:0.0.1
```
3. Open images, and find Permify.
4. Run Permify with the following credentials (optional setting)
   - Container Name: authorization-container
  Ports
   - **Local Host:** 3476
  Volumes
   - **Host Path:** choose the config file and folder
   - **Container Path:** /config
5. Test your connection.
   - Create an HTTP GET request ~ localhost:3476/v1/status/ping


## Why Permify?
You can use Permify any stage of your development for your authorization needs but Permify works best:
- If you need to refactor your authorization.
- If you’re managing authorization for growing micro-service infrastructure.
- If your authorization logic is cluttering your code base.
- If your data model is getting too complicated to handle your authorization within the service.
- If your authorization is growing too complex to handle within code or API gateway.
## Features
- Sync & coordinate your authorization data hassle-free.
- Get Boolean - Yes/No decision returns.
- Store your authorization data in-house with high availability & low latency.
- Easily model, debug & refactor your authorization logic.
- Enforce authorizations with a single request anywhere you call it.
- Low latency with parallel graph engine for enforcement check.
## Example
Permify helps you move & sync authorization data from your ListenDB to WriteDB with a single config file based on your
authorization model that you provide us in a YAML schema.
After configuration, you can check authorization with a simple call.
**Request**
```json
{
 “user”: “1",
 “action”: “push”,
 “object”: “repository:1"
}
```
**Response**
```json
{
 “can”: true,
 “debug”: “user 1 is a owner of organization 1”
}
```

## The Graph of Relations
The relation tuples of the ACL used by Permify can be represented as a graph of relations. This graph will help you understand the performance of check engine and the algorithms it uses.
<img width=“1756" alt=“filled graph” src=“https://user-images.githubusercontent.com/39353278/175956412-54801b34-1524-414d-8737-19d3b1abfb73.png”>
this graph is created by combining schema file and data.

## Performance
## API
### Check
Returns a decision about whether user can perform an action on a certain resource. For example, can the user do push on
a repository object?
**Path:** POST /v1/permissions/check
| Required | Argument | Type  | Default | Description                                        |
|----------|----------|---------|---------|-------------------------------------------------------------------------------------------|
| [x]   | user   | string | -    | the user or user set who wants to take the action. Examples: “1”, “organization:1#owners” |
| [x]   | action  | string | -    | the action the user wants to perform on the resource                   |
| [x]   | object  | string | -    | name and id of the resource. Example: “organization:1”                  |
| [ ]   | depth  | integer | 8    |                                              |
#### Example
Request
```json
{
 “user”: “1”,
 “action”: “push”,
 “object”: “repository:1”
}
```
Response
```json
{
 “can”: true,
 “debug”: “user 1 is a owner of organization 1"
}
```
### Write Custom Tuple
We examined how we created the tuple by listening to the tables. Permify allows to create custom tuples.
**Path:** POST /v1/relationships/write
| Required | Argument     | Type  | Default | Description |
|----------|-------------------|--------|---------|-------------|
| [x]   | entity     | string | -    |       |
| [x]   | object_id     | string | -    |       |
| [x]   | relation     | string | -    |       |
| [ ]   | userset_entity | string | -    |       |
| [x]   | userset_object_id | string | -    |       |
| [ ]   | userset_relation | string | -    |       |
#### Example
Request
```json
{
 “namespace”: “organization”,
 “object_id”: “1",
 “relation”: “admin”,
 “userset_namespace”: “”,
 “userset_object_id”: “1",
 “userset_relation”: “”
}
```
Response
```json
{
 “message”: “success”
}
```
### Delete Tuple
Delete relation tuple.
**Path:** POST /v1/relationships/delete
| Required | Argument     | Type  | Default | Description |
|----------|-------------------|--------|---------|-------------|
| [x]   | entity     | string | -    |       |
| [x]   | object_id     | string | -    |       |
| [x]   | relation     | string | -    |       |
| [ ]   | userset_entity | string | -    |       |
| [x]   | userset_object_id | string | -    |       |
| [ ]   | userset_relation | string | -    |       |
#### Example
Request
```json
{
 “entity”: “organization”,
 “object_id”: “1",
 “relation”: “admin”,
 “userset_entity”: “”,
 “userset_object_id”: “1",
 “userset_relation”: “”
}
```
Response
```json
{
 “message”: “success”
}
```
## Client SDKs
We are building SDKs to make installation easier, leave us a feedback on which SDK we should build first.
## Community
You can join the conversation at our [Discord channel](https://discord.gg/MJbUjwskdH). We love to talk about authorization and access control - we would
love to hear from you :heart:
If you like Permify, please consider giving us a :star:️

<h2 align="left">:heart: Let's get connected:</h2>

<p align="left">
<a href="https://discord.gg/MJbUjwskdH">
 <img alt="guilyx’s Discord" width="50px" src="https://user-images.githubusercontent.com/34595361/178992169-fba31a7a-fa80-42ba-9d7f-46c9c0b5a9f8.png" />
</a>
<a href="https://twitter.com/GetPermify">
  <img alt="guilyx | Twitter" width="50px" src="https://user-images.githubusercontent.com/43545812/144034996-602b144a-16e1-41cc-99e7-c6040b20dcaf.png"/>
</a>
<a href="https://www.linkedin.com/company/permifyco">
  <img alt="guilyx's LinkdeIN" width="50px" src="https://user-images.githubusercontent.com/43545812/144035037-0f415fc7-9f96-4517-a370-ccc6e78a714b.png" />
</a>
</p>


## License

Licensed under the Apache License, Version 2.0: http://www.apache.org/licenses/LICENSE-2.0
