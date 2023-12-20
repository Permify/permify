# Configuration File

Permify offers various options for configuring your Permify API Server.

Here is the example configuration YAML file with glossary below. You can also find
this [example config file](https://github.com/Permify/permify/blob/master/example.config.yaml) in Permify repo.

***Example config.yaml file***

```yaml
# The server section specifies the HTTP and gRPC server settings,
# including whether or not TLS is enabled and the certificate and
# key file locations.
server:
  rate_limit: 100
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

# The logger section sets the logging level for the service.
logger:
  level: info

# The profiler section enables or disables the pprof profiler and
# sets the port number for the profiler endpoint.
profiler:
  enabled: true
  port: 6060

# The authn section specifies the authentication method for the service.
authn:
  enabled: true
  method: preshared
  preshared:
    keys: []

# The tracer section enables or disables distributed tracing and sets the
# exporter and endpoint for the tracing data.
tracer:
  exporter: zipkin
  endpoint: http://localhost:9411/api/v2/spans
  enabled: true

# The meter section enables or disables metrics collection and sets the
# exporter and endpoint for the collected metrics.
meter:
  exporter: otlp
  endpoint: localhost:4318
  enabled: true

# The service section sets various service-level settings, including whether
# or not to use a circuit breaker, and cache sizes for schema, permission,
# and relationship data.
service:
  circuit_breaker: false
  watch:
    enabled: false
  schema:
    cache:
      number_of_counters: 1_000
      max_cost: 10MiB
  permission:
    bulk_limit: 100
    concurrency_limit: 100
    cache:
      number_of_counters: 10_000
      max_cost: 10MiB
  relationship:

# The database section specifies the database engine and connection settings,
# including the URI for the database, whether or not to auto-migrate the database,
# and connection pool settings.
database:
  engine: postgres
  uri: postgres://user:password@host:5432/db_name
  auto_migrate: false
  max_open_connections: 20
  max_idle_connections: 1
  max_connection_lifetime: 300s
  max_connection_idle_time: 60s
  garbage_collection:
    enabled: true
    interval: 200h
    window: 200h
    timeout: 5m

# distributed configuration settings
distributed:
  # Indicates whether the distributed mode is enabled or not
  enabled: true

  # The address of the distributed service.
  # Using a Kubernetes DNS name suggests this service runs in a Kubernetes cluster
  # under the 'default' namespace and is named 'permify'
  address: "kubernetes:///permify.default"

  # The port on which the service is exposed
  port: "5000"

```

## Options

<details><summary>server | Server Configurations</summary>
<p>

#### Definition

Server options to run Permify. (`grpc` and `http` available for now.)

#### Structure

```
├── server
    ├── rate_limit
    ├── (`grpc` or `http`)
    │   ├── enabled
    │   ├── port
    │   └── tls
    │       ├── enabled
    │       ├── cert
    │       └── key
```

#### Glossary

| Required | Argument                  | Default | Description                                                         |
|----------|---------------------------|---------|---------------------------------------------------------------------|
| [ ]      | rate_limit                | 100     | the maximum number of requests the server should handle per second. |
| [x]      | [ server_type ]           | -       | server option type can either be `grpc` or `http`.                  |
| [ ]      | enabled (for server type) | true    | switch option for server.                                           |
| [x]      | port                      | -       | port that server run on.                                            |
| [x]      | tls                       | -       | transport layer security options.                                   |
| [ ]      | enabled (for tls)         | false   | switch option for tls                                               |
| [ ]      | cert                      | -       | tls certificate path.                                               |
| [ ]      | key                       | -       | tls key pat                                                         |

#### ENV

| Argument                  | ENV                               | Type         |
|---------------------------|-----------------------------------|--------------|
| rate_limit                | PERMIFY_RATE_LIMIT                | int          |
| grpc-port                 | PERMIFY_GRPC_PORT                 | string       |
| grpc-tls-enabled          | PERMIFY_GRPC_TLS_ENABLED          | boolean      |
| grpc-tls-key-path         | PERMIFY_GRPC_TLS_KEY_PATH         | string       |
| grpc-tls-cert-path        | PERMIFY_GRPC_TLS_CERT_PATH        | string       |
| http-enabled              | PERMIFY_HTTP_ENABLED              | boolean      |
| http-port                 | PERMIFY_HTTP_PORT                 | string       |
| http-tls-key-path         | PERMIFY_HTTP_TLS_KEY_PATH         | string       |
| http-tls-cert-path        | PERMIFY_HTTP_TLS_CERT_PATH        | string       |
| http-cors-allowed-origins | PERMIFY_HTTP_CORS_ALLOWED_ORIGINS | string array |
| http-cors-allowed-headers | PERMIFY_HTTP_CORS_ALLOWED_HEADERS | string array |

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

| Required | Argument | Default | Description                                      |
|----------|----------|---------|--------------------------------------------------|
| [x]      | level    | info    | logger levels: `error`, `warn`, `info` , `debug` |

#### ENV

| Argument                  | ENV                             | Type   |
|---------------------------|---------------------------------|--------|
| log-level                 | PERMIFY_LOG_LEVEL               | string |

</p>
</details>

<details><summary>authn | Server Authentication</summary>
<p>

#### Definition

You can choose to authenticate users to interact with Permify API.

There are 2 authentication method you can choose:

* [Pre Shared Keys](#pre-shared-keys)
* [OpenID Connect](#openid-connect)

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

| Required | Argument | Default | Description                                                                                                          |
|----------|----------|---------|----------------------------------------------------------------------------------------------------------------------|
| [x]      | method   | -       | Authentication method can be either `oidc` or `preshared`.                                                           |
| [ ]      | enabled  | true    | switch option authentication config                                                                                  |
| [x]      | keys     | -       | Private key/keys for server authentication. Permify does not provide this key, so it must be generated by the users. |

#### ENV

| Argument              | ENV                           | Type         |
|-----------------------|-------------------------------|--------------|
| authn-enabled         | PERMIFY_AUTHN_ENABLED         | boolean      |
| authn-method          | PERMIFY_AUTHN_METHOD          | string       |
| authn-preshared-keys  | PERMIFY_AUTHN_PRESHARED_KEYS  | string array |


#### OpenID Connect

Permify supports OpenID Connect (OIDC). OIDC provides an identity layer on top of OAuth 2.0 to address the shortcomings
of using OAuth 2.0 for establishing identity.

With this authentication method, you be able to integrate your existing Identity Provider (IDP) to validate JSON Web
Tokens (JWTs) using JSON Web Keys (JWKs). By doing so, only trusted tokens from the IDP will be accepted for
authentication.

#### Structure

```
├── authn
|   ├── method
|   ├── enabled
|   ├── client-id
|   ├── issuer
```

#### Glossary

| Required | Argument  | Default | Description                                                                                                                                                                                                                       |
|----------|-----------|---------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| [x]      | method    | -       | Authentication method can be either `oidc` or `preshared`.                                                                                                                                                                        |
| [ ]      | enabled   | false   | switch option authentication config                                                                                                                                                                                               |
| [x]      | client_id | -       | This is the client ID of the application you're developing. It is a unique identifier that is assigned to your application by the OpenID Connect provider, and it should be included in the JWTs that are issued by the provider. |
| [x]      | issuer    | -       | This is the URL of the provider that is responsible for authenticating users. You will use this URL to discover information about the provider in step 1 of the authentication process.                                           |

#### ENV

| Argument              | ENV                           | Type         |
|-----------------------|-------------------------------|--------------|
| authn-enabled         | PERMIFY_AUTHN_ENABLED         | boolean      |
| authn-method          | PERMIFY_AUTHN_METHOD          | string       |
| authn-oidc-issuer     | PERMIFY_AUTHN_OIDC_ISSUER     | string       |
| authn-oidc-client-id  | PERMIFY_AUTHN_OIDC_CLIENT_ID  | string       |

</p>
</details>


<details><summary>tracer | Tracing Configurations</summary>
<p>

#### Definition

Permify integrated with [jaeger], [otlp], [signoz], and [zipkin] tacing tools to analyze performance and behavior of your
authorization when using Permify.

#### Structure

```
├── tracer
|   ├── exporter
|   ├── endpoint
|   ├── enabled
```

#### Glossary

| Required | Argument | Default | Description                                                                |
|----------|----------|---------|----------------------------------------------------------------------------|
| [x]      | exporter | -       | Tracer exporter, the options are `jaeger`, `otlp`, `signoz`, and `zipkin`. |
| [x]      | endpoint | -       | export uri for tracing data.                                               |
| [ ]      | enabled  | false   | switch option for tracing.                                                 |
| [ ]      | insecure | false   | Whether to use HTTP instead of HTTPs for exporting the traces.             |

#### ENV

| Argument             | ENV                           | Type         |
|----------------------|-------------------------------|--------------|
| tracer-enabled       | PERMIFY_TRACER_ENABLED        | boolean      |
| tracer-exporter      | PERMIFY_TRACER_EXPORTER       | string       |
| tracer-endpoint      | PERMIFY_TRACER_ENDPOINT       | string       |
| tracer-insecure      | PERMIFY_TRACER_INSECURE       | boolean      |

</p>
</details>

<details><summary>meter | Meter Configurations</summary>
<p>

#### Definition

Configuration for observing metrics; check count, cache check count and session information; Permify version, hostname,
os, arch.

#### Structure

```
├── meter
|   ├── exporter
|   ├── endpoint
|   ├── enabled
```

#### Glossary

| Required | Argument | Default | Description                                                  |
|----------|----------|---------|--------------------------------------------------------------|
| [x]      | exporter | -       | [otpl](https://opentelemetry.io/docs/collector/) is default. |
| [x]      | endpoint | -       | export uri for metric observation                            |
| [ ]      | enabled  | true    | switch option for meter tracing.                             |

#### ENV

| Argument           | ENV                     | Type         |
|--------------------|-------------------------|--------------|
| meter-enabled      | PERMIFY_METER_ENABLED   | boolean      |
| meter-exporter     | PERMIFY_METER_EXPORTER  | string       |
| meter-endpoint     | PERMIFY_METER_ENDPOINT  | string       |

</p>
</details>

<details><summary>database | Database Configurations</summary>
<p>

#### Definition

Configurations for the database that points out where your want to store your authorization data (relation tuples,
audits, decision logs, authorization model)

#### Structure

```
├── database
|   ├── engine
|   ├── uri
|   ├── auto_migrate
|   ├── max_open_connections
|   ├── max_idle_connections
|   ├── max_connection_lifetime
|   ├── max_connection_idle_time
|   ├──garbage_collection
|       ├──enable: true
|       ├──interval: 3m
|       ├──timeout: 3m
|       ├──window: 720h
```

#### Glossary

| Required | Argument                        | Default | Description                                                                                                       |
|----------|---------------------------------|---------|-------------------------------------------------------------------------------------------------------------------|
| [x]      | engine                          | memory  | Data source. Permify supports **PostgreSQL**(`'postgres'`) for now. Contact with us for your preferred database.  |
| [x]      | uri                             | -       | Uri of your data source.                                                                                          |
| [ ]      | auto_migrate                    | true    | When its configured as false migrating flow won't work.                                                           |                                           
| [ ]      | max_open_connections            | 20      | Configuration parameter determines the maximum number of concurrent connections to the database that are allowed. |
| [ ]      | max_idle_connections            | 1       | Determines the maximum number of idle connections that can be held in the connection pool.                        |
| [ ]      | max_connection_lifetime         | 300s    | Determines the maximum lifetime of a connection in seconds.                                                       |                 
| [ ]      | max_connection_idle_time        | 60s     | Determines the maximum time in seconds that a connection can remain idle before it is closed.                     |                
| [ ]      | enable (for garbage collection) | false   | Switch option for garbage collection.                                                                             |               
| [ ]      | interval                        | 3m      | Determines the run period of a Garbage Collection operation.                                                      |              
| [ ]      | timeout                         | 3m      | Sets the duration of the Garbage Collection timeout.                                                              |             
| [ ]      | window                          | 720h    | Determines how much backward cleaning the Garbage Collection process will perform.                                |                     

#### ENV

| Argument                                      | ENV                                                    | Type     |
|-----------------------------------------------|--------------------------------------------------------|----------|
| database-engine                               | PERMIFY_DATABASE_ENGINE                                | string   |
| database-uri                                  | PERMIFY_DATABASE_URI                                   | string   |
| database-auto-migrate                         | PERMIFY_DATABASE_AUTO_MIGRATE                          | boolean  |
| database-max-open-connections                 | PERMIFY_DATABASE_MAX_OPEN_CONNECTIONS                  | int      |
| database-max-idle-connections                 | PERMIFY_DATABASE_MAX_IDLE_CONNECTIONS                  | int      |
| database-max-connection-lifetime              | PERMIFY_DATABASE_MAX_CONNECTION_LIFETIME               | duration |
| database-max-connection-idle-time             | PERMIFY_DATABASE_MAX_CONNECTION_IDLE_TIME              | duration |
| database-garbage-collection-enabled           | PERMIFY_DATABASE_GARBAGE_ENABLED                       | boolean  |
| database-garbage-collection-interval          | PERMIFY_DATABASE_GARBAGE_COLLECTION_INTERVAL           | duration |
| database-garbage-collection-timeout           | PERMIFY_DATABASE_GARBAGE_COLLECTION_TIMEOUT            | duration |
| database-garbage-collection-window            | PERMIFY_DATABASE_GARBAGE_COLLECTION_WINDOW             | duration |

</p>
</details>

<details><summary>profiler | Performance Profiler Configurations</summary>
<p>

#### Definition

pprof is a performance profiler for Go programs. It allows developers to analyze and understand the performance
characteristics of their code by generating detailed profiles of program execution

#### Structure

```
├── profiler
|   ├── enabled
|   ├── port
```

#### Glossary

| Required | Argument | Default | Description                                   |
|----------|----------|---------|-----------------------------------------------|
| [ ]      | enabled  | true    | switch option for profiler.                   |
| [x]      | port     | -       | port that profiler runs on *(default: 6060)*. |

#### ENV

| Argument         | ENV                        | Type         |
|------------------|----------------------------|--------------|
| profiler-enabled | PERMIFY_PROFILER_ENABLED   | boolean      |
| profiler-port    | PERMIFY_PROFILER_PORT      | string       |

</p>
</details>

<details><summary>Distributed | Consistent hashing Configurations</summary>
<p>

#### Definition

A consistent hashing ring ensures data distribution that minimizes reorganization when nodes are added or removed, improving scalability and performance in distributed systems."

#### Structure

```
├── distributed
|   ├── enabled
|   ├── address
|   ├── port
```

#### Glossary

| Required | Argument    | Default | Description                          |
|----------|-------------|---------|--------------------------------------|
| [x]      | enabled     | false   | switch option for distributed.       |
| []       | address     | -       | address of the distributed service   |
| []       | port        | 5000    | port on which the service is exposed |


#### ENV

| Argument             | ENV                         | Type    |
|----------------------|-----------------------------|---------|
| distributed-enabled  | PERMIFY_DISTRIBUTED_ENABLED | boolean |
| distributed-address  | PERMIFY_DISTRIBUTED_ADDRESS | string  |
| distributed-port     | PERMIFY_DISTRIBUTED_PORT    | string  |

</p>
</details>

[jaeger]: https://www.jaegertracing.io/

[otlp]: https://opentelemetry.io/

[zipkin]: https://zipkin.io/

[signoz]: https://signoz.io/
