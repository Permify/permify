---
title: "Install with Brew"
---

# Brew With Configurations

This section shows how to intall and run Permify Service with using brew. 

### Install Permify

Open terminal and run following line,

```shell
brew install permify/tap/permify
```

### Run Permify Service 

To run the Permify Service, `permify serve` command should be run with configurations.

By default, the service is configured to listen on ports 3476 (HTTP) and 3478 (gRPC) and store the authorization data in memory rather then an actual database. You can override these with running the command with configuration flags. See all configuration options with running `permify serve --help` on terminal. 

Check out the [Centralize Authorization Data] section to learn how to organize this config YAML file and get more details about the configuration options.

[Centralize Authorization Data]:  ../getting-started/sync-data

### Configuration Flags

| Flag                                | Description                                                                         | Default              | 
|-------------------------------------|-------------------------------------------------------------------------------------|----------------------|
| --grpc-port                         | Port that server run on                                                             | 3478                 | 
| --grpc-tls-config-cert-path         | GRPC tls certificate path                                                           | -                    | 
| --grpc-tls-config-key-path          | GRPC tls key path                                                                   | -                    | 
| --http-cors-allowed-headers         | CORS allowed headers for http gateway                                               | [*]                  | 
| --http-cors-allowed-origins         | CORS allowed origins for http gateway                                               | [*]                  | 
| --http-enabled                      | Switch option for HTTP server                                                       | true                 | 
| --http-port                         | HTTP port address                                                                   | 3476                 | 
| --http-tls-config-cert-path         | HTTP tls certificate path                                                           | -                    | 
| --http-tls-config-key-path          | HTTP tls key path                                                                   | -                    |
| --log-level                         | Real time logs of authorization. Permify uses zerolog as a logger                   | debug                | 
| --profiler-enabled                  | Switch option for profiler                                                          | false                |
| --profiler-port                     | Profiler port address                                                               | -                    |
| --authn-enabled                     | Enable server authentication                                                        | false                |
| --authn-preshared-keys              | Preshared key/keys for server authentication.                                       | -                    |
| --tracer-enabled                    | Switch option for tracing                                                           | false                | 
| --tracer-endpoint                   | Export uri for tracing data                                                         | -                    | 
| --tracer-exporter                   | Can be; jaeger, signoz or zipkin. (integrated tracing tools)                        | -                    | 
| --meter-enabled                     | Switch option for metric                                                            | true                 |
| --meter-endpoint                    | Export uri for metric data                                                          | otlp                 |
| --meter-exporter                    | Can be; otlp. (integrated metric tools)                                             | telemetry.permify.co |
| --service-circuit-breaker           | Switch option for service circuit breaker                                           | false                | 
| --service-concurrency-limit         | Concurrency limit                                                                   | 100                  | 
| --database-engine                   | Data source. Permify supports PostgreSQL('postgres') for now.                       | memory               |
| --database-uri                      | Uri of your data source to store relation tuples and schema                         | -                    |
| --database-auto-migrate             | Auto migrate database tables                                                        | true                 |
| --database-max-open-connections     | Maximum number of parallel connections that can be made to the database at any time | 20                   |
| --database-max-idle-connections     | Maximum number of idle connections that can be made to the database at any time     | 1                    |
| --database-max-connection-lifetime  | Maximum amount of time a connection may be reused                                   | 300s                 |
| --database-max-connection-idle-time | Maximum amount of time a connection may be idle                                     | 60s      

### Test your connection.

You can test your connection with creating an HTTP GET request,

```shell
localhost:3476/healthz
```

You can use our Postman Collection to work with the API. Also see the [Using the API] section for details of core functions.

[Using the API]: ../api-overview/

[![Run in Postman](https://run.pstmn.io/button.svg)](https://www.postman.com/permify-dev/workspace/permify/collection)
[![View in Swagger](http://jessemillar.github.io/view-in-swagger-button/button.svg)](https://permify.github.io/permify-swagger/)

### Need any help ?

Our team is happy to help you get started with Permify, [schedule a call with an Permify engineer](https://meetings-eu1.hubspot.com/ege-aytin/call-with-an-expert).
