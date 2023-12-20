
# Relationship Based Access Control

Permify was designed and structured as a true [Relational Based Access Control(ReBAC)](https://www.permify.co/post/relational-based-access-control-models/) solution, so besides roles and attributes Permify also supports indirect permission granting through relationships.

Here are some common use cases where you can benefit from using ReBAC models in your Permify Schema.

- [Protecting Organizational-Wide Resources](#protecting-organizational-wide-resources)
- [Deeply Nested Hierarchies](#deeply-nested-hierarchies)
- [User Groups & Team Permissions](#user-groups--team-permissions)

## Protecting Organizational-Wide Resources

This example demonstrates grouping the users by organization and giving them access to organizational-wide resources. 

In this use case we'll follow a simplified version of Github's access control that shows how to model basic repository push, read and delete permissions with our authorization language DSL, [Permify Schema].

[Permify Schema]: ../getting-started/modeling

Before we get started, here's the final schema that we will create in this tutorial.

```perm
entity user {} 

entity organization {

    // organizational roles
    relation admin @user    
    relation member @user    

} 

entity repository {

    // represents repositories parent organization
    relation parent @organization 

    // represents user of this repository
    relation owner  @user           

    // permissions
    action push   = owner
    action read   = owner and (parent.admin or parent.member)
    action delete = parent.admin or owner

} 
```

### Schema Deconstruction

#### Entities

This schema consists of 3 entities, 

- `user`, represents users. This entity is empty because it's only responsible for referencing users.

```perm
  entity user {}
```

- `organization`, represents organization that user and repositories belongs. 

- `repository`, represents a repository in a github.

#### Relations

To define a relation, **relations** need to be created as entity attributes.

##### organization entity

In our schema we defined 2 relations in the organization entity: ``admin`` and ``member``. 

```perm

entity organization {

    relation admin @user    
    relation member @user    

} 

```

``admin`` indicates that the user got an administrative role in that organization and with the same logic ``member`` represents a default user that belongs to that organization.

##### repository entity

Repository entities have 2 relations: ``parent`` and ``owner``. Both of these relations represent actual database relations with other entities rather than a role-based approach similar to the **organization** entity above.

```perm
entity repository {

    relation parent @organization 
    relation owner @user           

} 
```

The ``parent`` relation represents the parent organization of a repository. And ``owner`` represents the specific user, the repository's owner.

#### Actions

Actions describe what relations, or relation's relation, can do. You can think of actions as entities' permissions. Actions define who can perform a specific action and in which circumstances.

Permify Schema supports ***and***, ***or***, ***and not*** and ***or not*** operators to define actions. 

##### repository actions

In our schema, we examined one of the main functionalities user can make on any GitHub repository. These are pushing to the repo, reading & viewing the repo, and deleting that repo. 

We can say only,

- Repository owners can  ``push`` to that repo.
- Repository owners, who have an admin or member role of the parent organization, can ``read``.
- Repository owners or admins of the parent organization can ``delete`` the repository.

```
entity repository {

    action push   = owner
    action read   = owner and (parent.admin or parent.member)
    action delete = parent.admin or owner

} 
```

Since `parent` represents the parent organization of a repository. It can reach repositories parent organization relations with comma. So, 

- ``parent.admin``
indicates admin role on organization

- ``parent.member`` 
indicates member of that organization.

### Sample Relational Tuples 

organization:2#admin@user:daniel

organization:54#member@user:ege

organization:12#member@user:jack

repository:34#parent@organization:54 

repository:68#owner@user:12

repository:12#owner@user:46


.
.
.

For more details about how relational tuples are created and stored in your preferred database, see [Relational Tuples].

[Relational Tuples]: ../getting-started/sync-data.md

For instance, you can define that a user has certain permissions because of their relation to other entities.

An example of this would be granting a manager the same permissions as their subordinates, or giving a user access to a resource because they belong to a certain group. This is facilitated by our relationship-based access control, which allows the definition of complex permission structures based on the relationships between users, roles, and resources.

## Deeply Nested Hierarchies 

This use case shows solving deeply nested hierarchies with the [Permify Schema]. 

We have a unique **action** usage for nested hierarchies, where parent and child entities can share permissions between them. Let's follow the below team project authorization model to examine this case.

[Permify Schema]: ../getting-started/modeling

Before we get started, here's the final schema that we will create in this tutorial.

```perm
entity user {}

entity organization {

    // organization user types
    relation admin @user
}

entity team {
    
    //refers to the organization that a team belongs to 
    relation org @organization

    // Only the organization administrator can edit
    action edit = org.admin
}

entity project {

    //refers to the team that a project belongs to 
    relation team @team

    // This action is responsible for nested permission inheritance
    // team.edit refers to the edit action on the team entity which we defined above 
    // This means that the organization admin, who can edit the team
    // can also edit the project related to the team.
    action edit = team.edit
}
```

### Sample Relational Tuples 

organization:1#admin@user:1

team:1#org@organization:1#...

project:1#team@team:1#...

Lets assume we created the above [relational tuples]. If we try to enforce `Can user:1 edit project:1?` we will get **Allow** since the `user:1` is an admin of the `organization:1` and `project:1` belongs to `team:1`, which belongs to `organization:1`.

[relational tuples]: ../getting-started/sync-data.md

Let's break down this case,

```perm
entity project {

   relation team @team

   action edit = team.edit
}
```

In the above `team.edit` points to the **edit** action in the **team** (that the project belongs to). That edit action on the team entity (`action edit = org.admin`) states that only admins of the **organization (which that team belongs to)** can edit. So our project inherits that action and conducts a result accordingly.

If we go back to our question: `Can user:1 edit project:1?` this will give an **Allow** result, because user:1 is an admin in an organization that the projects' parent team belongs to.

## User Groups & Team Permissions

This use case shows how to organize permissions based on groupings of users or resources. In this use case we'll follow a simple project management app with our authorization language, [Permify Schema].

[Permify Schema]: ../getting-started/modeling

Before we get started, here's the final schema that we will create in this tutorial.

```perm
entity user {}

entity organization {

    //organizational roles
    relation admin @user
    relation member @user

}

entity team {

    // represents owner or creator of the team
	relation owner @user

    // represents direct member of the team
	relation member @user

    // represents the organization that the team belongs to
    relation org @organization

    // organization admins or team owners can edit, delete the team details
    action edit = org.admin or owner
    action delete = org.admin or owner

    // to invite someone you need to be an organization admin and either an owner or member of this team
    action invite = org.admin and (owner or member)

    // only team owners can remove users
    action remove_user =  owner

}

entity project {

    // represents team and organization that a project belongs to
	relation team @team
    relation org @organization

    action view = org.admin or team.member
    action edit = org.admin or team.member
    action delete = team.member

}
```

### Schema Deconstruction

#### Entities

This schema consists of 4 entities, 

- `user`, represents users. This entity is empty because its only responsible for referencing users.

```perm
  entity user {}
```

- `organization`, represents an organization that contain teams.

- `team`, represents teams, which belong to an organization.

- `project`, represents projects that belong to teams.

#### Relations

##### organization entity

We can use **relations** to define roles.

The organization entity has 2 relations ``admin`` and ``member`` users. Think of these as organizational-wide roles.

```perm
entity organization {

    relation admin @user
    relation member @user

}

```

Roles (relations) can be scoped with different kinds of entities. But for simplicity, we follow a multi-tenancy approach, which demonstrates that each organization has its own roles.

##### team entity

The team entity has its own relations respectively,  ``owner``, ``member`` and ``org``

```perm
entity team {

	relation owner @user
	relation member @user
    relation org @organization

}
```

##### project entity

The project entity has ``team`` and ``org`` relations. Both these relations represent parent relationships with other entities, parent team and parent organization.

```perm
entity project {

	relation team @team
    relation org @organization

}
```

#### Actions

Actions describe what relations, or relation's relation, can do. You can think of actions as entities' permissions. Actions define who can perform a specific action and in which circumstances.

Permify Schema supports ***and***, ***or*** and ***not*** operators to define actions. 

##### team actions

- Only organization ***admin (admin role)*** and ***team owner*** can edit and delete team specific resources. 

- Moreover, to invite a colleague to a team you must have an organizational ***admin role*** and either be a ***owner*** or ***member*** of that team. 

- To remove users in team you must be an ***owner*** of that team. 

And these rules are defined in Permify Schema as:

```perm
entity team {

    action edit = org.admin or owner
    action delete = org.admin or owner

    action invite = org.admin and (owner or member)
    action remove_user =  owner

}
```

##### project actions

And here are the project actions. The actions consist of checking access for basic operations such as viewing, editing, or deleting project resources.

```perm
entity project {

    action view = org.admin or team.member
    action edit = org.admin or team.member
    action delete = team.member

}
```

### Sample Relational Tuples 

team:2#member@user:daniel

team:54#owner@user:daniel

organization:12#admin@user:jack

organization:51#member@user:jack

organization:41#member@team:42#member 

project:35#team@team:34#....


.
.
.
.
.


organization:41#member@team:42#member 

**--> represents members of team 42 are also members of organization 41**

project:35#team@team:34#....

**--> represents project 54 is in team 34**

## Need any help on Authorization ?

Our team is happy to help you get started with Permify. If you'd like to learn more about using Permify in your app or have any questions about this example, [schedule a call with one of our Permify engineers](https://meetings-eu1.hubspot.com/ege-aytin/call-with-an-expert). Alternatively you can join our [discord community](https://discord.com/invite/MJbUjwskdH) to discuss.