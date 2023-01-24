---
title: "Build a Team permissions system in Node.js app using Auth0 and Permify - Part 1"
description: "We will build a team permission system in ExpressJs with using Auth0 and Permify."
slug: team-permission-system-node-express-auth0-part-1
authors:
  - name: Ege Aytin
    image_url: https://user-images.githubusercontent.com/34595361/213848483-fe6f2073-18c5-46ef-ae60-8db80ae66b8d.png
    title: Permify Core Team
    email: ege@permify.co
tags: [nodejs, expressjs, auth0, role, OpenID Connect]
image: https://user-images.githubusercontent.com/34595361/213847473-2c865a2e-6872-4797-918c-ded8dd340068.jpeg
hide_table_of_contents: false
---

![rbac-vue-cover](https://user-images.githubusercontent.com/34595361/213848085-7eb83a3b-5bf6-4caa-a9eb-6d42973b813b.png)

In this article series we will build a team permission system in ExpressJs with using Auth0 and Permify.

<!--truncate-->

## Introduction

Working in a team is part of most B2B applications. As such, you need to build a robust auth system that lets an different user types has various degrees of access to resources according to their organization roles, the team their belonged even their role in the team.

For simplicity of this tutorial series, we'll structure our application with 4 user types:

- **Member:** Member of the organization and can only view teams.
- **Admin:** Administrator in an organization; can view, edit and delete the team resources.
- **Team Manager:** Can view, and edit resources of the team
- **Team Member:** Can view resources of the team.

This is Part 1 of the series and in this part we will:

1. Set up a backend server with [Express.js](https://expressjs.com/).
2. Set Up Authentication with using [Auth0 OpenID Connect](https://github.com/auth0/express-openid-connect).
3. Test it out!

In the second part of this series we will set up our authorization structure with using open source authorization service, [Permify](https://github.com/Permify/permify).

### Prerequisites

* [Node.js installed](https://nodejs.org/en/download/)
* [Create a Auth0 account](https://auth0.com/signup?place=header&type=button&text=sign%20up)

## Step 1: Create the backend Node.js server with Express.js

Let’s start by creating a basic server with *express.js*. Start by creating an empty directory and creating a package.json file with the help of the following command: 

```js
npm init -y
```

After you are done creating a package.json file let’s install some packages that we will need for authentication:

- [express](https://www.npmjs.com/package/express)
- [express-openid-connect](https://www.npmjs.com/package/express-openid-connect): Express middleware to protect web applications using [OpenID Connect](https://openid.net/connect/)
- [dotenv](https://www.npmjs.com/package/dotenv): Loads environment variables from .env file.
- [nodemon](https://www.npmjs.com/package/nodemon): It monitors for any changes in your source and automatically restarts your server.

You can download all of the packages with a single command:

```js
npm install express express-openid-connect dotenv nodemon --save
```

Now, let’s quickly create a basic express.js server. Create a new app.js file inside the root folder of our project. In this app.js file, we have just created a basic express server that is running on port 3000.

```js
// app.js 
const express = require("express"); 
const app = express(); 

const port = process.env.PORT || 3000; 
app.listen(port, () => {
   console.log(`Server running on port ${port}`); 
});
```

Test the server by running following command in your terminal, 

```js
nodemon app.js
```
![01_Server_running](https://user-images.githubusercontent.com/34595361/214090558-061474e7-dc68-4469-a1c4-751e9e4efa39.PNG)

## Step 2: Set Up Authentication with Auth0

We will create a simple login that will:

1. Let the user enter an email and password
2. After a user logs in it should be registered as a member user type.

We will use Auth0 to handle authentication and then add the [Express OpenID Connect library](https://openid.net/connect/) (that we installed earlier) to our app for login/logout workflows.

To get started with authentication, [signup](https://auth0.com/signup?place=header&type=button&text=sign%20up) for an Auth0 account. After signing up, you will be automatically redirected to the dashboard. On your dashboard click on **Create Application**.

![4_Auth0_dashboard](https://user-images.githubusercontent.com/34595361/214093320-ad95c910-e20a-4a3c-b5dc-51e26e1c3eed.PNG)

Create a new **Regular Web Application** which we will be using for our authentication. 

![5_Application_type_auth0](https://user-images.githubusercontent.com/34595361/214093329-d6adb780-3aa7-4cef-a63c-df6fd5930ddd.PNG)

You will be asked to select the technology you’re using. For the simplicity of this tutorial, we will skip the integration

![6_skip_integration_auth0](https://user-images.githubusercontent.com/34595361/214093331-bd4e8109-d850-46cb-af16-c429920787dd.png)

Once you are done, you will end up on the dashboard of your Auth0 application.
Go to settings and add the following **Application URIs** (You need to scroll down a little bit)

![7_dasboard_settings_auth0](https://user-images.githubusercontent.com/34595361/214093334-ca9d393e-d135-470c-8387-003f4f275c29.png)

**1. Configure Callback URL:**

This is where the user will be redirected after they complete their authentication.

Set the URL to *http://localhost:3000/callback* in the **Allowed Callback URLs** field for the application you just created. 

**2. Configure Logout URL:**

A logout URL is an application route to which Auth0 can redirect users when they log out.

Set the URL to http://localhost:3000 in the Allowed Logout URLs field for the application you just created. Make sure you save the changes after adding the URLs.

![8_URLS_auth0](https://user-images.githubusercontent.com/34595361/214093339-4fd77211-c2a2-4d7e-9a30-85345b71ac21.png)

Now that we have our application setup we will proceed further to configure the router with the following configuration keys that we will get from our Auth0 application dashboard.

The *Express OpenId Connect library* that we installed earlier provides an auth router in order to attach authentication routes to your application. We will need the following configuration keys in order to configure the router.

- **issuerBaseURL** - The Domain as a secure URL found in your Application settings
- **baseURL** - The URL where the application is served (since its test you can make it localhost:300)
- **clientID** - The Client ID found in your Application settings
- **secret** - A long, random string minimum of 32 characters long

Now, create a new *.env* file in the root of our project which will store all of our configuration keys.

```js
// .env 
ISSUER_BASE_URL = https://YOUR_DOMAIN_URL
CLIENT_ID = YOUR_CLIENT_ID 
BASE_URL = http://localhost:3000 
SECRET = LONG_RANDOM_VALUE 
```

You can get the **Domain** and **Client Id** from your application settings in Auth0 under the **“Basic Information”** section as shown in screenshot below.

![9_domain_clientID](https://user-images.githubusercontent.com/34595361/214093344-34cb8ae6-d1f3-44c1-a132-adf981b46102.png)

We can access these configuration keys in our **app.js** and make the **openid-connect** initialization as follows:

```js
// app.js 
require("dotenv").config(); 
const { auth } = require("express-openid-connect"); 

app.use( 
  auth({ 
    issuerBaseURL: process.env.ISSUER_BASE_URL, 
    baseURL: process.env.BASE_URL, 
    clientID: process.env.CLIENT_ID, 
    secret: process.env.SECRET, 
    idpLogout: true, 
  }) 
);
```

Here, we require a **dotenv** package that will reference the environment variables using the auth router we discussed earlier.

A user can now log in to our application after visiting the **/login** route.

After the completion of authentication, the user will be redirected to the home page that we don’t have set up yet. So, let’s quickly set up our root route.

```js
// app.js 
app.use( 
  auth({ 
    authRequired: false, 
    auth0Logout: true,
    issuerBaseURL: process.env.ISSUER_BASE_URL, 
    baseURL: process.env.BASE_URL, 
    clientID: process.env.CLIENT_ID, 
    secret: process.env.SECRET, 
    idpLogout: true, 
  }) 
); 

app.get("/", (req, res) => {      
   res.send(
       req.oidc.isAuthenticated() ? "Logged in" : "Logged  out"
   ); 
});
```

Also, note that we have set up two properties: 

- **authRequired: false** which will make sure that each route requires authentication, 

- **auth0Logout: true**

The last thing we want to do is to create a **/profile** route that will show the information about the user. The **/profile** route consists of **requiresAuth** middleware for routes that require authentication. 

Every route utilizing this middleware will check to see if there is a current, active user session, and if not, it will direct the user to log in.

That’s it, pat yourself on the back if you have made it till here 🎉

The final **app.js** file should look like this:

```js
// app.js
const express = require("express");
const app = express();
require("dotenv").config();
const { auth, requiresAuth } = require("express-openid-connect");

app.use(
  auth({
    authRequired: false,
    auth0Logout: true,
    issuerBaseURL: process.env.ISSUER_BASE_URL,
    baseURL: process.env.BASE_URL,
    clientID: process.env.CLIENT_ID,
    secret: process.env.SECRET,
    idpLogout: true,
  })
);

// req.isAuthenticated is provided from the auth router
app.get("/", (req, res) => {
  res.send(req.oidc.isAuthenticated() ? "Logged in" : "Logged out");
});

// The /profile route will show the user profile as JSON
app.get("/profile", requiresAuth(), (req, res) => {
  res.send(JSON.stringify(req.oidc.user));
});

const port = process.env.PORT || 3000;
app.listen(port, () => {
  console.log(`Server running on port ${port}`);
});
```

## Step 3: Test it out

Let’s test out what we just implemented, run the app on your terminal with the following 

```js
nodemon app.js
```

![10_localhostlogout](https://user-images.githubusercontent.com/34595361/214093345-8800930e-8880-4427-a43b-c8129f2c0921.PNG)

Now, visit **localhost:3000/login** route for authentication. Once you visit the route it redirects you to the Auth0 custom page for login.

![auth0_custom_page](https://user-images.githubusercontent.com/34595361/214096225-d760de89-4b01-4cb1-99d7-a4ca7fab1601.png)

Enter your details and you will be redirected back to the application

![11_logged_in](https://user-images.githubusercontent.com/34595361/214093349-32da632f-62cb-4af2-91b0-d78bac08a53a.PNG)

Since you logged in, now you can try to get your user details with the **/profile** endpoint 🎉

## Next Steps

That’s it for Part 1 of this tutorial where we created a basic server with express.js and then implemented authentication with Auth0.

This was just the authentication part for the 2 part series of **Build a Team permissions system using Auth0 and Permify.**

In the next part, we will dive deep into how to implement authorization with Permify on our NodeJS application.

If you have any questions or doubts, feel free to ask them :)
