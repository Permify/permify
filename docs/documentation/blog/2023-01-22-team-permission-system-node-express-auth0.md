---
title: "Build a Team permissions system in Node.js app using Auth0 and Permify - Part 2"
description: "We will build a team permission system in ExpressJs with using Auth0 and Permify."
slug: team-permission-system-node-express-auth0-part-2
authors:
  - name: Ege Aytin
    image_url: https://user-images.githubusercontent.com/34595361/213848483-fe6f2073-18c5-46ef-ae60-8db80ae66b8d.png
    title: Permify Core Team
    email: ege@permify.co
tags: [nodejs, expressjs, auth0, role, permissions, authorization, google zanzibar]
image: https://user-images.githubusercontent.com/34595361/213847473-2c865a2e-6872-4797-918c-ded8dd340068.jpeg
hide_table_of_contents: false
---

![rbac-vue-cover](https://user-images.githubusercontent.com/34595361/213848085-7eb83a3b-5bf6-4caa-a9eb-6d42973b813b.png)

This is Part 2 in the series of guides on building team permission system in Node.js app using Auth0 and Permify. 
<!--truncate-->

## Introduction

In the first part we set up our express.js server and handle authentication via Auth0. In this part we'll handle the authorization with using [Permify](https://github.com/Permify/permify). It is an open-source authorization service for creating and maintaining  access control in your applications.

In this part we will:

1. Build team permission authorization model with [Permify Schema](https://docs.permify.co/docs/getting-started/modeling).
2. Run and set up Permify authorization service.
3. Build endpoints with check permission middleware to secure our resources.
3. Test it out!

### Prerequisites

* [Docker installed](https://docs.docker.com/get-docker/)

## Step 1: Build team permission authorization model

Authorization model is basically the structure of set of rules that give users or services permission to access some data or perform a particular action. Before creating the authorization lets remember our user types and rules for this example. We have 4 different user types to create a simple team permission system:

- **Member:** Member of the organization and can only view teams.
- **Admin:** Administrator in an organization; can view, edit and delete the team resources.
- **Team Manager:** Can view, and edit resources of the team
- **Team Member:** Can view resources of the team.

To develop the above model we’ll use Permify authorization language called Permify Schema. It allows you to specify your entities, the relations between them, and access control options.

In particular, Permify Schema has:

- **Entities:** represents your main object.
- **Relations:** represents relationships between entities.
- **Actions:** describes what permissions the relations can do.

Permify has its own [playground](https://play.permify.co/) where you can create your Permify Schema. 

![playground-image](https://user-images.githubusercontent.com/34595361/214105218-90ce1ae1-9180-41ae-908a-29c2259291c4.png)

Let's create our authorization model according to our team permissions rules above. Copy and paste the following model to the **"Authorization Model"** section in the playground then click **Save** button on above. You can see the relations between entities and permissions on **Visualizer**

```perm
entity user {}

entity organization {

   // organizational user types
   relation admin @user
}

entity team {

   // represents owner or creator of the team
   relation manager @user

   // represents direct member of the team
   relation member @user @organization#member

   // reference for organization that team belong
   relation org @organization
}

entity document {

   // reference for team that team resource belongs
   relation team @team

   // reference for organization that resource belongs
   relation org @organization

   // permissions
   action view = team.member or team.manager or org.admin
   action edit = team.manager or org.admin
   action delete = team.manager or org.admin
}
```

### Breakdown of Schema:

#### Entities & Relations

#### User Entity
The user entity represents users. This is a mandatory entity in Permify Schema.

#### Organization Entity
This entity represents the organization to which the users and the teams belong. The organization entity has 1 user types, the admin. 

#### Team Entity
Organizations and users can have multiple teams, so each team is related with an organization and with users.
The team entity has 3 relations:
- **manager:** represents the owner or creator of team
- **member:** represents direct member of the team
- **org:** reference for organization that team belong

#### Document Entity
The resource entity has 2 relations

```perm
// reference for a team that team resource belongs
relation team @team
 
// reference for the organization that the resource belongs
relation org @organization
```

#### Actions
As we discussed above, actions describe what relations can do which means it defines who can perform a specific action, we can think of actions as permissions for entities. 

We only defined actions on documents for the sake of creating a simple use case for our tutorial. Lets examine document actions.

#### Document Actions

These actions actually represents the user types and rules we defined earlier, lets remember those:

- **Member:** Member of the organization and can only view teams.
- **Admin:** Administrator in an organization; can view, edit and delete the team resources.
- **Team Manager:** Can view, and edit resources of the team
- **Team Member:** Can view resources of the team.

So in Permify it can be achievable with following document actions. 

```perm
   action view = team.member or team.manager or org.admin
   action edit = team.manager or org.admin
   action delete = team.manager or org.admin
```

Lets look at the edit action, if we say we have an document with id 14: only user that is member of the team, which document:14 belongs **or** user has administrator role in organization can edit document:14.

## Step 2: Run and set up Permify authorization service.

Lets run our authorization service in our local environment. We’ll use docker for running our service. If you don't have docker installed on your computer you can easily get it from [here](https://docs.docker.com/get-docker/). Lets run following docker command in our terminal: 

### Run Permify service in local

```js
docker run -p 3476:3476 -p 3478:3478  ghcr.io/permify/permify serve
```

You should see following output on your terminal,

![terminal output](https://user-images.githubusercontent.com/34595361/214109434-51739ef2-6fb7-49c4-8ca2-c8b59ffe7c4a.png)

This will start Permify our authorization service with the default configuration options:
- Port 3476 is used to serve the REST API.
- Port 3478 is used to serve the GRPC Service.
- Authorization data stored in memory

For this tutorial we'll use REST API to manage authorization in our application. You can check our available endpoints from [Permify Swagger Docs](https://app.swaggerhub.com/apis-docs/permify/permify/latest#/)

:::caution
Production usage of Permify needs some other configurations when running this docker command; such as defining running options, selecting datastore to store authorization data, etc. But for simplicity of this tutorial we’ll skip those parts and use our local environment and store authorization data in memory.
:::

### Test our connection via Postman

Lets test our connection with creating an HTTP GET request - localhost:3476/healthz

![healthz-postman](https://user-images.githubusercontent.com/34595361/214111101-1ee1f950-8ad3-497d-898e-3b97e9b9339c.png)

### Configure authorization model 

We’ll use Permify access control checks to secure our endpoints but before that we need to configure our created authorization model to our authorization service and create some data to test it out.

Permify Schema needs to be sent to the [Write Schema API](https://docs.permify.co/docs/api-overview/write-schema) endpoint for configuration of your authorization model.

 Lets copy that schema from our playground using the **Copy** button

![copy-button-playground](https://user-images.githubusercontent.com/34595361/214111689-88f7fd21-d812-48f9-ab9a-998ac2fab9c1.png)

And use it in postman as body params to make a POST "/v1/schemas/write” request as following.

![schema-write](https://user-images.githubusercontent.com/34595361/214112081-4d10a549-f650-417c-b261-f3056872fa28.png)

yayy 🥳, we just completed the configuration of Permify authorization service. Now we have a running API that has authorization model configured and ready to use!

As next steps, we’ll build our endpoints and secure them with [Permify Check Request](https://docs.permify.co/docs/getting-started/enforcement).

## Step 3: Build API endpoints and secure them with Check Middleware

So at that point our Permify API running at port **3476** and our express server running at port **3000** in our local. 

Our express server can behave Permify as an authorization service which is abstracted from source code. And we’ll use this authorization service to protect our API endpoints. But before that we need to create a middleware to determine whether a user is authorized to perform a specific endpoint.

### Creating the check permission middleware

We will create a middleware that will take two parameters: the id of the resource and the permission type of the action as follows:

```js
const checkPermissions = (permissionType) => {
  return async (req, res, next) => {

    // get authenticated user id from auth0
    const userInfo = await req.oidc.user;
    console.log('userInfo', userInfo)
    
    // body params of Permify check request
    const bodyParams = {
      metadata: {
        schema_version: "",
        snap_token: "",
        depth: 20,
      },
      entity: {
        type: "document",
        id: req.params.id,
      },
      permission: permissionType,
      subject: {
        type: "user",
        id: userInfo.sid,
        relation: "",
      },
    };

    // performing the check request
    const checkRes = await fetch("http://localhost:3476/v1/permissions/check", {
      method: "POST",
      body: JSON.stringify(bodyParams),
      headers: { "Content-Type": "application/json" },
    })
    .catch((err) => {
      res.status(500).send(err);
    });
    
    let checkResJson = await checkRes.json()
    console.log('Check Result:', checkResJson)

    if (checkResJson.can == "RESULT_ALLOWED") {
        // if user authorized
        req.authorized = "authorized";
        next();
    } 

    // if user not authorized
    req.authorized = "not authorized";
    next();
  };
};
```

As you can see this middleware performs a check request inside with using "http://localhost:3476/v1/permissions/check" [Permify Check Request](https://docs.permify.co/docs/api-overview/check-api)

We need to pass some information such as; who's performing action, what is the specific action, etc via body params to endpoint: "http://localhost:3476/v1/permissions/check", and this endpoint will return a authorization decision result.

As you seen above the endpoints decision data is added to the req object as a property req.authorized. This can be used to determine whether the user is authorized to perform the action.

This middleware is used in the application's routing to ensure that only authorized users can access specific routes or execute specific operations.

### Build endpoints and secure them with checkPermissions

We’ll create following endpoints to test our authorization structure.

- GET /docs?id API route to view resource
- PUT /docs?id API route to edit resource
- DELETE /docs?id API route to delete resource

For the sake of simplicity, we'll not do any operations in endpoints, just check the access control for each route.

#### View a resource that belongs to a specific team

```js
// view the resource
app.get("/docs/:id", requiresAuth(), checkPermissions("view"), (req, res) => {
  
  /// Result
  res.send(`"User is ${req.authorized} to view document:${req.params.id}`);

});
```

#### Update a resource 
```js
// edit the resource
app.put("/docs/:id", requiresAuth(), checkPermissions("edit"), (req, res) => {
 
  // Result
  res.send(`"User is ${req.authorized} to edit document:${req.params.id}`);

});
```

#### Delete a resource 
```js
// delete the resource
app.delete("/docs/:id", requiresAuth(), checkPermissions("delete"), (req, res) => {

  // Result
  res.send(`"User is ${req.authorized} to delete document:${req.params.id}`);
  
});
```

## Step 4: Test it out

So thus far we build an endpoints that protected from unauthorized actions according to our authorization model, so lets see this endpoints in action.

Our permiy service is running on local:3476 so lets again run our express server with nodemon as follows:

```js
nodemon app.js
```

Since we handled authentication part we should see "Logged in" when in the home page - "localhost:3000/". If you're not authenticated please check out the steps in part 1 of this series to log in.













