# Simple Organizational Role Based Access Control

This example shows how to model simple role based access control for organizational roles and permissions with Permify's DSL, Permify Schema.

-------

## Full Schema

```perm
entity user {} 

entity organization {

    //roles 
    relation admin @user    
    relation member @user    
    relation manager @user    

    //grant access permissions
    action grant_manager_access = admin
    action remove_manager_access = admin
    action grant_admin_access = admin
    action remove_admin_access = admin

    //resource access permissions
    action view_files = admin or manager or member
    action edit_files = admin or manager
    action delete_file = admin 

} 
```

## Schema Deconstruction

### Entities

This examples consist 2 entity, 

- `user`, represents users (maybe corresponds as employees). This entity is empty because its only responsible for referencing users.

```perm
  entity user {}
```

- `organization`, representing organization that user (employees) belongs. It has several roles and permissions related with the spesific resources such as organization files. Moreover, it has permission for spesific action like granting access ability.

### Relations

#### organization entity

To define roles, **relations** needed to be created as entity attributes. In above schema we defined 3 roles respectively; admin, manager and member. 

```perm
entity organization {

    //roles 
    relation admin @user    
    relation member @user    
    relation manager @user    

}
```
### Actions

Actions describe what relations, or relation’s relation can do, think of actions as entities' permissions. Actions defines who can perform a specific action in which circumstances.

Permify Schema supports ***and***, ***or***, ***and not*** and ***or not*** operators to define actions. 

#### organization actions

In this example we define several actions for controling manager and admin grant access abilities, and some resource based 
actions such as organization files and organizations vendor's files.

```perm
entity organization {

    //grant access permissions
    action grant_manager_role = admin
    action remove_manager_role = admin
    action grant_admin_role = admin
    action remove_admin_role = admin

    //resource access permissions
    action view_files = admin or manager or member
    action edit_files = admin or manager and not member
    action delete_file = admin 

} 
```

Let's take a loot at some of actions:

- ``action grant_manager_role = admin``
indicates that only administrator have permission to grant manager role to a user.

- ``action edit_files = admin or manager`` 
indicates that only admin or manager have permission to edit files in organization.


## Example Relational Tuples for this case

organization:2#admin@user:daniel

organization:5#member@user:ashley

organization:17#manager@user:mert

organization:21#member@user:ege

.
.
.

For more details about how relational tuples created and stored your preferred database, see Permify [docs](https://docs.permify.co/docs/relational-tuples).


