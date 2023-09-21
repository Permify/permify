
# Tracing Tools

Permify has integrations with some of popular tracing tools to analyze performance and behavior of your authorization. These are:

- [Jaeger](https://www.jaegertracing.io/)
- [Signoz](https://signoz.io/)
- [Zipkin](https://zipkin.io/)

## Usage

### Set Up

Adding one of these tracing tools to your authorization system is quite simple, you just need to define it in the Permify configuration file as **tracer**. 

```yaml
tracer:
  exporter: 'zipkin'
  endpoint: 'http://172.17.0.4:9411/api/v2/spans'
  disabled: false
```
- ***exporter***: enter the tool name that you want to use. `jaeger` , `signoz` and `zipkin`.
- ***endpoint***: export url for tracing data.
- ***disabled***: switch option for tracing.

**Example YAML configuration file**

```yaml
app:
  name: ‘permify’
http:
  port: 3476
logger:
  log_level: ‘debug’
  rollbar_env: ‘permify’
tracer:
  exporter: 'zipkin'
  endpoint: 'http://172.17.0.4:9411/api/v2/spans'
  disabled: false
database:
  write:
    connection: 'postgres'
    database: 'morf-health-demo'
    uri: 'postgres://postgres:SphU4Uf3QXNntT@permify.us-east-1.rds.amazonaws.com:5432'
    pool_max: 2
```

After running Permify in your server, you should run Zipkin as well. If you're using docker here is the docker pull request for Zipkin:

```
docker run -d -p 9411:9411 openzipkin/zipkin
```
