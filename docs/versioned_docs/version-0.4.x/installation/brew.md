---
title: "Install with Brew"
---

# Brew With Configurations

This section shows how to install and run Permify Service with using brew. 

### Install Permify

Open terminal and run following line,

```shell
brew install permify/tap/permify
```

### Run Permify Service 

To run the Permify Service, `permify serve` command should be run with configurations.

By default, the service is configured to listen on ports 3476 (HTTP) and 3478 (gRPC) and store the authorization data in memory rather then an actual database. You can override these with running the command with configuration flags. 

### Configure With Using Flags

See all configuration flags with running,

```shell
permify serve --help
```

:::info Environment Variables
In addition to CLI flags, Permify also supports configuration via environment variables. You can replace any flags' argument with an environment variable by converting dashes into underscores and prefixing with PERMIFY_ (e.g. **--log-level** becomes **PERMIFY_LOG_LEVEL**). 
:::

### Configure With Using Config File

You can also configure Permify Service with using a configuration file.

```shell
 permify serve -c=config.yaml
```

or 

```shell
 permify serve --config=config.yaml
```

### Test your connection.

You can test your connection with creating an HTTP GET request,

```shell
localhost:3476/healthz
```

You can use our Postman Collection to work with the API. Also see the [Using the API] section for details of core functions.

[Using the API]: ../../api-overview/

[![Run in Postman](https://run.pstmn.io/button.svg)](https://www.postman.com/permify-dev/workspace/permify/collection)
[![View in Swagger](http://jessemillar.github.io/view-in-swagger-button/button.svg)](https://permify.github.io/permify-swagger/)

### Need any help ?

Our team is happy to help you get started with Permify, [schedule a call with an Permify engineer](https://meetings-eu1.hubspot.com/ege-aytin/call-with-an-expert).
