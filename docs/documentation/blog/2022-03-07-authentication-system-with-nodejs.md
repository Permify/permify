---
title: "Authentication system in Node using PassportJs, Express, and MongoDB"
description: "One of the toughest topics while building API is, for sure, implementing user authentication. In this piece, we will take a look at how to build a simple and solid authentication approach in NodeJS with PassportJs."
slug: authentication-system-with-nodejs
authors:
  - name: Ege Aytin
    image_url: https://user-images.githubusercontent.com/34595361/213848483-fe6f2073-18c5-46ef-ae60-8db80ae66b8d.png
    title: Permify Core Team
    email: ege@permify.co
tags: [nodejs, authorization, mongodb,permissions, express, ]
image: https://user-images.githubusercontent.com/34595361/213846772-1f42a7fd-e331-42c8-b9b8-d24a5eeb3458.png
hide_table_of_contents: false
---

<!-- ![Authentication system in Node using PassportJs, Express, and MongoDB](https://user-images.githubusercontent.com/34595361/213846772-1f42a7fd-e331-42c8-b9b8-d24a5eeb3458.png) -->

One of the toughest topics while building API is, for sure, implementing user authentication. In this piece, we will take a look at how to build a simple and solid authentication approach in NodeJS with PassportJs.

<!--truncate-->

We will create a demo express app that you will be able to use as an authentication boilerplate in your ExpressJs applications.

## Prerequisites

Before we start, let's go over the technologies we will be using:

- NodeJs
- ExpressJs
- Passport.js
- MongoDB and Mongoose

## What will we build? 

This is Part 1 in the series of guides on creating an authentication system using Node.js, Express for your SaaS app.

In this part we will build a token-based authentication with Passport. In particular, we will;

- Use JSON web tokens for token-based user authentication 
- Use the Passport module together with passport-local and passport-local-mongoose for setting up local authentication within your server.

Source Code: https://github.com/EgeAytin/node-express-authentication

## Passport.js overview

Passport is nothing but an authentication middleware that supports various strategies. It can be used for user authentication;

- Including a local strategy like using username and password.
- Even third-party authentication.
- Or using OAuth or OAuth 2.0, like using Facebook, Twitter, Google, and so on. 

For more detail check out their [documentation](https://www.passportjs.org/docs/).

## Step 1: Setting up our demo Express app

As a start, we need to create a demo express application. Let’s quickly creatine with the express app generator tool, express-generator.

Run the following command to create;

```js
npx express-generator node-express-authentication
```

For earlier Node versions, install the application generator as a global npm package, and then launch it.

For more - https://expressjs.com/en/starter/generator.html

### Packages we need to install:

In addition to predefined packages on express-generator, we will use the following packages in our project;

- passport
- passport-local
- passport-jwt
- passport-local-mongoose
- jsonwebtoken
- mongoose
- express-session

You can install them with a single line of command:

```js
npm install mongoose passport passport-local passport-jwt passport-local-mongoose jsonwebtoken express-session --save 
```

### Add Config.js

Let's add a file named config.js in the root folder of our project.

In this config.js file, I’ll specify the secret key that we are going to be using for signing my JSON Web Token.

And also I’ll specify a Mongo URL here, which will be the URL for my MongoDB server.

```js
module.exports = {
   'secretKey': '12345-67890-09876-54321',
   'mongoUrl' : 'mongodb://localhost:27017/your-server'
}
```

### Update app.js

So once we have completed creating the config.js file, we need to update the app.js file in order to initialize PassportJS and connect to our MongoDB server.

```js
var passport = require('passport');
var config = require('./config');
 
. . .
 
//db connection
const mongoose = require('mongoose');
const url = config.mongoUrl;
const connect = mongoose.connect(url);
 
connect.then((db) => {
   console.log("Connected correctly to server");
}, (err) => { console.log(err); });
 
. . .
 
//passport initialization
app.use(passport.initialize());
```

## Step 2: Creating User Schema and Model

We will create the User schema and model to use the passport-local-mongoose. Let's create a  models folder in the root folder and create user.js within:‍

Note: If you aren't familiar with MongoDB or Mongoose, I highly suggest reading this [article](https://www.freecodecamp.org/news/introduction-to-mongoose-for-mongodb-d2a7aa593c57/) first.

```js
var mongoose = require('mongoose');
var Schema = mongoose.Schema;

var passportLocalMongoose = require('passport-local-mongoose');

var User = new Schema({
    roles: [{
        type: String,
        required: false
    }],
    attributes: [{
        type: Object,
        required: false
    }],
});

User.plugin(passportLocalMongoose);

module.exports = mongoose.model('User', User);
```

We don't need to add username or password fields, because these would be automatically added in by the passport-local-mongoose plugin. So, I’ve just define role names and attributes fields for the authorization part which we cover in next chapter.

Note that to use passport-local-mongoose in our mongoose schema and model, we need to assign the User model to the Plugin:

```js
User.plugin(passportLocalMongoose);
```

## Step 3: Creating Authentication Strategy

Let's create a new file called authenticate.js in the project folder. We're going to use this file to store the authentication strategies that we will configure.

We will implement a strategy that has the following logic: If the Bearer token is included in the incoming request; then it will be extracted to authenticate the user based upon the token.

```js
var passport = require('passport');
var config = require('./config.js');
 
// User model
var User = require('./models/user');
 
// Strategies
var JwtStrategy = require('passport-jwt').Strategy;
var ExtractJwt = require('passport-jwt').ExtractJwt;
var LocalStrategy = require('passport-local').Strategy;
 
// Used to create, sign, and verify tokens
var jwt = require('jsonwebtoken');
 
// Local strategy with passport mongoose plugin User.authenticate() function
passport.use(new LocalStrategy(User.authenticate()));
 
// Required for our support for sessions in passport.
passport.serializeUser(User.serializeUser());
passport.deserializeUser(User.deserializeUser());
 
exports.getToken = function(user) {
   // This helps us to create the JSON Web Token
   return jwt.sign(user, config.secretKey,{expiresIn: 3600});
};
 
// Options to specify for my JWT based strategy.
var opts = {};
 
// Specifies how the jsonwebtoken should be extracted from the incoming request message
opts.jwtFromRequest = ExtractJwt.fromAuthHeaderAsBearerToken();
 
//Supply the secret key to be using within strategy for the sign-in.
opts.secretOrKey = config.secretKey;
 
// JWT Strategy
exports.jwtPassport = passport.use(new JwtStrategy(opts,
   // The done is the callback provided by passport
   (jwt_payload, done) => {
     
       // Search the user with jwt.payload ID field
       User.findOne({_id: jwt_payload._id}, (err, user) => {
           // Have error
           if (err) {
               return done(err, false);
           }
           // User exist
           else if (user) {
               return done(null, user);
           }
           // User doesn't exist
           else {
               return done(null, false);
           }
       });
   }));
 
// Verify an incoming user with jwt strategy we just configured above   
exports.verifyUser = passport.authenticate('jwt', {session: false});
```

## Step 4: Defining User Routes

Since our project was generated with an express-generator, the user's route was already created for us. For simplicity purposes, we will use the users' route, and won't add an extra route.

Let's open the user.js file in the routes folder, and add the necessary routes in order to complete the authentication flow.

We will add login, sign up, register and log out endpoints:

```js
var express = require('express');
var router = express.Router();
var passport = require('passport');
 
// User model
var User = require('../models/user');
 
// Parse Json
const bodyParser = require('body-parser');
router.use(bodyParser.json());
 
// Get our authenticate module
var authenticate = require('../authenticate');
 
// Get Users
router.get('/', function(req, res, next) {
 res.send('respond with a resource');
});
 
// Register User
router.post('/signup', (req, res, next) => {
 // Create User
 User.register(new User({username: req.body.username}),
   req.body.password, (err, user) => {
   if(err) {
     res.statusCode = 500;
     res.setHeader('Content-Type', 'application/json');
     res.json({err: err});
   }
   else {
     // Use passport to authenticate User
     passport.authenticate('local')(req, res, () => {
       res.statusCode = 200;
       res.setHeader('Content-Type', 'application/json');
       res.json({success: true, status: 'Registration Successful!'});
     });
   }
 });
});
 
// Login
router.post('/login', passport.authenticate('local'), (req, res) => {
 
 // Create a token
 var token = authenticate.getToken({_id: req.user._id});
 
 // Response
 res.statusCode = 200;
 res.setHeader('Content-Type', 'application/json');
 res.json({success: true, token: token, status: 'You are successfully logged in!'});
});
 
// Logout
router.get('/logout', (req, res) => {
 if (req.session) {
   req.session.destroy();
   res.clearCookie('session-id');
   res.redirect('/');
 }
 else {
   var err = new Error('You are not logged in!');
   err.status = 403;
   next(err);
 }
});
 
module.exports = router;
```

When the user is successfully authenticated on the “/login” endpoint, the token will be created by the server and sent back to the client or the user. 

So, the client will include the token in every subsequent incoming request in the authorization header with ExtractJWT.fromAuthHeaderAsBearerToken() method, which we defined in authenticate.js file.

I'll show you how this is done in the following Test With Postman section.

The authentication scheme that we have built so far can also be used for third-party authentication based on OAuth 2.0, etc.

We'll create a separate guide just for this. To get updated, join our newsletter below.

## Step 5: Verify User in incoming requests

So now, any place that we want to verify the user, We can use the verifyUser function which is exported from authenticate.js. 

How do we make use of this? 

We're going to implement it into the “/users” endpoint that is predefined with express generator, so lets update the routes/user.js :

```js
// Get Users
router.get('/', authenticate.verifyUser, (req, res, next) =>{
 // Get all records
 User.find({})
   .then((users) => {
       res.statusCode = 200;
       res.setHeader('Content-Type', 'application/json');
       // format result as json
       res.json(users);
   }, (err) => next(err))
   .catch((err) => next(err));
});
```

## Step 6: Test with Postman

We just simply implemented token-based authentication with Passport in the Express app. So let’s test our authentication layer with Postman.

### Start server
Let's start our server with npm start,

Note: Do not forget that your MongoDB server also needs to be started

Then check the index route on Postman as below,

![postman-test-1](https://user-images.githubusercontent.com/34595361/213846714-a7e09353-cda5-4a2c-800a-74310d946022.png)

As you can see we access our server index endpoint because we didn’t configure an authentication check for that route.

​​Now let's test our users index endpoint. We should get “unauthorized” as a result.

![postman-test-2](https://user-images.githubusercontent.com/34595361/213846716-d9e4fdb4-5210-42ad-a743-f21695d4aa6c.png)

### Register User

To register users to our mongo database we should send a payload object which consists of username and password fields on the body:

![postman-test-3](https://user-images.githubusercontent.com/34595361/213846719-aacf5b32-6bd0-44e2-b28c-dca9320d799e.png)

### Login and test authentication

Let's log in with the user we just created above:

![postman-test-4](https://user-images.githubusercontent.com/34595361/213846720-3414811e-e573-4207-9251-1ed0cfaaef31.png)

On successful login, our server will return the JWT. We will use this token to set the Authorization header on protected endpoint requests as below:

![postman-test-5](https://user-images.githubusercontent.com/34595361/213846721-ca160c30-4c03-486c-aee9-d04e3c29e6e0.png)

As you can see we filled the Authorization header field with Bearer -token-, and fetched all the users on our database.

That's the end of our demo project. We implemented a simple authentication mechanism with the Passport on Express app.

Do not forget that Passport is just an authentication middleware, hence it doesn't cover the other parts of authentication for us; such as API token mechanisms, password reset token mechanisms etc.

To sum up, this mechanism only can be a boilerplate, or a starter for your full authentication implementation unless you have a test or lightweight application. If you wonder what you need to consider when building a robust authentication in NodeJs take a look at this [article](https://hackernoon.com/your-node-js-authentication-tutorial-is-wrong-f1a3bf831a46).

Apart from these, if you have any questions or doubts, feel free to ask ping me :)

