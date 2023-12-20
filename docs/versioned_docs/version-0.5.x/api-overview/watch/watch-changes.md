import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

# Watch

The Permify Watch API acts as a real-time broadcaster that shows changes in the relation tuples.

The Watch API exclusively supports gRPC and works with PostgreSQL, given the track_commit_timestamp option is enabled. Please note, it doesn't support in-memory databases or HTTP communication.

# Requirements

- PostgreSQL database set up with track_commit_timestamp option enabled

## Enabling track_commit_timestamp on PostgreSQL

To ensure data consistency and synchronization between your application and Permify, enable track_commit_timestamp on
your PostgreSQL server. This can be done by executing the following options in your PostgreSQL:

### Option 1: SQL Command

1. Open your PostgreSQL command line interface.
2. Execute the following command:

    ```sql
    ALTER SYSTEM SET track_commit_timestamp = ON;
    ```

3. Reload the configuration with the following command:

    ```sql
    SELECT pg_reload_conf();
    ```

### Option 2: Editing postgresql.conf

1. Find and open the postgresql.conf file in a text editor. Its location depends on your PostgreSQL installation. Common
   locations are:
    - Debian-based systems: /etc/postgresql/[version]/main/postgresql.conf
    - Red Hat-based systems: /var/lib/pgsql/data/postgresql.conf

2. Add or modify the following line in the postgresql.conf file:
   ```
   track_commit_timestamp = on
   ```

3. Save and close the postgresql.conf file.
4. Reload the PostgreSQL configuration for the changes to take effect. This can be done via the PostgreSQL console:
    ```sql
    SELECT pg_reload_conf();
    ```    

   Or if you have command line access, use:

    ```bash
   sudo service postgresql reload
    ```

Please ensure you have the necessary permissions to execute these commands or modify the postgresql.conf file. Also, remember that changes in the postgresql.conf file will persist across restarts, while the SQL method may need to be reapplied depending on your PostgreSQL version and setup.

:::info
Important Configuration Requirement: To use the Watch API, it must be enabled in your configuration file. Add or modify the following lines:

```yaml
service:
  watch:
    enabled: true
```

:::

## Request

**Path:** POST /v1/watch/watch

[![View in Swagger](http://jessemillar.github.io/view-in-swagger-button/button.svg)](https://permify.github.io/permify-swagger/#/Watch/watch.watch)

| Required | Argument   | Type   | Default | Description                                                                                                                                                                                                                                                                                                                                   |
|----------|------------|--------|---------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| [x]      | tenant_id  | string | -       | identifier of the tenant, if you are not using multi-tenancy (have only one tenant) use pre-inserted tenant `t1` for this field.                                                                                                                                                                                                              |
| [ ]      | snap_token | string | -       | specifies the starting point for broadcasting changes. If a snap_token is provided, all changes following that specific snapshot will be broadcasted. If a snap_token is not provided, the Watch API will broadcast all changes that occur after the Watch API is initiated., see more details on [Snap Tokens](../../reference/snap-tokens). |


[//]: # (<Tabs>)

[//]: # (<TabItem value="go" label="Go">)

[//]: # ()
[//]: # (```go)

[//]: # ()
[//]: # (```)

[//]: # ()
[//]: # (</TabItem>)

[//]: # (<TabItem value="node" label="Node">)

[//]: # ()
[//]: # (```javascript)

[//]: # ()
[//]: # (```)

[//]: # ()
[//]: # (</TabItem>)

[//]: # (</Tabs>)

## Response

```json
{
   "changes": {
      "tuple_changes": [
         {
            "operation": "OPERATION_CREATE",
            "tuple": {
               "entity": {
                  "type": "organization",
                  "id": "1"
               },
               "relation": "admin",
               "subject": {
                  "type": "user",
                  "id": "56",
                  "relation": ""
               }
            }
         }
      ],
      "snap_token": "MgMAAAAAAAA="
   }
}
```


## Need any help ?

Our team is happy to help you get started with Permify. If you'd like to learn more about using Permify in your app or
have any questions about this
example, [schedule a call with one of our Permify engineer](https://meetings-eu1.hubspot.com/ege-aytin/call-with-an-expert).

