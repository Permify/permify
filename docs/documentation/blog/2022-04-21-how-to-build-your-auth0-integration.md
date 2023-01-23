---
title: "How to Build Auth0 Integration To Your Application"
description: "In this article, I will explain how to build an Auth0 integration. I will perform this integration through the actions that auth0 has just introduced and are currently in beta."
slug: how-to-build-your-auth0-integration
authors:
  - name: Tolga Ozen
    image_url: https://user-images.githubusercontent.com/34595361/213848541-8d4da803-8842-4adc-8125-1ca1838b51b9.jpeg
    title: Permify Core Team
    email: tolga@permify.co
tags: [auth0, permissions, authorization, authentication]
image: https://user-images.githubusercontent.com/34595361/213844539-ef15eb7a-ae67-4781-b83e-be077f0e8af4.png
hide_table_of_contents: false
---

![Building an Auth0 Integration](https://user-images.githubusercontent.com/34595361/213844539-ef15eb7a-ae67-4781-b83e-be077f0e8af4.png)

In this article, I will explain how to build an Auth0 integration. I will perform this integration through the actions that auth0 has just introduced and are currently in beta.

<!--truncate-->

## What Is Action?
Auth0 actions is a feature that allows you to customize auth0 behavior. With Actions, you can add basic custom logic to your login and identity flows tailored to your needs. Actions also allow you to connect external integrations that enhance your overall extensibility experience. It allows Auth0 customers to implement solutions like authentication or permissions management without writing any code.

For example, you can add an action to your app registration flow using a Marketplace Partner that specializes in authorization.

Actions Integrations are standalone functions executed at selected points on the Auth0 platform. Written in JavaScript, they are closed source; customers cannot change the code.

## Sample Application for Integration
To show you how to write an integration with a simple example, I chose to make an application that will send you an email when the user signs up.

We will use auth0 actions and amazon ses as email service to trigger this process. But you can use this flow for your own app integration.

**Prerequisites:**
- Auth0 account.
- AWS account


## Step 1: Create an Action
Log in your Auth0 account,

- Open Actions/Library from sidebar menu.
- Click Build Custom button on right corner.
- Select Post User Registration.
- Create action called ses.

![Create-Action](https://user-images.githubusercontent.com/34595361/213844594-6b702ce2-a338-4ecc-8b71-5a765cede5e1.png)

## Step 2: Add a Dependency

This is the field where we specify which npm packages the integration will depends on. We will use the node-ses package for our integration.

Select the action we created on the previous step. On dependency tab(the last icon) of the console,

- Click the button Add Dependency
- Enter name as node-ses
- Enter latest version of our node.js package (Latest version is 3.0.3**)

![Add a Dependency](https://user-images.githubusercontent.com/34595361/213844595-55b49959-6370-4000-a7d6-563214ce0d40.png)

## Step 3: Add Secrets

Secrets allow you to safely define secret values ​​as properties of the event.secrets object. We will be able to securely access the secrets we will define from within the code we will build.

Go to the Secrets tab (above the dependency icon). We will define aws key and secret. You can generate these keys on AWS IAM dashboard.

Since it is not the subject of this article, I will use the keys I created before.

### Define AWS SES Key

- Click the button Add Secret
- Enter key as AWS_SES_KEY
- Enter your AWS key to the value field.

![Add a Dependency](https://user-images.githubusercontent.com/34595361/213844596-334e99e7-ffde-4edd-8bcd-781b1ea63b89.png)

### Define AWS Secret

- Click the button Add Secret
- Enter key as AWS_SES_SECRET
- Enter your AWS secret to value field.

![Add a Dependency](https://user-images.githubusercontent.com/34595361/213844597-144787b0-6a4c-4e4c-a76a-fbddde4512e2.png)

## Step 4: Build your Custom Logic

In our case, it contains the onExecutePostUserRegistration function. Inside this function, we will write the code to run after the user registers a new account.

First of all, we call our package that we added as dependency here.

```js
/**
* Handler that will be called during the execution of a PostUserRegistration flow.
*
* @param {Event} event - Details about the context and user that has registered.
*/
exports.onExecutePostUserRegistration = async (event) => {
  const ses = require('node-ses');
};
```

then let's call the secrets we created with event.secrets and set them into the initial function of the client we use.

We should call the mail sending function in the node-ses library and write the information about which email (we need to verify on aws ses) will be sent to which email when the user is registered.

You can fill in here with your own information.

```js
/**
* Handler that will be called during the execution of a PostUserRegistration flow.
*
* @param {Event} event - Details about the context and user that has registered.
*/
exports.onExecutePostUserRegistration = async (event) => {
  const ses = require('node-ses');
  const client = ses.createClient({key: event.secrets.AWS_SES_KEY, secret: event.secrets.AWS_SES_SECRET});

  // Give SES the details and let it construct the message for you.
  client.sendEmail({
    to:   'tolga@example.com',
    from: 'info@example.com',
    subject: 'New user has registered in your app',
    message: `Hi, ${event.user.username} (${event.user.email}) registered your app.`,
    altText: 'plain text'
  });
};
```

if you're not ready to finalize everything in your Action, you can save it as a draft to work on later. Lastly, there's Deploy, which you use when you've tested your Action and it's ready for prime time!

## Test your Actions Integration

You'll see a "play button" icon for testing your Action before deploying it on the left.

![Test your Actions Integration](https://user-images.githubusercontent.com/34595361/213844598-1ce6bb9a-fc20-42e2-9a66-1a8459b4fb60.png)

After clicking, you will see the object created by Auth0 for testing and the run button under it. Your code will run when you press the Run button.

### Result

![Test Result](https://user-images.githubusercontent.com/34595361/213844600-5817c2db-1403-4aef-86d8-454f2a562611.png)

Now that we’re satisfied that the Action is working as expected, it’s time to deploy it. Select Deploy. Deploying an Action takes a snapshot of the Action at that time and records it as an Action Version.

## Submit your Actions Integration

Submit your Actions Integration to Auth0 for evaluation once you've created and properly tested it: Submit a service request on the Auth0 Marketplace.

You can fill in the required fields and send a request via this link.

https://autheco.atlassian.net/servicedesk/customer/portal/1/group/1/create/43

After an initial review of the request, Auth0 sends you a link to a GitHub repository with instructions on how to document, test, and submit your Actions Integration.

## Publish your Actions Integration

Auth0 will send you a preview of your action after reviewing your Actions Integration and you can confirm when to post it on the Auth0 Marketplace.

If you have any questions or doubts, feel free to ask them. For more content like this, join our newsletter.












