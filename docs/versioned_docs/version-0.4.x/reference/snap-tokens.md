# Snap Tokens & Zookies

A Snap Token is a token that consists of an encoded timestamp, which is used to ensure fresh results in access control checks.

## Why you should use Snap Tokens ?

Basically, you should use snap tokens both for consistency and performance. The main goal of Permify is to provide an authorization system that ensures excellent performance that can handle millions of requests from different environments while ensuring data consistency.

Performance standards can be achievable with caching. In Permify, the cache mechanism eliminates re-computing of access control checks that once occurred, unless any relationships of resources don't change.

Still, all caches suffer from the risk of becoming stale. If some schema update happens, or relations change then all of the caches should be updated according to it to prevent false positive or false negative results.

Permify avoids this problem with an approach of snapshot reads. Simply, it ensures that access control is evaluated at a consistent point in time to prevent inconsistency.

To achieve this, we developed tokens called Snap Tokens that consist of a timestamp that is compared in access checks to ensure that the snapshot of the access control is at least as fresh as the resource timestamp - basically its stored snap token.

## How to use Snap Tokens

Snap Tokens used in endpoints to represent the snapshot and get fresh results of the API's. It mainly used in [Write API] and [Check API].

The general workflow for using snap token is getting the snap token from the response of Write API request - basically when writing a relational tuple - then mapped it with the resource. One way of doing that is storing snap token in the additional column in your relational database.

Then this snap token can be used in endpoints. For example it can be used in access control check with sending via `snap_token` field to ensure getting check result as fresh as previous request.

```json
{
  "schema_version": "ce8siqtmmud16etrelag",
  "snap_token": "gp/twGSvLBc=",
  "entity": {
    "type": "repository",
    "id": "1"
  },
  "permission": "edit",
  "subject": {
    "type": "user",
    "id": "1"
  }
}
```

[Write API]: ../../api-overview/relationship/write-relationships
[Check API]: ../../api-overview/permission/check-api

#### All endpoints that used snap token

- [Write API](../../api-overview/relationship/write-relationships)
- [Check API](../../api-overview/permission/check-api)
- [Expand API](../../api-overview/permission/expand-api)

## More on Cache Mechanism

Permify implements several cache mechanisms in order to achieve low latency in scaled distributed systems. See more in the section [Cache Mechanisms](./cache.md)
