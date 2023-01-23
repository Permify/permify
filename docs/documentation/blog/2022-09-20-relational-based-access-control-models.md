---
title: "Relational Based Access Control Models"
description: "Relational based access control has gained its popularity over years among startups to large enterprises. Yet, some large tech companies are already starting to use ReBAC as their leading access control model. In 2019 Google published white paper of its relationship based global authorization system called Zanzibar, which handles authorization for YouTube, Drive, Google Cloud and all of Google's other products. "
slug: relational-based-access-control-models
authors:
  - name: Ege Aytin
    image_url: https://user-images.githubusercontent.com/34595361/213848483-fe6f2073-18c5-46ef-ae60-8db80ae66b8d.png
    title: Permify Core Team
    email: ege@permify.co
tags: [rebac, relationship access control, rbac, authorization, permissions]
image: https://user-images.githubusercontent.com/34595361/213843033-996c4f51-82f2-4501-bafc-3186c6ace200.png
hide_table_of_contents: false
---

![Relational Based Access Control Models (Thumbnail)](https://user-images.githubusercontent.com/34595361/213843033-996c4f51-82f2-4501-bafc-3186c6ace200.png)

Relational based access control has gained its popularity over years among startups to large enterprises. Yet, some large tech companies are already starting to use ReBAC as their leading access control model. In 2019 Google published white paper of its relationship based global authorization system called Zanzibar, which handles authorization for YouTube, Drive, Google Cloud and all of Google's other products. 

<!--truncate-->

After only a short time, Zanzibar based access control services have been created by the technical teams at Airbnb and Carta. These all contribute to the popularity of ReBAC for sure. 

Although ReBAC is best known for social networks because its core concept is about the network of relations, it can be applied beyond that. 

In this article, we’ll primarily focus on common usages of ReBAC. I’ll cover 3 highly used models of relationship-based access control with examples from familiar use cases. 

- Ownership
- Parent-Child & Hierarchies
- User Groups & Teams

But before diving in, let’s start with a quick intro about relationship-based access control and cover how it works.

## What is Relational Based Access Control ?
ReBAC is an access control model that defines permissions based on the relationships between entities and subjects of your system.

If we think of a simple blog application, the system typically allows post creators to edit or delete the post, which ensures that no ordinary user can make updates on a random blog.

![relations](https://user-images.githubusercontent.com/34595361/213843003-8c2fb4b2-6e12-43aa-bf8b-71defc4893b2.jpg)

In our blog example, the creator represents a relationship between the user (subject) and the post (entity). So, if user X has a creator relation with the post Y, then the system allows user X to edit or delete post Y.

Subject indicates the party that takes the action on resources. In most cases - likewise in this blog example - the subject will be a user, but it can also be another entity like a device, tenant or user set such as team members. We’ll cover all these use cases, but since we started with creator - which is a form of ownership - let's continue with it and introduce our first ReBAC model: ownership.

## Most Popular ReBAC Models

### Ownership
Ownership can occur in many places across different applications. The general idea is, users or entities have absolute permissions on their own data. Such examples are; users accessing their own profile settings in an application, tracking their own healthcare data, or viewing their own governmental information. All of these examples contained an ownership model.

We mentioned the example “Users can edit posts they created”. We can easily achieve this functionality with ReBAC. The biggest advantage of using relationship-based access control is that you can often use data that is already present in the application.

![ownership](https://user-images.githubusercontent.com/34595361/213843053-74ac060c-b051-418d-9556-b99cccb12420.png)

When a user creates a post you can store its identity in the post table as owner. Sounds easy right ?

#### Ownership v RBAC
Let’s look at that situation from a role based access control perspective. You can easily define specific roles or permissions to perform actions on posts, such as read, delete, edit etc. 

But what about the ownership? How to create a permission or role that defines users to be able to edit posts they created.

It’s clear that we can’t achieve this with a single role. A workaround solution would be assigning the “owner” role to the user when a post is created to differentiate it from other users. Moreover, we need to add specific roles to each post to compare the user and owner roles. This is the point where roles and permissions don't fit well, as opposed to ReBAC. As we mentioned this can be easily achieved through ReBAC with a simple Database design.

It’s important to notice that access control models can be used together in such scenarios. Although it's not reasonable to build ownership with roles and permissions. ReBAC ownership model works great with role based access control and it’s widely used. 

Let’s give an example of [deleting repository permission](https://docs.github.com/en/repositories/creating-and-managing-repositories/deleting-a-repository) on Github. You can delete any repository or fork if you're either an organization owner or have admin permissions for the repository or fork.

![delete-repo](https://user-images.githubusercontent.com/34595361/213843097-2d18ea30-571b-4108-b4f0-c901aded7134.png)

For performing a delete action, GitHub uses both RBAC and ReBAC. This combination indicates that either you need to have an admin role or you need to be owner of the organization to delete a repository.

### Parent-Child & Hierarchies
Since being the owner of some resource can be also described as the parent of some resource, the parent-child model often be confused with ownership. Although they seem the same at first glance, they’re quite different.

Parent-child model is relevant with the nested relations of a child's resources. The general idea is granting parents access to perform actions on their children’s resources.

Think of an organization, in which departments have their own resources like files, documents, etc… We want users to view resources if only they’re a member of the department. So this prevents different users from achieving other departments' files. Notice that it’s not sufficient to say **'members can view department files'** — instead, you need to specify the parent of the user as the department.

![parent-child](https://user-images.githubusercontent.com/34595361/213843139-5be5d3aa-1afc-4c20-9e82-3ce0b0c46b8b.jpg)

What if we want to add RBAC in this scenario ? Let’s say users that have admin or manager roles can view all of the files, organization-wide. Then we have a similar case that we cover on the ownership model.

Quick note before continuing: Generally this is a case we wouldn’t prefer in real life, since giving a "god mode” to a specific user role is breaking the [least privilege](https://delinea.com/what-is/least-privilege) access principle and it's a nightmare at scale systems.

![parent-child Flowchart](https://user-images.githubusercontent.com/34595361/213843181-0cb4fa35-ba74-4248-9ba9-34dbcfcfe4df.jpg)

When combining roles and relationship-based access control be careful about the priority of enforcement. Most of the time you’ll need to check whether the user has admin or manager role first. If a user has it, then they can be authorized to perform the action. Unless the user doesn’t have one of these roles, then the system needs to check the parent child relation to decide.

### User Groups & Teams
Grouping users helps to organize access control in a more structured way, especially at scale.

We examine the parent child relationship model with segmentation of resources (entities), we group resources (files) as in departments. 

Groups model more focused on segmentation of users or user sets rather than segmenting the resources. In particular, this model ensures that the user has a privilege to access to do some action on a resource based on its group, team etc.

Github repository access would be a great example for group specific privileges. Github enables you to 
[give a team access to a repository](https://docs.github.com/en/organizations/managing-user-access-to-your-organizations-repositories/managing-team-access-to-an-organization-repository#giving-a-team-access-to-a-repository) or change a team's level of access to a repository in your repository settings.

![give-team-access](https://user-images.githubusercontent.com/34595361/213843202-d4be8cef-24b9-4aab-8e67-4ee7b1d1aed2.png)

Giving access to a team simply means members of that team can also access your private repository. 

The enforcement workflow is the same as other models. Most of the time you check whether the user has the relevant role or permissions, then you need to move on checking parent-child relation. Is the user still unauthorized ? if yes, then you need to check for each group the user belongs to.

## Conclusion
We went through 3 common relationship based access control models  and tried to explain their structure with basic real world examples.

Many applications already work with a database that contains a network of entities that have relationships with each other. We can clearly see this in relational databases. 

That's why most of the systems have already adopted relational based access control for the sake of their natural structure.

I try to keep things simple for demonstration purposes. If you have any questions or doubts, feel free to ask them.
