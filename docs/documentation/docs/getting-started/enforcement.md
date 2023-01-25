---
sidebar_position: 4
---

# Access Control Check

In Permify, you can perform access control checks as both [resource specific] and [subject specific] (data filtering) with single API calls.

A simple [resource based] access check takes form of ***Can the subject U perform action X on a resource Y ?***. A real world example would be: *can user:1 edit document:2* where the right side of the ":" represents identifier of the entity.

On the other hand [subject based] access check takes form of  ***Which resources does subject U perform an action X ?*** This option is best for filtering data or bulk permission checks. 

[resource based]: ../api-overview/permission/check-api.md
[subject based]: ../api-overview/permission/lookup-entity.md

## Performance & Availability

Permify designed to answer these authorization questions efficiently and with minimal complexity while providing low latency with:
- Using its parallel graph engine. 
- Storing the relationships between resources beforehand in Permify data store: [writeDB], rather than providing these relationships at “check” time.
- Implementing permission caching to not recompute repeated permission checks, and in memory cache to store authorization schema.
- Using [Snap Tokens](/docs/reference/snap-tokens) to achieve consistency and high performance in cache.

Performance and availability of the API calls - especially access checks - are crucial for us and we're ongoingly improving and testing it with various methods.   

:::info
We would love to create a test environment for you in order to test Permify API and see performance and availability of it. [Schedule a call with one of our Permify engineers](https://meetings-eu1.hubspot.com/ege-aytin/call-with-an-expert).
:::

[writeDB]: ../getting-started/sync-data.md

## How Access Decisions Evaluated?

Access decisions are evaluated by stored [relational tuples] and your authorization model, [Permify Schema]. 

In high level, access of an subject related with the relationships created between the subject and the resource. You can define this relationships in Permify Schema then create and store them as relational tuples, which is basically your authorization data. 

Permify Engine to compute access decision in 2 steps, 
1. Looking up authorization model for finding the given action's ( **edit**, **push**, **delete** etc.) relations.
2. Walk over a graph of each relation to find whether given subject ( user or user set ) is related with the action. 

Let's turn back to above authorization question ( ***"Can the user 3 edit document 12 ?"*** ) to better understand how decision evaluation works. 

[relational tuples]: /docs/getting-started/sync-data
[Permify Schema]:  /docs/getting-started/modeling

When Permify Engine receives this question it directly looks up to authorization model to find document `‍edit` action. Let's say we have a model as follows

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

As we can see above: only users with an admin role in an organization, which `document:12` belongs, and owners of the `document:12` can edit. Permify runs two concurrent queries for **parent.admin** and **owner**:

**Q1:** Get the owners of the `document:12`.

**Q2:** Get admins of the organization where `document:12` belongs to.

Since edit action consist **or** between owner and parent.admin, if Permify Engine found user:3 in results of one of these queries then it terminates the other ongoing queries and returns authorized true to the client.

Rather than **or**, if we had an **and** relation then Permify Engine waits the results of these queries to returning a decision. 

## Need any help ?

:::info
Bulk permission check or with other name data filtering is a common use case we have seen so far. If you have a similar use case we would love to hear from you. Join our [discord](https://discord.gg/JJnMeCh6qP) to discuss or [schedule a call with one of our Permify engineers](https://meetings-eu1.hubspot.com/ege-aytin/call-with-an-expert).
:::