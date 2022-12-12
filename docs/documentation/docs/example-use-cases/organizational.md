

# Organizations & Hierarchies

Group your users by organization with giving them access organzational-wide resources. In this use case we'll follow a simplified version of Github's access control that shows how to model basic repository push, read and delete permissions with Permify's DSL, [Permify Schema].

[Permify Schema]: /docs/getting-started/modeling

-------

## Full Schema

```perm
entity user {} 

entity organization {

    // organizational roles
    relation admin @user    
    relation member @user    

} 

entity repository {

    // represents repositories parent organization
    relation    parent   @organization 

    // represents user of this repository
    relation    owner    @user           

    // permissions
    action push   = owner
    action read   = owner and (parent.admin or parent.member)
    action delete = parent.admin or owner

} 
```

## Schema Deconstruction

### Entities

This schema consists 3 entities, 

- `user`, represents users. This entity is empty because its only responsible for referencing users.

```perm
  entity user {}
```

- `organization`, represents organization that user and repositories belongs. 

- `repository`, represents a repository in a github.

### Relations

To define relation, **relations** needed to be created as entity attributes.

#### organization entity

In our schema we defined 2 relation in organization entity, respectively; ``admin`` and ``member`` 

```perm

entity organization {

    relation admin @user    
    relation member @user    

} 

```

``admin`` indicates that the user got an administrative role in that organization and with the same logic ``member`` represents the default user that belongs to that organization.

#### repository entity

Repository entities have 2 relations, these are ``parent`` and ``owner``. Both of these relations represents actual database relations with other entities rather than a role-based approach likewise to the **organization** entity above.

```perm
entity repository {

    relation    parent   @organization 
    relation    owner    @user           

} 
```

``parent`` relation represents the parent organization with a repository. And ``owner`` represents the specific user, the repository's owner.

### Actions

Actions describe what relations, or relation’s relation can do, think of actions as entities' permissions. Actions defines who can perform a specific action in which circumstances.

Permify Schema supports ***and***, ***or***, ***and not*** and ***or not*** operators to define actions. 

#### repository actions

In our schema, we examined one of the main functionalities can the user make on any GitHub repository. These are pushing to the repo, reading & viewing the repo, and deleting that repo. 

We can say only,

- Repository owners can  ``push`` to that repo.
- Repository owners, who also need to have an administrative role or be an owner of the parent organization, can ``read``.
- Repository owners or administrative roles in an organization can ``delete`` the repository.

```
entity repository {

    action push   = owner
    action read   = owner and (parent.admin or parent.member)
    action delete = parent.admin or owner

} 
```

Since ``parent` represents the parent organization of repository. It can reach repositories parent organization relations with comma. So, 

- ``parent.admin``
indicates admin role on organization

- ``parent.member`` 
indicates member of that organization.

## Example Relational Tuples 

organization:2#admin@user:daniel

organization:54#member@user:ege

organization:12#member@user:jack

repository:34#parent@organization:54 

repository:68#owner@user:12

repository:12#owner@user:46


.
.
.

For more details about how relational tuples created and stored your preferred database, see [Relational Tuples].

[Relational Tuples]: ../getting-started/sync-data.md

## Need any help ?

Our team is happy to help you get started with Permify. If you'd like to learn more about using Permify in your app or have any questions about this example, [schedule a call with one of our Permify engineer](https://meetings-eu1.hubspot.com/ege-aytin/call-with-an-expert).
