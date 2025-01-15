
# Authorization As A Service | Permify

Getting authorization right is tough, no matter how you've set up your architecture. You're gonna need a solid plan to handle permissions between services, all while keeping it separate from your applications main code. 

In a monolithic app, you can abstract authorization from your app using [authorization libraries](https://permify.co/post/open-source-authorization-libraries/). This involves building a permission system for each individual application or service that is directly connected with the database.

This approach works well until you have several applications with many services. Managing multiple authorization systems for each application is not a scalable approach, as you can imagine.

So due to this, at some point, most companies tend to design these systems as abstract entities, such as a centralized engine, that cater apps that has many services. But its not an easy process for [several reasons](#building-an-authorization-service-is-hard).

Authorization as a service means outsourcing your app's permission management to streamline authorization in your applications. Beyond the clear advantage of saving valuable development time, [it also significantly enhances visibility, scalability, and flexibility](#benefits-of-using-an-authorization-service) within your authorization journey.

[Permify] is an **centralized authorization service** that offers a variety of binding and crafting options to secure your applications. It works in run time and respond to all authorization questions from any of your apps.

![authz-service](https://user-images.githubusercontent.com/34595361/196884110-147862c9-3657-4f07-831c-3e0d0e39eccf.png)

[Permify]: https://github.com/Permify/permify

## Building an Authorization Service is Hard

Building a centralized authorization service yourself is a hard process, and there are several reasons for that.

Although centralizing authorization is good in so many ways it has one big tradeoff. These centralized engines are stateless, meaning they don’t store data. They just behave as an engine to manage functionality such as performing access checks.

For instance; in order to make an access check and compute a decision, you need to load the authorization data and relations from the database and other services. In this case, querying the data needed for access check evaluation presents a significant downside in terms of performance and scalability.

Loading and processing authorization data is especially painful for access checks which come from different environments and services. Also, the authorization service which will be accessed by nearly every other service must be at least as available as the rest of your stack.

So for a centralized authorization service to operate smoothly, this systems needs to have to be fast, consistent, and available all times. 

Another point is, you probably need to have an additional service to to store your authorization data model, which generally includes saving and updating essential permissions like roles, attributes or relationships. This service should manage the entirety of authorization policies, providing administrators the flexibility to adjust these policies when necessary.

## Benefits of using an Authorization Service - Permify

### Move & Iterate Faster 
Avoid the hassle of building your a new authorization system, save time and money by leveraging existing, battle-tested code that has been developed by a team rather than starting from scratch. 

You can get started quickly with a [simple API](../api-overview.md) that you can easily integrate into your application to move and iterate faster.

### Scale As You Wish
Permify based on [Google Zanzibar], which is the global authorization system used at Google for handling authorization for hundreds of its services and products including; YouTube, Drive, Calendar, Cloud and Maps. 



Zanzibar system achieved more than 95% of the access checks responded in 10 milliseconds and has maintained more than 99.999% availability for the 3 year period. 

Permify applies proven techniques that Google used. We’re trying to make Zanzibar available to everyone to use and benefit in their applications and services

:::success Metrics
Currently, Permify can achieve response times of up to **10ms** for access control checks, with handling up to **1 trillion access requests** per second. Thanks to our state-of-the-art [parallel graph engine](https://docs.permify.co/docs/api-overview/permission/check-api/#how-access-decisions-evaluated) and various [cache mechanisms](https://docs.permify.co/docs/reference/cache/) that we operate.
:::

[Google Zanzibar]: https://permify.co/post/google-zanzibar-in-a-nutshell

### Gain Visibility Across Teams
Enterprise-grade authorizations require robust and fine-grained permissions as well as being able to observe and work on these permissions as a group. 

Yet, code-level authorization logic and distributed authorization data among multiple services make it harder to change permissions and keep them up to date all the time. 

Permify is designed to abstract authorization logic from your code and make authorization available to everyone including non-technical people in your organization. 

### Be Extendable, At Any Time
Products quickly changes due to never-ending user requirements as the company scales. It's so common that oldest authorization systems will fall short and needs to be changed in the road. 

Refactoring existing authorization systems is hard because generally these systems sit at the heart of your product. 

Permify has an extendable authorization language that allows you to update the current authorization model easily, securely, and without affecting production. 

After it's tested and ready to go, you can switch new version of your model without breaking a sweat.

### Audit Your Authorization and Ensure Security
Protect your data, prevent unauthorized access and ensure your customers security. 

Permify can help you with things like fraud detection, real-time transaction monitoring, and even risk assessment with various functions that can be used easily with single API calls.

## Need any help on Authorization ?

Our team is happy to help you get started with Permify. If you'd like to learn more about using Permify or how it might fit into your authorization workflow, [schedule a consultation call with one of our account executivess](https://calendly.com/d/cj79-kyf-b4z).

