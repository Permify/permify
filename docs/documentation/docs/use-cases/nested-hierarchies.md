
# Nested Hierarchies 

This use case shows solving deeply nested hierarchies with [Permify Schema]. We have a unique **action** usage for nested hierarchies, where parent and child entities can share permissions between them. Let's follow the below team project authorization model to examine this case.

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

## Sample Relational Tuples 

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

## Need any help ?

Our team is happy to help you get started with Permify. If you'd like to learn more about using Permify in your app or have any questions about this example, [schedule a call with one of our Permify engineer](https://meetings-eu1.hubspot.com/ege-aytin/call-with-an-expert).

