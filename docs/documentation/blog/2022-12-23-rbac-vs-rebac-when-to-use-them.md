---
title: "RBAC vs ReBAC: When to use them"
description: "RBAC vs ReBAC: When to use them"
slug: rbac-vs-rebac-when-to-use-them
authors:
  - name: Ege Aytin
    image_url: https://user-images.githubusercontent.com/34595361/213848483-fe6f2073-18c5-46ef-ae60-8db80ae66b8d.png
    title: Permify Core Team
    email: ege@permify.co
tags: [rbac, rebac, roles, permissions, authorization]
image: https://user-images.githubusercontent.com/34595361/213828169-bab1cf06-402b-4868-8711-218c16198736.png
hide_table_of_contents: false
---

![RBAC vs  ReBAC](https://user-images.githubusercontent.com/34595361/213828169-bab1cf06-402b-4868-8711-218c16198736.png)

RBAC and ReBAC are two well-known access control models. Although their high usage they also have certain disadvantages, such as RBAC's inability to provide dynamic behavior and flexibility, and ReBAC's inability to give the same ease of permission management as RBAC.

<!--truncate-->

In this article, we will compare the two commonly used authorization models, Role-based access control (RBAC) and Relationship-Based Access Control (ReBAC). We’ll briefly look at both of them and we'll explain how it works and when you should use them.

We will also see how we can combine these authorization models and make them work together.

Let’s see them in action!

## What Is Role-Based Access Control?

Role-based access control (RBAC) is the concept of assigning permissions to users based on their role. It allows data access to be easily authorized or denied based on a single property i.e the role. For example, think of organization employees with different positions, each employee's authorization can be based on several factors, such as authority, responsibility, and job competency, etc.

With RBAC you can control users’ authority based on their role in an organization. Generally, lower-level employees don’t have access to sensitive data or they can’t perform high-level tasks within an organization. It helps secure a company's sensitive data as employees are only given the required data to carry out their jobs.

The core principle of RBAC is to assign employees enough resources to do their job. Suppose a team of different roles (Engineering, Marketing, Finance, HR) working on a large project. The Engineering team might only have access to AWS, GitHub, etc. while the finance team might only have access to financial tools and resources. The same goes for the Marketing and HR team as they might have access only to their required resources.

Access management is a lot easier with RBAC unless and until an organization strictly adheres to role requirements. The policies may easily be transferred to a new position or deleted from the role group rather than needing to be altered each time a person quits the company or changes employment. Additionally, depending on the organizational position they play, new hires might have access rather rapidly.

Now let’s take a look at what ReBAC (Relational-Based Access Control) has to offer.

## What Is Relational-Based Access Control?

You might have started to hear more about Relationship-Based Access Control (ReBAC) in Identity and Access Management (IAM) space. Let’s explore what REBAC really ReBAC is.

Organizing permissions to users based on relationships between resources means a relationship-based authorization or ReBAC. An example would be to permit only post authors to modify or remove blog posts, ensuring that a regular user cannot update a blog at random.

It’s this relationship mechanism that allows authors to modify blog posts. The relationship determines how access is granted.

Let’s take a look at a couple of more real-world examples with relationship-based access control models:

Ownership: You can delete a repository on GitHub if you are the owner or have admin permissions for a repository.
Parent-Child and Hierarchies: You can create an issue in a GitHub repo if you’re a contributor to a parent repository.
User Groups and Teams: You can also create a private repository on GitHub and you can give the team access to a repository or change a team's level of access to a repository in your repository settings.
In each of the examples above, the connections between the objects serve as a description of the authorization logic.

You can find more about Relational-Based Access Control [here](/blog/).

## RBAC vs ReBAC: When to use them

RBAC (role-based access control) is one of the most popular models for defining permissions.

In the traditional RBAC model, roles are given to users, and these roles correspond to particular permissions on various resources. This might work with common use cases within an organization. But It happens frequently that certain users may need access to resources that don't quite fit the stated roles, resulting in the creation of new roles. As the organization grows big in size the number of resources as well as roles will increase exponentially resulting in role explosion.

This problem of role explosion can be resolved with the help of ReBAC which allows us to display permissions as connections between entities. 

ReBAC also enables us to define role-role and resource-resource relationships. This gives us the ability to define permissions based on relationships and enable role inheritance. 

Many times neither RBAC nor ReBAC will be the perfect solution to cover all the use cases you need. That’s the reason why most organizations combine these solutions. 

Let’s see RBAC combined with ReBAC in action.

### RBAC combined with ReBAC
As we saw in the example of ReBAC “Users can edit posts they created”. This can be easily achieved with the help of  Relational-Based Access Control. 

If we want to implement this with Role-based access control (RBAC) we surely cannot achieve that with a single role. The solution with the help of RBAC will be to assign the owner role to the user at the time of the creation of a post to differentiate it from other users. Additionally, we must assign specific roles to each post in order to compare the user and owner roles.

It’s where roles and permissions don’t work well. In such scenarios, access control models can be combined and used together. The ReBAC ownership model works great with role-based access control and it’s widely used. 

When combining RBAC, the rules can be changed to “Users can edit posts they created if they are the owner of the post or they have an admin role”. We use both RBAC and ReBAC to accomplish this action.

### When to choose ReBAC over RBAC?
Imagine a “data analyst” who needs access to a resource from a different team (ex. finance). In this case, the “data analyst” is a single user and you can’t simply assign them the finance role as you do not want them to have access to all the finance sheets present in an organization. Now, you have no option but to create a new role with a miscellaneous name (DA_FT_FIX_TEMP) which you will no longer remember or often forget to change later. 

For this, RBAC is too coarse-grained. You cannot afford to simply create a new role for each user every time they need access to a resource from different teams.

Some system administrators add more roles to their systems to increase granularity. As we discussed earlier, this can result in role explosions where there are hundreds or even thousands of roles to keep track of. To overcome this problem, you need ReBAC (Relationship-Based Access Control) which might be the perfect fit for your need for fine-grained permissions.

Another example that can be achieved by ReBAC: If an "admin" role inherits from a "manager" role, any new permissions granted to the "manager" will be inherited by the "admin." Traditional RBAC cannot accomplish this.

## Make a Smart Decision
If you require fine-grained permissions, ReBAC solutions like those based on Google's Zanzibar are an excellent option. Permify is an open-source authorization service inspired by Google Zanzibar.

The definition of access controls based on roles (RBAC) and attributes (ABAC) is widely established, but when numbers and complexity rise, there is a large overhead.


