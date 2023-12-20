# Cache Mechanisms 

This section showcases the cache mechanisms that Permify uses.

## Schema Cache

Schemas are stored in an in-memory cache based on their versions. If a version is specified in the request metadata, it will be searched for in the in-memory cache. If not found, it will query the database for the version and store it in the cache. If no version information is given in the metadata, versions will be assumed to be alphanumeric and sorted in that order, and Permify will request the head version and check if it exists in the memory cache.

The size of this can be determined through the Permify configuration. Here is an example configuration:
service:

```yaml 
…
  schema:
    cache:
      number_of_counters: 1_000
      max_cost: 10MiB
…
``` 

The cache library used is: https://github.com/dgraph-io/ristretto

## Data Cache

Permify applies the MVCC (Multi Version Concurrency Control) pattern for Postgres, creating a separate database snapshot for each write and delete operation. This both enhances performance and provides a consistent cache.

An example of a cache key is:
check_{tenant_id}_{schema_version}:{snapshot_token}:{check_request}

Permify hashes each request and searches for the same key. If it cannot find it, it runs the check engine and writes to the cache, thus creating a consistently working hash.

The size of this can also be determined via the Permify configuration. Here’s an example:
service:

```yaml 
  …
  permission:
    bulk_limit: 100
    concurrency_limit: 100
    cache:
      number_of_counters: 10_000
      max_cost: 10MiB
  …
``` 

The cache library used is: https://github.com/dgraph-io/ristretto

Note: Another advantage of the MVCC pattern is the ability to historically store data. However, it has a downside of accumulation of too many relationships. For this, we have developed a garbage collector that will delete old data at a time period you specify.

## Distributed Cache

Permify does provide a distributed cache across availability zones (within an AWS region) via **Consistent Hashing**. Permify uses Consistent Hashing across its distributed instances for more efficient use of their individual caches. 

This would allow for high availability and resilience in the face of individual nodes or even entire availability zone failure, as well as improved performance due to data locality benefits.

Consistent Hashing is a distributed hashing scheme that operates independently of the number of objects in a distributed hash table. This method hashes according to the nodes’ peers, estimating which node a key would be on and thereby ensuring the most suitable request goes to the most suitable node, effectively creating a natural load balancer.

### How Consistent Hashing Operates in Permify

With a single instance, when an API request is made, request and corresponding response stored in its corresponding local cache.

If we have more than one Permify instance consistent hashing activates on API calls, hashes the request, and outputs a unique key representing the node/instance that will store the request's data. Suppose it stored in the instance 2, subsequent API calls with the same hash will retrieve the response from the instance 2, regardless of which instance that API called from.

Using this consistent hashing approach, we can effectively utilize individual cache capacities. Adding more instances automatically increases the total cache capacity in Permify.

You can learn more about consistent hashing from the following blog post: [Introducing Consistent Hashing](https://itnext.io/introducing-consistent-hashing-9a289769052e)

:::info
Note, however, that while the consistent hashing approach will distribute keys evenly across the cache nodes, it's up to the application logic to ensure the cache is used effectively (i.e., that it reads from and writes to the cache appropriately).
:::

Here is an example configuration:

```yaml 
distributed:
  # Indicates whether the distributed mode is enabled or not
  enabled: true

  # The address of the distributed service.
  # Using a Kubernetes DNS name suggests this service runs in a Kubernetes cluster
  # under the 'default' namespace and is named 'permify'
  address: "kubernetes:///permify.default:5000"

  # The port on which the service is exposed
  port: "5000"
``` 

Additional to that we’re using a [circuit breaker](https://blog.bitsrc.io/circuit-breaker-pattern-in-microservices-26bf6e5b21ff) pattern to detect and handle failures when the underlying database is unavailable. It prevents unnecessary calls when the database is down and handles the process on the rebooting phase.

## Need any help ?

Our team is happy help you to structure right architecture for your permission system. Feel free to [schedule a call with one of our Permify engineer](https://meetings-eu1.hubspot.com/ege-aytin/call-with-an-expert).



