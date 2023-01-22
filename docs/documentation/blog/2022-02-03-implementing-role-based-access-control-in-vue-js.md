---
title: "Implementing Role-Based Access Control in VueJS"
description: "In this article, I’ll share an effective approach to implement RBAC in VueJs applications."
slug: implementing-role-based-access-control-in-vue-js
authors:
  - name: Ege Aytin
    image_url: https://user-images.githubusercontent.com/34595361/213848483-fe6f2073-18c5-46ef-ae60-8db80ae66b8d.png
    title: Permify Core Team
    email: ege@permify.co
tags: [role, permissions, vue, security]
image: https://user-images.githubusercontent.com/34595361/213848085-7eb83a3b-5bf6-4caa-a9eb-6d42973b813b.png
hide_table_of_contents: false
---

![rbac-vue-cover](https://user-images.githubusercontent.com/34595361/213848085-7eb83a3b-5bf6-4caa-a9eb-6d42973b813b.png)

Implementing client-side authorization is one of the toughest topics for frontend developers. Not just because it's complicated, but also it takes time to build it, especially for Business SaaS applications.

Wide range of best practices for different tech stacks (Frameworks, languages etc.)
Different access control needs of your users.
Flexibility and maintenance.

<!--truncate-->

Although there are many ways to implement authorization & access control, I’ll share an effective approach to implement RBAC in VueJs applications. To keep this article simple and easy to understand, we’re going through a demo project generated with the Vue CLI.

## Creating Vue demo project:

From your Terminal or Command Prompt, execute the following command to create a new project:

```js
vue create vue-rbac-demo
```

Firstly, we need to clear our application in order to start fresh. So after the Vue project is generated your folder structure should look like following unless you didn’t choose features manually while creating your app.

![rbac-vue-1](https://user-images.githubusercontent.com/34595361/213848080-0a038003-24f6-4645-9e0e-9832f4f692ba.png)

## Restructuring the project:

Just remove folders, assets, and components. Open up a folder called views and create;

- **ContactDetails.vue** file as our protected page
- **Home.vue** as a public page
- **Unauthorized.vue** file for redirecting on forbidden access attempts

**ContactDetails.vue**
```js
<template>
  <div>
    Contact's Sensitive Information
  </div>
</template>

<script>
export default {
  name: 'ContactDetails',
}
</script>
```

**Home.vue**
```js
<template>
  <div>
    Home
  </div>
</template>

<script>
export default {
  name: 'Home',
}
</script>
```

**Unauthorized.vue**
```js
<template>
  <div>
    <p>This action is unauthorized </p>
    <router-link to="/">Back to home page</router-link>
  </div>
  
</template>

<script>
export default {
  name: 'Unauthorized',
}
</script>
```

After these operations, we need to define our application routers by creating a new folder called router, then add **index.js** within it to create routes.

Firstly we need to install **“vue-router”** with following command:

```js
npm install vue-router@3
```

**Note:** *Since we are using default Vue 2.0 project, I added the vue router version 3. If you're using  Vue 3.0 you can install latest version of it. You can find the latest changes and migrating from Vue 2 informations on this page ‍--> [router.vuejs.org/guide/migration/](https://router.vuejs.org/guide/migration/)*

After “vue-router” package installed, we will create router/index.js as follows:

```js
import Vue from 'vue'
import VueRouter from 'vue-router'

Vue.use(VueRouter)

const routes = [
   {
    path: '/',
    name: 'Home',
    component: () => import('../views/Home.vue'),
    meta: {
      authRequired: 'false',
    },
  },
  {
    path: '/contact-details/:id',
    name: 'ContactDetails',
    component: () => import('../views/ContactDetails.vue'),
    meta: {
      authRequired: 'true',
    },
  },
  {
    path: '/unauthorized',
    name: 'Unauthorized',
    component: () => import('../views/Unauthorized.vue'),
    meta: {
      authRequired: 'false',
    },
  }
]

const router = new VueRouter({
  mode: 'history',
  base: process.env.BASE_URL,
  routes
})

export default router
```

To keep things simple and clear, I just define the minimum required routes. As you can see I change the default mode to history mode in order to discard the ‘#/‘ situation.

Also, I add “authRequired” meta attribute to each route to understand which page is protected. If “authRequired” meta is ‘false’ we don’t need to check the authorization of the user, like on the home page.

After setting up the router, Let’s add it to main.js with the following code:

```js
import Vue from 'vue'
import App from './App.vue'
import router from './router'

Vue.config.productionTip = false

new Vue({
  router,
  render: h => h(App)
}).$mount('#app')
```

As a result of clearing and updating our demo project, our folder structure should look like below:

![rbac-vue-2](https://user-images.githubusercontent.com/34595361/213848082-6bc76074-b207-4195-9708-afc74d6665dc.png)

## Creating User object with role and permissions

In order to control access checks, we need logged-in users’ roles and permissions. In real-world applications, there are a couple of ways to get and set the users’ roles and permissions from your server API.

However since this is a demo app; I’ll create a sample user JSON, and use its roles and permissions in order to perform access checks.

Let's create user.json file inside our source folder, and add the following sample JSON data:

```json
{
    "id": "g729ad9sf4q3e4kd1ya4",
    "email": "marlee.jenkins@sample.com",
    "first_name": "Marlee",
    "last_name": "Jenkins",
    "avatar_url": "https://i.pravatar.cc/150?img=8",
    "role":  {
        "id": "2",
        "key": "manager",
        "permissions": [
            {
                "id": "5",
                "key": "create-contact"
            },
            {
                "id": "6",
                "key": "update-contact"
            }
        ]
    },
    "created_at": "2021-09-26T13:27:16.436575Z",
    "updated_at": "2021-11-08T13:39:28.905851Z",
    "verified": true
}
```

After adding user.json, we'll update our **App.js** file, lets make It a simple link page that behaves like a standard navigation bar with default styles as below:

```vue
<template>
  <div id="app">
    <div id="nav">
      <router-link to="/">Home</router-link> |
      <router-link :to="{ name: 'ContactDetails', params: { id: user.id }}">Contact Details</router-link>
    </div>
    <router-view/>
  </div>
</template>

<script>
import user from './user.json'

export default {
  data() {
    return {
      user: user,
    };
  }
};
</script>

<style>
#app {
  font-family: Avenir, Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  text-align: center;
  color: #2c3e50;
}
</style>
```


To start our project let's execute npm serve (or yarn serve whichever you use). You should see a page similar to the below:

![rbac-vue-3](https://user-images.githubusercontent.com/34595361/213848084-8b6e5020-153e-4968-b910-d2533368ed02.png)

## Access Control Using Permissions and Roles in Vue

In the previous steps, we reorganize our project and create a sample user JSON. Now we can implement access control in our frontend based on users’ roles and permissions.

Let's implement access control for both our routes and UI elements.

### Access Control In Route Level

We’ll implement access control for our application’s routes by checking certain conditions before every route change, let’s create fake conditions for users to access the contact details page:

* Admin and super admin roles can directly access contact details.
* If the user is the contact itself, he/she can access their own details as expected.
* Manager role can access contact details only if the user has permissions “create-contact” and “update-contact”.

Users can access the contact details page if conditions hold true for them.

The vue-router package provides beforeEach() method which we use to perform our access checks. **beforeEach()** is a navigation guard responsible for protecting routes on navigation change, as you can tell by the name, it invokes before every route change.

So let’s import **user.js** in router/index.js file and add the beforeEach() method below our router instance as follows:

```js
router.beforeEach((to, from, next) => {
  //check page is protected or not
  if (to.meta.authRequired === 'true') {

    //get contact's id
    const contactId = to.params.id

    //access check
    if (
      //if user is admin or super admin
      user.role === 'super_admin' ||
      user.role === 'admin' ||
      //if user is the contact itself
      user.id === contactId ||
      //if user is manager and has necessary permissions
      user.role === 'manager' &&
      user.role.permissions.some(p => p.key === 'create-contact') &&
      user.role.permissions.some(p => p.key === 'update-contact')
    ) {
      return next()
    } else {
      router.push({
        name: 'Unauthorized'
      })
    }
  } else {
    return next()
  }
});
```

### Access Control In UI Layers and Components

In the previous step, we cover the access checks at route level, The second common way is flagging components or UI layers.

Simply, we want to show/hide certain elements on our page depending on users’ permissions. For instance, we want to show certain components only to the users who met the conditions. Otherwise, we won’t show you those specific components.

We can use created property to implement access control logic to our components, also we need to set a condition to UI elements to perform hide/show actions. We can easily do that with v-if directive.

To see this component-level access check in action lets add a components folder and create CreateContact.vue file with the following code:,

```vue
<template>
  <div v-if="isAuthorized">
    <button type="button">Create Contact</button>
  </div>
  <div v-else>
    <div>You don't have permission to see create contact button!</div>
  </div>
</template>

<script>
import user from '../user.json'
export default {
  data() {
    return {
      isAuthorized: false,
    };
  },
  async created() {
    if (
      //if user is admin or super admin
      user.role === "super_admin" ||
      user.role === "admin" ||
      //if have necessary permission
      user.role.permissions.some((p) => p.key === "create-contact")
    ) {
      this.isAuthorized = true;
    } 
  },
};
</script>
```

To see our create contact component in action we'll add it to Home Page as follows:

```vue
<template>
  <div>
    <p>Home Page</p>
    <CreateContact/>
  </div>
</template>

<script>
import CreateContact from "../components/CreateContact"
export default {
  name: 'Home',
  components:{CreateContact}
}
</script>
```

## Conclusion

Why you probably need to outsource Access Control
You can use many different methods instead of this. For instance, if you want more dynamic approach, using computed property instead of created might be a good solution for you.

Apart from these, as you noticed we created fake scenario conditions and handle their access checks with if-else.

Adding everywhere this kind of heavy loaded if checks might not be an ideal case for a real-world application. Also, you should get / fetch every information that you check on “if” statement. In this demo app we just gave user.json sample object, and didn’t care that too much.

I try to keep things simple for demonstration purposes. But developing an RBAC mechanism on scale up application is much tougher than it should be.

If you have any questions or doubts, feel free to ask them. Also you can find the source code of the application here: [github.com/EgeAytin/vue-rbac-demo](https://github.com/EgeAytin/vue-rbac-demo)