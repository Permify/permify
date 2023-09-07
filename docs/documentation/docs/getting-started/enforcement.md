---
sidebar_position: 4
---

# Interacting With The API

Permify API provides various functionalities around authorization such as performing access checks, reading and writing relation tuples, expanding your permissions (schema actions), and more.

We structured Permify API in 4 core parts:

- [PermissionService]: Consists access control requests, permission checks and their configurations.
- [RelationshipService]: Authorization data operations such as creating, deleting and reading relational tuples.
- [SchemaService]: Modeling and Permify Schema related functionalities including configuration and auditing.
- [TenancyService]: Consists tenant operations such as creating, deleting and listing.

Permify exposes its APIs via both [gRPC](https://buf.build/permify/permify/docs/main:base.v1) - with [go] and [nodeJS] client options - and [REST](https://restfulapi.net/).

[PermissionService]: ../api-overview/permission
[RelationshipService]: ../api-overview/relationship
[SchemaService]: ../api-overview/schema
[TenancyService]: ../api-overview/tenancy
[go]: https://github.com/Permify/permify-go
[nodeJS]: https://github.com/Permify/permify-node

[![Run in Postman](https://run.pstmn.io/button.svg)](https://www.postman.com/permify-dev/workspace/permify/collection)
[![View in Swagger](http://jessemillar.github.io/view-in-swagger-button/button.svg)](https://permify.github.io/permify-swagger/)

:::info Integration with a Service Mesh
Our software does not include built-in support for service meshes (eg. Istio). 

However, since it communicates using standard protocols like gRPC and HTTP, it is compatible with Istio and similar service meshes. Users will need to configure their service mesh setup manually to manage traffic for our software within their deployment environment.
:::

## Core Paths

- Configure your authorization model with [Schema Write](../api-overview/schema/write-schema.md)
- Write relational tuples with [Write Relationships](../api-overview/relationship/write-relationships.md)
- Read relation tuples and filter them with [Read API](../api-overview/relationship/read-api.md)
- Check access with [Check API](../api-overview/permission/check-api.md)
- Check entities permissions with [Lookup Entity](../api-overview/permission/lookup-entity.md)
- Check subject permissions with [Lookup Subject](../api-overview/permission/lookup-subject.md)
- Delete relation tuples with [Delete Tuple](../api-overview/relationship/delete-relationships.md)
- Expand schema actions with [Expand API](../api-overview/permission/expand-api.md)

## Authentication

You can secure APIs with our authentication methods; **Open ID Connect** or **Pre Shared Keys**. They can be configurable with flags or using configuration yaml file. See more details how to enable authentication from [Configuration Options](../reference/configuration.md)

To access the endpoints after enabling authentication, it's necessary to provide a Bearer Token for identification. If your using golang or nodeJs client library, an authentication token can be provided via interceptors. You can find details in the clients' documentation.

## Availability of the Service

For our dedicated instance service we do have **99.9%** level of availability and to assure this level of availability, we employ several strategies:

1. **Redundancy:** We deploy our system across multiple Availability Zones in a region, ensuring that it remains operational even if one zone experiences issues.
2. **Load Balancing:** Load balancers are used to distribute traffic across multiple instances of the service, ensuring that no single instance becomes a bottleneck.
3. **Auto-Scaling:** Our system is capable of scaling automatically based on the incoming load, ensuring that we have sufficient capacity to handle any increase in traffic.
4. **Data Replication:** Our PostgreSQL database replicates data to ensure its availability even in the event of a single-node failure.
5. **Backup and Recovery:** Regular backups are maintained, and our system supports a robust recovery strategy in case of significant failures.
6. **Monitoring & Alerts:** Using tools like Amazon CloudWatch, we monitor the health and performance of our system and can quickly respond to any detected issues.

## Service Credits for Availability Failures

In case of availability failures, Permify's Service Level Agreement (SLA) provides for Service Credits which are applied as a discount on your future bills:

- If uptime is less than 99.95% but above or equal to 99.0%, you get a 10% Service Credit.
- If uptime is less than 99.0%, you get a 25% Service Credit.
- If uptime is less than 95.0%, you get a 100% Service Credit.

These credits are your sole remedy for any availability failures under our SLA.

## Performance & Latency

Permify designed to answer these authorization questions efficiently and with minimal complexity while providing low latency with:
- Using its parallel graph engine. 
- Storing the relationships between resources beforehand in Permify data store: [writeDB], rather than providing these relationships at “check” time.
- Implementing permission caching to not recompute repeated permission checks, and in memory cache to store authorization schema.
- Using [Snap Tokens](../reference/snap-tokens) to achieve consistency and high performance in cache.

Permify implements several cache mechanisms in order to achieve low latency in scaled distributed systems. See more on the section [Cache Mechanisms](../reference/cache.md) 

[writeDB]: ../getting-started/sync-data.md

## Request Rate Limits

Default rate limit is set to 100 requests per second. However, users can adjust this based on their specific needs following our documentation. 

- doc: https://docs.permify.co/docs/reference/configuration
- algorithm: https://en.wikipedia.org/wiki/Token_bucket

## Need any help ?

Our team is happy to help you get started with Permify. If you'd like to learn more about using Permify in your app or have any questions about this example, [schedule a call with one of our Permify engineer](https://meetings-eu1.hubspot.com/ege-aytin/call-with-an-expert).
