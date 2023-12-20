---
sidebar_position: 1
---

# Role Based Access Control

Want to implement roles and permissions in your application? Permify fully covers you at that point. The example below shows how to model simple role based access controls for organizational roles and permissions with our authorization language, [Permify Schema].

[Permify Schema]: ../../getting-started/modeling

Before we get started, here's the final schema that we will create in this tutorial.

```perm
entity user {} 

entity organization {

    //roles 
    relation admin @user    
    relation member @user    
    relation manager @user    
    relation agent @user  

    //organization files access permissions
    action view_files = admin or manager or (member not agent)
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

This schema consists of 2 entities, 

- `user`, represents users (maybe corresponds to employees). This entity is empty because it's only responsible for referencing users.

```perm
  entity user {}
```

- `organization`, represents the organization the user (employees) belongs. It has several roles and permissions related to the specific resources such as organization files and vendor files.

### Relations

#### organization entity

We can use **relations** to define roles. In this example, we have 4 organization wide roles: admin, manager, member, and agent. 

```perm
entity organization {

    //roles 
    relation admin @user    
    relation member @user    
    relation manager @user 
    relation agent @user     

}
```

Roles (relations) can be scoped to different kinds of entities. But for simplicity, we follow a multi-tenancy approach, which demonstrates each organization has its own roles.

### Actions

Actions describe what relations, or relation's relation, can do. You can think of actions as entities' permissions. Actions define who can perform a specific action and in which circumstances.

Permify Schema supports ***and***, ***or***, ***and not*** and ***or not*** operators to define actions. 

#### organization actions

In our schema, we define several actions for controlling access permissions on organization files and organization vendor's files.

```perm
entity organization {

    //organization files access permissions
    action view_files = admin or manager or (member not agent)
    action edit_files = admin or manager
    action delete_file = admin 

    //vendor files access permissions
    action view_vendor_files = admin or manager or agent
    action edit_vendor_files = admin or agent
    action delete_vendor_file = agent

} 
```

let's take a look at some of the actions:

- ``action edit_files = admin or manager`` 
indicates that only the admin or manager has permission to edit files in the organization.

- ``action view_files = admin or manager or (member not agent)``
indicates that the admin, manager, or members (without having the agent role) can view organization files.



## Example Relational Tuples for this case

organization:2#admin@user:daniel

organization:5#member@user:ashley

organization:17#manager@user:mert

organization:21#agent@user:ege

.
.
.

For more details about how relational tuples are created and stored in your preferred database, see [Relational Tuples].

[Relational Tuples]: ../getting-started/sync-data.md

## Need any help ?

Our team is happy to help you get started with Permify. If you'd like to learn more about using Permify in your app or have any questions about this example, [schedule a call with one of our Permify engineers](https://meetings-eu1.hubspot.com/ege-aytin/call-with-an-expert). Alternatively you can join our [discord community](https://discord.com/invite/MJbUjwskdH) to discuss.

