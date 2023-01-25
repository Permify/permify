import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

# Create Tenant

Permify provide it's own authorization language to model common patterns of easily. We called the authorization model Permify Schema and it can be created on our [playground](https://play.permify.co/) as well as in any IDE or text editor. 

We also have a [VS Code extension](https://marketplace.visualstudio.com/items?itemName=Permify.perm) to ease modeling Permify Schema with code snippets and syntax highlights. Note that on VS code the file with extension is ***".perm"***.

:::caution Use Playground For Testing
If you're planning to test Permify manually, maybe with an API Design platform such as [Postman](https://www.postman.com/), [Insomnia](https://insomnia.rest/), etc; we're suggesting using our playground to create model. Because Permify Schema needs to be configured (send to API) in Permify API in a **string** format. Therefore, created model should be converted to **string**. 

Although, it could easily be done programmatically, it could be little challenging to do it manually. To help on that, we have a button on the playground to copy created model to the clipboard as a string, so you get your model in string format easily.

![copy-btn](https://user-images.githubusercontent.com/34595361/198015792-a7f0d727-a1a5-4039-b0be-d097321b8d53.png)
:::

Permify Schema needed to be send to API endpoint **/v1/schemas/write"** for configuration of your authorization model on Permify API.

## Request

**POST** "/v1/schemas/write"**

| Required | Argument | Type | Default | Description |
|----------|-------------------|--------|---------|-------------|
| [x]   | tenant_id | string | - | identifier of the tenant, if you are not using multi-tenancy (have only one tenant) use pre-inserted tenant `t1` for this field.
| [x]   | schema | string | - | Permify Schema as string|

<Tabs>
<TabItem value="go" label="Go">

```go
sr, err: = client.Schema.Write(context.Background(), &v1.SchemaWriteRequest {
    Schema: `
    "entity user {}\n\n    entity organization {\n\n        relation admin @user\n        relation member @user\n\n        action create_repository = (admin or member)\n        action delete = admin\n    }\n\n    entity repository {\n\n        relation owner @user\n        relation parent @organization\n\n        action push = owner\n        action read = (owner and (parent.admin and parent.member))\n        action delete = (parent.member and (parent.admin or owner))\n    }"
    `,
})
```

</TabItem>
<TabItem value="node" label="Node">

```javascript
client.schema.write({
    schema: `
    "entity user {}\n\n    entity organization {\n\n        relation admin @user\n        relation member @user\n\n        action create_repository = (admin or member)\n        action delete = admin\n    }\n\n    entity repository {\n\n        relation owner @user\n        relation parent @organization\n\n        action push = owner\n        action read = (owner and (parent.admin and parent.member))\n        action delete = (parent.member and (parent.admin or owner))\n    }"
    `
}).then((response) => {
    // handle response
})
```

</TabItem>
<TabItem value="curl" label="cURL">

```curl
curl --location --request POST 'localhost:3476/v1/tenants/{tenant_id}/schemas/write' \
--header 'Content-Type: application/json' \
--data-raw '{
    "schema": "entity user {}\n\n    entity organization {\n\n        relation admin @user\n        relation member @user\n\n        action create_repository = (admin or member)\n        action delete = admin\n    }\n\n    entity repository {\n\n        relation owner @user\n        relation parent @organization\n\n        action push = owner\n        action read = (owner and (parent.admin and parent.member))\n        action delete = (parent.member and (parent.admin or owner))\n }"
}'
```
</TabItem>
</Tabs>

## Example Request on Postman
**POST** "/v1/schemas/write"**

**Example Request on Postman:**

![permify-schema](https://user-images.githubusercontent.com/34595361/197405641-d8197728-2080-4bc3-95cb-123e274c58ce.png)

## Need any help ?

Our team is happy to help you get started with Permify. If you'd like to learn more about using Permify in your app or have any questions about this example, [schedule a call with one of our Permify engineer](https://meetings-eu1.hubspot.com/ege-aytin/call-with-an-expert).