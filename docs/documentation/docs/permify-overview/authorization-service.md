
# What is Authorization Service?

Authorization is an important part of software development. There are many different ways to implement authorization, but it's important for all apps to have some form of it in order to protect the user from malicious actors and unauthorized access attempts.

An authorization service is a module that allows you to manage access to your application and ease the development and maintenance of your authorization system. It works in run time and respond to all authorization questions from any of your apps.

![authz-service](https://user-images.githubusercontent.com/34595361/196884110-147862c9-3657-4f07-831c-3e0d0e39eccf.png)

[Permify] is a fully open source authorization service that offers a variety of binding and crafting options to secure your applications.

[Permify]: https://github.com/Permify/permify

## Why should I use Authorization Service instead of doing from scratch?

### Move & Iterate Faster 
Avoid the hassle of building your a new authorization system, save time and money by leveraging existing, battle-tested code that has been developed by a team rather than starting from scratch. You can started quickly with a simple API that you can easily integrate into your application to move and iterate faster.

### Do Not Reinvent The Wheel
Permify based on [Google Zanzibar], which is the global authorization system used at Google for handling authorization for hundreds of its services and products including; YouTube, Drive, Calendar, Cloud and Maps. Building a scalable and performent authorization system is hard and needs a quite engineering time. Zanzibar system achieved more than 95% of the access checks responded in 10 milliseconds and has maintained more than 99.999% availability for the 3 year period. Permify applies proven techniques that Google used. We’re trying to make Zanzibar available to everyone to use and benefit in their applications and services.

[Google Zanzibar]: https://www.permify.co/post/google-zanzibar-in-a-nutshell

### Gain Visibility Across Teams
Enterprise-grade authorizations require robust and fine-grained permissions as well as being able to observe and work on these permissions as a group. Yet, code-level authorization logic and distributed authorization data among multiple services make it harder to change permissions and keep them up to date all the time. Permify is designed to abstract authorization logic from your code and make authorization available to everyone including non-technical people in your organization. 

### Be Extendable, At Any Time
Products quickly changes due to never-ending user requirements as the company scales. It's so common that oldest authorization systems will fall short and needs to be changed in the road. Refactoring existing authorization systems is hard because generally these systems sit at the heart of your product. Permify has an extendable authorization language that allows you to update the current authorization model easily, securely, and without affecting production. After it's tested and ready to go, you can switch new version of your model without breaking a sweat.

### Audit Your Authorization and Ensure Security
Protect your data, prevent unauthorized access and ensure your customers security. Permify can help you with things like fraud detection, real-time transaction monitoring, and even risk assessment with various functions that can be used easily with single API calls.

## Cases that can benefit from An Authorization Service:

- If you already have an identity/auth solution and want to plug in fine-grained authorization on top of that.
- If you want to create a unified access control mechanism to use across your individual applications.
- If you want to make future-proof authorization system and don't want to spend engineering effort for it.
- If you’re managing authorization for growing micro-service infrastructure.
- If your authorization logic is cluttering your code base.
- If your data model is getting too complicated to handle your authorization within the service.
- If your authorization is growing too complex to handle within code or API gateway.

