---
sidebar_position: 1
---

# Modeling Authorization

Permify was designed and structured as a true ReBAC solution, so besides roles and attributes Permify also supports indirect permission granting through relationships. 

With Permify, you can define that a user has certain permissions because of their relation to other entities. An example of this would be granting a manager the same permissions as their subordinates, or giving a user access to a resource because they belong to a certain group. 

This is facilitated by our relationship-based access control, which allows the definition of complex permission structures based on the relationships between users, roles, and resources.

## Permify Schema

Permify has its own language that you can model your authorization logic with it. The language allows to define arbitrary relations between users and objects, such as owner, editor, commenter or roles like admin, manager, member and also dynamic attributes such as boolean variables, IP range, time period, etc.

![modeling-authorization](https://raw.githubusercontent.com/Permify/permify/master/assets/permify-dsl.gif)

You can define your entities, relations between them and access control decisions with using Permify Schema. It includes set-algebraic operators such as intersection and union for specifying potentially complex access control policies in terms of those user-object relations.

Here’s a simple breakdown of our schema.

![permify-schema](https://user-images.githubusercontent.com/34595361/183866396-9d2850fc-043f-4254-aa4c-ee2c4172afb8.png)

Permify Schema can be created on our [playground](https://play.permify.co/) as well as in any IDE or text editor. We also have a [VS Code extension](https://marketplace.visualstudio.com/items?itemName=Permify.perm) to ease modeling Permify Schema with code snippets and syntax highlights. Note that on VS code the file with extension is **_".perm"_**.

## Developing a Schema

This guide will show how to develop a Permify Schema from scratch with a simple example, yet it will show almost every aspect of our modeling language.

We'll follow a simplified version of the GitHub access control system, where teams and organizations have control over the viewing, editing, or deleting access rights of repositories. 

Before start I want to share the  full implementation of simple Github access control example with using Permify Schema.

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
    action read =  org.admin and (owner or maintainer or org.member)
    action delete = parent.admin or owner

}
```

:::info
You can start developing Permify Schema on [VSCode]. You can install the extension by searching for **Perm** in the extensions marketplace.

[vscode]: https://marketplace.visualstudio.com/items?itemName=Permify.perm

:::

## Defining Entities 

The very first step to build Permify Schema is creating your Entities. Entity is an object that defines your resources that held role in your permission system.

Think of entities as tables in your database. We are strongly advice to name entities same as your database table name that its corresponds. In that way you can easily model and reason your authorization as well as eliminating the error possibility.

You can create entities using `entity` keyword. Let's create some entities according to our example GitHub authorization logic."

```perm
entity user {}

entity organization {}

entity team {}

entity repository {}
```

Entities has 2 different attributes. These are;

- **relations**
- **actions or permissions**

## Defining Relations

Relations represent relationships between entities. It's probably the most critical part of the schema because Permify mostly based on relations between resources and their permissions. 

Keyword **_relation_** need to used to create a entity relation with name and type attributes.

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

### Roles and User Types

You can define user types and roles within the entity. If you specifically want to define a global role, such as `admin`, we advise defining it at the entity with the most global hierarchy, such as an organization. Then, spread it to the rest of the entities to include it within permissions. 

For the sake of simplicity, let's define only 2 user types in an organization, these are administrators and direct members of the organization.

```perm
entity organization {

    relation admin  @user
    relation member @user

}
```

### Parent-Child Relationship

→ Let's say teams can belong organizations and can have a member inside of it as follows,

```perm
entity organization {

    relation admin  @user
    relation member @user

}

entity team {

    relation parent  @organization
    relation member  @user

}
```

The parent relation is indicating the organization the team belongs to. This way we can achieve **parent-child relationship** within these entities.

### Ownership

In Github workflow, organizations and users can have multiple repositories, so each repository is related with an organization and with users. We can define repository relations as as follows.

```perm
entity repository {

    relation  parent @organization

    relation  owner  @user
    relation  maintainer @user @team#member

}
```

The owner relation indicates the creator of the repository, that way we can achieve **ownership** in Permify.

### Multiple Relation Types

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


With `organization:1#viewer@organization:O1#member` all members of the organization O1 will have the right to perform the relevant action.

In other words, all members in O1 now end up having the relevant `viewer` relation.

You can think of these definitions as a precaution taken against creating undesired user set relationships.
:::

Defining multiple relation types totally optional. The goal behind it to improve validation and reasonability. And for complex models, it allows you to model your entities in a more structured way.

## Defining Actions and Permissions

Actions describe what relations, or relation’s relation can do. Think of actions as permissions of the entity it belongs. So actions defines who can perform a specific action on a resource in which circumstances. 

The basic form of authorization check in Permify is **_Can the user U perform action X on a resource Y ?_**.

### Intersection and Exclusion

The Permify Schema supports **`and`**, **`or`** and **`not`** operators to achieve permission **intersection** and **exclusion**. The keywords **_action_** or **_permission_** can be used with those operators to form rules for your authorization logic.

#### Intersection

Lets get back to our github example and create a read action on repository entity to represent usage of **`and`** &, **`or`** operators,

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

→ If we examine the `read` action rules; user that is `organization admin` and following users can read the repository: `owner` of the repository, or `maintainer`, or `member` of the organization which repository belongs to.

:::info Permission Keyword
The same `read` can also be defined using the **permission** keyword, as follows:

```perm
 permission read =  org.admin and (owner or maintainer or org.member)
```

Using `action` and `permission` will yield the same result for defining permissions in your authorization logic. See why we have 2 keywords for defining an permission from the [Nested Hierarchies](#nested-hierarchies) section.
:::

#### Exclusion

After this point, we'll move beyond the GitHub example and explore more advanced abilities of Permify DSL. 

Before delving into details, let's examine the **`not`** operator and conclude [Intersection and Exclusion](#intersection-and-exclusion) section.

Here is the **post** entity from our sample [Instagram Authorization Structure](./examples/google-docs.md)example,

```perm
entity post {
    // posts are linked with accounts.
    relation account @account

    // comments are limited to people followed by the parent account.
    attribute restricted boolean

    ..
    ..

    // users can comment and like on unrestricted posts or posts by owners who follow them.
    action comment = account.following not restricted
    action like = account.following not restricted
}
```

As you can see from the comment and like actions, a user tagged with the `restricted` attribute — details of defining attributes can be found in the [Attribute Based Permissions (ABAC)](#attribute-based-permissions-abac) section — won't be able to like or comment on the specific post.

This is a simple example to demonstrate how you can exclude users, resources, or any subject from permissions using the **`not`** operator.

### Permission Union

Permify allows you to set permissions that are effectively the union of multiple permission sets. 

You can define permissions as relations to union all rules that permissions have. Here is an simple demonstration how to achieve permission union in our DSL, you can use actions (or permissions) when defining another action (or permission) like relations,

```perm
   action edit =  member or manager
   action delete =  edit or org.admin
```

The `delete` action inherits the rules from the `edit` action. By doing that, we'll be able to state that only organization administrators and any relation capable of performing the edit action (member or manager) can also perform the delete action.

Permission union is super beneficial in scenarios where a user needs to have varied access across different departments or roles.

### Nested Hierarchies 

The reason we have two keywords for defining permissions (`action` and `permission`) is that while most permissions are based on actions (such as view, read, edit, etc.), there are still cases where we need to define permissions based on roles or user types, such as admin or member.

Additionally, there may be permissions that need to be inherited by child entities. Using the `permission` keyword in these cases is more convenient and provides better reasoning of the schema. 

Here is a simple example to demonstrate inherited permissions.

Let's examine a small snippet from our [Facebook Groups](./examples/google-docs.md) real world example. Let's create a permission called 'view' in the comment entity (which represents the comments of the post in Facebook Groups)

Users can only view a comment if:

- The user is the owner of that comment 
**or**
- The user is a member of the group to which the comment's post belongs.

```perm
// Represents a post in a Facebook group
entity post {

    ..
    ..

    // Relation to represent the group that the post belongs to
    relation group @group

    // Permissions for the post entity
    
    ..
    ..
    permission group_member = group.member
}

// Represents a comment on a post in a Facebook group
entity comment {

    // Relation to represent the owner of the comment
    relation owner @user

    // Relation to represent the post that the comment belongs to
    relation post @post
    relation comment @comment

    ..
    ..

    // Permissions 
    action view = owner or post.group_member

    ..
    ..
}
```

The `post.group_member` refers to the members of the group to which the post belongs. We defined it as action in **post** entity as,

```perm
permission group_member = group.member
```

Permissions can be inherited as relations in other entities. This allows to form nested hierarchical relationships between entities. 

In this example, a comment belongs to a post which is part of a group. Since there is a **'member'** relation defined for the group entity, we can use the **'group_member'** permission to inherit the **member** relation from the group in the post and then use it in the comment.

### Recursive ReBAC

With Permify DSL, you can define recursive relationship-based permissions within the same entity.

As an example, consider a system where there are multiple organizations within a company, some of which may have a parent-child relationship between them.

As expected, organization members are also granted permission to view their organization details. You can model that as follows: 

```perm
entity user {}

entity organization {
    relation parent @organization
    relation member @user @organization#member

    action view = member or parent.member
}
```

Let's extend the scenario by adding a rule allowing parent organization members to view details of child organizations. Specifically, a member of **Organization Alpha** could view the details of **Organization Beta** if **Organization Beta** belongs to **Organization Alpha**.

![modeling-authorization](https://user-images.githubusercontent.com/58391988/279456032-485a0aef-b83b-4257-af48-0fcbe6fa2e64.png)

First authorization schema that we provide won't solve this issue because `parent.member` accommodate single upward traversal in a hierarchy. 

Instead of `parent.member` we can call the parent view permission on the same entity - `parent.view` to achieve multiple levels of upward traversal, as follows: 

```perm
entity user {}

entity organization {
    relation parent @organization
    relation member @user @organization#member

    action view = member or parent.view
}
```

This way, we achieve a recursive relationship between parent-child organizations.

:::note
*Credits to [Léo](https://github.com/LeoFVO) for the illustration and for [highlighting](https://github.com/Permify/permify/issues/790) this use case.*
:::

## Attribute Based Permissions (ABAC)

To support Attribute Based Access Control (ABAC) in Permify, we've added two main components into our schema language: `attribute` and `rule`.

### Defining Attributes

Attributes are used to define properties for entities in specific data types. For instance, an attribute could be an IP range associated with an organization, defined as a string array:

```perm
attribute ip_range string[]
```

Here are the all attribute types that you use when defining an `attribute`.

```perm
// A boolean attribute type
boolean

// A boolean array attribute type.
boolean[]

// A string attribute type.
string

// A string array attribute type.
string[]

// An integer attribute type.
integer 

// An integer array attribute type.
integer[]

// A double attribute type.
double

// A double array attribute type.
double[]
```

### Defining Rules

Rules are structures that allow you to write specific conditions for the model. You can think rules as simple functions of every software language have. They accept parameters and are based on condition to return a true/false result.

In the following example schema, a rule could be used to check if a given IP address falls within a specified IP range:

```perm
entity user {}

entity organization {
	
	relation admin @user

	attribute ip_range string[]

	permission view = check_ip_range(request.ip, ip_range) or admin
}

rule check_ip_range(ip string, ip_range string[]) {
	ip in ip_range
}
```

:::info Syntax 
We design our schema language based on [Common Expression Language (CEL)](https://github.com/google/cel-go). So the syntax looks nearly identical to equivalent expressions in C++, Go, Java, and TypeScript. 

Please let us know via our [Discord channel](https://discord.gg/n6KfzYxhPp) if you have questions regarding syntax, definitions or any operator you identify not working as expected.
:::

Let's examine some of common usage of ABAC with small schema examples.

### Boolean - True/False Conditions

For attributes that represent a binary choice or state, such as a yes/no question, the `Boolean` data type is an excellent choice.

```perm
entity post {
		attribute is_public boolean
		
		permission view = is_public
}
```

:::caution
⛔ If you don’t create the related attribute data, Permify accounts boolean as `FALSE`
:::

### Text & Object Based Conditions

String can be used as attribute data type in a variety of scenarios where text-based information is needed to make access control decisions. Here are a few examples:

- **Location:** If you need to control access based on geographical location, you might have a location attribute (e.g., "USA", "EU", "Asia") stored as a string.
- **Device Type**: If access control decisions need to consider the type of device being used, a device type attribute (e.g., "mobile", "desktop", "tablet") could be stored as a string.
- **Time Zone**: If access needs to be controlled based on time zones, a time zone attribute (e.g., "EST", "PST", "GMT") could be stored as a string.
- **Day of the Week:** In a scenario where access to certain resources is determined by the day of the week, the string data type can be used to represent these days (e.g., "Monday", "Tuesday", etc.) as attributes!

```perm
entity user {}

entity organization {
	
	relation admin @user

	attribute location string[]

	permission view = check_location(request.current_location, location) or admin
}

rule check_location(current_location string, location string[]) {
	current_location in location
}
```

:::caution
⛔ If you don’t create the related attribute data, Permify accounts string as `""`
:::

### Numerical Conditions

#### Integers

Integer  can be used as attribute data type in several scenarios where numerical information is needed to make access control decisions. Here are a few examples:

- **Age:** If access to certain resources is age-restricted, an age attribute stored as an integer can be used to control access.
- **Security Clearance Level:** In a system where users have different security clearance levels, these levels can be stored as integer attributes (e.g., 1, 2, 3 with 3 being the highest clearance).
- **Resource Size or Length:** If access to resources is controlled based on their size or length (like a document's length or a file's size), these can be stored as integer attributes.
- **Version Number:** If access control decisions need to consider the version number of a resource (like a software version or a document revision), these can be stored as integer attributes.

```perm
entity content {
    permission view = check_age(request.age)
}

rule check_age(age integer) {
		age >= 18
}
```

:::caution
⛔ If you don’t create the related attribute data, Permify accounts integer as `0`
:::

#### Double - Precise numerical information

Double can be used as attribute data type in several scenarios where precise numerical information is needed to make access control decisions. Here are a few examples:

- **Usage Limit:** If a user has a usage limit (like the amount of storage they can use or the amount of data they can download), and this limit needs to be represented with decimal precision, it can be stored as a double attribute.
- **Transaction Amount:** In a financial system, if access control decisions need to consider the amount of a transaction, and this amount needs to be represented with decimal precision (like $100.50), these amounts can be stored as double attributes.
- **User Rating:** If access control decisions need to consider a user's rating (like a rating out of 5 with decimal points, such as 4.7), these ratings can be stored as double attributes.
- **Geolocation:** If access control decisions need to consider precise geographical coordinates (like latitude and longitude, which are often represented with decimal points), these coordinates can be stored as double attributes.

```perm
entity user {}

entity account {
    relation owner @user
    attribute balance double

    permission withdraw = check_balance(request.amount, balance) and owner
}

rule check_balance(amount double, balance double) {
	(balance >= amount) && (amount <= 5000)
}
```

:::caution
⛔ If you don’t create the related attribute data, Permify accounts double as `0.0`
:::

See more details on [Attribute Based Access Control](#attribute-based-permissions-abac) section to learn our approach on ABAC as well as how it operates in Permify. you can see more comprehensive ABAC examples in the [Example ABAC Use Cases](../use-cases/abac/#example-use-cases) section in related page.

## More Comprehensive Examples

You can check out more comprehensive schema examples from the [Real World Examples](../examples.md) section.

Here is what each example focuses on,

*  [Google Docs]: how users can gain direct access to a document through **organizational roles** or through **inherited/nested permissions**.
*  [Facebook Groups]: how users can perform various actions based on the **roles and permissions within the groups** they belong.
*  [Notion]: how **one global entity (workspace) can manage access rights** in the child entities that belong to it.
*  [Instagram]: how **public/private attributes** play role in granting access to specific users.
*  [Mercury]: how **attributes and rules interact within the hierarchical relationships**.

[Google Docs]:./examples/google-docs.md
[Facebook Groups]:./examples/facebook-groups.md
[Notion]:./examples/notion.md
[Instagram]:./examples/instagram.md
[Mercury]:./examples/mercury.md