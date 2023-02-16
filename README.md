
<h1 align="center">
    <img src="https://raw.githubusercontent.com/Permify/permify/master/assets/permify-logo.svg" alt="Permify logo" width="336px" /><br />
    Permify - Open Source Authorization Service
</h1>

<p align="center">
    <a href="https://github.com/Permify/permify" target="_blank"><img src="https://img.shields.io/github/go-mod/go-version/Permify/permify?style=for-the-badge&logo=go" alt="Permify Go Version" /></a>&nbsp;
    <a href="https://goreportcard.com/report/github.com/Permify/permify" target="_blank"><img src="https://goreportcard.com/badge/github.com/Permify/permify?style=for-the-badge&logo=go" alt="Permify Go Report Card" /></a>&nbsp;
    <a href="https://github.com/Permify/permify" target="_blank"><img src="https://img.shields.io/github/license/Permify/permify?style=for-the-badge" alt="Permify Licence" /></a>&nbsp;
    <a href="https://discord.gg/MJbUjwskdH" target="_blank"><img src="https://img.shields.io/discord/950799928047833088?style=for-the-badge&logo=discord&label=DISCORD" alt="Permify Discord Channel" /></a>&nbsp;
    <a href="https://github.com/Permify/permify/pkgs/container/permify" target="_blank"><img src="https://img.shields.io/github/v/release/permify/permify?include_prereleases&style=for-the-badge" alt="Permify Release" /></a>&nbsp;
    <a href="https://img.shields.io/github/commit-activity/m/Permify/permify?style=for-the-badge" target="_blank"><img src="https://img.shields.io/github/commit-activity/m/Permify/permify?style=for-the-badge" alt="Permify Commit Activity" /></a>&nbsp;
    <a href="https://img.shields.io/github/actions/workflow/status/Permify/permify/release.yml?style=for-the-badge" target="_blank"><img src="https://img.shields.io/github/actions/workflow/status/Permify/permify/release.yml?style=for-the-badge" alt="GitHub Workflow Status" /></a>&nbsp;
    <a href="https://play.permify.co" target="_blank"><img src="https://img.shields.io/badge/Try-Permify%20Playground-blueviolet?style=for-the-badge" alt="Try Permify" /></a>
</p>

<p align="center">
    <img src="https://raw.githubusercontent.com/Permify/permify/master/assets/permify-dsl.gif" alt="Permify - Open source authorization as a service" />
</p>

## What is Permify?
[Permify](https://github.com/Permify/permify) is an **open-source authorization service** for creating and maintaining fine-grained authorizations in your applications.

With Permify you can easily structure your authorization model, store authorization data in your own servers securely, and interact with Permify API to handle all authorization questions from any of your applications.

Permify's data model is inspired by Google’s consistent, global authorization system, [Google Zanzibar Paper](https://storage.googleapis.com/pub-tools-public-publication-data/pdf/41f08f03da59f5518802898f68730e247e23c331.pdf).

**Looking for Permify Managed Service?** See our [pricing plans](https://permify.co/pricing/1).

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

### Cases that can benefit from Permify

- If you already have an identity/auth solution and want to plug in fine-grained authorization on top of that.
- If you want to create a unified access control mechanism to use across your individual applications.
- If you want to make future-proof authorization system and don't want to spend engineering effort for it.
- If you’re managing authorization for growing micro-service infrastructure.
- If your authorization logic is cluttering your code base.
- If your data model is getting too complicated to handle your authorization within the service.
- If your authorization is growing too complex to handle within code or API gateway.

## How it works

Permify stores access control relations on a database you choose and performs authorization checks based on the stored relations. We called that database Write Database (WriteDB) and it behaves as a centralized data source for your authorization system. You can model your authorization with Permify's DSL - Permify Schema and perform access checks with a single API call anywhere on your stack.

![Value Chain Schema](https://user-images.githubusercontent.com/34595361/186108668-4c6cb98c-e777-472b-bf05-d8760add82d2.png)

## QuickStart

You can quickly start Permify on your local with running the docker command below:

```shell
docker run -p 3476:3476 -p 3478:3478  ghcr.io/permify/permify serve
```

This will start Permify with the default configuration options: 
* Port 3476 is used to serve the REST API.
* Port 3478 is used to serve the GRPC Service.
* Authorization data stored in memory.

See [all of the options] that you can use to set up and deploy Permify in your servers.

[all of the options]: https://docs.permify.co/docs/installation

### Test your connection

You can test your connection with creating an GET request,

```shell
localhost:3476/healthz
```

## Getting Started

- [Set Up & Run Permify Service] in your environment.
- Follow a guide to model your authorization using [Permify Schema].
- See our [playground], build your authorization logic and test it with sample data.
- Take a look at the overview of [Permify API].
- See [our article] to examine Google Zanzibar In A Nutshell.

[Set Up & Run Permify Service]: https://docs.permify.co/docs/installation
[Permify Schema]: https://docs.permify.co/docs/getting-started/modeling
[playground]: https://play.permify.co/
[Permify API]: https://docs.permify.co/docs/api-overview
[our article]: https://permify.co/post/google-zanzibar-in-a-nutshell

[![Run in Postman](https://run.pstmn.io/button.svg)](https://www.postman.com/permify-dev/workspace/permify/collection)
[![View in Swagger](http://jessemillar.github.io/view-in-swagger-button/button.svg)](https://permify.github.io/permify-swagger)

## Community & Support
We would love to hear from you :heart:

You can get immediate help on our Discord channel. This can be any kind of question-related to Permify, authorization, or authentication and identity management. We'd love to discuss anything related to access control space.

For feature requests, bugs, or any improvements you can always open an issue. 

### Want to Contribute? Here are the ways to contribute to Permify

* **Contribute to codebase:** We're collaboratively working with our community to make Permify the best it can be! You can develop new features, fix existing issues or make third-party integrations/packages. 
* **Improve documentation:** Alongside our codebase, documentation one of the most significant things in our open-source journey. We're trying to give the best DX possible to explain ourselfs and Permify. And you can help on that with importing resources or adding new ones.
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

## Stargazers

[![Stargazers repo roster for permify/permify](https://reporoster.com/stars/permify/permify)](https://github.com/permify/permify/stargazers)
