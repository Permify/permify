---
sidebar_position: 1
---

# Guide

Permify is an open-source authorization service that you can run in your environment and works as an API. This guide shows how to set up Permify in your servers and use it across your applications. Set up and implementation consists of 4 steps,

1. [Set Up & Run Permify Service](#run-permify-api)
2. [Model your Authorization with Permify's DSL, Permify Schema](#model-your-authorization-with-permify-schema)
3. [Migrate and Store Authorization Data as Relational Tuples](#store-authorization-data-as-relational-tuples)
4. [Perform Access Check](#perform-access-check)

:::info Talk to an Permify Engineer
Want to walk through this guide 1x1 rather than docs ? [schedule a call with an Permify engineer](https://meetings-eu1.hubspot.com/ege-aytin/call-with-an-expert).
:::

## Set Up Permify Service

You can run Permify Service with two options: 

- [Run From Container](#run-from-container)  
- [Install With Brew](#install-with-brew). 

### Run From Container

Installation needs some configuration such as defining running options, selecting datastore to store authorization data and more. 

However, If you want to play around with Permify without doing any configurations, you can quickly start Permify on your local with running the command below:

```shell
docker run -p 3476:3476 -p 3478:3478  ghcr.io/permify/permify serve
```

This will start Permify with the default configuration options: 
* Port 3476 is used to serve the REST API.
* Port 3478 is used to serve the GRPC Service.
* Authorization data stored in memory.

See [Container With Configurations] section to get more details about the configuration options and learn the full integration to run Permify Service from container.

[Container With Configurations]: ../container

### Install With Brew

Firstly, open terminal and run following line,

```shell
brew install permify/tap/permify
```

After the brew installation, the `serve` command should be used to run Permify. However, if you want to start Permify without doing any configurations, you can run the command without config flags as follows:

```shell
permify serve
```

This will start Permify with the default configuration options: 
* Port 3476 is used to serve the REST API.
* Port 3478 is used to serve the GRPC Service.
* Authorization data stored in memory.

You can override these configurations with running the command with configuration flags. See all configuration options with running `permify serve --help` on terminal. 

Check out the [Brew With Configurations] section to learn full implementation with configurations.

[Brew With Configurations]: ../brew

### Test your connection

You can test your connection with creating an HTTP GET request,

```shell
localhost:3476/healthz
```

You can use our Postman Collection to work with the API. Also see the [Using the API] section for details of core functions.

[Using the API]: ../../api-overview

[![Run in Postman](https://run.pstmn.io/button.svg)](https://god.gw.postman.com/run-collection/16122080-54b1e316-8105-4440-b5bf-f27a05a8b4de?action=collection%2Ffork&collection-url=entityId%3D16122080-54b1e316-8105-4440-b5bf-f27a05a8b4de%26entityType%3Dcollection%26workspaceId%3Dd3a8746c-fa57-49c0-83a5-6fcf25a7fc05)
[![View in Swagger](http://jessemillar.github.io/view-in-swagger-button/button.svg)](https://app.swaggerhub.com/apis-docs/permify/permify/latest)


## Model your Authorization with Permify Schema

After installation completed and Permify server is running, next step is modeling authorization with Permify's authorization language - [Permify Schema]-  and condigure it to Permify API. 

You can define your entities, relations between them and access control decisions of each actions with using [Permify Schema].

### Creating your authorization model

Permify Schema can be created on our [playground](https://play.permify.co/) as well as in any IDE or text editor. We also have a [VS Code extension](https://marketplace.visualstudio.com/items?itemName=Permify.perm) to ease modeling Permify Schema with code snippets and syntax highlights. Note that on VS code the file with extension is ***".perm"***.

:::caution Use Playground For Testing
If you're planning to test Permify manually, maybe with an API Design platform such as [Postman](https://www.postman.com/), [Insomnia](https://insomnia.rest/), etc; we're suggesting using our playground to create model. Because Permify Schema needs to be configured (send to API) in Permify API in a **string** format. Therefore, created model should be converted to **string**. 

Although, it could easily be done programmatically, it could be little challenging to do it manually. To help on that, we have a button on the playground to copy created model to the clipboard as a string, so you get your model in string format easily.

![copy-btn](https://user-images.githubusercontent.com/34595361/198015792-a7f0d727-a1a5-4039-b0be-d097321b8d53.png)

:::

Let's create our authorization model. We'll be using following a simple user-organization authorization case for this guide. 

```perm
entity user {} 

entity organization {

    relation admin @user    
    relation member @user     
    
    action view_files = admin or member
    action edit_files = admin

} 
```

We have 2 entities these are **"user"** and **"organization"**. Entities represents your main tables. We strongly advise naming entities the same as your original database entities. 

Lets roll back our example, 

- The `user` entity represents users. This entity is empty because it's only responsible for referencing users.

- The `organization` entity has its own relations (`admin` and `member`) which related with user entity. This entity also  has 2 actions, respectively:
  - Organization member and admin can view files.
  - Only admins can edit files.

:::info
For implementation sake we'll not dive more deep about modeling but you can find more information about modeling on [Modeling Authorization with Permify] section. Also can check out [example use cases] to better understand some basic use cases modeled with Permify Schema. 

[Modeling Authorization with Permify]: ../../getting-started/modeling
[example use cases]: ../../example-use-cases/simple-rbac
:::

### Configuring Permify Schema to API 

After modeling completed, you need to send Permify Schema - authorization model - to API endpoint **/v1/schemas/write"** for configuration of your authorization model on Permify API.

#### Path : ** POST "/v1/schemas/write"**
| Required | Argument | Type | Default | Description |
|----------|-------------------|--------|---------|-------------|
| [x]   | schema | string | - | Permify Schema as string|

**Example Request on Postman:**

![permify-schema](https://user-images.githubusercontent.com/34595361/197405641-d8197728-2080-4bc3-95cb-123e274c58ce.png)

## Store Authorization Data as Relational Tuples

After you completed configuration of your authorization model via Permify Schema. Its time to add authorizations data to see Permify in action. Permify stores your authorization data in a database you prefer. We called that database as WriteDB, and you can configure it when running Permify Service. 

:::info
If your Permify Service running default configurations, authorization data will be stored in memory. 
:::

If you set up Permify Service from container you can both configure WriteDB with using [configuration yaml file](https://github.com/Permify/permify/blob/master/example.config.yaml) and configuration flags. On the other hand, If you're using brew to install and run Permify you can only use the configuration flags.

### Create Relational Tuples

You can create relational tuples as authorization rules at this writeDB by using `/v1/relationships/write` endpoint.

For our guide let's grant one of the team members (Ashley) an admin role. 

**Request:** POST - `/v1/relationships/write` 

```json
{
    "schema_version": "",
    "tuples": [
        {
       "entity": {
        "type": "organization",
        "id": "1" //Organization identifier
        },
        "relation": "admin",
        "subject": {
            "type": "user",
            "id": "1", //Ashley's identifier
            "relation": ""
        }
    }
    ]
}
```

**Created relational tuple:** organization:1#admin@1

**Semantics:** User 1 (Ashley) has admin role on organization 1.

:::info
You can find more detailed explanation from [Move & Synchronize Authorization Data] section.

[Move & Synchronize Authorization Data]: ../../getting-started/sync-data
:::

### Performing Access Control Check

You can check authorization with
single API call. This check request returns a decision about whether user can perform an action on a certain resource.

Access decisions generated according to relational tuples, which stored in your database (writeDB) and [Permify Schema] action conditions.

[Permify Schema]: ../getting-started/modeling

## Perform Access Check

Finally we're ready to control authorization. Lets perform an example access check via [check] API. 

[check]: ../../api-overview/check-api

***Can the user 45 view files on organization 1 ?***

### Path: 

POST /v1/permissions/check

| Required | Argument | Type | Default | Description |
|----------|----------|---------|---------|-------------------------------------------------------------------------------------------|
| [x]   | entity | object | - | name and id of the entity. Example: organization:1”.
| [x]   | action | string | - | the action the user wants to perform on the resource |
| [x]   | subject | object | - | the user or user set who wants to take the action  |
| [ ]   | schema_version | string | - | get results according to given schema version|
| [ ]   | depth | integer | 8 | |

### Request

```json
{
  "metadata": {
    "schema_version": "",
    "snap_token": "",
    "depth": 20
  },
  "entity": {
    "type": "organization",
    "id": "1"
  },
  "permission": "view_files",
  "subject": {
    "type": "user",
    "id": "45",
    "relation": ""
  },
}
```

### Response

```json
{
  "can": "RESULT_ALLOW",
  "metadata": {
    "check_count": 0
  }
}
```

See [Access Control Check] section for learn how access checks works and access decisions evaluated in Permify

[Access Control Check]: ../../getting-started/enforcement

## Need any help ?

Our team is happy to help you get started with Permify. If you struggle with installation or have any questions, [schedule a call with one of our Permify engineers](https://meetings-eu1.hubspot.com/ege-aytin/call-with-an-expert). Alternatively you can join our [discord community](https://discord.com/invite/MJbUjwskdH) to discuss.