# Github Example Access Control 

This example shows how to model basic github push, read and delete repository access control with Permify's DSL, Permify Schema.

-------

## Full Schema

```perm
entity user {} 

entity organization {

    relation admin @user    
    relation member @user    

} 

entity repository {

    relation    parent   @organization 
    relation    owner    @user           

    action push   = owner
    action read   = owner and (parent.admin or parent.member)
    action delete = parent.admin or owner

} 
```

## Schema Deconstruction

### Entities

This examples consist 2 entity, 

- `user`, represents users. This entity is empty because its only responsible for referencing users.

```perm
  entity user {}
```

- `organization`, represents organization that user and repositories belongs. 

- `repository`, represents a repository in a github.

### Relations

To define relation, **relations** needed to be created as entity attributes.

#### organization entity

In above schema we defined 2 relation in organization entity, respectively; ``admin`` and ``member`` 

```perm

entity organization {

    relation admin @user    
    relation member @user    

} 

```

``admin`` indicates that the user got an administrative role in that organization and with the same logic ``member`` represents the default user that belongs to that organization.

#### repository entity

Repository entity have 2 relations too, these are ``parent`` and ``owner``. Boht of these relations represents actual database elations with other entities rather than role based approach likewise in organization entity.

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

In this example we examined one of the main functionalities can user made on any github repository. These are pushing to repo, reading & viewving repo and deleting that repo. If we think that is a private repository,

We can say only,

- Repository owners can  ``push`` to that repo.
- Repository owners, whom is also need to have administrative role or be an owner of parent organization, can ``read``.
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

For more details about how relational tuples created and stored your preferred database, see Permify [docs](https://docs.permify.co/docs/relational-tuples).

## Need any help ?

Our team is happy to help you get started with Permify. If you'd like to learn more about using Permify in your app or have any questions about this example, [schedule a call with one of our Permify engineer](https://calendly.com/ege-permify/30min).
