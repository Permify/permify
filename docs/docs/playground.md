---
sidebar_position: 6
---

# Using Permify Playground

You can use our [Playground] to create and test your authorization schema in a browser.

Our playground consists 3 main sections,

- [Schema (Authorization Model)](#schema-authorization-model)
- [Authorization Data](#authorization-data)
- [Enforcement](#enforcement-access-check-scenarios)

Let's examine these sections by following a simple example.

[Playground]: https://play.permify.co/

## Schema (Authorization Model)

You can create your authorization model in this section with using our domain specific language.

You can define your entities, relations between them and access control decisions with using Permify Schema. We already have a couple of use cases and example that you can choose to see how authorization can be structured. Also, you can check our docs to [learn more about how to model authorization](./getting-started/modeling.md) in Permify.

To demonstrate how the playground works, let's create a simple authorization model as follows. This model should be selected as the default when you open the playground.

```perm
entity user {}

entity organization {

    // organizational roles
    relation admin @user
    relation member @user
}

entity repository {

    // represents repositories parent organization
    relation parent @organization

    // represents owner of this repository
    relation owner  @user

    // permissions
    permission edit   = parent.admin or owner
    permission delete = owner
}
```

We have 2 permissions, `edit` for access of editing repository and `delete` for access of deleting repository.

Repositories has parent child relation with organizations. The `parent` relation in the repository entity represents that parent child association, while ownership of the repository is represented with the `owner` relation.

Organizations can have organizational wide roles such as admin and member, which defined as `admin` and `member` relation in organization entity.

:::info Automatic Saving for Schema Changes
Schema changes are captured automatically, and other sections update accordingly. Some delays may occur at times; please feel free to reach out if these delays hinder your testing process.
:::

### Visualizer

We get loads of feedback about the observability and reasonability of the authorization model across teams and colleagues.

So we put a simple visualizer that shows how your authorization structure looks at a high level. In particular, you can examine relations between entities and their permissions.

![relational-tuples](https://github.com/Permify/permify/assets/34595361/f8b77c18-dd46-461c-9408-392b642cc900)

## Authorization Data

You can create sample authorization data to test your authorization logic. In Permify, authorization data stored as tuples and these tuples stored in a database that you preferred.

The basic tuple takes the form of:

`‍entity # relation @ user`

So the entity can be any entity that you defined in your model. If we look up our example it can be an organization or repository (since the user is empty). The relation can be one of the defined relations in the selected entity.

The user is basically the user or user set in our system. Let's say we want make the **user 1** `admin` in **organization 1** then we need to create an example relational tuple according to our model as follows:

`‍organization:1#admin@user:1`

To create a relation tuple in playground just hit the **Add Relationship** button.

![create-tuple-empty](https://github.com/Permify/permify/assets/34595361/33b85fe7-25e2-400d-8055-94d305023d8c)

You can choose entity, relation and the subject (user or user set) with entering identifier to create sample data. Let's create the relation tuple `‍organization:1#admin@user:1` as follows.

![create-tuple-user](https://github.com/Permify/permify/assets/34595361/016d6f9e-955a-4c39-ab55-21a9fd6dffd9)

Let's add one more relation tuple to perform a sample access check. I want to add repository:1 into organization:1 - `‍repository:1#parent@organization:1#...` as follows:

![create-tuple-parent](https://github.com/Permify/permify/assets/34595361/42daf251-818a-4bd2-8790-1c8656cd497f)

Created tuples shown in the **Data** section as follows.

![authorization-data](https://github.com/Permify/permify/assets/34595361/ccc25da1-5212-425d-b604-6a31a8f9555f)

## Enforcement (Access Check Scenarios)

Finally as we have a sample data let's perform an access check!

The YAML in the Enforcement section represents a test scenario for conducting access checks. This scenario-based testing process provides the ability to execute complex access scenarios in a single place.

Let's name our scenario **"admin_access_test"** and create tests to check:

- Whether user:1 (admin) can edit repository:1? 
- Whether user:1 (admin) can delete repository:1?

Below is the YAML scenario covering these two tests:

![scenario-check](https://github.com/Permify/permify/assets/34595361/934add02-6b6a-45ed-9b5b-6a2539778fcf)

In the above YAML structure,

#### entity

Represents the resource for which we want to check access - `repository:1`

#### subject

Represents the subject that performs the action or grants access - `user:1`.

#### assertions

Assertions stands for defining the expected result for specific action or an permission. In our case we're evaluating access for edit action.

Since organization:1 is parent of repository:1 ( `‍repository:1#parent@organization:1#...` ) and user:1 has an admin role in organization:1 ( `‍organization:1#admin@user:1` ) user:1 should allow to edit the repository:1 because the we define rule of the edit permission as:

`‍permission edit = parent.admin or owner`

:::note
which `‍parent.admin`‍ indicates admin in the organization that repository belongs to. 
:::

So user:1 should be able to edit resource:1, therefore expected outcome for that access request is true.
- `edit: true`

On the other hand, user:1 should't be able to delete resource:1, because only owners can. Therefore expected outcome for that is false.
- `delete: false`

:::info Create More Advanced Scenarios
For simplicity, we've created a basic scenario. However, you can create more advanced scenarios using our validation YAML structure.

To learn how to use this syntax for complex scenarios, refer to the [Creating Test Scenarios](../getting-started/testing#creating-test-scenarios) section in [Testing & Validation](./getting-started/testing.md) page.
:::

Let's click the Run button to execute our scenario.

![scenario-check-true](https://github.com/Permify/permify/assets/34595361/a90c042f-e0f8-46a0-9800-383620226acd)

Let's change the expected outcome as false (`edit: false`) and hit the **Run** button again we'll see an error message.

![scenario-check-false](https://github.com/Permify/permify/assets/34595361/9f9768bf-c534-4b1d-9447-e55cab2dafca)

As we seen above this is how you can model your authorization and test it with sample data in Permify Playground.

## Need any help ?

Our team is happy to help you get started with Permify. If you'd like to learn more about using Permify in your app or have any questions about this example, [schedule a consultation call with one of our account executives](https://calendly.com/d/cj79-kyf-b4z).
