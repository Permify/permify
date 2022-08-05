

<h1 align="center">
    <img src="https://raw.githubusercontent.com/Permify/permify/master/assets/permify-logo.svg" alt="Permify logo" width="336px" /><br />
    Permify - Open Source Authorization Service
</h1>

<p align="center">
    <a href="https://github.com/Permify/permify" target="_blank"><img src="https://img.shields.io/github/go-mod/go-version/Permify/permify?style=for-the-badge&logo=go" alt="Permify Go Version" /></a>&nbsp;
    <a href="https://goreportcard.com/report/github.com/Permify/permify" target="_blank"><img src="https://goreportcard.com/badge/github.com/Permify/permify?style=for-the-badge&logo=go" alt="Permify Go Report Card" /></a>&nbsp;
    <a href="https://github.com/Permify/permify" target="_blank"><img src="https://img.shields.io/github/license/Permify/permify?style=for-the-badge" alt="Permify Licence" /></a>&nbsp;
    <a href="https://discord.gg/MJbUjwskdH" target="_blank"><img src="https://img.shields.io/discord/950799928047833088?style=for-the-badge&logo=discord&label=DISCORD" alt="Permify Discord Channel" /></a>&nbsp;
    <a href="https://hub.docker.com/repository/docker/permify/permify" target="_blank"><img src="https://img.shields.io/docker/v/permify/permify?style=for-the-badge&logo=docker&label=docker" alt="Permify Docker Image Version" /></a>&nbsp;
</p>

<p align="center">
    <img src="https://raw.githubusercontent.com/Permify/permify/master/assets/permify-demo-github.gif" alt="Permify - Open source authorization as a service"  width="820px" />
</p>

## What is Permify?

Permify is an open-source authorization service for creating and maintaining fine-grained authorizations. You can run Permify container with docker and it works as a Rest API.

Permify converts and syncs your authorization data as relation tuples into your preferred database. And you can check authorization with single request based on those tuples.

Data model is inspired by [Google Zanzibar White Paper](https://storage.googleapis.com/pub-tools-public-publication-data/pdf/41f08f03da59f5518802898f68730e247e23c331.pdf).

## Why Permify?

You can use Permify any stage of your development for your authorization needs but Permify works best:

- If you want to create unified control mechanism for individual applications.
- If you need to refactor your authorization.
- If your data model is getting too complicated to handle your authorization within the service.
- If you’re managing authorization for growing micro-service infrastructure.
- If your authorization logic is cluttering your code base.
- If your authorization is growing too complex to handle within code or API gateway.

## Features

- Sync & coordinate your authorization data hassle-free.
- Get Boolean - Yes/No decision returns.
- Store your authorization data in-house with high availability & low latency.
- Easily model, debug & refactor your authorization logic.
- Enforce authorizations with a single request anywhere you call it.
- Low latency with parallel graph engine for enforcement check.

## Example Access Check

Permify helps you convert & sync authorization data to a database you point at with a YAML config file. And after you model your authorization with Permify's DSL - Permify Schema, you can perform access checks with a single call anywhere on your app. Access decisions made according to stored relational tuples.

**Request**

```json
{
  "entity": {
    "type": "repository",
    "id": "1"
  },
  "action": "read",
  "subject": {
    "type":"user",
    "id": "1"
  }
}
```

***Can the user 1 push on a repository 1 ?***

**Response**

```json
{
  "can": false, // main decision
  "decisions": { // descion logs
    "organization:1#member": {
      "prefix": "not",
      "can": false,
      "err": null
    },
    "repository:1#owner": {
      "prefix": "",
      "can": true,
      "err": null
    }
  }
}
```
## Getting Started

- [Install Permify] with running Permify container using docker.
- Follow a guide to model your authorization using [Permify Schema].
- Learn how Permify [moves & syncs your authorization data].
- Take a look at the overview of [Permify API].

[Install Permify]: https://docs.permify.co/docs/installation
[Permify Schema]: https://docs.permify.co/docs/getting-started/modeling
[moves & syncs your authorization data]: https://docs.permify.co/docs/getting-started/sync-data
[Permify API]: https://docs.permify.co/docs/api-overview

## Client SDKs

We are building SDKs to make installation easier, leave us a feedback on which SDK we should build first.

[//]: # (Stargazers)

[//]: # (-----------)

[//]: # ()
[//]: # ([![Stargazers repo roster for @Permify/permify]&#40;https://reporoster.com/stars/Permify/permify&#41;]&#40;https://github.com/Permify/permify/stargazers&#41;)

## Community & Support
You can join the conversation at our [Discord channel](https://discord.gg/MJbUjwskdH). We love to talk about authorization and access control - we would
love to hear from you :heart:

If you like Permify, please consider giving us a :star:️

<h2 align="left">:heart: Let's get connected:</h2>

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

## License

Licensed under the Apache License, Version 2.0: http://www.apache.org/licenses/LICENSE-2.0
