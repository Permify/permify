---
title: "Cookie Management in ExpressJS to Authenticate Users"
description: "Express.js is a widely used NodeJs framework by far and if you’re familiar with it, probably cookie management is not a very painful concept for you."
slug: cookie-management-in-expressjs-to-authenticate-users
authors:
  - name: Ege Aytin
    image_url: https://user-images.githubusercontent.com/34595361/213848483-fe6f2073-18c5-46ef-ae60-8db80ae66b8d.png
    title: Permify Core Team
    email: ege@permify.co
tags: [cookie, nodejs, authentication, authorization]
image: https://user-images.githubusercontent.com/34595361/213847158-108069ba-44cb-4fe6-9659-260f6749ef04.jpeg
hide_table_of_contents: false
---

![cookie-express-cover](https://user-images.githubusercontent.com/34595361/213847158-108069ba-44cb-4fe6-9659-260f6749ef04.jpeg)

Express.js is a widely used NodeJs framework by far and if you’re familiar with it, probably cookie management is not a very painful concept for you.

<!--truncate-->

Although there are numerous use cases of the cookies; session management, personalization, etc. We will create a demo app that demonstrates a simple implementation of cookie management in your Express.js apps to authenticate users.

Note: You can find the source code of the application here. → https://github.com/EgeAytin/express-cookies

Before we create our demo app, let's talk a little bit about Cookies;

## So what are HTTP cookies?

HTTP cookies are small pieces of data that are sent from a web server and stored on the client side.

To set up a cookie on the client's side, the server sends a response with the Set-Cookie header.
When the client receives the response message from the server containing the Set-Cookie header, it'll set up the cookie on the client-side.

![cookie-express-1](https://user-images.githubusercontent.com/34595361/213847155-8d94e041-3ace-4628-9486-e3f696cd6668.png)

Such that each subsequent request going from the client-side will explicitly include;

- A header field called “Cookie”
- An actual header that contains the value.
- The cookie information that has been sent by the server in the response message.

Actually, this is enough for scope of our article but If you want to learn more about browser cookies, I recommend reading this [article](https://www.digitalocean.com/community/tutorials/js-what-are-cookies).

## Step 1: Setting up our demo Express app

For a kick-off, we need to create a demo express application where we can implement our cookie management. To quickly create one I’ll use the express app generator tool, express-generator.

Run the following command to create,

```js
npx express-generator express-cookie
```

For earlier Node versions, install the application generator as a global npm package and then launch it, for more check out [express documentation](https://expressjs.com/en/starter/generator.html).

All necessary starter modules and middleware that we will use should already be generated with express-generator, your project folder structure should look like below

All necessary starter modules and middleware that we will use should already be generated with express-generator, your project folder structure should look like below

![cookie-express-2](https://user-images.githubusercontent.com/34595361/213847156-01e9b71c-c403-482e-a549-f4cbd2835d62.png)

## Step 2: Create Basic Authentication middleware

To demonstrate the use of cookies for authentication, we won’t need to implement a fully-fledged authentication system.

So for simplicity, I will use Basic Authentication. The very basic mechanism that will enable us to authenticate users.

### How does Basic authentication work?

When the server receives the request, the server will extract authorization information from the client's request header. And then, use that for authenticating the client before allowing access to the various operations on the server-side.

If this client request does not include the authorization information, then the server will challenge the client, they're asking for the client to submit this information with the user name and password fields.

So, every request message originating from a client should include the encoded form of the username and password in the request header that goes from the client to the server-side.

Open your app.js and add our auth middleware with the logic above as follows:

```js

. . .
function auth (req, res, next) {
//server will extract authorization information from the client's request header
var authHeader = req.headers.authorization;
  if (!authHeader) {
    var err = new Error('You are not authenticated!');
    res.setHeader('WWW-Authenticate', 'Basic');
    err.status = 401;
    next(err);
    return;
 }
 
//If this client request does not include the authorization information
var auth = new Buffer.from(authHeader.split(' ')[1], 'base64').toString().split(':');
var user = auth[0];
var pass = auth[1];
 
//static credential values
if (user == 'admin' && pass == 'password') {
    next(); // user authorized
} else {
    var err = new Error('You are not authenticated!');
    res.setHeader('WWW-Authenticate', 'Basic');    
    err.status = 401;
    next(err);
}
}
app.use(auth);
. . .
```
**Note:** *You should add auth middleware on top of the routers so that the authorization middleware can be triggered correctly when a request is received.*

### Cookie-based authentication

We want only authenticated users to access various operations on the server-side. Here’s how  cookie-based workflow should work;

- The first time that the user tries to access the server, we will expect the user to authorize himself/herself.
- Thereafter, we will set up the cookie on the client-side from the server.
- Subsequently, the client doesn't have to keep sending the basic authentication information. Instead, the client will need to include the cookie in the outgoing request.

## Step 3: Setting the cookie on the Client-Side

Express has a cookie property on the response object,  so we do not need to implement any other library, lets send user name as cookie:

```js
// sentUserCookie creates a cookie which expires after one day
const sendUserCookie = (res) => {
   // Our token expires after one day
   const oneDayToSeconds = 24 * 60 * 60;
   res.cookie('user', 'admin', { maxAge: oneDayToSeconds});
};
```

### Getting the Cookies on request

We will use **cookie-parser** middleware to handle cookies. If you open app.js you will notice that the cookie-parser is already included in our express application, because we generated our project with **express-generator**.

![cookie-express-3](https://user-images.githubusercontent.com/34595361/213847157-0e49971c-1cc0-4c5f-94a4-65f256518411.png)

Note: If you need to explicitly install cookie-parser, the installation command is: 

```js
npm install cookie-parser.
```

cookie-parser parses cookie header and attach on request, so we can access cookies with: req.cookie

Check out the [source code](https://github.com/expressjs/cookie-parser/blob/master/index.js) of the cookie-parser for more information.

## Step 4: Auth mechanism with cookies

We looked at how we can get and set cookies, let's modify our auth middleware for creating a simple authentication mechanism with cookies;

```js
function auth (req, res, next) {
   //check client has user cookie
   if (!req.cookie.user) {
     //get authorization
     var authHeader = req.headers.authorization;
     if (!authHeader) {
         var err = new Error('You are not authenticated!');
         res.setHeader('WWW-Authenticate', 'Basic');             
         err.status = 401;
         next(err);
         return;
     }
     //If this client request does not include the authorization information
     var auth = new Buffer.from(authHeader.split(' ')[1], 'base64').toString().split(':');
     var user = auth[0];
     var pass = auth[1];
     if (user == 'admin' && pass == 'password') {
         sendUserCookie(res)
         next();  // user authorized
     } else {
         var err = new Error('You are not authenticated!');
         res.setHeader('WWW-Authenticate', 'Basic');             
         err.status = 401;
         next(err);
     }
   }
   else {
       //client request has cookie, check is valid
       if (req.cookie.user === 'admin') {
           next();
       }
       else {
           var err = new Error('You are not authenticated!');
           err.status = 401;
           next(err);
       }
   }
 }
```

## Conclusion 

In a nutshell, we build a mechanism that requests some information on the browser for authentication. Afterward, we examine how to persist these auth information by using cookies to prevent resending of auth information.

Now expanding this further, if your server wants to track information about your client, then the server may explicitly set up a session tracking mechanism. Cookies are small, and can't store lots of information.

Now, if we want a lot more information to be tracked about a client on the server-side, then express-sessions enable us to do that. We’ll explore them in the following articles.
