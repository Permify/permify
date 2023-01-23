---
title: "Fine-Grained Access Control: Where RBAC falls short"
description: "Deciding an access control structure is important for businesses since security plays an important role and authorization mechanisms sit at the heart of the system. So most businesses today consider authorization aspects meticulously. And these considerations lead to one common question: “How granular should the access control be?”In this article, we’ll primarily focus on this question and examine; what is fine-grained access control, where it is used, it's importance, and how to choose the right authorization granularity for your company."
slug: fine-grained-access-control-where-rbac-falls-short
authors:
  - name: Ege Aytin
    image_url: https://user-images.githubusercontent.com/34595361/213848483-fe6f2073-18c5-46ef-ae60-8db80ae66b8d.png
    title: Permify Core Team
    email: ege@permify.co
tags: [fine grained access control, authorization, security, rbac, abac, rebac]
image: https://user-images.githubusercontent.com/34595361/213831693-57ebef70-ae5e-4df4-a36b-effcebae8c9a.png
hide_table_of_contents: false
---

![Fine-Grained Access Control: Where RBAC falls short](https://user-images.githubusercontent.com/34595361/213831693-57ebef70-ae5e-4df4-a36b-effcebae8c9a.png)

Role based access control (RBAC) is one of the traditional methods to restrict system access to unauthorized users. It is so common that almost all products somehow use roles for their access control.

Despite its high usage, using roles alone is  too static and coarse grained to handle more advanced authorization cases. That's why most companies choose fine grained access control models and techniques over coarse grained RBAC. To give a quick comparison, with RBAC data access may simply be permitted or forbidden based on a single property, i.e. role. In contrast, fine-grained access control gives organizations the ability to manage access based on more than one attribute.

<!--truncate-->

Deciding an access control structure is important for businesses since security plays an important role and authorization mechanisms sit at the heart of the system. So most businesses today consider authorization aspects meticulously. And these considerations lead to one common question: “How granular should our access control be?”

As a member of a team that builds an open-source authorization service for creating and maintaining authorizations, I’ll primarily focus on how to choose the right authorization granularity for your applications.

But before that let’s understand what is fine-grained access control and where it is used briefly.

## What is Fine-Grained Access Control?

Fine-grained access control is a method to control who can access the data based on multiple - and sometimes dynamic - properties. 

As compared to coarse-grained access control, which relies specifically upon a single property (e.g. role) for data protection, fine-grained access control defines permissions based on multiple properties. These properties can be a user's job role, time of day, user-groups & teams, resource parent-child relations or even organization hierarchies.

Coarse-grained RBAC only provides a static authorization technique, which is not sufficient for granular and dynamic access control needs and potentially increases security risk for those cases. For example If an employee needs some temporary access to a resource that is beyond the scope of their assigned role, there’s no simple way to allow it with a coarse-grained RBAC model. It must either allow access to a resource 24/7 or not at all.

Continuing our example, with fine-grained access control it's possible to grant access to resources only during your working hours of the day which reduces unnecessary risk factors in the company. We can easily achieve this time based access control with [attribute based access control (ABAC)](https://www.okta.com/blog/2020/09/attribute-based-access-control-abac/).

To be clear out, using RBAC does not mean your authorization model is coarse grained unless you use solely RBAC and not combine any conditions other than role of users to perform access controls. 

As an example of that, we frequently come across combining role based access control with other access control models such as ABAC and ReBAC. Let's look at an example from one of our writings. Imagine an organization where each department has its own set of files and we want users to access a file if only:

- they’re a member of the department, which file belongs to,

- they have an admin or manager role in organizational wide,

- they’re a creator, basically owner of that file.

![Group 1551](https://user-images.githubusercontent.com/34595361/213831815-63a26c4b-92e5-4e86-8354-c053cc8737e6.png)

Considering the example above, we combined role based access control (user's that has admin or manager role) and relation based access control (member of the department - parent & child membership of resource).

The important thing here, when checking multiple access control models, be careful about the priority of enforcement. We cover this priority issue and the whole examples of using role-based with relation-based access control in the [article](https://www.permify.co/post/relational-based-access-control-models). So for the sake of the topic, I’m leaving this here and continuing.

## Why is Fine-Grained Access Control Important?

Fine-grained access control is crucial because it enables data with various access requirements to coexist in the same storage area without posing a threat to security or compliance.

Fine-grained access control is typically used in cloud computing where large amounts of data types and data resources are stored together but each of the data items must be accessed based on different criteria. As you can guess, fine-grained authorization is more secure than coarse-grained authorization, because it narrows down access to data resources.

For example: a user might be blocked from accessing the data from mobile devices at a particular time period.

### Common Use Cases for Fine-Grained Access Control

Some common use cases for fine-grained access control:

#### 1. To control who can read, edit, move or delete the data
Fine-grained access control gives you complete control over the actions surrounding the data.

Imagine three employees with different roles. The first employee should be authorized for read-only access and not edit or delete the data. While one of the employees can be authorized to read, edit, move or delete the data. The third employee should be authorized to access the data.

With coarse-grained access control, the authorization to data can be completely permitted or forbidden based on the role. But with fine grained access control, in this example attribute based access control(ABAC) is used, you can define more granular and dynamic permissions among users and resources.

#### 2. Remote Access to data
With the rise of remote work, it is difficult to maintain control over the data as people work from home at different hours. Fine-grained access controls allow companies to implement data control based on factors like location, time of day, and more.

For example, You may be able to limit data access according to users working hours.

For today's cloud-based business environments, fine-grained systems are appropriate; however, it is crucial to find a balance between a company's security requirements and granting regulated access to organizational assets to unidentified third parties.

#### 3. Third-Party Access
Fine-grained systems are ideal for managing third-party access. For example: In B2B business it may require giving temporary third-party access to company assets. At the same time, it is important to maintain security. The fine-grained solution can be used to allow third-party read-only access keeping the assets secured.

## Which should be chosen, coarse-grained or fine-grained?

We can examine this question with the following aspects of authorization.

### Scalability
The coarse-based RBAC does not scale well as the businesses grow and more roles are added to the organization it becomes complex for data teams to manage RBAC which might create a role explosion is a situation that makes it difficult for administrators to manage access permissions.

When it comes to fine-grained access control, the access is easily managed at scale as administrators can create attributes that take several factors into account, giving them much more control over the data.

### Implementation
Implementation of fine-grained policies requires a significant amount of time and expertise which cannot be afforded by every business. Also, the wrong implementation can cost businesses a loss of time.

Coarse-grained access control is quicker to set up and is widely adopted (in the case of RBAC) which makes it easier to implement for small to medium-sized companies.

### Security
Fine-grained access control ensures data integrity which means that only people who are authorized based on conditions can access the sensitive data. Instead of completely blocking sensitive data, you can make it work by setting multiple conditions that a user has to meet to gain access.With fine-grained access control, sign-in attempts can be blocked from suspicious IP addresses which cannot be achieved by coarse-grained access control.

### Maintenance
The multidimensional approach of fine-grained access control makes data access more dynamic than the static coarse-based access control architecture. Once allocated to system components, it requires less maintenance after being defined by data teams.

## Conclusion 

Final thoughts, you should use fine-grained access control if your organization has varying access requirements for different resources and you want to increase data security restrictions to speed up collaboration & business.You can choose coarse-grained access control if your organization is smaller in size and roles are manageable for secure access of data.

If you aren’t sure which model is the best for your organization? Let us help you. [Permify](https://github.com/Permify/permify) is an open-source authorization service for creating and maintaining fine-grained authorizations in your applications.






