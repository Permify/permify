---
id: api-overview
title: API Overview
sidebar_label: Using the API
slug: /api-overview
---

# Overview

Permify API provides various functionalities around authorization such as performing access checks, reading and writing relation tuples, expanding your permissions (schema actions), and more.

We structured Permify API in 3 core parts:

- [PermissionService]: Consists access control requests and options.
- [RelationshipService]: Authorization data operations such as creating, deleting and reading relational tuples.
- [SchemaService]: Modeling and Permify Schema related functionalities including configuration and auditing.
- [TenancyService]: Consists tenant operations such as creating, deleting and listing.

Permify exposes its APIs via both [gRPC](https://buf.build/permify/permify/docs/main:base.v1) - with [go] and [nodeJS] client options - and [REST](https://restfulapi.net/). 

[PermissionService]: ./api-overview/permission
[RelationshipService]: ./api-overview/relationship
[SchemaService]: ./api-overview/schema
[TenancyService]: ./api-overview/tenancy

[go]: https://github.com/Permify/permify-go
[nodeJS]: https://github.com/Permify/permify-node

[![Run in Postman](https://run.pstmn.io/button.svg)](https://www.postman.com/permify-dev/workspace/permify/collection)
[![View in Swagger](http://jessemillar.github.io/view-in-swagger-button/button.svg)](https://permify.github.io/permify-swagger/)

## Core Paths

- Configure your authorization model with [Schema Write](./api-overview/schema/write-schema.md)
- Write relational tuples with [Write Relationships](./api-overview/relationship/write-relationships.md)
- Read relation tuples and filter them with [Read API](./api-overview/relationship/read-api.md)
- Check access with [Check API](./api-overview/permission/check-api.md)
- Check entities permissions with [Lookup Entity](./api-overview/permission/lookup-entity.md)
- Delete relation tuples with [Delete Tuple](./api-overview/relationship/delete-relationships.md)
- Expand schema actions with [Expand API](./api-overview/permission/expand-api.md)
- Get permissions of your resources with [Schema Lookup](./api-overview/permission/schema-lookup.md)

## Authentication

You can secure APIs with our authentication methods; **Open ID Connect** or **Pre Shared Keys**. They can be configurable with flags or using configuration yaml file. See more details how to enable authentication from [Configuration Options](http://localhost:3000/docs/reference/configuration)

To access the endpoints after enabling authentication, it's necessary to provide a Bearer Token for identification. If your using golang or nodeJs client library, an authentication token can be provided via interceptors. You can find details in the clients' documentation.

## Need any help ?

Our team is happy to help you get started with Permify. If you'd like to learn more about using Permify in your app or have any questions about this example, [schedule a call with one of our Permify engineer](https://meetings-eu1.hubspot.com/ege-aytin/call-with-an-expert).

