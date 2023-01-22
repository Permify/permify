---
title: "How to Get Authorization Right"
description: "In the past few months, we have talked to over 100 engineers from both Fortune 500 companies and startups about their approach to authorization."
slug: why-decouple-authorizations
authors:
  - name: Fred Dogan
    image_url: https://user-images.githubusercontent.com/34595361/213848632-4a98f25b-df49-4ee1-ab53-785de24c8388.jpeg
    title: Permify Core Team
    email: firat@permify.co
  - name: Ege Aytin
    image_url: https://user-images.githubusercontent.com/34595361/213848483-fe6f2073-18c5-46ef-ae60-8db80ae66b8d.png
    title: Permify Core Team
    email: ege@permify.co
  - name: Tolga Ozen
    image_url: https://user-images.githubusercontent.com/34595361/213848541-8d4da803-8842-4adc-8125-1ca1838b51b9.jpeg
    title: Permify Core Team
    email: tolga@permify.co
tags: [permissions, rbac, abac, authorization, access control]
image: https://user-images.githubusercontent.com/34595361/213843696-69e75aa1-750e-4ccc-b1a5-b16ad7d3fa0e.png
hide_table_of_contents: false
---

![how-to-get-authz-right](https://user-images.githubusercontent.com/34595361/213843696-69e75aa1-750e-4ccc-b1a5-b16ad7d3fa0e.png)

In the past few months, we have talked to over 100 engineers from both Fortune 500 companies and startups about their approach to authorization. The complaints were playing in tune:

- Everyone hated syncing and moving authorization data.
- Most engineers agreed that modeling is hard, especially when it comes to never ending product requirements.
- No one liked the fact that authorization logic is cluttering the code base, and creating technical debt.
- And last but not least, Many developers told us testing & auditing haunt them at nights.
Even though everyone ran into similar problems, there was no consensus over a solution. Each team keeps reinventing the wheel.

<!--truncate-->

Some teams spend months to clean out their technical debt, and build a full fledged authorization service. While others keep adding new systems, and end up maintaining several authorization models. (Dare you to try preventing conflict between these systems, and you’ll start crying.)

Permify came to being when we had these same problems ourselves. We keep hitting similar issues over and over again while building products for both ourselves and clients.

For months we have tested different models, and ideas. We build stripe-like API, an [OPA](https://www.openpolicyagent.org/) control plane. We have used [Z3](http://theory.stanford.edu/~nikolaj/programmingz3.html#sec-intro), [Google Zanzibar](https://research.google/pubs/pub48190/), and more to come up with a flexible system.

Yes, building an unified authorization is hard. Here’s why, and how to properly solve it.

For people who have limited time, let’s boil it down to the 3 main factors.

## Overview

**The Curse: Modeling the Logic**

Almost all products start with a simple authorization system, it’s easy to spin a roles table into your database. But as your company grows, your requirements quickly change with never-ending user requests.

Now your bare minimum system falls short. And you have an ever-changing product with new stakeholders.

It’s challenging to refactor a simple model for these new complex use cases. But it’s also hard to design a complex model that’s easy to start with.

**The Debt: Designing the Architecture**

Authorization decisions consist of logic and authorization data. And most of the authorization data is a subset of application data. So, it’s challenging to orchestrate this data when authorization logic is separated from your core application.

And if it’s not, then you’ll end up cluttering your code base with authorization logic.

**The Ambiguity: Implementing the Enforcement**

The major problem with the enforcement is implementation.

Since authorization checks occur in so many places; like user interfaces, routers, API endpoints, database queries… It is a tedious and high effort task.  So, choosing where to enforce authorization, and loading the authorization data is hard.

## The Curse: Modeling the Logic

Modeling represents authorization logic. Basically the conditions in which a user can perform an action on a resource. For example, only owners and admins can edit the posts.

Depending on your use case, It often consists of roles, attributes and relationships.

### 3 Types of Access Control

- **Role based access control (RBAC)** is a simple system where you set access rights based on roles and permissions of a user.

- **Relational based Access control (ReBAC)** defines access policies based on relationships between resources. For instance, allowing only the users who are part of Team X to create documents.

- ‍**Attribute based access control (ABAC)** is an approach that makes access checks according to resource and user attributes. It determines access by user characteristics, object characteristics, action types.

### Why is Modeling hard ?

Almost all products start with a simple authorization system, It’s easy to spin up a couple role tables into your database. But things can easily go south. Your product quickly changes due to never-ending user requirements as the Company scales. And the simple authorization system will fall short.

Let’s assume you begin with a simple RBAC with 3 global roles; Admin, Editor and Member. But as the product grows, you’ll be adding new features.

And now larger customers want to have granularity, multi-tenancy, and more... These 3 simple roles are not enough anymore.

So, what to do? Maybe, you can find workaround solutions. If you have a monolith this can work for a while. But if you have a micro-service architecture, this will mean more technical debt for each service. And may result in a conflict between different authorization systems.

You can start to refactor your authorization system at that point, which is pretty common. We witnessed that teams struggle for a couple months to refactor their rigid legacy models. Why spend precious engineering resources to continuously tinkering with this logic?

Bottom line is that it is challenging to design a simple system that can accomodate the needs of different stakeholders. But it’s also quite cumbersome to create a complex model that’s easy to start with.

So there is always a dilemma for engineering teams when thinking about authorization. Let’s sort out some possible solutions.

#### Libraries

You can use existing open-source libraries like CanCanCan for Ruby, or CASL for JS. Simply roll your model in one of these. For instance in CanCanCan, you can model your logic like this;

It can be easier to start a complex model with these libraries. And they especially make sense when your application is monolith.

But as soon as your team starts taking the micro-service approach, you’ll end up with multiple different languages. And perhaps different models for each micro-service. This will make the authorization logic inconsistent, since there is no unified model for each service.

One can come up with your own library and build a standard policy format. But this can be a project on its own.

#### Centralized Authorization Services

You can use centralized authorization services like Permify, among others. In centralized services usually the models are represented as policies and schemas which define rules and relations between entities.

Since authorization data is centralized, the logic is consistent with the data model. Choosing a centralized service can be overkill if you have a monolithic architecture. But it is much more suitable to micro-services based architectures.

#### Centralized Decision Points

Centralized decision point is a locus where policy rules have been resolved, evaluated, and combined to make a decision. It’s pretty much the same as centralized service without centralized data.

But there is no single point where authorization data is stored. So, your model is unaware of your data. This forces you to send data where enforcement happens in line with your rules. And the caller has to know the model well.

This can easily breed several problems like latency, and overload.

## The Debt: Designing the Architecture

The authorization decision consists of two parts, logic and data.

Authorization logic is basically access control rules and policies that can be reasoned about. In other words, whether a user can take an action on a given condition or not.

On the other hand, In order to return any kind of authorization decision you need to have the authorization data.

For instance, only post owners can edit posts. This represents the logic. So, we need user and post data to make a decision about this logic.

And most of the authorization data is a subset of application data. So, [it’s challenging to orchestrate this data](https://medium.com/airbnb-engineering/himeji-a-scalable-centralized-system-for-authorization-at-airbnb-341664924574) when logic is separated from your core application.

The one common best practice would be decoupling authorization logic from your core application if you don’t have simple monolithic structure. This will keep your data unified through multiple services.

### Architecture Patterns

There are multiple alternatives for your authorization architecture. Depending on your needs; you can separate authorization data, or logic. You can replace decision points.

And this creates several combinations which makes it hard to decide. So, let’s talk about some patterns for different structures;

#### For Monoliths

Monoliths are easier to handle since authorization data and decision logic lives inside the structure. Most of the authorization data depends on the application data, So you can keep data where authorization happens.

Since you can make direct DB read, there is no concern about data synchronization.

For this structure, it makes sense to use a library. You don’t need a language agnostic approach, and you have no concerns about unifying the logic. Since it’s defacto unified.

But monoliths can still be challenging if you move some authorization data outside of the application.

For example, you can integrate a third party identity provider. In this case, you may need to fetch data from external resources for every decision point. Or store authorization data into JWT token to make it available.

#### For Microservices

Most of the teams didn’t face an issue with their authorization system until breaking their monolith structure to multiple services.

Let's examine which patterns can be followed if you have microservices approach;

**1. Centralized Authorization**

You have multiple services, and it sounds logical to centralize your authorizations. Create an abstract service that is only responsible for authorization, keep its data accordingly and feed all of other services on access decisions. It seems easy right ? Unfortunately it’s not.

Decision points may require additional data in order to return a decision. And centralized services need centralized authorization data that’s almost always up to date. Now the problem is fetching and picking the application data you need.

**Attaching Data to Request**
![Attaching-Data-To-Request](https://user-images.githubusercontent.com/34595361/213843996-a7679ea6-de07-4f3b-828d-dfc1742646fc.png)

You can attach necessary data to authorization requests. But this means the caller has to know the authorization logic. And it could lead to sending more data than necessary.

Which will eventually cause latency and availability issues. Especially if you have a complex authorization logic.

**Orchestrating Authorization Data** 
![Orchestrating-Authorization-Data](https://user-images.githubusercontent.com/34595361/213844056-a0dbb82a-bdac-44f5-9ef3-0d66730982e7.png)

In other words, centralizing authorization data. But you have to figure out how to sync data from your microservices, keep all the data relevant and manage conflicts.

It also means you should remodel all the data into an unified model for the authorization service.

And if you have services at scale, you need to be cautious about performance issues. How will you meet the latency requirements? You can set up a great cache. However it has its own trade-offs, like keeping cache updated.

Also you have one more thing to keep an eye on: availability. This service has to be available at all times, but that is pretty much ipso facto for microservices.

Overall this is an approach that seems to be optimal, but has lots of implementation and maintenance efforts within.

**2. Each Service Owns the Data & Logic**

![Each-Service-Owns-Data-Logic](https://user-images.githubusercontent.com/34595361/213844087-4deda543-57bf-43a2-81f7-6363b81f6951.png)

Let’s say you don’t want any problems regarding scalability. You can treat each micro service like a monolith and build authorization systems within. So basically each micro service has its own authorization system.

When each service has its own authorization logic and data, scalability is no more a problem. But there are 2 main problems with this approach:

If some parts of your application share the same spot for multiple microservices it can overlap, and you need to keep them in sync. Also adding the same logic multiple times can be hard and inefficient.

Since each micro-service probably has a different tech stack, you’ll need to implement authorization logic in each service. So you must carry out this process for each unique programming language and framework that you use in your system. And this can make implementation and maintenance exhausting.

## The Ambiguity: Implementing the Enforcement

Enforcement is how you exercise authorization decisions. You can enforce authorization in so many places. UI, routes, API endpoints, DB queries.

There is no single pattern to implement enforcement. So it’s hard to decide where to make enforcement, and how to load authorization data.

### Deciding Where to Make Enforcement?

Since enforcement can happen in so many places, defining where to make access checks is tricky. You can go all the way down in the stack, and make enforcement at resource level. Now your checks are scattered, and they happen in so many places. You have fine-grained authorization, but you end up with high implementation effort.

For instance you can check your authorization at row level.

Another problem with this pattern is that it leads to technical debt. Since you check permissions in so many places, at some point authorization starts to hinder the development processes. Even adding a feature can be tough because of authorization.

On the other hand, you can push all the way up, and exercise at API Gateway. And it’s tougher to manage authorization data for access checks. And creates a single point of failure.

### Loading Necessary Application Data

Let's take a look at a simple enforcement example similar to push repository at Github. Allow to push the library if the user is resource owner, or a maintainer.

There is isUserMaintainer function which fetches maintainers, and decides if a given user is a maintainer.

As you’ll need to compare multiple cases to enforce access decisions. Loading the application data from different services can pollute your code,and make it harder to maintain.

As your authorization gets fine-grained, you’ll end up with more complicated logic. This will mess up your code base which will cause latency and availability issues.

### Availability

Authorization plays a significant role in the reliability of the product. Because almost each enforcement point is a single point of failure for the application workflows.

There are 2 main reasons:

- Most of the time enforcement occurs at the beginning or middle of the workflow.
Unfortunately there is no adequate error handling on failed enforcement.

- Enforcement decides whether someone can do something on a resource. So it’s in its nature to occur before showing the actual result or user interfaces.

Being not in the final point of the workflow creates a potential problem for availability. Because if the enforcement fails, user cannot continue further. Or in the worst case scenario, end up accessing somewhere it’s not allowed to.

Another issue  is a lack of error handling process on a failed enforcement. For instance, you want to check

- Can a user perform a specific action on the given resource as A role, or a resource owner.

When authorization service failed on this operation. Either your application will fail to perform the action. Or you’ll end up giving access to unauthorized users.

### Latency

Latency is a highly important topic about enforcement. To explain where latency can create problems, let’s examine the below example.

In this particular example the isUserMaintainer is where we should focus regarding the latency. Because this is the step where authorization data can be queried from a database, an external service or any other source you need.

So, if you have a complex enforcement which needs multiple data sources to make access checks. The fetching data and executing business logic will increase response times.

## Conclusion - It’s Hard But Not Impossible

When we started building Permify, I was a little bit skeptical myself. Authorization has been around for so long. There must have been a unified solution that came out earlier than us.

And there wasn’t one.There wasn’t a consensus, a perfect solution. It is hard to create one, harder than we thought.

Some people are in favor of the Access Control Lists. Some go to the fringes, and build full-fledged ABAC models. Many others prefer simpler models.

You can argue that it's even impossible to build a unified solution. For many, it seems like a hail mary. Nevertheless, I know it’s hard but not impossible!

### How to Get Authorizations Right? 

If you're interested in building authorization system, or talking about authorization in general we'd glad to have you in our community Join us at Discord. And if you're frustrated with messy authorization system, you can  or schedule a 1-o-1 with one of our engineers, or join our waitlist to start using Permify.

**Sources:**
* https://research.google/pubs/pub48190/
* https://www.permify.co/post/why-decouple-authorizations
* https://medium.com/building-carta/authz-cartas-highly-scalable-permissions-system-782a7f2c840f
* https://medium.com/airbnb-engineering/himeji-a-scalable-centralized-system-for-authorization-at-airbnb-341664924574
* https://dbconvert.com/blog/postgresql-change-data-capture-cdc/amp/**