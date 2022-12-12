
# User Groups

This use case shows how to organize permissions based on groupings of users or resources. In this use case we'll follow a simple project management app with Permify's DSL, [Permify Schema].

[Permify Schema]: /docs/getting-started/modeling

-------

## Full Schema

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

## Schema Deconstruction

### Entities

This schema consists 4 entity, 

- `user`, represents users. This entity is empty because its only responsible for referencing users.

```perm
  entity user {}
```

- `organization`, represents organization that contain teams.

- `team`, represents teams, which belongs to a organization.

- `project`, represents projects that belongs teams.

### Relations

#### organization entity

We can use **relations** to define roles.

The organization entity has 2 relations ``admin`` and ``member`` users. Think of these as organizational-wide roles.

```perm
entity organization {

    relation admin @user
    relation member @user

}

```

Roles (relations) can be scoped with different kinds of entities. But for simplicity, we follow a multi-tenancy approach, which demonstrates each organization has its own roles.

#### team entity

The eeam entity has its own relations respectively,  ``owner``, ``member`` and ``org``

```perm
entity team {

	relation owner @user
	relation member @user
    relation org @organization

}
```

#### project entity

Project entity has  ``team`` and ``org`` relations. Both these relations represents parent relationship with other entites, parent team and parent organization.

```perm
entity project {

	relation team @team
    relation org @organization

}
```

### Actions

Actions describe what relations, or relation’s relation can do, think of actions as entities' permissions. Actions defines who can perform a specific action in which circumstances.

Permify Schema supports ***and***, ***or***, ***and not*** and ***or not*** operators to define actions. 

#### team actions

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

#### project actions

And there are the project actions below. It consists of checking access for basic operations such as viewing, editing, or deleting project resources.

```perm
entity project {

    action view = org.admin or team.member
    action edit = org.admin or team.member
    action delete = team.member

}
```

## Example Relational Tuples 

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

For more details about how relational tuples created and stored your preferred database, see [Relational Tuples].

[Relational Tuples]: ../getting-started/sync-data.md

## Need any help ?

Our team is happy to help you get started with Permify. If you'd like to learn more about using Permify in your app or have any questions about this example, [schedule a call with one of our Permify engineer](https://meetings-eu1.hubspot.com/ege-aytin/call-with-an-expert).

