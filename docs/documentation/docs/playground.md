---
sidebar_position: 6
---

# Permify Playground

You can use our [Playground] to create and test your authorization in a browser. Our playground consists 4 sections; Authorization Model, Visualizer, Authorization Data and Enforcement. Let's examine these sections by following a simple example.

[Playground]: https://play.permify.co/

## Authorization Model

You can create your authorization model in this section with using Permify authorization language, Permify Schema. You can define your entities, relations between them and access control decisions with using Permify Schema. We already have a couple of use cases and example that you can choose to see how authorization can be structured with Permify Schema. Also, you can check our docs to learn more about how to model authorization in Permify.

To demonstrate how playground works, let's choose the "empty" option from our dropdown to create a simple authorization model as follows:

![relational-tuples](https://user-images.githubusercontent.com/34595361/193245391-6ff7cd21-69e3-4b8e-9fa8-d28c9045fe16.png)

We have 2 permissions these are editing repository and deleting repository. Repository has parent child relation with organizations. Lastly organizations can have organizational wide roles such as admin and member. After completing your authorization model you can just save it with hitting the save button and start testing it.

## Visualizer

We get loads of feedback about the observability and reasonability of the authorization model across teams and colleagues. So we put a simple visualizer that shows how your authorization structure looks at a high level. In particular, you can examine relations between entities and their permissions. Here is a visualization for example model that we created above.

![relational-tuples](https://user-images.githubusercontent.com/34595361/193245587-ff794d53-c142-44fb-959b-5c4546dd73c1.png)

## Authorization Data

You can create sample authorization data to test your authorization logic. In Permify, authorization data stored as relation tuples and these tuples stored in a database that you preferred. The basic relation tuple takes the form of:

`‍entity # relation @ user`

So the entity can be any entity that you defined in your model. If we look up our example it can be an organization or repository (since the user is empty). The relation can be one of the defined relations in the selected entity. Lastly, the user is basically the user or user set in our system. Let's say we want make user 1 admin in organization 1 then we need to create an example relational tuple according to our model as follows:

`‍organization:1#admin@user:1`

To create a relation tuple in playground just hit the "new" button and a pop up will open.

![relational-tuples](https://user-images.githubusercontent.com/34595361/193246047-a6c425bd-b417-4054-b1a0-9352e8f30ded.png)

You can choose entity, relation and the subject (user or user set) with entering identifier to create sample data. Let's create the relation tuple `‍organization:1#admin@user:1` as follows.

![relational-tuples](https://user-images.githubusercontent.com/34595361/193246036-691cb4ab-a589-4856-887e-7f412a2bb32d.png)

And this created tuple shown in the Authorization Data section as follows.

![relational-tuples](https://user-images.githubusercontent.com/34595361/193246251-ffbb5c8d-944b-4b87-ae50-82a7c2d575e2.png)

Let's add one more relation tuple to perform a sample access check. I want to add repository:1 into organization:1 as follows:

![relational-tuples](https://user-images.githubusercontent.com/34595361/193246717-cce0dc69-f10b-4e3a-8a85-ed846373a154.png)

Created relational tuple after this will be: "repository:1#parent@organization:1#..." We used “...”  when subject type is different from user entity. #… represents a relation that does not affect the semantics of the tuple.

## Enforcement ( Access Checks)
Finally as we have a sample data lets perform an access check from the right below. Let's check whether user:1 can edit the repository:1. Since organization:1 is parent of repository:1 ( `‍repository:1#parent@organization:1#...` ) and user:1 has an admin role in organization:1 ( `‍organization:1#admin@user:1` ) user:1 should allow to edit the repository:1 because the we define rule of the edit permission action as:

`‍action edit = owner or parent.admin`

which parent.admin indicates admin in the organization that repository belongs to. So let's type **"can user:1 edit repository:1"** and hit the check button to get result.

![relational-tuples](https://user-images.githubusercontent.com/34595361/193246742-4df97b34-5e94-4132-9c7c-8d184ccc32f4.png)


Let's try to get unauthorized result. Type "can user:1 delete repository:1" on the question input. Since only owners can delete the repository this access check will result as unauthorized.

![relational-tuples](https://user-images.githubusercontent.com/34595361/193246754-86332f18-a483-479b-a0cf-62703c38a2f4.png)

As we seen above this is how you can model your authorization and test it with sample data in Permify Playground. Check out our docs for different modeling use cases, creating and storing relational tuples and more. 

## Need any help ?

Our team is happy to help you get started with Permify. If you'd like to learn more about using Permify in your app or have any questions about this example, [schedule a call with one of our Permify engineer](https://meetings-eu1.hubspot.com/ege-aytin/call-with-an-expert).

