---
sidebar_position: 4
---

# Access Control Check

In Permify, you can perform access control checks as both [resource specific] and [subject specific] (data filtering) with single API calls.

A simple [resource specific] access check takes form of ***Can the subject U perform action X on a resource Y ?***. A real world example would be: can user:1 edit document:2 where the right side of the ":" represents identifier of the entity.

On the other hand [subject specific] access check takes form of  ***Which resources does subject U perform an action X ?*** This option is best for filtering data or bulk permission checks. For example you list some resources with pagination and want to get the exact resource list of user:1 can delete on each page.

[resource specific]: #resource-specific-check-request
[subject specific]: #subject-specific-data-filtering-check-request

## Performance & Availability

Permify designed to answer these authorization questions efficiently and with minimal complexity while providing low latency with:
- Using its parallel graph engine. 
- Storing the relationships between resources and subjects beforehand in Permify data store: [writeDB], rather than providing these relationships at “check” time.
- Using in memory cache to store authorization schema.

Performance and availability of the API calls - especially access checks - are crucial for us and we're ongoingly improving and testing it with various methods.   

:::info
We would love to create a test environment for you in order to test Permify API and see performance and availability of it. [Schedule a call with one of our Permify engineers](https://meetings-eu1.hubspot.com/ege-aytin/call-with-an-expert).
:::

[writeDB]: ../getting-started/sync-data.md

## Resource Specific Check Request

Let's follow an simplified access control decision for examining the resource specific [check] request.

[check]:  https://app.swaggerhub.com/apis-docs/permify/permify/latest

***Can the user 3 edit document 12 ?***

#### Path: 

POST /v1/permissions/check

| Required | Argument | Type | Default | Description |
|----------|----------|---------|---------|-------------------------------------------------------------------------------------------|
| [x]   | entity | object | - | id and type of the entity. Example: document:1”.
| [x]   | action | string | - | the action the subject wants to perform on the resource |
| [x]   | subject | object | - | The user or user set that wants to take the action  |
| [ ]   | depth | integer | 20 | Timeout limit when if recursive database queries got in loop|
| [ ]   | schema_version | string | - | Version of the schema|

#### Request

```json
{
  "entity": {
    "type": "document", 
    "id": "12"
  },
  "action": "edit",
  "subject": {
    "type":"user",
    "id": "3"
  }
}
```

#### Response

```json
{
  "can": false, // check decision
  "decisions": { // decision logs
    "organization:1#admin": {
      "prefix": "not",
      "can": false,
      "err": null
    },
    "document:1#owner": {
      "prefix": "",
      "can": false,
      "err": null
    }
  }
}
```

### How Access Decisions Evaluated?

Access decisions are evaluated by stored [relational tuples] and your authorization model, [Permify Schema]. 

In high level, access of an subject related with the relationships created between the subject and the resource. You can define this relationships in Permify Schema then create and store them as relational tuples, which is basically your authorization data. 

Permify Engine to compute access decision in 2 steps, 
1. Looking up authorization model for finding the given action's ( **edit**, **push**, **delete** etc.) relations.
2. Walk over a graph of each relation to find whether given subject ( user or user set ) is related with the action. 

Let's turn back to above authorization question ( ***"Can the user 3 edit document 12 ?"*** ) to better understand how decision evaluation works. 

[relational tuples]: /docs/getting-started/sync-data
[Permify Schema]:  /docs/getting-started/modeling

When Permify Engine recieves this question it ireclty looks up to authorization model to find document `‍edit` action. Let's say we have a model as follows

```perm
entity user {}
        
entity organization {

    // organizational roles
    relation admin @user
    relation member @user
}

entity document {

    // represents documents parent organization
    relation parent @organization
    
    // represents owner of this document
    relation owner  @user
    
    // permissions
    action edit   = parent.admin or owner
    action delete = owner
} 
```

Which has a directed graph as follows:

![relational-tuples](https://user-images.githubusercontent.com/34595361/193418063-af33fe81-95ed-4615-9d86-b50d4094ad8e.png)

As we can see above: only users with an admin role in an organization, which `document:12` belongs, and owners of the `document:12` can edit. Permify runs two concurent queries for **parent.admin** and **owner**:

**Q1:** Get the owners of the `document:12`.

**Q2:** Get admins of the organization where `document:12` belongs to.

Since edit action consist **or** between owner and parent.admin, if Permify Engine found user:3 in results of one of these queries then it terminates the other ongoing queries and returns authorized true to the client.

Rather than **or**, if we had an **and** relation then Permify Engine waits the results of these queries to returning a decision. 

## Subject Specific (Data Filtering) Check Request

For this access check you can ask questions in form of “Which resources can user:X do action Y?” And you’ll get a SQL query without any conditions (filter, pagination or sorting etc) attached to it. 

You can add conditions depending on your needs after getting the query response. So if you have a list with pagination, after getting the core SQL query from our API request you can add pagination filters to it.

**Example:**

Let's follow an simplified access control decision for examining the subject specific check request.

***Which documents can user:2 edit ?***

#### Path: /v1/permissions/lookup-query

#### Request

```json
{
    "entity": "document",
    "action": "edit",
    "subject": {
        "type": "user",
        "id": "1", 
    }
}
```

#### Response

```sql
  SELECT * FROM docs where IN docs.owner_id = 1
```
### Column Notation 

Let's say we have a model as follows

```perm
entity user {}

entity organization {

   relation admin @user

} `table:organizations`

entity document {

   relation parent @organization `column:parent_id`
   relation owner @user `column:owner_id`
   relation shared @user

   action edit = owner 

} `table:documents`

```

As you notice above we have a new syntax at end of the entity relations: `column:...`. This representation means if you identified the specific relation in the database table (which references the entity) then you can define it to get a leaner SQL query. For example;

```perm
relation owner @user `column:owner_id`
```

indicates that, we have an owner relation in document betwen user. Plus owner of any specific document also kept in real application database document table with **owner_id** column. Since we got owner_id specified, when application client ask a question like 

***Which documents can user:2 edit ?*** 

Permify can return the SQL query as: 

`SELECT * FROM documents where IN docs.owner_id = 1`

On the application side you can use this SQL query with any pagination or sorting conditions as you wish to get related documents.

### Without Specifiying Columns 

What about the scenario we didn't specify the **owner_id** on schema ? Then Permify engine will find which documents are user:2 is owner then return query as:

`SELECT * FROM docs where docs.id IN (1,23,7,8,544,33,4,5,6)`

This query has needed more computational time on the Permify engine side as well as it's not suitable for using with pagination or sorting because Permify engine basically adds all documents that `user:2` can edit.

It's important to define columns of the relations that are already kept in your database to optimize performance and the decreasing the computational effort of the Permify engine.

### How SQL generated?

To show workflow of generating SQL query, let's add one more rule to our edit action and update document entity as:

```perm
entity document {

   relation parent @organization `column:parent_id`
   relation owner @user `column:owner_id`
   relation shared @user

   action edit = owner or parent.admin

} `table:documents`
```

And lets assume we have stored following relational tuples in our writeDB:

```
document:3#owner@user:2
document:1#parent@organization:1#...
organization:1#admin@user:2

```

Where `user:2` is the **admin** of the `organization:1` to which `document:1` belongs. And `user:2` is also the **owner** of `document:3`. To sum up, `user:2` could edit both `document:1` and `document:3`.

And let's ask our access question again; ***Which documents can user:2 edit ?*** 

#### 1) Checking the action permissions on Model

After the API call, our engine first checks the configured Permify Schema ( your authorization model ) and find the push action statement, which is: 

```perm
action edit = owner or parent.admin
```

So, result will be the list of documents that user:2 is either owner or an parent.admin ( parent referencing the organization so its referring the admin relation in organization )

#### 2) Checking the action permissions on Model

If we start with owner action, 

```perm
relation  owner  @user `column:owner_id`
```

Crucial part in here is determining whether the column ( `column:owner_id` ) is defined in the owner relation. So above the column defined as owner_id which indicates the exact same column name in your database. 

#### 3) Building the SQL query from authorization data

Afterwards, our engine will find the column and ask the question. ***which the documents where user:2 is owner?***. Basically it returns the documents that owner id is 2. 

But since the **parent.admin** ( referencing the organizational admin ) can also have access to edit with **or** operation. So we need to extend our SQL a little with the question:

**which the organizations where user:2 is admin?**

And with combining these two questions we come up with a query:

**Semantics :** Return the documents, where `user:2` is the owner and where `user:2` has the admin relation in the organization that these documents belongs.

When we look at the column of parent relation we see the `parent_id:1`, so we need to check the `parent_id` column.

```perm
relation parent @organization `column:parent_id`
```

Only stored data related with this is  `document:1#parent@organization:1#…` So our engine will build the query according to organization.id = 1. 

#### SQL  

```sql
select * from documents INNER JOIN organizations
ON documents.parent_id = organizations.id 
where owner_id = 2 or organizations.id = 1
```

**Note:** We can have multiple relation tuples, especially if you have multi-tenancy with more than one organization in the application. So in that case ***“organizations.id = 1”*** will be updated as

***organizations.id in (2, 12, 57, etc)***

Finally, our API endpoint returns the result SQL. This endpoint will return a query without any sort, filter, etc. After you got the response,  you can customize the query with the needed conditions, like below:

```sql
select * from documents INNER JOIN organizations
ON documents.parent_id = organizations.id 
where owner_id = 2 or organizations.id = 1 limit 10 offset 15
```
In this way, we can decouple the condition functionalities from the data filtering. 

:::info
Bulk permission check or with other name data filtering is a common use case we have seen so far. If you have a similar use case we would love to hear from you. Join our [discord](https://discord.gg/JJnMeCh6qP) to discuss or [schedule a call with one of our Permify engineers](https://meetings-eu1.hubspot.com/ege-aytin/call-with-an-expert).
:::