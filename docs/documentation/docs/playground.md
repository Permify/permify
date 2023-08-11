---
sidebar_position: 6
---

# Using Permify Playground

You can use our [Playground] to create and test your authorization in a browser. Our playground consists 4 sections; Authorization Model, Visualizer, Authorization Data and Enforcement. Let's examine these sections by following a simple example.

[Playground]: https://play.permify.co/

## Authorization Model

You can create your authorization model in this section with using Permify authorization language, Permify Schema. You can define your entities, relations between them and access control decisions with using Permify Schema. We already have a couple of use cases and example that you can choose to see how authorization can be structured with Permify Schema. Also, you can check our docs to learn more about how to model authorization in Permify.

To demonstrate how playground works, let's choose the "empty" option from our dropdown to create a simple authorization model as follows:

![authorization-model](https://github.com/Permify/permify/assets/39353278/308cfbac-0db3-4349-ae38-4cfe64bc6732)

We have 2 permissions these are editing repository and deleting repository. Repository has parent child relation with organizations. Lastly organizations can have organizational wide roles such as admin and member. After completing your authorization model you can just save it with hitting the save button and start testing it.

## Visualizer

We get loads of feedback about the observability and reasonability of the authorization model across teams and colleagues. So we put a simple visualizer that shows how your authorization structure looks at a high level. In particular, you can examine relations between entities and their permissions. Here is a visualization for example model that we created above.

![relational-tuples](https://github.com/Permify/permify/assets/39353278/a80a39b3-5139-4f13-9395-bdf1f9296c49)

## Authorization Data

You can create sample authorization data to test your authorization logic. In Permify, authorization data stored as relation tuples and these tuples stored in a database that you preferred. The basic relation tuple takes the form of:

`‍entity # relation @ user`

So the entity can be any entity that you defined in your model. If we look up our example it can be an organization or repository (since the user is empty). The relation can be one of the defined relations in the selected entity. Lastly, the user is basically the user or user set in our system. Let's say we want make user 1 admin in organization 1 then we need to create an example relational tuple according to our model as follows:

`‍organization:1#admin@user:1`

To create a relation tuple in playground just hit the "new" button and a pop up will open.

![create-tuple-empty](https://github.com/Permify/permify/assets/39353278/abea768e-8721-4957-a2a8-4dc8eff9d6bc)

You can choose entity, relation and the subject (user or user set) with entering identifier to create sample data. Let's create the relation tuple `‍organization:1#admin@user:1` as follows.

![create-tuple-user](https://github.com/Permify/permify/assets/39353278/2525cb4e-8014-4871-849b-77df80efa577)

And this created tuple shown in the Authorization Data section as follows.

![authorization-data](https://github.com/Permify/permify/assets/39353278/d415c6bf-6b00-457c-8a95-dea846e96125)

Let's add one more relation tuple to perform a sample access check. I want to add repository:1 into organization:1 as follows:

![create-tuple-parent](https://github.com/Permify/permify/assets/39353278/d8ea5e8d-c487-4cdf-91cd-48f039949046)

Created relational tuple after this will be: "repository:1#parent@organization:1".

## Enforcement ( Access Checks)
Finally as we have a sample data lets perform an access check from the right below. Let's check whether user:1 can edit the repository:1. Since organization:1 is parent of repository:1 ( `‍repository:1#parent@organization:1#...` ) and user:1 has an admin role in organization:1 ( `‍organization:1#admin@user:1` ) user:1 should allow to edit the repository:1 because the we define rule of the edit permission as:

`‍permission edit = owner or parent.admin`

which parent.admin indicates admin in the organization that repository belongs to. So let's type **"user:1 edit repository:1"** and hit the check button to get result.

![relational-tuples](https://github.com/Permify/permify/assets/39353278/e7724518-f641-4f6a-9d93-d4dc89b1f409)


Let's try to get unauthorized result. Type "user:1 delete repository:1" on the question input. Since only owners can delete the repository this access check will result as unauthorized.

![relational-tuples](https://github.com/Permify/permify/assets/39353278/51bd7e2e-a4c1-4df5-8a1f-9eec7999872a)

As we seen above this is how you can model your authorization and test it with sample data in Permify Playground. Check out our docs for different modeling use cases, creating and storing relational tuples and more. 

## Need any help ?

Our team is happy to help you get started with Permify. If you'd like to learn more about using Permify in your app or have any questions about this example, [schedule a call with one of our Permify engineer](https://meetings-eu1.hubspot.com/ege-aytin/call-with-an-expert).

