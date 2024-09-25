[permify-banner](https://github.com/user-attachments/assets/a4ff3ba8-8b2e-40cd-85f6-e6dac7a0359b)

<h1 align="center">
    Permify - Open Source Fine-Grained Authorization
</h1>
<p align="center">
    Implement fine-grained, scalable and extensible access controls within minutes to days instead of months. <br>
    Inspired by Google’s consistent, global authorization system, <a href="https://permify.co/post/google-zanzibar-in-a-nutshell/" target="_blank">Zanzibar</a>
</p>

<p align="center">
    <a href="https://github.com/Permify/permify" target="_blank"><img src="https://img.shields.io/github/go-mod/go-version/Permify/permify?style=for-the-badge&logo=go" alt="Permify Go Version" /></a>&nbsp;
    <a href="https://goreportcard.com/report/github.com/Permify/permify" target="_blank"><img src="https://goreportcard.com/badge/github.com/Permify/permify?style=for-the-badge&logo=go" alt="Permify Go Report Card" /></a>&nbsp;
    <a href="https://github.com/Permify/permify" target="_blank"><img src="https://img.shields.io/github/license/Permify/permify?style=for-the-badge" alt="Permify Licence" /></a>&nbsp;
    <a href="https://discord.gg/n6KfzYxhPp" target="_blank"><img src="https://img.shields.io/discord/950799928047833088?style=for-the-badge&logo=discord&label=DISCORD" alt="Permify Discord Channel" /></a>&nbsp;
    <a href="https://github.com/Permify/permify/pkgs/container/permify" target="_blank"><img src="https://img.shields.io/github/v/release/permify/permify?include_prereleases&style=for-the-badge" alt="Permify Release" /></a>&nbsp;
    <a href="https://img.shields.io/github/commit-activity/m/Permify/permify?style=for-the-badge" target="_blank"><img src="https://img.shields.io/github/commit-activity/m/Permify/permify?style=for-the-badge" alt="Permify Commit Activity" /></a>&nbsp;
    <a href="https://img.shields.io/github/actions/workflow/status/Permify/permify/release.yml?style=for-the-badge" target="_blank"><img src="https://img.shields.io/github/actions/workflow/status/Permify/permify/release.yml?style=for-the-badge" alt="GitHub Workflow Status" /></a>&nbsp;
    <a href="https://scrutinizer-ci.com/g/Permify/permify/?branch=master" target="_blank"><img src="https://img.shields.io/scrutinizer/quality/g/Permify/permify/master?style=for-the-badge" alt="Scrutinizer code quality (GitHub/Bitbucket)" /></a>&nbsp;
    <a href='https://coveralls.io/github/Permify/permify?branch=master'><img alt="Coveralls" src="https://img.shields.io/coverallsCoverage/github/Permify/permify?style=for-the-badge"></a>
</p>        

![permify-centralized](https://github.com/user-attachments/assets/e1c22244-1fa4-4bc3-8b7a-bdfb97610c5f)

## What is Permify?

[Permify](https://github.com/Permify/permify) is an open-source authorization service for easily building and managing fine-grained, scalable, and extensible access controls for your applications and services. Inspired by Google’s consistent, global authorization system, [Google Zanzibar](https://storage.googleapis.com/pub-tools-public-publication-data/pdf/41f08f03da59f5518802898f68730e247e23c331.pdf)

Our service makes authorization more secure and adaptable to changing needs, allowing you to get it up and running in just a few minutes to a couple of days—no need to spend months building out entire piece of infrastructure.

It works in run time and can respond to any type of access control checks (can user X view document Y?, which posts can members of team Y edit?, etc.) from any of your apps and services in tens of milliseconds.

### With Permify, you can:

🧪 **Centralize & Standardize Your Authorization**: Abstract your authorization logic from your codebase and application logic to easily reason, test, and debug your authorization. Behave your authorization as a sole entity and move faster with in your core development.

🔮 **Build Granular Permissions For Any Case You Have:** You can create granular (resource-specific, hierarchical, context aware, etc) permissions and policies using Permify's domain specific language that is compatible with RBAC, ReBAC and ABAC.

🔐 **Set Authorization For Your Tenants By Default**: Set up isolated authorization logic and custom permissions for your vendors/organizations (tenants) and manage them in a single place.

🚀 **Scale Your Authorization As You Wish:** Achieve lightning-fast response times down to 10ms for access checks with a proven infrastructure inspired by Google Zanzibar.

## Getting Started 

- Follow a guide to model your authorization using [Permify's Authorization Language].
- See our [Playground], build your authorization logic and test it with sample data.
- Explore overview of [Permify API] and learn how to interact with it.
- See [our article] to examine [Google Zanzibar](https://storage.googleapis.com/pub-tools-public-publication-data/pdf/41f08f03da59f5518802898f68730e247e23c331.pdf) in a nutshell.
- Explore our [SDK samples] for hands-on examples.

[Permify's Authorization Language]: https://docs.permify.co/getting-started/modeling
[playground]: https://play.permify.co/
[Permify API]: https://docs.permify.co/api-reference
[our article]: https://permify.co/post/google-zanzibar-in-a-nutshell
[SDK samples]: https://github.com/Permify/permify/tree/master/sdk

### QuickStart

You can quickly start Permify on your local with running the docker command below:

```shell
docker run -p 3476:3476 -p 3478:3478  ghcr.io/permify/permify
```

This will start Permify with the default configuration options: 
* Port 3476 is used to serve the REST API.
* Port 3478 is used to serve the GRPC Service.
* Authorization data stored in memory.

See [all of the options] that you can use to set up and deploy Permify in your servers.

[all of the options]: https://docs.permify.co/setting-up

#### Test your connection

You can test your connection with creating a GET request,

```shell
localhost:3476/healthz
```

## Community ♥️

We would love to hear from you!

Get the latest product updates, receive immediate assistance from our team members, and feel free to ask any questions about Permify or authorization in a broader context by joining our conversation on Discord!

<a href="https://discord.gg/n6KfzYxhPp" target="_blank"><img src="https://img.shields.io/badge/Join%20Our%20Discord!-blueviolet?style=for-the-badge" alt="Join Our Discord" /></a>&nbsp;

## Contributing

The open source community thrives on contributions, offering an incredible space for learning, inspiration, and creation. Your contributions are immensely valued and appreciated!

Here are the ways to contribute to Permify:

* **Contribute to codebase:** We're collaboratively working with our community to make Permify the best it can be! You can develop new features, fix existing issues or make third-party integrations/packages. 
* **Improve documentation:** Alongside our codebase, documentation one of the most significant part in our open-source journey. We're trying to give the best DX possible to explain ourselves and Permify. And you can help on that with importing resources or adding new ones.
* **Contribute to playground:** Permify playground allows you to visualize and test your authorization logic. You can contribute to our playground by improving its user interface, fixing glitches, or adding new features.

### Bounties 
[![Open Bounties](https://img.shields.io/endpoint?url=https%3A%2F%2Fconsole.algora.io%2Fapi%2Fshields%2Fpermify%2Fbounties%3Fstatus%3Dopen&style=for-the-badge)](https://console.algora.io/org/permify/bounties?status=open)

We have a list of [issues](https://github.com/Permify/permify/labels/%F0%9F%92%8E%20Bounty) where you can contribute and gain bounty award! Bounties will be awarded for fixing issues via accepted Pull Requests (PR).

Before start please see our [contributing guide](https://github.com/Permify/permify/blob/master/CONTRIBUTING.md).

## Roadmap

You can find Permify's Public Roadmap [here](https://github.com/orgs/Permify/projects/1)!

## Contributors ♥️

<a href="https://github.com/permify/Permify/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=permify/Permify&anon=1" />
</a>

## Communication Channels

If you like Permify, please consider giving us a :star:

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
