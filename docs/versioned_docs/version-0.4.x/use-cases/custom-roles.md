
# Custom Roles

This document highlights a solution for custom roles with [Permify Schema]. In this tutorial, we will create custom **admin** and **member** roles in a project. Then set the permissions of these roles according to their capabilities on the dashboard and tasks.

[Permify Schema]: ../../getting-started/modeling

Before we get started, here's the final schema that we will create in this tutorial.

```perm
entity user {}

entity role {
    relation assignee @user
}

entity dashboard {
    relation view @role#assignee
    relation edit @role#assignee
}

entity task {
    relation view @role#assignee
    relation edit @role#assignee
}
```

This schema encompasses several crucial elements to structure a custom role-based access control system. The role entity serves as a particularly important component, as it enables the creation of multiple custom roles. These roles may vary according to the needs of the application and could include roles like **admin**, **editor**, or **member**, among others.

Once these custom roles have been established, they can be assigned to other entities in the system. Specifically, in this schema, these roles are attached to the dashboard and task entities. Each of these entities, dashboard and task, has pre-defined permissions associated with them. These permissions, defined within the schema or model, could represent various operations such as **view**, **edit**, and so forth.

With this setup, it's possible to map these pre-defined permissions of the dashboard and task entities to the custom roles that have been created. This implies that specific permissions, for instance, **view** and **edit** for a dashboard or a task, could be assigned to a particular custom role.

Based on this model, the example relationships are as follows. With these relationships, custom roles such as **admin** and **member** have been created.

## Relationships

dashboard:project-progress#view@role:admin#assignee

dashboard:project-progress#view@role:member#assignee

dashboard:project-progress#edit@role:admin#assignee

task:website-design-review#view@role:admin#assignee

task:website-design-review#view@role:member#assignee

task:website-design-review#edit@role:admin#assignee

Together with these relationships and the model, a view has been created for the **project-progress** dashboard and the **website-design-review** task as shown in the table below.

| permission         | admin | member  |
|--------------------|-------|---------|
| **dashboard:view** | ✅     | ✅       |
| **dashboard:edit** | ✅     | ⛔       |
| **task:view**      | ✅     | ✅       |
| **task:edit**      | ✅     | ⛔       |


Subsequently, you can make authorization decisions by assigning these custom roles to the users that you have created.

role:member#assignee@user:1

When we write these relationship, the final situation will be as follows.

`Can user:1 view dashboard:project-progress?` gives **Allow** result since the `user:1` is assignee of `role:member` and `role:member` has `dashboard:project-progress#view` permission.

`Can user:1 view task:website-design-review?` gives **Denied** result since the `user:1` is not assignee of `role:admin`.


## Need any help ?

Our team is happy to help you get started with Permify. If you'd like to learn more about using Permify in your app or have any questions about this example, [schedule a call with one of our Permify engineer](https://meetings-eu1.hubspot.com/ege-aytin/call-with-an-expert).

