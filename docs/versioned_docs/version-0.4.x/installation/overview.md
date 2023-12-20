---
sidebar_position: 1
---

# Guide

This guide shows you how to set up Permify in your servers and use it across your applications.

:::info Minimum Requirements
PostgreSQL: Version 13.8 or higher
:::

Please ensure your system meets these requirements before proceeding with the following steps:

1. [Set Up & Run Permify Service](#set-up-permify-service)
2. [Model your Authorization with Permify's DSL, Permify Schema](#model-your-authorization-with-permify-schema)
3. [Manage and Store Authorization Data as Relational Tuples](#store-authorization-data-as-relational-tuples)
4. [Perform Access Check](#perform-access-check)

:::info Talk to an Permify Engineer
Want to walk through this guide 1x1 rather than docs ? [schedule a call with an Permify engineer](https://meetings-eu1.hubspot.com/ege-aytin/call-with-an-expert).
:::

## Set Up Permify Service

You can run Permify Service with various options but in that tutorial we'll run it via docker container. 

### Run From Docker Container

Production usage of Permify needs some configurations such as defining running options, selecting datastore to store authorization data and more. 

However, for the sake of this tutorial we'll not do any configurations and quickly start Permify on your local with running the docker command below:

```shell
docker run -p 3476:3476 -p 3478:3478  ghcr.io/permify/permify serve
```

This will start Permify with the default configuration options: 
* Port 3476 is used to serve the REST API.
* Port 3478 is used to serve the GRPC Service.
* Authorization data stored in memory.

:::info
You can examine [Deploy using Docker] section to get more about the configuration options and learn the full integration to run Permify Service from docker container.

[Deploy using Docker]: ../container
:::

### Test your connection

You can test your connection with creating an HTTP GET request,

```shell
localhost:3476/healthz
```

You can use our Postman Collection to work with the API. Also see the [Using the API] section for details of core endpoints.

[Using the API]: ../../api-overview

[![Run in Postman](https://run.pstmn.io/button.svg)](https://www.postman.com/permify-dev/workspace/permify/collection)
[![View in Swagger](http://jessemillar.github.io/view-in-swagger-button/button.svg)](https://permify.github.io/permify-swagger/)

## Model your Authorization with Permify Schema

After installation completed and Permify server is running, next step is modeling authorization with Permify authorization language - [Permify Schema]-  and configure it to Permify API. 

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
[example use cases]: ../../use-cases/simple-rbac
:::

### Configuring Schema via API 

After modeling completed, you need to send Permify Schema - authorization model - to [Write Schema API](../api-overview/schema/write-schema.md) for configuration of your authorization model on Permify authorization service.

:::caution Before Continue on Writing Schema
You'll see **tenant_id** parameter almost all Permify APIs including Write Schema. With version 0.3.x Permify became a tenancy based authorization infrastructure, and supports multi-tenancy by default so its a mandatory parameter when doing any operations.

We provide a pre-inserted tenant - **t1** - for ones that don't need/want to use multi-tenancy. So, we will be passing **t1** to all tenant id parameters throughout this guidance. <!-- For more details about Multi Tenancy usage and structure of Permify see [Multi Tenancy Section](./aws.md). -->
:::

#### Example HTTP Request on Postman: 

| Required | Argument | Type | Default | Description |
|----------|-------------------|--------|---------|-------------|
| [x]   | tenant_id | string | - | identifier of the tenant, if you are not using multi-tenancy (have only one tenant) use pre-inserted tenant `t1` for this field.
| [x]   | schema | string | - | Permify Schema as string|

**POST /v1/tenants/{tenant_id}/schemas/write**

![permify-schema](https://user-images.githubusercontent.com/34595361/214457054-19b141ac-6bfa-4db4-aeab-f7b7149c3351.png)

## Store Authorization Data as Relational Tuples

After you completed configuration of your authorization model via Permify Schema. Its time to add authorizations data to see Permify in action. 

### Create Relational Tuples

You can create relational tuples as authorization rules at this writeDB by using [Write Relationships API](../api-overview/relationship/write-relationships.md)

For our guide let's grant one of the team members (Ashley) an admin role.

#### Example HTTP Request on Postman: 

| Required | Argument | Type | Default | Description |
|----------|-------------------|--------|---------|-------------|
| [x]   | tenant_id | string | - | identifier of the tenant, if you are not using multi-tenancy (have only one tenant in your system) use pre-inserted tenant **t1** for this field.
| [x]   | tuples | array | - | Can contain multiple relation tuple object|
| [x]   | entity | object | - | Type and id of the entity. Example: "organization:1”|
| [x]   | relation | string | - | Custom relation name. Eg. admin, manager, viewer etc.|
| [x]   | subject | string | - | User or user set who wants to take the action. |
| [ ]   | schema_version | string | 8 | Version of the schema |

**POST /v1/tenants/{tenant_id}relationships/write** 

```json
{
    "metadata": {
        "schema_version": ""
    },
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

![write-relationships](https://user-images.githubusercontent.com/34595361/214458203-8264e141-642d-48b0-9242-416bbf6f8795.png)

**Created relational tuple:** organization:1#admin@user:1

**Semantics:** User 1 (Ashley) has admin role on organization 1.

:::tip
In ideal production usage Permify stores your authorization data in a database you prefer. We called that database as WriteDB, and you can configure it with using [configuration yaml file](https://github.com/Permify/permify/blob/master/example.config.yaml) or CLI flag options. 

But in this tutorial Permify Service running default configurations on local, so authorization data will be stored in memory. You can find more detailed explanation how Permify stores authorization data in [Managing Authorization Data] section.

[Managing Authorization Data]: ../../getting-started/sync-data
:::

## Perform Access Check

Finally we're ready to control authorization. Access decision results computed according to relational tuples and the stored model, [Permify Schema] action conditions.

Lets get back to our example and perform an example access check via [Check API]. We want to check whether an specific user has an access to view files in a organization.

[Check API]: ../../api-overview/permission/check-api
[Permify Schema]: ../../getting-started/modeling

#### Example HTTP Request: 

***Can the user 45 view files on organization 1 ?***

**POST /v1/tenants/{tenant_id}/permissions/check**

| Required | Argument       | Type     | Default | Description                                                                                                                                       |
|----------|----------------|----------|---------|---------------------------------------------------------------------------------------------------------------------------------------------------|
| [x]      | tenant_id      | string   | -       | identifier of the tenant, if you are not using multi-tenancy (have only one tenant in your system) use pre-inserted tenant **t1** for this field. |
| [x]      | entity         | object   | -       | name and id of the entity. Example: organization:1.                                                                                               | 
| [x]      | action         | string   | -       | the action the user wants to perform on the resource                                                                                              |
| [x]      | subject        | object   | -       | the user or user set who wants to take the action                                                                                                 |
| [ ]      | schema_version | string   | -       | get results according to given schema version                                                                                                     |
| [ ]      | depth          | integer  | 8       | -                                                                                                                                                 |

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

[Access Control Check]: ../api-overview/permission/check-api.md

## Need any help ?

Our team is happy to help you get started with Permify. If you struggle with installation or have any questions, [schedule a call with one of our Permify engineers](https://meetings-eu1.hubspot.com/ege-aytin/call-with-an-expert). Alternatively you can join our [discord community](https://discord.com/invite/MJbUjwskdH) to discuss.