---
id: api-overview
title: API Overview
sidebar_label: Using the API
slug: /api-overview
---

# Overview

Permify API provides various functionalities around authorization such as performing access checks, reading and writing relation tuples, expanding your permissions (schema actions), and more.

We structured Permify API in 3 core parts; *modeling authorization*, *storing authorization data* and *enforcement*. Therefore, Permify API has sections that represent the functionalities of these core parts.

- **Permission Section**: Consist enforcement requests and options.
- **Relationship Section**: Authorization data operations such as creating, deleting and reading relational tuples.
- **Schema Section**: Modeling and Permify Schema related functionalities including configuration and auditing.

Permify exposes its APIs via both [gRPC](https://buf.build/permify/permify/docs/main:base.v1) and [REST](https://restfulapi.net/).

[![Run in Postman](https://run.pstmn.io/button.svg)](https://god.gw.postman.com/run-collection/16122080-54b1e316-8105-4440-b5bf-f27a05a8b4de?action=collection%2Ffork&collection-url=entityId%3D16122080-54b1e316-8105-4440-b5bf-f27a05a8b4de%26entityType%3Dcollection%26workspaceId%3Dd3a8746c-fa57-49c0-83a5-6fcf25a7fc05)
[![View in Swagger](http://jessemillar.github.io/view-in-swagger-button/button.svg)](https://app.swaggerhub.com/apis-docs/permify/permify/latest)

## Core Paths

- Check access with [Check API](./api-overview/check-api.md)
- Configure your authorization model with [Schema Write](./api-overview/write-schema.md)
- Write relational tuples with [Write API](./api-overview/write-relationships.md)
- Read relation tuples and filter them with [Read API](./api-overview/read-api.md)
- Delete relation tuples with [Delete Tuple](./api-overview/delete-relationships.md)
- Expand schema actions with [Expand API](./api-overview/expand-api.md)
- Get permissions of your resources with [Schema Lookup](./api-overview/schema-lookup.md)

## Need any help ?

Our team is happy to help you get started with Permify. If you'd like to learn more about using Permify in your app or have any questions about this example, [schedule a call with one of our Permify engineer](https://meetings-eu1.hubspot.com/ege-aytin/call-with-an-expert).

