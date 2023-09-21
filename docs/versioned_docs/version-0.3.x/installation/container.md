---
title: "Docker Container"
---

# Deploy using Docker

This section shows how to run Permify Service from a docker container. You can run Permify service from a container with following steps.

## Run following line on Terminal

```shell
docker run -p 3476:3476 -p 3478:3478 -v {YOUR-CONFIG-PATH}:/config ghcr.io/permify/permify serve
```

This will start an API server with the configuration options that pointed out on the **{YOUR-CONFIG-PATH}**.

### Configure With a YAML file

This config path - `{YOUR-CONFIG-PATH}:/config` - addresses the [config yaml file](../reference/configuration.md), where you can configure running options of the Permify Server as well as define the ***database*** to store your authorization related data. 

:::info Talk to an Permify Engineer
By default, the container is configured to listen on ports 3476 (HTTP) and 3478 (gRPC) and store the authorization data in memory rather than an actual database.
:::

### Configure With Using Flags

Alternatively, you can set configuration options with the respected flags when running the command. See all configuration flags with running,

```shell
docker run -p 8080:8080 ghcr.io/permify/permify serve -help
```

:::info Environment Variables
In addition to CLI flags, Permify also supports configuration via environment variables. You can replace any flags' argument with an environment variable by converting dashes into underscores and prefixing with PERMIFY_ (e.g. **--log-level** becomes **PERMIFY_LOG_LEVEL**). 
:::

### Test your connection.

You can test your connection with creating an HTTP GET request,

```shell
localhost:3476/healthz
```

You can use our Postman Collection to work with the API. Also see the [Using the API] section for details of core functions.

[Using the API]: ../api-overview.md

[![Run in Postman](https://run.pstmn.io/button.svg)](https://www.postman.com/permify-dev/workspace/permify/collection)
[![View in Swagger](http://jessemillar.github.io/view-in-swagger-button/button.svg)](https://permify.github.io/permify-swagger/)


### Need any help ?

Our team is happy to help you get started with Permify, [schedule a call with an Permify engineer](https://meetings-eu1.hubspot.com/ege-aytin/call-with-an-expert).
