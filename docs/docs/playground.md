---
sidebar_position: 6
---

# Using Permify Playground

You can use our [Playground] to create and test your authorization in a browser. Our playground consists 4 sections,

- Schema (Authorization Model) 
- Authorization Data
- Visualizer 
- Enforcement

Let's examine these sections by following a simple example.

[Playground]: https://play.permify.co/

## Schema (Authorization Model)

You can create your authorization model in this section with using Permify authorization language, Permify Schema. 

You can define your entities, relations between them and access control decisions with using Permify Schema. We already have a couple of use cases and example that you can choose to see how authorization can be structured. Also, you can check our docs to learn more about how to model authorization in Permify.

To demonstrate how the playground works, let's create a simple authorization model as follows. This model should be selected as the default when you open the playground.

![authorization-model](https://github.com/Permify/permify/assets/34595361/9da0957c-a6ee-4dd7-81ff-693a98b3d4d1)

We have 2 permissions, `edit` for editing repository and `delete` for deleting repository. 

Repository has parent child relation with organizations. The `parent` relation in the repository entity represents that parent child association, while ownership of the repository is represented with the `owner` relation.

Organizations can have organizational wide roles such as admin and member, which defined as `admin` and `member` relation in organization entity.

:::info Automatic Saving for Schema Changes
Schema changes are captured automatically, and other sections update accordingly. Some delays may occur at times; please feel free to reach out if these delays hinder your testing process.
:::

## Visualizer

We get loads of feedback about the observability and reasonability of the authorization model across teams and colleagues. 

So we put a simple visualizer that shows how your authorization structure looks at a high level. In particular, you can examine relations between entities and their permissions. Here is a visualization for example model that we created above.

![relational-tuples](https://github.com/Permify/permify/assets/39353278/a80a39b3-5139-4f13-9395-bdf1f9296c49)

## Authorization Data

You can create sample authorization data to test your authorization logic. In Permify, authorization data stored as relation tuples and these tuples stored in a database that you preferred. 

The basic relation tuple takes the form of:

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

## Enforcement (Access Checks)

Finally as we have a sample data let's perform an access check!

In Playground you should create an scenario in order to perform access checks. This scenario based testing process gives ability to perform complex access scenarios in a single place.

Hit the **New Scenario** button in the right side and a pop up will open. You can enter name and description of the scenario in here.

![new-scenario-popup](https://github.com/Permify/permify/assets/34595361/c9c50da5-e3d8-4bc2-9599-985092006358)

<!-- Lets check editing access of the repository,After creating a new scenario it will shown as follows:

![empty-scenario](https://github.com/Permify/permify/assets/34595361/cf137f9c-a96a-47d2-9bf2-4e47054c2131) -->

Let's check **whether user:1 can edit the repository:1** as follows:

![scenario-check](https://github.com/Permify/permify/assets/34595361/0168f013-45a0-49fe-8164-3f5f5311f15c)

In the above YAML structure,

#### entity
Represents the resource for which we want to check access - `repository:1` 

#### subject
Represents the subject that performs the action or grants access - `user:1`.

#### context 
Refers to additional data provided during an access check to be evaluated for the access decision. It's primarily used for dynamic access checks, such as those involving time, date, or IP address, etc. 

In our case, we leave it empty as null.For our case we leave it empty as null. You can check the details from the [Contextual Tuples](./reference/contextual-tuples.md) section.

#### assertions

Assertions stands for defining the expected result for specific action or an permission. In our case we're evaluating access for edit action.

Since organization:1 is parent of repository:1 ( `‍repository:1#parent@organization:1#...` ) and user:1 has an admin role in organization:1 ( `‍organization:1#admin@user:1` ) user:1 should allow to edit the repository:1 because the we define rule of the edit permission as:

`‍permission edit = parent.admin or owner`

which `‍parent.admin`‍ indicates admin in the organization that repository belongs to. So user:1 should be able to edit resource:1, therefore expected outcome for that access request is true - `edit: true`

:::info Create More Advanced Scenarios
For simplicity, we've created a basic scenario. However, you can create more advanced scenarios using our validation YAML structure.

To learn how to use this syntax for complex scenarios, refer to the [Creating Test Scenarios](../getting-started/testing#creating-test-scenarios) section in [Testing & Validation](./getting-started/testing.md) page.
:::

Let's click the Run button to execute our scenario. The scenario name should turn green once the scenario result is confirmed as correct.

![scenario-check-true](https://github.com/Permify/permify/assets/34595361/208e1761-f202-449d-a9e0-498ab0d4ce6d)

Let's change the expected outcome as false (`edit: false`) and hit the **Run** button again we'll see an error message.

![scenario-check-false](https://github.com/Permify/permify-validate-action/assets/34595361/28a206ca-f7cb-42a8-a8c4-a18376ebf8f3)

As we seen above this is how you can model your authorization and test it with sample data in Permify Playground. 

## Need any help ?

Our team is happy to help you get started with Permify. If you'd like to learn more about using Permify in your app or have any questions about this example, [schedule a call with one of our Permify engineer](https://meetings-eu1.hubspot.com/ege-aytin/call-with-an-expert).

