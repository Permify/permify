---
title: "Docker Container"
---

# Run using Docker

This section shows how to run Permify using our docker container. You can run Permify using Docker with following command.

## Run in a terminal

```shell
docker run -p 3476:3476 -p 3478:3478 -v {YOUR-CONFIG-PATH}:/config ghcr.io/permify/permify serve
```

This will start a Permify server with the configuration that is in **{YOUR-CONFIG-PATH}**.

### Configure with a YAML file

This config path - `{YOUR-CONFIG-PATH}` - should contain the [config yaml file](../reference/configuration.md), where you can configure the Permify Server as well as define the ***database*** to store your authorization related data in.

:::info Talk to an Permify Engineer
By default, the container is configured to listen on ports 3476 (HTTP) and 3478 (gRPC) and store the authorization data in memory rather than an actual database.
:::

### Configure Using Flags

Alternatively, you can set configuration options using flags when running the command. See all the configuration flags by running,

```shell
docker run -p 3476:3476 -p 3478:3478 ghcr.io/permify/permify serve -help
```

:::info Environment Variables
In addition to CLI flags, Permify also supports configuration via environment variables. You can replace any flag with an environment variable by converting dashes into underscores and prefixing with PERMIFY_ (e.g. **--log-level** becomes **PERMIFY_LOG_LEVEL**). 
:::

### Test your connection.

You can test your connection by making an HTTP GET request,

```shell
localhost:3476/healthz
```

You can use our Postman Collection to work with the API. Also see the [Using the API] section for details of core functions.

[Using the API]: ../api-overview.md

[![Run in Postman](https://run.pstmn.io/button.svg)](https://www.postman.com/permify-dev/workspace/permify/collection)
[![View in Swagger](http://jessemillar.github.io/view-in-swagger-button/button.svg)](https://permify.github.io/permify-swagger/)


### Need any help ?

Our team is happy to help you get started with Permify, [schedule a call with a Permify engineer](https://meetings-eu1.hubspot.com/ege-aytin/call-with-an-expert).
