
<div align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="https://github.com/Permify/permify/raw/master/assets/logo-permify-dark.png">
    <img alt="Permify logo" src="https://github.com/Permify/permify/raw/master/assets/logo-permify-light.png" width="40%">
  </picture>
<h1 align="center">
   Permify - Open Source Fine-Grained Authorization
</h1>
</div>
<p align="center">
    Implement fine-grained, scalable and extensible access controls within minutes to days instead of months. <br>
    Inspired by Google’s consistent, global authorization system, <a href="https://permify.co/post/google-zanzibar-in-a-nutshell/" target="_blank">Zanzibar</a>
</p>

<p align="center">
    <a href="https://trendshift.io/repositories/5027" target="_blank"><img src="https://trendshift.io/api/badge/repositories/5027" alt="Permify%2Fpermify | Trendshift" style="width: 250px; height: 55px;" width="250" height="55"/></a>
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

![permify-centralized](https://github.com/user-attachments/assets/124eaa43-5d33-423d-a258-5d6f4afbc774)

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

## Permify Cloud vs Self-hosted?

Permify is [open-source authorization service](https://permify.co/) and we have a free and self-hosted solution called **Permify Community Edition (CE)**. Here are the differences between Permify managed hosting in the cloud and the Permify CE:

|  | Permify Cloud  | Permify Community Edition |
| ------------- | ------------- | ------------- |
| **Infrastructure management** | Easy and convenient. It takes minutes to start permissions systems deployed on secure Permify infrastructure with a high availability, backups, security and maintenance all done for you by us. We manage everything so you don’t have to worry about anything and can focus on your core development. | You do it all yourself. You need to get a server and you need to manage your infrastructure. You are responsible for installation, maintenance, upgrades, server capacity, uptime, backup, security, stability, consistency, latency and so on.|
| **Release schedule** | Continuously developed and improved with new features and updates multiple times per week. | It's a long-term release published four times per year, so the latest features and improvements won’t be immediately available.|
| **Premium features** | All features available as listed in [our pricing plans](https://permify.co/pricing/). | Selected premium features, such as observability dashboards and data synchronization are not available as we aim to maintain a protective barrier around our cloud offering.|
| **Deployment regions** | You can select your preferred region, supported by **AWS**, **GCP**, or **Azure** to deploy your authorization system. Disaster recovery zones are strategically located to replicate data across regions, ensuring rapid recovery and continuous service during any incident. We also provide **SLAs** to ensure availability and latency. | You have full control and can host your instance on any server in any country of your choice. This includes hosting on personal servers or with cloud providers. |
| **Data privacy** | Permify Cloud is **SOC2 and GDPR compliant**, ensuring adherence to stringent data protection standards. You can check out our [Trust Center](https://trust.permify.co/) for comprehensive insights into our data management, security measures, and compliance practices. | Data privacy management is your responsibility. While you have full control over your data, it is up to you to implement and maintain necessary compliance measures, such as GDPR or SOC 2, as well as other security protocols. |
| **Premium support** | Real support delivered by real human beings who build and maintain Permify. | Premium support is not included. CE is community supported only.|
| **Costs** | There’s a cost associated with providing an authorization service, so we base our pricing on the number of monthly active users you have. Your payments fund the further development of Permify. | You need to pay for your server, CDN, backups, and other costs associated with running the infrastructure you need. |

Interested in trying out Permify Cloud? Our team is happy to help. [Schedule a quick demo](https://permify.co/book-demo/) session with our experts.

### QuickStart

You can quickly start Permify on your local with running the docker command below:

```shell
docker run -p 3476:3476 -p 3478:3478 ghcr.io/permify/permify serve
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

## 🚀 Performance

We conducted a load test on **Permify** using **1000 VUs (Virtual Users)** and **10,000 RPS (Requests per Second)**. The results demonstrate strong performance and reliability under heavy load, with **0% request failures and consistently low latency**.  

| **Metric**                     | **Value / Stats**                                                              |
|-------------------------------|--------------------------------------------------------------------------------|
| **Total Checks**              | ✅ 100.00% (75,369 out of 75,369)                                               |
| **Data Received**             | 15 MB (145 kB/s)                                                               |
| **Data Sent**                 | 21 MB (203 kB/s)                                                               |
| **Dropped Iterations**        | 271,664 (2,688.45/s)                                                           |
| **HTTP Request Duration**     | avg = 10.14ms · p(90) = 14.3ms · p(95) = 26.34ms · max = 295.29ms              |
| **HTTP Request Waiting Time** | avg = 9.96ms · p(90) = 14.1ms · p(95) = 26.17ms · max = 295.21ms              |
| **HTTP Request Failed**       | ❌ 0.00% (0 out of 75,369)                                                      |
| **Total HTTP Requests**       | 75,369 (745.87/s)                                                              |
| **Virtual Users (VUs)**       | 46 avg (min = 13, max = 1000)                                                  |

📄 **[Full Performance Test Report →](/docs/performance-test/README.md)**

## Community ♥️

Permify is a [Cloud Native Computing Foundation](https://www.cncf.io/) member and a community-driven project supported by companies worldwide, from startups to Fortune 500 enterprises.

Your feedback helps shape the future of Permify, and we'd love to hear from you!

Share your use case, get the latest product updates, and feel free to ask any questions about Permify or authorization in a broader context by joining our conversation on Discord!

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
