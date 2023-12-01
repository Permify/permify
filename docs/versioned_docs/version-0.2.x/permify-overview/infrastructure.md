
# Where does Permify fit into your environment?

Permify is a simply GRPC service that responsible for managing and authorization in your environment. This section shows where and how does Permify fit into your environment with examining Permify's architecture, deployment patterns and the usage with the authentication and identity providers.

## Architecture

Permify stores access control relations on a database you choose and performs authorization checks based on the stored relations. We called that database Write Database - **WriteDB** - and it behaves as a centralized data source for your authorization system.

You can model your authorization with Permify's DSL - Permify Schema and your applications can talk to Permify API over REST API or GRPC Service to perform access control checks, read or query authorization-related data, or make changes to data stored in WriteDb. 

![relational-tuples](https://user-images.githubusercontent.com/34595361/186108668-4c6cb98c-e777-472b-bf05-d8760add82d2.png)

### Permify with Authentication 

Authentication involves verifying that the person actually is who they purport to be, while authorization refers to what a person or service is allowed to do once inside the system.

To clear out, Permify doesn't handle authentication or user management. Permify behave as you have a different place to handle authentication and store relevant data. Authentication or user management solutions (AWS Cognito, Auth0, etc) only can feed Permify with user information (attributes, identities, etc) to provide more consistent authorization across your stack. 

### Permify with Identity Providers

Identity providers help you store and control your users’ and employees’ identities in a single place. 

Let’s say you build a project management application. And a client wants to connect this application via SSO. You need to connect your app to Okta. And your client can control who can access the application, and which group of authorization types they can have. But as a maker of this project management app. You need to build the permissions and then map to Okta. 

What we do is, help you build these permissions and eventually map anywhere you want.

## Deployment Patterns

There are two main deployment patterns that you can follow, integrate Permify into your applications as a sidecar or using Permify as a service across your applications. Despite for both of these deployment patterns implementation is same - running Permify API in a environment you choose - the architectural aspects and usages differs. So let's examine them both.

### Permify As A Service

Permify can be deployed as a sole service that abstracts authorization logic from core applications and behaves as a single source of truth for authorization. Gathering authorization logic in a central place offers important advantages over maintaining separate access control mechanisms for individual applications. See the [What is Authorization Service] Section for a detailed explanation of those advantages.

[What is Authorization Service]: ../authorization-service

![load-balancer](https://user-images.githubusercontent.com/34595361/201173835-6f6b67cd-d65b-4239-b695-04ecf1bad5bc.png)

Since multiple applications could interact with the Permify Service on that pattern, preventing bottleneck for Permify endpoints and providing high availability is important. As shown from above schema, you can horizontally scale Permify Service with positioning Permify instances behind of a load balancer. 

### Using Permify as a Sidecar

Permify can be used as a sidecar as well. In this deployment model, each application uses its own Permify instance and manages its own specific authorization. 

![load-balancer](https://user-images.githubusercontent.com/34595361/201466158-951d5111-843d-4ed2-a4e6-82f2f8edf16a.png)

Although unified authorization offers many advantages, using the sidecar model ensures high performance and availability plus avoids the risk of a single point of failure of the centered authorization mechanism.


