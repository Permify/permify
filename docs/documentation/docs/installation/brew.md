---
title: "Brew (With Conf)"
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

[Centralize Authorization Data]:  /docs/getting-started/sync-data

### Configuration Flags

| Flag | Description | Default | 
|--------------------------|----------| ----------|
|  --authn-enabled     | Enable server authentication | false | 
|  --authn-preshared-keys   | Preshared key/keys for server authentication. | - | 
|  --database-engine     | Data source. Permify supports PostgreSQL('postgres') for now. |  memory | 
|  --database-name    | Custom database name |  - |
|  --max_open_connections   | Max connection pool size | 20 | 
|  --database-uri   | Uri of your data source to store relation tuples | - | 
|  --grpc-port  | Port that server run on | 3478 | 
|  --grpc-tls-config-cert-path   | GRPC tls certificate path | - | 
|  --grpc-tls-config-key-path | GRPC tls key path | - | 
|  -h or --help  | Help for serve | no value | 
|  --http-cors-allowed-headers  | CORS allowed headers for http gateway | [*] | 
|  --http-cors-allowed-origins  | CORS allowed origins for http gateway | [*] | 
|  --http-enabled  | Switch option for HTTP server | true | 
|  --http-port  |  HTTP port address | 3476 | 
|  --http-tls-config-cert-path   | HTTP tls certificate path | - | 
|  --http-tls-config-key-path | HTTP tls key path | - | 
|  --log-level | Real time logs of authorization. Permify uses zerolog as a logger | debug| 
|  --tracer-enabled | Switch option for tracing | false | 
|  --tracer-endpoint | Export uri for tracing data | - | 
|  --tracer-exporter | Can be; jaeger, signoz or zipkin. (integrated tracing tools) | - | 

### Test your connection.

You can test your connection with creating an HTTP GET request,

```shell
localhost:3476/healthz
```

You can use our Postman Collection to work with the API. Also see the [Using the API] section for details of core functions.

[Using the API]: ../api-overview/

[![Run in Postman](https://run.pstmn.io/button.svg)](https://god.gw.postman.com/run-collection/16122080-54b1e316-8105-4440-b5bf-f27a05a8b4de?action=collection%2Ffork&collection-url=entityId%3D16122080-54b1e316-8105-4440-b5bf-f27a05a8b4de%26entityType%3Dcollection%26workspaceId%3Dd3a8746c-fa57-49c0-83a5-6fcf25a7fc05)
[![View in Swagger](http://jessemillar.github.io/view-in-swagger-button/button.svg)](https://app.swaggerhub.com/apis-docs/permify/permify/latest)

### Need any help ?

Our team is happy to help you get started with Permify, [schedule a call with an Permify engineer](https://meetings-eu1.hubspot.com/ege-aytin/call-with-an-expert).
