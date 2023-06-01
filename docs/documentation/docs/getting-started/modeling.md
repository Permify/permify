---
sidebar_position: 1
---

# Modeling Authorization

Permify has its own language that you can model your authorization logic with it. The language allows to define arbitrary relations between users and objects, such as owner, editor, commenter or roles like user types such as admin, manager, member, etc.

![modeling-authorization](https://raw.githubusercontent.com/Permify/permify/master/assets/permify-dsl.gif)

## Permify Schema

You can define your entities, relations between them and access control decisions with using Permify Schema. It includes set-algebraic operators such as intersection and union for specifying potentially complex access control policies in terms of those user-object relations.

Here’s a simple breakdown of our schema.

![permify-schema](https://user-images.githubusercontent.com/34595361/183866396-9d2850fc-043f-4254-aa4c-ee2c4172afb8.png)

Permify Schema can be created on our [playground](https://play.permify.co/) as well as in any IDE or text editor. We also have a [VS Code extension](https://marketplace.visualstudio.com/items?itemName=Permify.perm) to ease modeling Permify Schema with code snippets and syntax highlights. Note that on VS code the file with extension is **_".perm"_**.

## Developing a Schema

This guide will show how to develop a Permify Schema from scratch with a simple example, yet it will show almost every aspect of our modeling language.

We'll follow a simplified version of github access control system. To see completed model you can jump directly to [Github Example](#github-example).

:::info
You can start developing Permify Schema on [VSCode]. You can install the extension by searching for **Perm** in the extensions marketplace.

[vscode]: https://marketplace.visualstudio.com/items?itemName=Permify.perm

:::

### Defining Entities

The very first step to build Permify Schema is creating your Entities. Entity is an object that defines your resources that held role in your permission system.

Think of entities as tables in your relationship database. We are strongly advice to name entities same as your database table name that its corresponds. In that way you can easily model and reason your authorization as well as eliminating the error possibility.

You can create entities using `entity` keyword. Since we're following example of simplified github access control, lets create some of our entities as follows.

```perm
entity user {}

entity organization {}

entity team {}

entity repository {}
```

Entities has 2 different attributes. These are;

- **relations**
- **actions (or permissions)**

### Defining Relations

Relations represent relationships between entities. It's probably the most critical part of the schema because Permify mostly based on relations between resources and their permissions. Keyword **_relation_** need to used to create a entity relation with name and type attributes.

**Relation Attributes:**

- **name:** relation name.
- **type:** relation type, basically the entity it’s related to (e.g. user, organization, document, etc.)

An example relation takes form of,

```perm
relation [name] @[type]
```

Lets turn back to our example and define our relations inside our entities:

#### User Entity

→ The user entity is a mandatory entity in Permify. It generally will be empty but it will used a lot in other entities as a relation type to referencing users.

```perm
entity user {}
```

#### Organization Entity

→ For the sake of simplicity let's define only 2 user types in an organization, these are administrators and direct members of the organization.

```perm
entity organization {

    relation admin  @user
    relation member @user

}
```

#### Team Entity

→ Let's say teams can belong organizations and can have a member inside of it as follows,

```perm
entity team {

    relation parent  @organization
    relation member  @user

}
```

The parent relation is indicating the organization the team belongs to. This way we can achieve **parent-child relationship** inside this entity.

#### Repository Entity

→ Organizations and users can have multiple repositories, so each repository is related with an organization and with users. We can define repository relations as as follows.

```perm
entity repository {

    relation  parent @organization

    relation  owner  @user
    relation  maintainer @user @team#member

}
```

The owner relation indicates the creator of the repository, that way we can achieve **ownership** in Permify.

**Defining Multiple Relation Types**

As you can see we have new syntax above,

```perm
    relation maintainer @user @team#member
```

When we look at the maintainer relation, it indicates that the maintainer can be an `user` as well as this user can be a `team member`.

:::info
You can use **#** to reach entities relation. When we look at the `@team#member` it specifies that if the user has a relation with the team, this relation can only be the `member`. We called that feature locking, because it basically locks the relation type according to the prefixed entity.

Actual purpose of feature locking is to giving ability to specify the sets of users that can be assigned.

For example:

```perm
    relation viewer @user
```

When you define it like this, you can only add users directly as tuples (you can find out what relation tuples is in next section):

- organization:1#viewer@user:U1
- organization:1#viewer@user:U2

However, if you define it as:

```perm
    relation viewer @user @organization#member
```

You will then be able to specify not only individual users but also members of an organization:

- organization:1#viewer@user:U1
- organization:1#viewer@user:U2
- organization:1#viewer@organization:O1#member

You can think of these definitions as a precaution taken against creating undesired user set relationships.
:::

Defining multiple relation types totally optional. The goal behind it to improve validation and reasonability. And for complex models, it allows you to model your entities in a more structured way.

### Defining Actions and Permissions

Actions describe what relations, or relation’s relation can do. Think of actions as permissions of the entity it belongs. So actions defines who can perform a specific action on a resource in which circumstances. So, the basic form of authorization check in Permify is **_Can the user U perform action X on a resource Y ?_**.

#### Intersection and Exclusion

The Permify Schema supports **`and`**, **`or`** and **`not`** operators to achieve permission **intersection** and **exclusion**. The keywords **_action_** or **_permission_** can be used with those operators to form rules for your authorization logic.

Lets get back to our github example and create some permissions on repository entity,

```perm
entity repository {

    relation  parent   @organization

    relation  owner @user
    relation  maintainer @user @team#member

    ..
    ..

    action push = owner

}
```

→ `action push = owner or maintainer` indicates only the repository owner or maintainers can push to
repository.

:::info
The same `push` can also be defined using the **permission** keyword, as follows:

```perm
permission push = owner
```

Using `action` and `permission` will yield the same result for defining permissions in your authorization logic.

The reason we have two keywords for defining permissions is that while most permissions are based on actions (such as view, read, edit, etc.), there are still cases where we need to define permissions based on roles or user types, such as admin or member.

Additionally, there may be permissions that need to be inherited by child entities. Using the `permission` keyword in these cases is more convenient and provides better reasoning of the schema.

See **Real World Examples Section** for examining the contextualize difference between using permission and action keyword. You can start with observing [Google Docs Simplified](./examples/google-docs.md) example.
:::

For this tutorial we'll continue with `action` keyword,

```perm
entity repository {

    relation  parent   @organization

    relation  owner @user
    relation  maintainer @user @team#member


    ..
    ..

    action read =  org.admin and (owner or maintainer or org.member)

}
```

→ Let's examine the `read` action rules; user that is `organization admin` and following users can read the repository: `owner` of the repository, or `maintainer`, or `member` of the organization which repository belongs to.

:::info Permission Union 

Permify allows you to set permissions that are effectively the union of multiple permission sets. 

You can define permissions as relations to union all rules that permissions have. Here is an simple demonstration how to achieve permission union in our DSL, you can use actions (or permissions) when defining another action (or permission) like relations,

```perm
   action edit =  member or manager
   action delete =  edit or org.admin
```

The `delete` action inherits the rules from the `edit` action. By doing that, we'll be able to state that only organization administrators and any relation capable of performing the edit action (member or manager) can also perform the delete action.

Permission union is super beneficial in scenarios where a user needs to have varied access across different departments or roles.
::: 

### Full Schema

Here is full implementation of simple Github access control example with using Permify Schema.

```perm
entity user {}

entity organization {

    relation admin @user
    relation member @user

    action create_repository = admin or member
    action delete = admin

}

entity team {

    relation parent  @organization
    relation member  @user

    action edit = member or parent.admin

}

entity repository {

    relation parent @organization

    relation owner @user
    relation maintainer @user @team#member

    action push   = owner or maintainer
    action read   = (owner or maintainer or parent.member) and parent.admin
    action delete = parent.admin or owner

}
```

## Permission Capabilities

1. **Permission Union:**

Permify allows you to set permissions that are effectively the union of multiple permission sets. For example, if a user belongs to multiple roles, each with their own permissions, the user’s effective permissions will be the union of all permissions of the roles they belong to. This is beneficial in scenarios where a user needs to have varied access across different departments or roles.

2. **Permission Indirection: (ReBAC)**

Permify supports indirect permission granting through relationships. For instance, you can define that a user has certain permissions because of their relation to other entities.

An example of this would be granting a manager the same permissions as their subordinates, or giving a user access to a resource because they belong to a certain group. This is facilitated by our relationship-based access control, which allows the definition of complex permission structures based on the relationships between users, roles, and resources.

We utilize ReBAC and Google Zanzibar to create natural linkage between business units, functions, and entities of an organization.

## Real World Examples

This example shows almost all aspects of the Permify Schema.

You can check out more schema examples from the [Real World Examples](./examples) section with their detailed examination.
