import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

# List Schema

Models written to Permify using the [write schema API](./write-schema.md) can be listed using this API with the timestamps at which the models were created. 

Request needs to be made to the API endpoint **/v1/schemas/list** to list all the models.

## Request

```javascript
POST /v1/tenants/{tenant_id}/schemas/list
```

[![View in Swagger](http://jessemillar.github.io/view-in-swagger-button/button.svg)](https://permify.github.io/permify-swagger/#/Schema/schemas.list)

| Required | Argument | Type | Default | Description |
|----------|-------------------|--------|---------|-------------|
| [x]   | tenant_id | string | - | identifier of the tenant, if you are not using multi-tenancy (have only one tenant) use pre-inserted tenant `t1` for this field.
| [ ]   | page_size | string | - | Number of schema versions to be fetched |
| [ ]   | continuous_token | string | - | Continuation token for subsequent pages to be fetched |

<Tabs>
<TabItem value="go" label="Go">

```go
sr, err: = client.Schema.List(context.Background(), &v1.SchemaListRequest {
    TenantId: "t1",
    PageSize: "10",
    ContinuousToken: "",
})
```

</TabItem>
<TabItem value="curl" label="cURL">

```curl
curl --location --request POST 'localhost:3476/v1/tenants/{tenant_id}/schemas/read' \
--header 'Content-Type: application/json' \
--data-raw '{
    "page_size": "10",
    "continuous_token": ""
}'
```
</TabItem>
</Tabs>

## Example Request on Postman
**POST** "/v1/tenants/{tenant_id}/schemas/list"

**Example Request on Postman:**
![permify-schema](https://github.com/Permify/permify/assets/30985448/aa73c993-e808-496e-bebc-f91ced3a3399)

