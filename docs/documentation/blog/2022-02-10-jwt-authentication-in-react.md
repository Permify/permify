---
title: "JWT Authentication in React"
description: "In this article, we’re gonna build a demo app which demonstrates how to manage authentication in React.js using JWT."
slug: jwt-authentication-in-react
authors:
  - name: Ege Aytin
    image_url: https://user-images.githubusercontent.com/34595361/213848483-fe6f2073-18c5-46ef-ae60-8db80ae66b8d.png
    title: Permify Core Team
    email: ege@permify.co
tags: [react, authorization, JWT, token, authentication]
image: https://user-images.githubusercontent.com/34595361/213847473-2c865a2e-6872-4797-918c-ded8dd340068.jpeg
hide_table_of_contents: false
---

![jwt-react-cover](https://user-images.githubusercontent.com/34595361/213847473-2c865a2e-6872-4797-918c-ded8dd340068.jpeg)

In this article, we’re gonna build a demo app which demonstrates how to manage authentication in React.js using JWT.

<!--truncate-->

Why these two? simply React is the widely used frontend framework (personally my favorite), and JSON Web Token, is the most used authentication protocol on the web.

Because it’s such a common way to manage authentication in client side applications with JWT.  I’ll try to highlight best practices rather than how to implement it.

### Packages we will be use in this app:

- **react-router-dom** for defining routes
- **axios** for api calls

Note: You can find the source code of the application here. → https://github.com/EgeAytin/react-jwt-auth

## Short overview of JWT

First of all, it is useful to shortly talk about what is JWT as a reminder. JWT, or JSON Web Token, is a web protocol used to share security information between client and a server.

In a standard web application, private API requests contain JWT that is generated as a result of the verification of the user information, thus allowing these users to reach protected data and access services.

![jwt-react-example](https://user-images.githubusercontent.com/34595361/213847475-f4d300c2-2ffc-4489-b9de-9a6fe0edd4c7.jpeg)

You can also check out this [stackoverflow thread](https://stackoverflow.com/questions/31687442/where-do-i-need-to-use-jwt) of where to use JWT for a more detailed approach.

## Step 1: Create the Project

So let's create our demo app with create-react prompt:

```js
npx create-react-app react-jwt-auth
```

After creating the project, we should set the scene for implementing JWT authentication to our application, we need:

- **Router** to implement pages, we will use react-router for this,
- **Login page** which we will get user information and send login request to set JWT token
- **Home page** which just be accessible for authenticated users.

### Defining project routes

We will use react router for defining routes, so let's install it with following command:‍

```js
yarn add react-router-dom
```

If you’re using npm it’s fine, in this project I used yarn. Here’s the npm command for installing:

```js
npm install react-router-dom
```

We need to set history property to benefit from **react-router-dom** Router component. Let's add a helpers folder inside our source folder, then add **history.js** inside it:

**history.js**
```js
import { createBrowserHistory } from 'history';
 
export const history = createBrowserHistory();
```

After that, let’s create a file called **routes.js** in source folder to add our routes:

**routes.js**
```js
import React from "react";
import { Redirect, Switch, Route, Router } from "react-router-dom";
 
//history
import { history } from './helpers/history';
 
//pages
import HomePage from "./pages/HomePage"
import LoginPage from "./pages/Login"
 
function Routes() {
   return (
       <Router history={history}>
           <Switch>
               <Route
                   exact
                   path="/"
                   component={HomePage}
               />
               <Route
                   path="/login"
                   component={LoginPage}
               />
               <Redirect to="/" />
           </Switch>
       </Router>
   );
}
 
export default Routes
```

### Updating App.js

After defining the routes of our project, we need to import our Routers component in App.js. I cleared the whole App.js, and just left the Routers component for simplicity.

### Defining Private Route

There are a couple of ways to track authenticated users at the route level with react-router. Here I’ll share an effective and flexible approach which I used over many projects.

We’ll create a Route Guard component as an authentication middleware.

For a more clear explanation, RouteGuard is a route component that you can use instead of 'react-router-dom' Route component for access control management of pages.

So let's implement it.

## Step 2: Creating RouteGuard Component

First I’ll create folder called components, and add **RouteGuard.js** inside.

**RouteGuard.js**

```js
import React from 'react';
import { Route, Redirect } from 'react-router-dom';
 
const RouteGuard = ({ component: Component, ...rest }) => {
 
   function hasJWT() {
       let flag = false;
 
       //check user has JWT token
       localStorage.getItem("token") ? flag=true : flag=false
      
       return flag
   }
 
   return (
       <Route {...rest}
           render={props => (
               hasJWT() ?
                   <Component {...props} />
                   :
                   <Redirect to={{ pathname: '/login' }} />
           )}
       />
   );
};
 
export default RouteGuard;
```

Note: I used react-router-dom version 5 in this article. if you’re using the latest version of it, you should check upgrading from v5 document. V6 offers new features so things change a little bit, for ex. Instead of using render and component props you should use children.

In our RouteGuard component, we return react-router-dom Route component with a conditional operator inside of the render property.

If the user hasn’t got a token on its local storage, we redirect the user to the login page.

As you noticed our hasJWT() function just checks if local storage has JWT or not. You can do various operations inside of this function in order to check authentication, and authorization as well.

For simplicity, I use localStorage to store JWT token. However, it's not the best place to store it in enterprise level apps for security concerns. For more information about  “where to store your JWT tokens” check out this [article](https://javascript.plainenglish.io/where-to-store-the-json-web-token-jwt-4f76abcd4577?gi=75941aed56aa).

Also for  more future proof authentication management I strongly suggest using redux and redux middlewares such as redux thunk.  We’ll create a separate guide just for redux authentication management.

### Updating our Routes.js file

Home page should be accessed only if the user is logged in, in other words the user has a JWT token. So let's import RouteGuard and update our Route component of the Home Page in our routes.js.

**route.js**
```js
import React from "react";
import { Redirect, Switch, Route, Router } from "react-router-dom";
import RouteGuard from "./components/RouteGuard"
 
//history
import { history } from './helpers/history';
 
//pages
import HomePage from "./pages/HomePage"
import LoginPage from "./pages/Login"
 
function Routes() {
   return (
       <Router history={history}>
           <Switch>
               <RouteGuard
                   exact
                   path="/"
                   component={HomePage}
               />
               <Route
                   path="/login"
                   component={LoginPage}
               />
               <Redirect to="/" />
           </Switch>
       </Router>
   );
}
 
export default Routes
```

## Step 3: Creating Home and Login Page

Create pages folder inside our source folder, and add Home.js:

**Home.js**
```js
function HomePage() {
   return (
     <div>
         Home Page
     </div>
   );
}

export default HomePage;
```

We define a simple template for the Home page just for demonstration. In our Login page, we’ll make a sign in request to a server, and set our JWT token to local storage.

Let's set up our Login form:

### Login Request with “axios”

As I mentioned above we need to make a login request to get JWT Token which will be used in private api requests.

We’ll use axios in order to make API calls, so lets install axios with yarn:

yarn add axios

Then let’s import axios to our Login page, and update the handleSubmit function:

**handleSubmit**
```js
const handleSubmit = (email, pass) => {
   //reqres registered sample user
   const loginPayload = {
     email: 'eve.holt@reqres.in',
     password: 'cityslicka'
   }
 
   axios.post("https://reqres.in/api/login", loginPayload)
     .then(response => {
       //get token from response
       const token  =  response.data.token;
 
       //set JWT token to local
       localStorage.setItem("token", token);
 
       //set token to axios common header
       setAuthToken(token);
 
//redirect user to home page
       window.location.href = '/'
     })
     .catch(err => console.log(err));
 };

```

Since we’re haven’t got a real server to make requests, I used reqres fake API to handle the server side of this demo app. You can check out their [documentation here](https://reqres.in/).

In the handleSubmit() function; we make a login request with predetermined email and password values.

Then we get the token from the response, set it to our local storage and axios common header with setAuthorizationToken() helper method which we’ll cover in following section.

And finally redirect the authenticated user to the home page.

Note: These user values are already registered in reqres API, I don’t want to include registration process to this demo app since it's kind of out of scope.

### Set axios common headers with setAuthToken()

In order to use JWT on each private request, we need to add them to the request header as expected. Axios has a header common set option, and we’ll use that with a helper method called setAuthToken().

```js
import axios from 'axios';
 
export const setAuthToken = token => {
   if (token) {
       axios.defaults.headers.common["Authorization"] = `Bearer ${token}`;
   }
   else
       delete axios.defaults.headers.common["Authorization"];
}
```

**Note:** *axios interceptors can be used to do the same operation; especially if you have other config values, or any kind of condition before every request is sent.*

Finally, we’ll need to update App.js as follow to check JWT token right after application is created.

```js
//check jwt token
 const token = localStorage.getItem("token");
 if (token) {
     setAuthToken(token);
 }
```

That’s it, we just simply implemented JWT authentication in a demo React application. With the last operation, our axios requests header will include the Bearer token to protect our application private API’s.

I try to keep things simple for demonstration purposes of JWT usage on the client side. For a complete authentication mechanism, you need to implement registration, verification and logout processes with different services and packages for scaling applications. Apart from this, if you liked this article and wonder more about auth check out our github. Additionally if you have any questions or doubts, feel free to ask them.