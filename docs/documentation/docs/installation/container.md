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

This config path - `{YOUR-CONFIG-PATH}:/config` - addresses [config yaml file](#configuration-file), where you can configure running options of the Permify Server as well as define the ***database to store your authorization data***. 

:::info Talk to an Permify Engineer
By default, the container is configured to listen on ports 3476 (HTTP) and 3478 (gRPC) and store the authorization data in memory rather than an actual database.
:::

## Configuration File

Here is the example configuration YAML file with descriptions below. You can also find this [example config file](https://github.com/Permify/permify/blob/master/example.config.yaml) in Permify repo.

***Example config.yaml file***

```yaml
server:
  http:
    enabled: true
    port: 3476
    tls:
      enabled: true
      cert: /etc/letsencrypt/live/yourdomain.com/fullchain.pem
      key: /etc/letsencrypt/live/yourdomain.com/privkey.pem
  grpc:
    port: 3478
    tls:
      enabled: true
      cert: /etc/letsencrypt/live/yourdomain.com/fullchain.pem
      key: /etc/letsencrypt/live/yourdomain.com/privkey.pem

logger:
  level: 'debug'

authn:
  enabled: false
  keys: []

tracer:
  exporter: 'zipkin'
  endpoint: 'http://localhost:9411/api/v2/spans'
  enabled: false

meter:
  exporter: 'otlp'
  endpoint: 'localhost:4318'
  enabled: true

service:
  circuit_breaker: false
  concurrency_limit: 100

profiler:
  enabled: true
  port: 6060

database:
  engine: 'postgres'
  uri: 'postgres://user:password@host:5432/db_name'
  auto_migrate: true
  max_open_connections: 100
  max_idle_connections: 1
  max_connection_lifetime: 300s
  max_connection_idle_time: 60s
```

### Options

<details><summary>server | Server Configurations</summary>
<p>

#### Definition
Server options to run Permify. (`grpc` and `http` available for now.)

#### Structure
```
├── server
    ├── (`grpc` or `http`)
    │   ├── enabled
    │   ├── port
    │   └── tls
    │       ├── enabled
    │       ├── cert
    │       └── key
```

#### Glossary

| Required | Argument | Default | Description |
|----------|----------|---------|---------|
| [x]   | [ server_type ] | - | server option type can either be `grpc` or `http`.
| [ ]   | enabled (for server type) | true | switch option for server.  |
| [x]   | port | - | port that server run on.
| [x]   | tls | - | transport layer security options. |
| [ ]   | enabled (for tls) | false | switch option for tls  |
| [ ]   | cert | - | tls certificate path.  |
| [ ]   | key | - | tls key pat  |

</p>
</details>

<details><summary>logger | Logging Options</summary>
<p>

#### Definition
Real time logs of authorization. Permify uses [zerolog] as a logger.

[zerolog]: https://github.com/rs/zerolog

#### Structure
```
├── logger
    ├── level
```

#### Glossary

| Required | Argument | Default | Description |
|----------|----------|---------|---------|
| [x]   | level  | info | logger levels: `error`, `warn`, `info` , `debug`

</p>
</details>

<details><summary>authn | Server Authentication</summary>
<p>

#### Definition

You can choose to authenticate users to interact with Permify API.

There are 2 authentication method you can choose: 

* [Multi Tenant](#multi-tenant)
* [Pre Shared Keys](#pre-shared-keys)


#### Multi Tenant

You can add tenant ID in your JWT token and pass the secret key to Permify, so Permify can authenticate based on tenant access.

#### Structure
```
├── authn
|   ├── method
|   ├── enabled
|   ├── private_token
|   ├── algorithms
```

#### Glossary

| Required | Argument | Default | Description |
|----------|----------|---------|---------|
| [x]   | method | - | Authentication method can be either `multitenant` or `preshared`.
| [ ]   | enabled | false | switch option authentication config  |
| [x]   | private_token | - | The secret key of your JWTs
| [x]   | algorithms | - | Hashing algorithms or signing methods you choose, here are the options: `HS256`, `HS384` ,`HS512`, `RS256`, `RS384`, `RS512`, `ES256`, `ES384`, `ES512`, `Ed25519` |

#### Pre Shared Keys

On this method, you must provide a pre shared keys in order to identify yourself.

#### Structure
```
├── authn
|   ├── method
|   ├── enabled
|   ├── keys
```

#### Glossary

| Required | Argument | Default | Description |
|----------|----------|---------|---------|
| [x]   | method | - | Authentication method can be either `multitenant` or `preshared`.
| [ ]   | enabled | true | switch option authentication config  |
| [x]   | keys | - | Private key/keys for server authentication. Permify does not provide this key, so it must be generated by the users.

</p>
</details>

* **tracer** (optional)
  * **exporter:** Permify integrated with [jaeger] , [signoz] and [zipkin] tacing tools. See our [change log] about tracing performance of your authorization.
  * **endpoint:** export uri for tracing data.
  * **enabled:** switch option for tracing. *(default: false)*

* **meter** (optional)
  * **exporter:** [otpl](https://opentelemetry.io/docs/collector/) is default.
  * **endpoint:** export uri to observe metrics; check count, cache check count and session information; Permify version, hostname, os, arch. 
  * **enabled:** switch option for meter tracing. *(default: true)*

* **database** : Points out where your want to store your authorization data (relation tuples, audits, decision logs, authorization model )
  * **engine:** Data source. Permify supports **PostgreSQL**(`'postgres'`) for now. Contact with us for your preferred database. *(default: memory)*
  * **uri:** Uri of your data source.
  * **auto_migrate** : When its false migrating flow won't work *(default: true)*
  * **max_open_connections:** configuration parameter determines the maximum number of concurrent connections to the database that are allowed. *(default: 20)*
  * **max_idle_connections:** which determines the maximum number of idle connections that can be held in the connection pool.  *(default: 1)*
  * **max_connection_lifetime:** configuration parameter determines the maximum lifetime of a connection in seconds.,  *(default: 300s)*
  * **max_connection_idle_time:** configuration parameter determines the maximum time in seconds that a connection can remain idle before it is closed.  *(default: 60s)*

* **profiler** : pprof is a performance profiler for Go programs. It allows developers to analyze and understand the performance characteristics of their code by generating detailed profiles of program execution
  * **enabled:** switch option for profiler. *(default: true)*
  * **port:** port that profiler runs on *(default: 6060)*

[jaeger]: https://www.jaegertracing.io/
[zipkin]: https://zipkin.io/
[signoz]: https://signoz.io/
[change log]: https://www.permify.co/change-log/integration-with-tracing-tools-jaeger-signoz-and-zipkin

### Configure With Using Flags

Alternatively, you can set configuration options with the respected flags when running the command. See all configuration flags with running,

```shell
docker run -p 8080:8080 ghcr.io/permify/permify serve -help
```

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
