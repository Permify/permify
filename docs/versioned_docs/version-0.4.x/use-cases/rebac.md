
# Relationship Based Access Control

Permify has designed and structured as a true [Relationship Based Access Control(ReBAC)](https://permify.co/post/relationship-based-access-control-rebac/) solution, so besides roles and attributes Permify also supports indirect permission granting through relationships.

Here are some common use cases where you can benefit from using ReBAC models in your Permify Schema.

- [Protecting Organizational-Wide Resources](#protecting-organizational-wide-resources)
- [Deeply Nested Hierarchies](#deeply-nested-hierarchies)
- [User Groups & Team Permissions](#user-groups--team-permissions)

## Protecting Organizational-Wide Resources

This example demonstrate grouping the users by organization with giving them access organizational-wide resources. 

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

This schema consists 3 entities, 

- `user`, represents users. This entity is empty because its only responsible for referencing users.

```perm
  entity user {}
```

- `organization`, represents organization that user and repositories belongs. 

- `repository`, represents a repository in a github.

#### Relations

To define relation, **relations** needed to be created as entity attributes.

##### organization entity

In our schema we defined 2 relation in organization entity, respectively; ``admin`` and ``member`` 

```perm

entity organization {

    relation admin @user    
    relation member @user    

} 

```

``admin`` indicates that the user got an administrative role in that organization and with the same logic ``member`` represents the default user that belongs to that organization.

##### repository entity

Repository entities have 2 relations, these are ``parent`` and ``owner``. Both of these relations represents actual database relations with other entities rather than a role-based approach likewise to the **organization** entity above.

```perm
entity repository {

    relation    parent   @organization 
    relation    owner    @user           

} 
```

``parent`` relation represents the parent organization with a repository. And ``owner`` represents the specific user, the repository's owner.

#### Actions

Actions describe what relations, or relation’s relation can do, think of actions as entities' permissions. Actions defines who can perform a specific action in which circumstances.

Permify Schema supports ***and***, ***or***, ***and not*** and ***or not*** operators to define actions. 

##### repository actions

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

For more details about how relational tuples created and stored your preferred database, see [Relational Tuples].

[Relational Tuples]: ../getting-started/sync-data.md

For instance, you can define that a user has certain permissions because of their relation to other entities.

An example of this would be granting a manager the same permissions as their subordinates, or giving a user access to a resource because they belong to a certain group. This is facilitated by our relationship-based access control, which allows the definition of complex permission structures based on the relationships between users, roles, and resources.

## Deeply Nested Hierarchies 

This use case shows solving deeply nested hierarchies with [Permify Schema]. 

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
    
    //refers to organization that team belongs to 
    relation org @organization

    // Only the organization administrator can edit
    action edit = org.admin
}

entity project {

    //refers to team that project belongs to 
    relation team @team

    // This action responsible for nested permission inheritance
    // team.edit refers edit action on the team entity which we defined above 
    // Semantics of this is: Only the organization administrator, who has the 
    // team, to which this project belongs can edit.
    action edit = team.edit
}
```

### Sample Relational Tuples 

organization:1#admin@user:1

team:1#org@organization:1#...

project:1#team@team:1#...

Lets assume we created above [relational tuples]. If we try to enforce `Can user:1 edit project:1?` we will get **Allow** result since the `user:1` is organizational admin and `project:1` belongs to `team:1`, which belongs to `organization:1`.

[relational tuples]: ../getting-started/sync-data.md

Let's break down this case,

```perm
entity project {

   relation team @team

   action edit = team.edit
}
```

Above `team.edit` points out the **edit** action in the **team** (that project belongs to). 

And edit action on the team entity: `action edit = org.admin` states that only **organization (which that team belongs to) admins** can edit. So our project inherits that action and conducts a result accordingly.

If we roll back to our enforcement: `Can user:1 edit project:1?` gives **Allow** result, because user:1 is admin in an organization that the projects' parent team belongs to.

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

    // reference for organization that team belong
    relation org @organization

    // organization admins or owners can edit, delete the team details
    action edit = org.admin or owner
    action delete = org.admin or owner

    // to invite someone you need to be admin and either owner or member of this team
    action invite = org.admin and (owner or member)

    // only owners can remove users
    action remove_user =  owner

}

entity project {

    // references for team and organization that project belongs
	relation team @team
    relation org @organization

    action view = org.admin or team.member
    action edit = org.admin or team.member
    action delete = team.member

}
```

### Schema Deconstruction

#### Entities

This schema consists 4 entity, 

- `user`, represents users. This entity is empty because its only responsible for referencing users.

```perm
  entity user {}
```

- `organization`, represents organization that contain teams.

- `team`, represents teams, which belongs to a organization.

- `project`, represents projects that belongs teams.

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

Roles (relations) can be scoped with different kinds of entities. But for simplicity, we follow a multi-tenancy approach, which demonstrates each organization has its own roles.

##### team entity

The eeam entity has its own relations respectively,  ``owner``, ``member`` and ``org``

```perm
entity team {

	relation owner @user
	relation member @user
    relation org @organization

}
```

##### project entity

Project entity has  ``team`` and ``org`` relations. Both these relations represents parent relationship with other entites, parent team and parent organization.

```perm
entity project {

	relation team @team
    relation org @organization

}
```

#### Actions

Actions describe what relations, or relation’s relation can do, think of actions as entities' permissions. Actions defines who can perform a specific action in which circumstances.

Permify Schema supports ***and***, ***or*** and ***not*** operators to define actions. 

##### team actions

- Only organization ***admin (admin role)*** and ***team owner*** can perform editing and deleting team spesific resources. 

- Moreover, for inviting a colleague to a team you must have ***admin role*** and either be a ***owner*** or ***member*** on that team. 

- To remove users in team you must be a ***owner*** of that team. 

And these rules reflects Permify Schema as:

```perm
entity team {

    action edit = org.admin or owner
    action delete = org.admin or owner

    action invite = org.admin and (owner or member)
    action remove_user =  owner

}
```

##### project actions

And there are the project actions below. It consists of checking access for basic operations such as viewing, editing, or deleting project resources.

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

organiation:41#member@team:42#member 

project:35#team@team:34#....


.
.
.
.
.


organization:41#member@team:42#member 

**--> represents members of team 42 also members in organization 41**

project:35#team@team:34#....

**--> represents project 54 is in team 34**

## Need any help on Authorization ?

Our team is happy to help you anything about authorization. If you'd like to learn more about using Permify in your app or have any questions, [schedule a call with one of our founders](https://meetings-eu1.hubspot.com/ege-aytin/call-with-an-expert).