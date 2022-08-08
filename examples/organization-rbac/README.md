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
    relation agent @user  

    //organization files access permissions
    action view_files = admin or manager or (member and not agent)
    action edit_files = admin or manager
    action delete_file = admin 

    //vendor files access permissions
    action view_vendor_files = admin or manager or agent
    action edit_vendor_files = admin or agent
    action delete_vendor_file = agent

} 
```

## Schema Deconstruction

### Entities

This examples consist 2 entity, 

- `user`, represents users (maybe corresponds as employees). This entity is empty because its only responsible for referencing users.

```perm
  entity user {}
```

- `organization`, representing organization that user (employees) belongs. It has several roles and permissions related with the spesific resources such as organization files and vendor files.

### Relations

#### organization entity

To define roles, **relations** needed to be created as entity attributes. In above schema we defined 4 roles respectively; admin, manager, member and agent. 

```perm
entity organization {

    //roles 
    relation admin @user    
    relation member @user    
    relation manager @user 
    relation agent @user     

}
```
### Actions

Actions describe what relations, or relation’s relation can do, think of actions as entities' permissions. Actions defines who can perform a specific action in which circumstances.

Permify Schema supports ***and***, ***or***, ***and not*** and ***or not*** operators to define actions. 

#### organization actions

In this example we define several actions for controling access permissions on organization files and organizations vendor's files.

```perm
entity organization {

    //organization files access permissions
    action view_files = admin or manager or (member and not agent)
    action edit_files = admin or manager
    action delete_file = admin 

    //vendor files access permissions
    action view_vendor_files = admin or manager or agent
    action edit_vendor_files = admin or agent
    action delete_vendor_file = agent

} 
```

Let's take a loot at some of actions:

- ``action edit_files = admin or manager`` 
indicates that only admin or manager have permission to edit files in organization.

- ``action view_files = admin or manager or (member and not agent)``
indicates that admin, manager or members (without having the agent role) can view organization files.


## Example Relational Tuples for this case

organization:2#admin@user:daniel

organization:5#member@user:ashley

organization:17#manager@user:mert

organization:21#agent@user:ege

.
.
.

For more details about how relational tuples created and stored your preferred database, see Permify [docs](https://docs.permify.co/docs/relational-tuples).

## Need any help ?

Our team is happy to help you get started with Permify. If you'd like to learn more about using Permify in your app or have any questions about this example, [schedule a call with one of our Permify engineer](https://calendly.com/ege-permify/30min).

