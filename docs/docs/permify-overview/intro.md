---
sidebar_position: 1
---

# What is Permify?

[Permify](https://github.com/Permify/permify) is a **open source authorization service** for creating and maintaining fine-grained authorizations while ensuring least privilege across your organization.

With Permify, you can easily structure your authorization model, store authorization data in your preferred database, and interact with the Permify API to handle all authorization queries from your applications or services.

Permify inspired by Google’s consistent, global authorization system, [Google Zanzibar](https://storage.googleapis.com/pub-tools-public-publication-data/pdf/41f08f03da59f5518802898f68730e247e23c331.pdf). 

## A true ReBAC solution to ensure least privilege

Permify has designed and structured as a true ReBAC solution, so besides roles and traditional permissions Permify also supports indirect permission granting through relationships. 

For instance, you can define that a user has certain permissions because of their relation to other entities. An example of this would be granting a manager the same permissions as their subordinates, or giving a user access to a resource because they belong to a certain group. This is facilitated by our relationship-based access control, which allows the definition of complex permission structures based on the relationships between users, roles, and resources.

Our goal is to create a robust, flexible, and easily auditable authorization system that establishes a natural linkage between permissions across the business units, functions, and entities of an organization.

## Key Features

🛡️ **Production ready** authorization API that serve as **gRPC** and **REST**

🔮 Domain Specific Authorization Language - Permify Schema - to **easily model** your authorization

🔐 Database Configuration to store your permissions **in house** with **high availability**

✅ Perform access control checks and get answers **down to 10ms** with **parallel graph engine**

💪 Battle tested, robust **authorization architecture and data model** based on [Google Zanzibar](https://storage.googleapis.com/pub-tools-public-publication-data/pdf/41f08f03da59f5518802898f68730e247e23c331.pdf)

⚙️ Create custom permissions for your **tenants**, and manage them in single place with **Multi Tenancy**

⚡ Analyze **performance and behavior** of your authorization with tracing tools [jaeger], [signoz] or [zipkin]

[jaeger]: https://www.jaegertracing.io/
[signoz]: https://signoz.io/
[zipkin]: https://zipkin.io/

## Features Beyond Zanzibar

We’re trying to make [Zanzibar](https://storage.googleapis.com/pub-tools-public-publication-data/pdf/41f08f03da59f5518802898f68730e247e23c331.pdf) available to everyone to use and benefit in their applications and services. So that we utilize Zanzibar features and add new features on top of it to achieve robust permission systems. Here are some additional features that we have, 

- **Multi-Tenancy Support** - It enables users to create a custom authorization model for different applications, all managed within a single Permify instance.

- **Testing Framework - Permify Validate** - This enhances the testability of authorization logic. It includes features like scenario-based validation actions, policy coverage analysis, and IDL parser Integration to achieve end-to-end validation for the desired authorization schema.

- **Data Filtering** - In Zanzibar typical access check has the form of **"Does user U has relation R to object O?”** and yields true or false response. Additional to that, we have data filtering endpoints that let you ask questions in the form of **“Which resources can user:X do action Y?”** or **“Which user(s) can edit doc:Y”**. As a response to this, you’ll get a entity results in the format of a string array or as a streaming response depending on the endpoint you're using.

## Getting Started

In Permify, authorization divided into 3 core aspects; **modeling**, **storing authorization data** and **access checks**.  

- See how to [Model your Authorization] using Permify Schema.
- Learn how Permify [Store Authorization Data] as relations.
- Perform an [Access Checks] anywhere in your stack.

[Model your Authorization]: ../getting-started/modeling
[Store Authorization Data]: ../getting-started/sync-data
[Access Checks]: ../getting-started/enforcement

This document explains how Permify handles these aspects to provide a robust and scalable authorization system for your applications. For the ones that want trying out and examine it instantly, 

<div className="getting-started-grid" >
    <a href="https://play.permify.co/">
        <div className="btn-thumb">
            <div className="thumbnail">
                <img src="https://uploads-ssl.webflow.com/61bb34defcff34f786b458ce/6332bb38106ffd85102bb3bc_Screen%20Shot%202022-09-27%20at%2011.58.27.png"/>
            </div>
           <div className="thumb-txt">Use our Playground to test your authorization in a browser. </div>
        </div>
    </a>
    <a href="https://docs.permify.co/docs/installation/overview">
        <div className="btn-thumb">
            <div className="thumbnail">
                 <img src="https://user-images.githubusercontent.com/34595361/199695094-872d50fc-c33b-4d15-ad1d-a3899911a16a.png"/>
            </div>
            <div className="thumb-txt">Set up Permify Service in your environment</div>
        </div>
    </a>
    <a href="https://www.permify.co/post/google-zanzibar-in-a-nutshell">
        <div className="btn-thumb">
            <div className="thumbnail">
                <img src="https://uploads-ssl.webflow.com/61bb34defcff34f786b458ce/634520d7859cd419ec89f9ef_Google%20Zanzibar%20in%20a%20Nutshell-1.png"/>
            </div>
            <div className="thumb-txt">Examine Google Zanzibar In A Nutshell</div>
        </div>
    </a>
</div>

## Community & Support

We would love to hear from you :heart:

You can get immediate help on our Discord channel. This can be any kind of question-related to Permify, authorization, or authentication and identity management. We'd love to discuss anything related to access control space.

For feature requests, bugs, or any improvements you can always open an issue. 

### Want to Contribute? Here are the ways to contribute to Permify

* **Contribute to codebase:** We're collaboratively working with our community to make Permify the best it can be! You can develop new features, fix existing issues or make third-party integrations/packages. 
* **Improve documentation:** Alongside our codebase, documentation one of the most significant part in our open-source journey. We're trying to give the best DX possible to explain ourselfs and Permify. And you can help on that with importing resources or adding new ones.
* **Contribute to playground:** Permify playground allows you to visualize and test your authorization logic. You can contribute to our playground by improving its user interface, fixing glitches, or adding new features.

You can find more details about contributions on [CONTRIBUTING.md](https://github.com/Permify/permify/blob/master/CONTRIBUTING.md).

## Communication Channels

If you like Permify, please consider giving us a :star:

<p align="left">
<a href="https://discord.gg/MJbUjwskdH">
 <img height="70px" width="70px" alt="permify | Discord" src="https://user-images.githubusercontent.com/39353278/187209316-3d01a799-c51b-4eaa-8f52-168047078a14.png" />
</a>
<a href="https://twitter.com/GetPermify">
  <img height="70px" width="70px" alt="permify | Twitter" src="https://user-images.githubusercontent.com/39353278/187209323-23f14261-d406-420d-80eb-1aa707a71043.png"/>
</a>
<a href="https://www.linkedin.com/company/permifyco">
  <img height="70px" width="70px" alt="permify | Linkedin" src="https://user-images.githubusercontent.com/39353278/187209321-03293a24-6f63-4321-b362-b0fc89fdd879.png" />
</a>
</p>

## Roadmap

You can find Permify's Public Roadmap [here](https://github.com/orgs/Permify/projects/1)!

## Need any help on Authorization ?

Our team is happy to help you anything about authorization. Moreover, if you'd like to learn more about using Permify in your app or have any questions, [schedule a call with one of our founders](https://meetings-eu1.hubspot.com/ege-aytin/call-with-an-expert).