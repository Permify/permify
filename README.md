

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


Permify is an open-source authorization service that you can run with docker and works on a Rest API.

Permify converts, coordinate, and sync your authorization data as relation tuples into your preferred database. And you can check authorization with single request based on those tuples and [Permify Schema](https://github.com/Permify/permify/blob/master/assets/content/MODEL.md), where you model your authorization.

Data model is inspired by [Google Zanzibar White Paper](https://storage.googleapis.com/pub-tools-public-publication-data/pdf/41f08f03da59f5518802898f68730e247e23c331.pdf).

## Getting Started
Permify consists of 3 main parts; modeling authorization, synchronizing authorization data and access checks.

- [Modeling Authorization]
- [Move & Synchronize Authorization Data]
- [Access Checks]

[Modeling Authorization]: https://github.com/Permify/permify/blob/master/assets/content/MODEL.md
[Move & Synchronize Authorization Data]: https://github.com/Permify/permify/blob/master/assets/content/SYNC.md
[Access Checks]: https://github.com/Permify/permify/blob/master/assets/content/ENFORCEMENT.md

## Installation

### Container (Docker)

#### With terminal

1. Open your terminal.
2. Run following line.

```shell
docker run -d -p 3476:3476 --name permify-container -v {YOUR-CONFIG-PATH}:/config permify/permify:0.0.1
```

3. Test your connection.
    - Create an HTTP GET request ~ localhost:3476/v1/status/ping

#### With docker desktop

Setup docker desktop, and run service with the following steps;

1. Open your docker account.
2. Open terminal and run following line

```shell
docker pull permify/permify:0.0.1
```

3. Open images, and find Permify.
4. Run Permify with the following credentials (optional setting)
    - Container Name: authorization-container
      Ports
    - **Local Host:** 3476
      Volumes
    - **Host Path:** choose the config file and folder
    - **Container Path:** /config
5. Test your connection.
    - Create an HTTP GET request ~ localhost:3476/v1/status/ping

## Why Permify?

You can use Permify any stage of your development for your authorization needs but Permify works best:

- If you need to refactor your authorization.
- If you’re managing authorization for growing micro-service infrastructure.
- If your authorization logic is cluttering your code base.
- If your data model is getting too complicated to handle your authorization within the service.
- If your authorization is growing too complex to handle within code or API gateway.

## Features

- Sync & coordinate your authorization data hassle-free.
- Get Boolean - Yes/No decision returns.
- Store your authorization data in-house with high availability & low latency.
- Easily model, debug & refactor your authorization logic.
- Enforce authorizations with a single request anywhere you call it.
- Low latency with parallel graph engine for enforcement check.

## Example

Permify helps you move & sync authorization data from your ListenDB to WriteDB with a single config file based on your
authorization model that you provide us in a YAML schema.
After configuration, you can check authorization with a simple call.
**Request**

```json
{
  "user": "1",
  "action": "push",
  "object": "repository:1"
}
```

**Response**

```json
{
  "can": false, // main decision
  "decisions": { // decision logs
    "repository:1#parent.admin": {
      "can": false,
      "err": null
    },
    "repository:1#parent.member": {
      "can": false,
      "err": null
    }
  }
}
```

Check out [Permify API](https://github.com/Permify/permify/blob/master/assets/content/API.md) for more details.

## Client SDKs

We are building SDKs to make installation easier, leave us a feedback on which SDK we should build first.

[//]: # (Stargazers)

[//]: # (-----------)

[//]: # ()
[//]: # ([![Stargazers repo roster for @Permify/permify]&#40;https://reporoster.com/stars/Permify/permify&#41;]&#40;https://github.com/Permify/permify/stargazers&#41;)

## Community
You can join the conversation at our [Discord channel](https://discord.gg/MJbUjwskdH). We love to talk about authorization and access control - we would
love to hear from you :heart:
If you like Permify, please consider giving us a :star:️

<h2 align="left">:heart: Let's get connected:</h2>

<p align="left">
<a href="https://discord.gg/MJbUjwskdH">
 <img alt="guilyx’s Discord" width="50px" src="https://user-images.githubusercontent.com/34595361/178992169-fba31a7a-fa80-42ba-9d7f-46c9c0b5a9f8.png" />
</a>
<a href="https://twitter.com/GetPermify">
  <img alt="guilyx | Twitter" width="50px" src="https://user-images.githubusercontent.com/43545812/144034996-602b144a-16e1-41cc-99e7-c6040b20dcaf.png"/>
</a>
<a href="https://www.linkedin.com/company/permifyco">
  <img alt="guilyx's LinkdeIN" width="50px" src="https://user-images.githubusercontent.com/43545812/144035037-0f415fc7-9f96-4517-a370-ccc6e78a714b.png" />
</a>
</p>

## License

Licensed under the Apache License, Version 2.0: http://www.apache.org/licenses/LICENSE-2.0
