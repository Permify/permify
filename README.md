
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
    <a href="https://scrutinizer-ci.com/g/Permify/permify/?branch=master" target="_blank"><img src="https://img.shields.io/scrutinizer/quality/g/Permify/permify/master?style=for-the-badge" alt="Scrutinizer code quality (GitHub/Bitbucket)" /></a>&nbsp;
    <a href='https://coveralls.io/github/Permify/permify?branch=master'><img alt="Coveralls" src="https://img.shields.io/coverallsCoverage/github/Permify/permify?style=for-the-badge"></a>
</p>

![Permify - Open source authorization as a service](https://github.com/Permify/permify/assets/39353278/06262e07-84ba-4a1c-b859-870344396600)

## Join Our Team

Permify is on the lookout for engineers eager to tackle challenges in authorization. Join us!

<a href="http://permify.co/company/career/" target="_blank"><img src="https://img.shields.io/badge/We%20Are%20Hiring!-blue?style=for-the-badge" alt="We are hiring" /></a>&nbsp;

## What is Permify?

[Permify](https://github.com/Permify/permify) is a open-source authorization service for creating and managing fine-grained permissions in your applications and services. Inspired by Google’s consistent, global authorization system, [Google Zanzibar](https://storage.googleapis.com/pub-tools-public-publication-data/pdf/41f08f03da59f5518802898f68730e247e23c331.pdf)

Our goal is to make Google's Zanzibar available to everyone and help them build robust, flexible, and easily auditable authorization systems that perform well in scaled environments.

### With Permify, you can:

🔮 Create permissions and policies using [Permify's flexible authorization language](https://docs.permify.co/docs/getting-started/modeling) that is compatible with traditional roles and permissions (RBAC), arbitrary relations between users and objects (ReBAC), and attributes (ABAC).

🔐 [Manage and store authorization data](https://docs.permify.co/docs/getting-started/sync-data) in your preferred database with high availability and consistency.

✅ [Interact with the Permify API](https://docs.permify.co/docs/getting-started/enforcement) to perform access checks, filter your resources with specific permissions, perform bulk permission checks for various resources, and more.

🧪 Test your authorization logic with [Permify's schema testing](https://docs.permify.co/docs/getting-started/testing). You can conduct scenario-based testing, policy coverage analysis, and IDL parser integration to achieve end-to-end validations for your desired authorization schema.

⚙️ Create custom and isolated authorization models for different applications using Permify [Multi-Tenancy](https://docs.permify.co/docs/use-cases/multi-tenancy) support, all managed within a single place, Permify instance.

## Getting Started 

- Follow a guide to model your authorization using [Permify's Authorization Language].
- See our [Playground], build your authorization logic and test it with sample data.
- Explore overview of [Permify API] and learn how to interact with it.
- See [our article] to examine [Google Zanzibar](https://storage.googleapis.com/pub-tools-public-publication-data/pdf/41f08f03da59f5518802898f68730e247e23c331.pdf) in a nutshell.

[Permify's Authorization Language]: https://docs.permify.co/docs/getting-started/modeling
[playground]: https://play.permify.co/
[Permify API]: https://docs.permify.co/docs/api-overview
[our article]: https://permify.co/post/google-zanzibar-in-a-nutshell

### QuickStart

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

#### Test your connection

You can test your connection with creating an GET request,

```shell
localhost:3476/healthz
```

## Community ♥️

We would love to hear from you!

Get the latest product updates, receive immediate assistance from our team members, and feel free to ask any questions about Permify or authorization in a broader context by joining our conversation on Discord!

<a href="https://discord.gg/JJnMeCh6qP" target="_blank"><img src="https://img.shields.io/badge/Join%20Our%20Discord!-blueviolet?style=for-the-badge" alt="Join Our Discord" /></a>&nbsp;

## Contributing

[![Open Bounties](https://img.shields.io/endpoint?url=https%3A%2F%2Fconsole.algora.io%2Fapi%2Fshields%2Fpermify%2Fbounties%3Fstatus%3Dopen)](https://console.algora.io/org/permify/bounties?status=open)

The open source community thrives on contributions, offering an incredible space for learning, inspiration, and creation. Your contributions are immensely valued and appreciated!

Here are the ways to contribute to Permify:

* **Contribute to codebase:** We're collaboratively working with our community to make Permify the best it can be! You can develop new features, fix existing issues or make third-party integrations/packages. 
* **Improve documentation:** Alongside our codebase, documentation one of the most significant part in our open-source journey. We're trying to give the best DX possible to explain ourselves and Permify. And you can help on that with importing resources or adding new ones.
* **Contribute to playground:** Permify playground allows you to visualize and test your authorization logic. You can contribute to our playground by improving its user interface, fixing glitches, or adding new features.

Before start please see our [contributing guide](https://github.com/Permify/permify/blob/master/CONTRIBUTING.md).

### Good First Issues

Don't know where to start? We have a list of [good first issues](https://github.com/Permify/permify/blob/master/CONTRIBUTING.md) consisting of small features and bugs. This presents an excellent starting point to gain experience, and get familiar with our contribution process.

## Roadmap

You can find Permify's Public Roadmap [here](https://github.com/orgs/Permify/projects/1)!

## Stargazers

[![Stargazers repo roster for permify/permify](https://reporoster.com/stars/permify/permify)](https://github.com/permify/permify/stargazers)

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
