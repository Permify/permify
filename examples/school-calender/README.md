# Simple School Calender Use Case

This example shows how to model simple school calender with Permify's DSL, Permify Schema.

-------

## Full Schema

```perm
entity user {}

entity student {
	relation self @user
	relation teacher @user
}

entity class {
	relation member @student

    action view_calender = member.self or member.teacher
}
```

## Schema Deconstruction

### Entities

This examples consist 2 entity, 

- `user`, represents users. This entity is empty because its only responsible for referencing users.

```perm
  entity user {}
```

- `student`, representing students. student and class entities have many to many relation.

- `class`, representing class that students belongs.

### Relations

To define relation, **relations** needed to be created as entity attributes.

#### student entity

In above schema we defined 2 relation in user entity, respectively; ``self`` and ``student`` 

```perm
entity student {

	relation self @user
	relation teacher @user

}

```

Self describes student itself as a user and teacher represents the teacher that student take class from. 

#### class entity

Class entity has only one relation, which is ``member``. It represents the member of the class. Basically, students that taking that spesific class.

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

can access to calender of that spesific class.

```perm
entity class {

   action view_calender = member.self or member.teacher

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

For more details about how relational tuples created and stored your preferred database, see Permify [docs](https://docs.permify.co/docs/relational-tuples).


