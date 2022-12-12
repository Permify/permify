
# Parent Child Relationships

See how parent child relations can model in Permify. In this use case we'll follow a simple student-calendar management system with Permify's DSL, [Permify Schema].

[Permify Schema]: /docs/getting-started/modeling

-------

## Full Schema

```perm
entity user {}

entity student {

	// refers student itself
	relation self @user

	// teacher of the student
	relation teacher @user
}

entity class {

	// refers class member
	relation member @student

	// calender view permission
    action view_calendar = member.self or member.teacher
}
```

## Schema Deconstruction

### Entities

This schema consists 3 entity, 

- `user`, represents users. This entity is empty because it's only responsible for referencing users.

```perm
  entity user {}
```

- `student`, representing students. student and class entities have many to many relation.

- `class`, representing class that students belongs.

### Relations

To define relation, **relations** needed to be created as entity attributes.

#### student entity

In our schema above, we defined 2 relation in user entity, respectively; ``self`` and ``student`` 

```perm
entity student {

	relation self @user
	relation teacher @user

}

```

**self** describes student itself, and **teacher** represents the teacher that students take a class from. 

#### class entity

The class entity has only one relation, which is ``member``. It represents the member of the class. Basically, students whom taking that specific class.

```perm
entity class {

	relation member @student

}
```

### Actions

Actions describe what relations, or relation’s relation can do, think of actions as entities' permissions. Actions defines who can perform a specific action in which circumstances.

Permify Schema supports ***and***, ***or***, ***and not*** and ***or not*** operators to define actions. 

#### class actions

Think each class has a calendar that shows necessary times and dates of lectures, deadlines etc.

We want only,

- Students that take that class 
- Teachers, whom is teacher of the student that takes that specific class (class member). 

can access to calendar of that spesific class.

```perm
entity class {

   action view_calendar = member.self or member.teacher

}
```

Since ``member` represents the relation with student entitiy. It can reach its relations with comma. So, 

- ``member.self``
indicates student itself, whom takes that class.

- ``member.teacher`` 
indicates teacher of student, whom takes that class.

## Example Relational Tuples 

student:2#self@user:daniel

student:54#self@user:daniel

student:12#teacher@user:jack

student:34#teacher@user:jack

class:68#member@student:54

class:12#member@student:34


.
.
.

For more details about how relational tuples created and stored your preferred database, see [Relational Tuples].

[Relational Tuples]: ../getting-started/sync-data.md

## Need any help ?

Our team is happy to help you get started with Permify. If you'd like to learn more about using Permify in your app or have any questions about this example, [schedule a call with one of our Permify engineer](https://meetings-eu1.hubspot.com/ege-aytin/call-with-an-expert).

