---
sidebar_position: 1
---

# What is Permify?

[Permify](https://github.com/Permify/permify) is an **open source authorization service** for creating fine-grained and scalable authorization systems.

With Permify, you can easily structure your authorization model, store authorization data in your preferred database, and interact with the Permify API to handle all authorization queries from your applications or services.

Permify is inspired by Google’s consistent, global authorization system, [Google Zanzibar](https://permify.co/post/google-zanzibar-in-a-nutshell/).

### Motivation

Our goal is to make **Google's Zanzibar** available to everyone and help them to build robust, flexible, and easily auditable authorization system that establishes a [natural linkage between permissions](https://permify.co/post/relationship-based-access-control-rebac/) across the business units, functions, and entities of an organization.

## Key Features

🛡️ **Production ready** authorization API that serve as **gRPC** and **REST**.

🔮 Domain Specific Authorization Language to **easily model** your authorization. Supporting RBAC, ReBAC, ABAC and more.

🔐 Database Configuration to store your permissions with **high availability** and **low latency**.

✅ Perform access control checks and get answers **down to 10ms** with our various cache mechanisms that we operate.

💪 Battle tested, robust **authorization architecture and data model** based on [Google Zanzibar](https://storage.googleapis.com/pub-tools-public-publication-data/pdf/41f08f03da59f5518802898f68730e247e23c331.pdf).

⚙️ Create custom permissions for your **tenants**, and manage them in a single place with **Multi Tenancy**.

⚡ Analyze **performance and behavior** of your authorization with tracing tools [jaeger], [signoz] or [zipkin].

[jaeger]: https://www.jaegertracing.io/
[signoz]: https://signoz.io/
[zipkin]: https://zipkin.io/

## Getting Started

In Permify, authorization is divided into 3 core aspects; **modeling**, **storing authorization data** and **access checks**.  

- See how to [Model your Authorization] using Permify Schema.
- Learn how Permify will [Store Authorization Data] as relations.
- Perform [Access Checks] anywhere in your stack.

[Model your Authorization]: ../../getting-started/modeling
[Store Authorization Data]: ../../getting-started/sync-data
[Access Checks]: ../../getting-started/enforcement

This document explains how Permify handles these aspects to provide a robust and scalable authorization system for your applications. For the ones that want to try it out and examine it instantly, 

<div className="getting-started-grid" >
    <a href="https://play.permify.co/">
        <div className="btn-thumb">
            <div className="thumbnail">
                <img src="https://uploads-ssl.webflow.com/61bb34defcff34f786b458ce/6332bb38106ffd85102bb3bc_Screen%20Shot%202022-09-27%20at%2011.58.27.png"/>
            </div>
           <div className="thumb-txt">Use our Playground to test your authorization in a browser. </div>
        </div>
    </a>
    <a href="../../installation/overview">
        <div className="btn-thumb">
            <div className="thumbnail">
                 <img src="https://user-images.githubusercontent.com/34595361/199695094-872d50fc-c33b-4d15-ad1d-a3899911a16a.png"/>
            </div>
            <div className="thumb-txt">Set up Permify Service in your environment</div>
        </div>
    </a>
    <a href="https://permify.co/post/google-zanzibar-in-a-nutshell/">
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

For feature requests, bugs, or any improvements you can always open an [issue](https://github.com/permify/permify/issues). 

### Want to Contribute? Here are the ways to contribute to Permify

* **Contribute to codebase:** We're collaboratively working with our community to make Permify the best it can be! You can develop new features, fix existing issues or make third-party integrations/packages. 
* **Improve documentation:** Alongside our codebase, documentation is an important part of our open-source journey. We're trying to give the best DX possible to explain ourselves and Permify. And you can help with that by importing resources or adding new ones.
* **Contribute to playground:** Permify playground allows you to visualize and test your authorization logic. You can contribute to our playground by improving its user interface, fixing glitches, or adding new features.

You can find more details about contributions on [CONTRIBUTING.md](https://github.com/Permify/permify/blob/master/CONTRIBUTING.md).

## Communication Channels

If you like Permify, please consider giving us a :star: on [github](https://github.com/permify/permify)

<p align="left">
<a href="https://discord.gg/n6KfzYxhPp">
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

Our team is happy to help you get started with Permify. If you'd like to learn more about using Permify or how it might fit into your authorization workflow, [schedule a consultation call with one of our account executivess](https://calendly.com/d/cj79-kyf-b4z).

