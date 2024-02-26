import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

# Read Schema

When a model is written to Permify using the [write schema API](./write-schema.md) a schema version will be returned by the API. That schema version can be used to inspect the schema.

Permify Schema needed to be send to API endpoint **/v1/schemas/read** for configuration of your authorization model on Permify API.

## Request

```javascript
POST /v1/tenants/{tenant_id}/schemas/read
```

[![View in Swagger](http://jessemillar.github.io/view-in-swagger-button/button.svg)](https://permify.github.io/permify-swagger/#/Schema/schemas.read)

| Required | Argument | Type | Default | Description |
|----------|-------------------|--------|---------|-------------|
| [x]   | tenant_id | string | - | identifier of the tenant, if you are not using multi-tenancy (have only one tenant) use pre-inserted tenant `t1` for this field.
| [ ]   | schema_version | string | - | Permify Schema version to read|

<Tabs>
<TabItem value="go" label="Go">

```go
sr, err: = client.Schema.Read(context.Background(), &v1.SchemaReadRequest {
    TenantId: "t1",
	Metadata: &v1.SchemaReadRequestMetadata{
		SchemaVersion: "cnbe6se5fmal18gpc66g",
	},
})
```

</TabItem>
<TabItem value="curl" label="cURL">

```curl
curl --location --request POST 'localhost:3476/v1/tenants/{tenant_id}/schemas/read' \
--header 'Content-Type: application/json' \
--data-raw '{
    "metadata": {
        "schema_version": "cnbe6se5fmal18gpc66g"
    }
}'
```
</TabItem>
</Tabs>

## Example Request on Postman
**POST** "/v1/tenants/{tenant_id}/schemas/read"**

**Example Request on Postman:**

![permify-schema](https://github.com/Permify/permify/assets/30985448/a6944e3d-6a58-4489-b16f-da2fdf5f60f2)
