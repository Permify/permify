import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

# Context (Dynamic Permissions)

## What is it ?

Contextual tuples are relations that can be dynamically added to permission request operations. When you send these relations along with your requests, they get processed alongside existing relations in the database and will return a result.

You can utilize Contextual Tuples in authorization checks that depend on certain dynamic or contextual data (such as location, time, IP address, etc) that have not been written as traditional Permify relation tuples.

## Use Case

Let's give an example to better understand the usage of Contextual Tuples aka dynamic permissions in access checks.

Consider you're modeling an permission system for an internal application that belongs to an multi regional organization.

### Authorization Model

In that application an employee that belongs to HR department can view details of another employee if:

1. If he/she is an Manager in HR department
2. Connected via the branch's internal network or through the branch's VPN

As you notice we can model the rule **1.** easily with our existing schema language, which gives ability to define arbitrary relations between users and objects such as manager of HR entity, as follows,

```perm
entity user {}

entity organization {

    relation employee @user
    relation hr_manager @user @organization#employee

}
```

But to create the `view_employee` permission in the organization entity, we need to consider not only whether the employee is a manager but also check the IP address.

At this point, traditional relation tuples of Permify are insufficient since network address is an dynamic variable that cannot be added as static relations.

So, to incorporate the IP address into our authorization model we will use Contextual Tuples and send dynamic relations values when sending the access check request.

Let's extend our authorization model with adding contextual entities and relations to create the `view_employee` action.

```perm
entity user {}

entity organization {

    relation employee @user
    relation hr_manager @user @organization#employee

    relation ip_address_range @ip_address_range

    action view_employee = hr_manager and ip_address_range.user

}

entity ip_address_range {
    relation user @user
}
```

A quick breakdown we define **type** for contextual variable `ip_address_range` and related them with user. Afterwards call that dynamic entities inside our organization entity and form the `view_employee` permission as follows:

```perm
action view_employee = hr_manager and ip_address_range.user
```

### Dynamic Access Check

Since we cannot create relation statically for `ip_address_range` we need to send ip value on runtime, specifically when performing access control check.

So let's say user Ashley trying to view employee X. And lets assume that,

- She has a **manager** relation in HR department with the tuple `organization:1#hr_manager@user:1`
- She connected to VPN which connected to network 192.158.1.38 - which is Branch's internal network.

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
    Permission: "view_employee",
    Subject: &v1.Subject {
        Type: "user",
        Id: "1",
    },
    Context: *v1.Context {
        Tuples: []*v1.Tuple{
		    {
		        Entity: &v1.Entity {
			        Type: "organization",
                    Id: "1",
                },
		        Relation: "ip_address_range",
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
                Relation: "user",
                Subject: &v1.Subject {
                    Type: "user",
                    Id: "1",
                },
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
client.permission
  .check({
    tenantId: "t1",
    metadata: {
      snapToken: "",
      schemaVersion: "",
      depth: 20,
    },
    entity: {
      type: "organization",
      id: "1",
    },
    permission: "view_employee",
    subject: {
      type: "user",
      id: "1",
    },
    context: {
      tuples: [
        {
          entity: {
            type: "organization",
            id: "1",
          },
          relation: "ip_address_range",
          subject: {
            type: "ip_address_range",
            id: "192.158.1.38",
          },
        },
        {
          entity: {
            type: "ip_address_range",
            id: "192.158.1.38",
          },
          relation: "user",
          subject: {
            type: "user",
            id: "1",
          },
        },
      ],
    },
  })
  .then((response) => {
    if (response.can === PermissionCheckResponse_Result.RESULT_ALLOWED) {
      console.log("RESULT_ALLOWED");
    } else {
      console.log("RESULT_DENIED");
    }
  });
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
    "type": "organization",
    "id": "1"
  },
  "permission": "view_employee",
  "subject": {
    "type": "user",
    "id": "1",
    "relation": ""
  },
  "context": {
    "tuples": [
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
  }
}'
```

</TabItem>
</Tabs>

A quick note,

When you use contextual tuples, the cache system will not be operational. This is because the cache system is written along with snapshots and if contextual tuples are written, using the cache would lead to incorrect results.

Hence, to prevent this, the use of the cache is blocked at the time of the request. See more in the section [Permify Cache Mechanisms](./cache.md)

## Need any help ?

Our team is happy to help you get started with Permify. If you'd like to learn more about using Permify in your app or have any questions about this example, [schedule a call with one of our Permify engineer](https://meetings-eu1.hubspot.com/ege-aytin/call-with-an-expert).
