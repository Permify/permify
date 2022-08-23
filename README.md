
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
</p>
<p align="center">
    <img src="https://raw.githubusercontent.com/Permify/permify/master/assets/permify-demo-github.gif" alt="Permify - Open source authorization as a service"  width="820px" />
</p>

## What is Permify?

Permify is an open-source authorization service for creating and maintaining fine-grained authorizations. You can run Permify image container and it works as a Rest API.

Permify converts and syncs your authorization data as relation tuples into your preferred database. And you can check authorization with single request based on those tuples.

Data model is inspired by [Google Zanzibar White Paper](https://storage.googleapis.com/pub-tools-public-publication-data/pdf/41f08f03da59f5518802898f68730e247e23c331.pdf).

## Permify works best:

- If you already have an identity/auth solution and want to plug in fine-grained authorization on top of that.
- If you want to create a unified access control mechanism for individual applications.
- If you’re managing authorization for growing micro-service infrastructure.
- If your authorization logic is cluttering your code base.
- If your data model is getting too complicated to handle your authorization within the service.
- If your authorization is growing too complex to handle within code or API gateway.

### Features

🔐 Convert & store authorization data **in house** with high availability.

🔮 Easily model and refactor your authorization with **Permify's DSL, Permify Schema**.

📝 **Audit & Reason** your access control hassle-free with user interface.

🩺 Analyze **performance and behavior** of your authorization with tracing tools [jaeger], [signoz] or [zipkin].

✅ Low latency with **parallel graph engine** on access checks.

[jaeger]: https://www.jaegertracing.io/
[signoz]: https://signoz.io/
[zipkin]: https://zipkin.io/

## How it works

![Value Chain Schema](https://user-images.githubusercontent.com/34595361/186108668-4c6cb98c-e777-472b-bf05-d8760add82d2.png)

## Getting Started

- [Install Permify] with running Permify container using docker.
- Follow a guide to model your authorization using [Permify Schema].
- Learn how Permify [centralize & stores your authorization data].
- Take a look at the overview of [Permify API].

[Install Permify]: https://docs.permify.co/docs/installation
[Permify Schema]: https://docs.permify.co/docs/getting-started/modeling
[centralize & stores your authorization data]: https://docs.permify.co/docs/getting-started/sync-data
[Permify API]: https://docs.permify.co/docs/api-overview


## Community & Support
Join our [Discord channel](https://discord.gg/MJbUjwskdH) for issues, feature requests, feedbacks or anything else. We love to talk about authorization and access control :heart:

<p align="left">
<a href="https://discord.gg/MJbUjwskdH">
 <img alt="permify | Discord" width="50px" src="https://user-images.githubusercontent.com/34595361/178992169-fba31a7a-fa80-42ba-9d7f-46c9c0b5a9f8.png" />
</a>
<a href="https://twitter.com/GetPermify">
  <img alt="permify | Twitter" width="50px" src="https://user-images.githubusercontent.com/43545812/144034996-602b144a-16e1-41cc-99e7-c6040b20dcaf.png"/>
</a>
<a href="https://www.linkedin.com/company/permifyco">
  <img alt="permify | Linkedin" width="50px" src="https://user-images.githubusercontent.com/43545812/144035037-0f415fc7-9f96-4517-a370-ccc6e78a714b.png" />
</a>
</p>

## Contributing 
Want to contribute ? 

See: [CONTRIBUTING.md](https://github.com/Permify/permify/blob/master/CONTRIBUTING.md).

## Stargazers

[![Stargazers repo roster for permify/permify](https://reporoster.com/stars/permify/permify)](https://github.com/permify/permify/stargazers)

## License

Licensed under the Apache License, Version 2.0: http://www.apache.org/licenses/LICENSE-2.0
