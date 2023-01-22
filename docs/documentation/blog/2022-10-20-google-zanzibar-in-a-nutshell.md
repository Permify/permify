---
title: "Google Zanzibar In A Nutshell"
description: "In this article we’ll examine  Zanzibar, which is the global authorization system used at Google for handling authorization for hundreds of its services and products including; YouTube, Drive, Calendar, Cloud and Maps."
slug: google-zanzibar-in-a-nutshell
authors:
  - name: Ege Aytin
    image_url: https://user-images.githubusercontent.com/34595361/213848483-fe6f2073-18c5-46ef-ae60-8db80ae66b8d.png
    title: Permify Core Team
    email: ege@permify.co
tags: [google zanzibar, authorization, spannerdb, fine-grained access control, rbac, abac, rebac]
image: https://user-images.githubusercontent.com/34595361/213842888-45cbb19e-758d-4b3a-a161-ab3cd1a3e3c3.png
hide_table_of_contents: false
---

![Google Zanzibar in a Nutshell-1](https://user-images.githubusercontent.com/34595361/213842888-45cbb19e-758d-4b3a-a161-ab3cd1a3e3c3.png)

In this article we’ll examine [Zanzibar](https://research.google/pubs/pub48190/), which is the global authorization system used at Google for handling authorization for hundreds of its services and products including; YouTube, Drive, Calendar, Cloud and Maps.

<!--truncate-->

Google published Zanzibar back in 2019, and in a short time it gained attention quickly. In fact some companies like Airbnb and Carta started to shift their legacy authorization structure to Zanzibar style systems.

Additional to shifts from large tech companies, Zanzibar based solutions increased over the time. All disclosure; Permify is an authorization system based on Zanzibar. 

In this article we’ll look up what differs Zanzibar from other permission systems and why companies have been adopting Zanzibar style solutions ever since the paper was published.

Reading the original paper is quite enjoyable but I want to simplify the Zanzibar paper, and explain the why, what and how's in a nutshell. I love to start with the reason. So, let’s look up why Zanzibar is needed by Google.

## Why is Zanzibar Needed?

So as we established, Google uses Zanzibar for its product to handle authorization. But what is Authorization? Simply, It is the process of controlling who can do, own or access a system in an application. Authorization systems can branch off due to application and user needs. Eventually, when things grow, authorization is a hard piece to solve for various reasons.

We examined authorization complexity by its 3 core aspects; modeling, architecture and enforcement. 

The problem with modeling is most of the time it starts simple but gets uglier over time due to growth and never ending user requirements. In particular, It’s hard to create permissions for ever changing environments with multiple different use cases of individual applications.

In Google’s case, they have a large number of applications and services with different permission needs; Drive has its own sharing authorization, Youtube has private/public video access controls, Google Cloud mostly relies on role based access control(RBAC) etc. So, Google needs a flexible modeling structure where they can create and extend combinations of different access control approaches in a common place.

About enforcement; it's challenging because it happens in so so many places. And access checks usually contain data from different services or applications. So, it’s becoming harder to make decisions fast and reliable.

Access checks between resources from one application in another is very common in Google. For instance, Google Calendar might have the right authorization for Google Meets. So for Google It's important to have a unified authorization system that makes it easier for applications to interoperate.

This interpolation goes more depth from enforcement to architectural problems actually. A common architectural approach of handling authorization is building a permission system for each individual application that is connected directly with the database. Since authorizations is a system critical part, and a single point of failure; an error or just a simple mistake on authorization systems can cause security frauds, showing false positive or false negative results to the users.

Moreover, when you have multiple applications, maintaining multiple authorization mechanisms; it becomes harder. And due to this, generally these systems are designed as abstract entities, a library or a centralized engine that cater to many individual applications and services.

Google also took this approach before building Zanzibar. They have a common authorization library that is used across their applications and services as a centralized permission engine.

“One permission system to rule them all!” mantra is great and abstracting permission systems from application has one big tradeoff. These libraries or permission engines don’t store data. They just behave as an engine to manage functionality such as performing access checks. And on these functionalities, there is a trade off and a down side.

For instance; in order to make an access check and compute a decision, you need to load the authorization data and relations from the database and other services. In this case, there is a huge downside in terms of performance and scalability.

Loading and processing authorization data is especially painful given the fact that Google computes billions of access checks which come from different environments and applications from worldwide. When you count that, these checks have to be fast, consistent, and available all times. So that a unified authorization system without the data management is error prone at Google scale.

Google needed a solution that is super fast, secure, and scalable while providing consistency and reliability across all applications. Actually, lets see major goals of Zanzibar system with quotes from the paper:

> **Correctness:** It must ensure consistency of access control decisions to respect user intentions.
Flexibility: It must support a rich set of access control policies as required by both consumer and enterprise applications.‍

> **Low latency:** It must respond quickly because authorization checks are often in the critical path of user interactions. Low latency at the tail is particularly important for serving search results, which often require tens to hundreds of checks.‍

> **High availability:** It must reliably respond to requests because, in the absence of explicit authorizations, client services would be forced to deny their users access.‍

> **Large scale:** It needs to protect billions of objects shared by billions of users. It must be deployed around the globe to be near its clients and their end users

We’ll examine how Zanzibar achieves these goals. But first, let's understand what Zanzibar is.

## What is Zanzibar ?

Zanzibar is an authorization service that stores permissions as ACL styled relationships and performs access decisions according to these relations. Zanzibar unifies authorization to serve individual applications and services.

Unified authorization system has various advantages as we mentioned above. But abstracting the authorization from the core application has one massive challenge; data management. Not surprisingly, Zanzibar has a unique way of solution to it.

### Data Model
Abstracting authorization data eliminates data loading issues on each enforcement action. Think of a simple access control question “Can User X view document Y ?” In the traditional way we had to check our policy or access rules and then need to load the necessary data for the decision. In Zanzibar, you already have all of the required information stored as relational tuples to make decisions quickly. 

When using Zanzibar, you tell Zanzibar about the activities that are related to authorization data. Let's assume we have a rule that enforces “only document creators can view the document”. Sounds reasonable? In that case, you need to feed Zanzibar with actions in your system about document creation. Such as “user X created document Y”, “user Z created document T”, etc. Then Zanzibar stores this information as a relational tuple in the centralized data source. 

Alon Yao, one of the software engineers in Airbnb wrote a post about how Airbnb moved its legacy authorization system to a Zanzibar styled system called Himeji. And here is how they approached synchronization of authorization data.

> We created Himeji, a centralized authorization system based on Zanzibar, which is called from the data layer. It stores permissions data and performs the checks as a central source of truth. Instead of fanning out at read time, we write all permissions data when resources are mutated. We fan out on the writes instead of the reads, given our read-heavy workload.

Relational tuples are similar to individual ACL collections of object user or object object relations and take the form of  **“user U has relation R to object O”**. Zanzibar represents of the form of these relational tuples as:

```
⟨tuple⟩ ::= ⟨object⟩’#’⟨relation⟩‘@’⟨user⟩

⟨object⟩ ::= ⟨namespace⟩‘:’⟨object id⟩ 

⟨user⟩ ::= ⟨user id⟩ | ⟨userset⟩

⟨userset⟩ ::= ⟨object⟩‘#’⟨relation⟩
```

Although it seems complex at first glance, actually It's quite straightforward. Let's deconstruct a tuple `object#relation@user`

In our document example,  when user 1 created document 2 you need to send the object and the subject relation. If the Zanzibar system were a person you'd say **“Hey Zanzibar user:1 is owner of document:2”** . And then this data stored as  `document:2#owner@user:1`

Users can also be a user set like team members. This allows nested group relations in access decisions. As an example think a system where we want to group members can viewer relation with a document, the relation tuple can be created according to this is 

`doc:1#viewer@group:2#member`

So the Zanzibar derived system unifies and stores a collection of relational tuples as authorization data.

### Access Control Checks 
Considering a scaled system like Google as you might expect there will be tons of relational tuples created. Currently Zanzibar manages trillions of tuples with thousands of namespaces in it. So all of these relational tuples and authorization model combines and build up a relations graph, which is mainly used for the access control check.

![Access-Graph](https://user-images.githubusercontent.com/34595361/213842820-8920066c-eec8-468b-9465-202464813a44.png)

In Zanzibar, evaluating access decisions operate as a walkthrough of this directed graph. For instance, if we ask “does user:1 is owner to the document:1?” and if the owner relation between user:1 and document:1 can be accessible on the directed graph Zanzibar concludes that user:1 is authorized. 

How can this relation be stored and be present in the graph? Simply, the application client should send the data to writeAPI when user:1 creates a document:1. We’ll look up write function in below.

### Functionality
Zanzibar has five core methods, these are; read, write, watch, check and expand. Expand and check are authorization related functions. Others are more related with authorization data, and relational tuples.

If we need to give a quick explanation for each method; 

- **Read (2.4.1)**, which allows for directly querying the graph data, it can be used to filter authorization data, relational tuples that are stored.

- **Write (2.4.2)**, basically allows you to write relational tuples to the data storage. In a standard relational database based application which implements Zanzibar style system, relational tuples can write into both the application’s database, and the Zanzibar system within the same flow. 

- **Watch (2.4.3)**, as we can understand the functionality from its name, it used to watch the relation tuple changes in the graph data.

- **Check (2.4.4)**, probably the most obvious one. Check is used in enforcement. In a Zanzibar based system you can check access in the form of “Can subject S have permission P on resource R”. The subject and resource should be specified. The check engine will calculate the decision by walking through to the graph of relations.

- **Expand (2.4.5)**, used to improve observability and reasonability of access permissions. It returns user set, which takes form of <object#relation> in Zanzibar, in a tree format of given resource permissions.

### Data Consistency

Zanzibar is a service that is constantly being accessed with millions of access checks so it must be secure. And when thinking about this kind of system where data is actively managed, security is mainly related to stored and managed data consistency.

And these inconsistent data can cause false negative or false positive decision results on authorization checks. So, if some schema update happens, or relations change then all applications that use Zanzibar should update their cache according to it to prevent false positive or false negative results.

Zanzibar avoids this problem with an approach of snapshot reads. Basically it ensures that enforcement is evaluated at a consistent point of time to prevent inconsistency. Zanzibar team developed tokens called Zookies that consist of a timestamp which is compared in access checks to ensure that the snapshot of the enforcement is at least as fresh as the resource version’s timestamp.

### Providing Low Latency at High Scale
Zanzibar stores more than two trillion ACLs and performs millions of authorization checks per second. Although these numbers are huge It is kind of expectable when thinking of Zanzibar used by Google products like Youtube, Drive, Calendar and more. Since the products that are using Zanzibar are distributed, enforcement for any object can come from anywhere in the world. So that Zanzibar doesn't spread ACL data geographically.

Therefore, Zanzibar replicates all ACL data in tens of geographically distributed data centers and distributes load across thousands of servers around the world. More specifically, in the paper there is a section called Experience (see 2.4.1) which mentions that Zanzibar – “distributes this load across more than 10,000 servers organized in several dozen clusters around the world using [Spanner](https://research.google/pubs/pub39966/), which is Google's scalable, multi-version, globally-distributed, and synchronously-replicated database.”

It is no doubt that Zanzibar applies various techniques to hit the low latency and high availability goals in this globally distributed environment. There is a serious cache mechanism that works which allows to get results from the local replicated data. More about caching from the paper,

> It handles hot spots on normalized data by caching final and intermediate results, and by deduplicating simultaneous requests. It also applies techniques such as hedging requests and optimizing computations on deeply nested sets with limited denormalization.

And results ? More than 95% of the access checks responded in 10 milliseconds and has maintained more than 99.999% availability for the 3 year period.

## Conclusion
Google’s Zanzibar has a major goal of creating a planet scaled permission system that is fast, secure and available all time. As [Permify](https://github.com/Permify/permify), an open source authorization service that is based on Zanzibar, we’re trying to make Zanzibar available to everyone to use and benefit in their applications and services.
