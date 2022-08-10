# Basic Project Management Use Case

This example shows how to model simple project management system with Permify's DSL, Permify Schema.

-------

## Full Schema

```perm
entity user {}

entity organization {

    relation admin @user
    relation member @user

}

entity team {

	relation owner @user
	relation member @user
    relation org @organization

    action edit = org.admin or owner
    action delete = org.admin or owner

    action invite = org.admin and (owner or member)
    action remove_user =  owner

}

entity project {

	relation team @team
    relation org @organization

    action view = org.admin or team.member
    action edit = org.admin or team.member
    action delete = team.member

}
```

## Schema Deconstruction

### Entities

This examples consist 4 entity, 

- `user`, represents users. This entity is empty because its only responsible for referencing users.

```perm
  entity user {}
```

- `organization`, represents organization that contain teams.

- `team`, represents teams, which belongs to a organization.

- `project`, represents projects that belongs teams.

### Relations

To define relation, **relations** needed to be created as entity attributes.

#### organization entity

Organization entity has 2 relations ``admin`` and ``member`` users. Think of these as user roles.

```perm
entity organization {

    relation admin @user
    relation member @user

}

```

#### team entity

Team entity has its own relations respectively,  ``owner``, ``member`` and ``org``


```perm
entity team {

	relation owner @user
	relation member @user
    relation org @organization

}
```

#### project entity

Project entity has  ``team`` and ``org`` relations.  ``team``. Both these relations represents parent relationship with other entites, parent team and parent organization.

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

Only organization admin (admin role) and team owner can perform editing and deleting the team item. 

Moreover, for inviting a colleague to a team you must have admin role and either be a owner or member on that team. To remove users in team you must be a owner of that team. And these rules reflects Permify Schema as, 

```perm
entity team {

    action edit = org.admin or owner
    action delete = org.admin or owner

    action invite = org.admin and (owner or member)
    action remove_user =  owner

}
```

#### project actions

And there are the project actions below. It consist checking access for basic operations such as viewving, editing or deleting a project item.

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

For more details about how relational tuples created and stored your preferred database, see Permify [docs](https://docs.permify.co/docs/relational-tuples).

## Need any help ?

Our team is happy to help you get started with Permify. If you'd like to learn more about using Permify in your app or have any questions about this example, [schedule a call with one of our Permify engineer](https://calendly.com/ege-permify/30min).

