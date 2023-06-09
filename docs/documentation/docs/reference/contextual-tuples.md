import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

# Contextual Tuples

Contextual tuples are relations that can be dynamically added to permission request operations. When you send these relations along with your requests, they get processed alongside existing relations in the database and will return a result. Examples of these might be dynamic IP addresses, different locations, or time-based requests, where these can be beneficial.

:::info
Note: When you use contextual tuples, the cache system will not be operational. This is because the cache system is written along with snapshots and if contextual tuples are written, using the cache would lead to incorrect results. Hence, to prevent this, the use of the cache is blocked at the time of the request.
:::

#### IP Range

```perm
entity organization {

    relation ip_address_range @ip_address_range
    relation admin  @user
    
    permission edit = ip_address_range.user or admin
}

entity ip_address_range {
    relation user @user
}
```


<Tabs>
<TabItem value="go" label="Go">

```go
cr, err: = client.Permission.Check(context.Background(), &v1.PermissionCheckRequest {
    TenantId: "t1",
    Metadata: &v1.PermissionCheckRequestMetadata {
        SnapToken: ""
        SchemaVersion: ""
        Depth: 20,
    },
    Entity: &v1.Entity {
        Type: "organization",
        Id: "1",
    },
    Permission: "edit",
    Subject: &v1.Subject {
        Type: "user",
        Id: "1",
    },
    ContextualTuples: []*v1.Tuple{
		{
		    Entity: &v1.Entity {
			    Type: "organization",
                Id: "1",
            },
		    relation: "ip_address_range",
		    Subject: &v1.Subject {
			    Type: "ip_address_range",
                Id: "192.158.1.38",
            },
        },
        {
            Entity: &v1.Entity {
                Type: "ip_address_range",
                Id: "192.158.1.38",
            },
            relation: "user",
            Subject: &v1.Subject {
                Type: "user",
                Id: "1",
            },
        },
    }

    if (cr.can === PermissionCheckResponse_Result.RESULT_ALLOWED) {
        // RESULT_ALLOWED
    } else {
        // RESULT_DENIED
    }
})
```

</TabItem>
<TabItem value="node" label="Node">

```javascript
client.permission.check({
    tenantId: "t1", 
    metadata: {
        snapToken: "",
        schemaVersion: "",
        depth: 20
    },
    entity: {
        type: "repository",
        id: "1"
    },
    permission: "edit",
    subject: {
        type: "user",
        id: "1"
    },
    contextualTuples: [
        {
            entity: {
                type: "organization",
                id: "1"
            },
            relation: "ip_address_range",
            subject: {
                type: "ip_address_range",
                id: "192.158.1.38",
            }
        },
        {
            entity: {
                type: "ip_address_range",
                id: "192.158.1.38"
            },
            relation: "user",
            subject: {
                type: "user",
                id: "1",
            }
        }
    ]
}).then((response) => {
    if (response.can === PermissionCheckResponse_Result.RESULT_ALLOWED) {
        console.log("RESULT_ALLOWED")
    } else {
        console.log("RESULT_DENIED")
    }
})
```

</TabItem>
<TabItem value="curl" label="cURL">

```curl
curl --location --request POST 'localhost:3476/v1/tenants/{tenant_id}/permissions/check' \
--header 'Content-Type: application/json' \
--data-raw '{
  "metadata":{
    "snap_token": "",
    "schema_version": "",
    "depth": 20
  },
  "entity": {
    "type": "repository",
    "id": "1"
  },
  "permission": "edit",
  "subject": {
    "type": "user",
    "id": "1",
    "relation": ""
  },
  "contextual_tuples": [
    {
      "entity": {
        "type": "organization",
        "id": "1"
      },
      "relation": "ip_address_range",
      "subject": {
        "type": "ip_address_range",
        "id": "192.158.1.38"
      }
    },
    {
      "entity": {
        "type": "ip_address_range",
        "id": "192.158.1.38"
      },
      "relation": "user",
      "subject": {
        "type": "user",
        "id": "1"
      }
    }
  ]
}'
```

</TabItem>
</Tabs>




## Need any help ?

Our team is happy to help you get started with Permify. If you'd like to learn more about using Permify in your app or have any questions about this example, [schedule a call with one of our Permify engineer](https://meetings-eu1.hubspot.com/ege-aytin/call-with-an-expert).