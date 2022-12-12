---
sidebar_position: 4
---

# Testing & Validation

Testing is critical process when building and maintaining an authorization system. This page explains how to ensure the new authorization model and related authorization data works as expected in Permify.

Assuming that you're familiar with creating an authorization model and forming relation tuples in Permify. If not, we're strongly advising you to examine them before testing.

We provide a GitHub action repository called [permify-validate-action] for testing and validation. This repository runs the permify validate command on the created schema validation yaml file that consists of schema (authorization model) and relationships (sample authorization data) and assertions (sample check queries and results).

:::info 
If you don't know how to create Github action workflow and add a action to it, you can examine [related page](https://docs.github.com/en/actions/quickstart) on Github docs.
:::

## Usage 

### Adding action to your workflow

After adding [permify-validate-action] to your Github Action workflow, you need to define the schema validation yaml file as, 

- **With local file:**
```yaml
steps:
- uses: "permify/permify-validate-action@v1"
  with:
    validationFile: "test.yaml"
```

- **With external url:**
```yaml
steps:
- uses: "permify/permify-validate-action@v1"
  with:
    validationFile: "https://gist.github.com/permify-bot/bb8f95acb64525d2a41688ae0a6f4274"
```

:::info 
If you don't know how to create Github action workflow and add a action to it, you can examine [quickstart page](https://docs.github.com/en/actions/quickstart) on Github docs.
:::

### Creating Schema Validation File 

Below you can examine an example schema validation yaml file. It consists 3 parts; `schema` which is your new authorization model, sample `relationships` to test your model and lastly `assertions` which is the test access check questions and expected response - true or false.

```yaml
schema: >-
    entity user {}

    entity organization {

        relation admin @user
        relation member @user

        action create_repository = (admin or member)
        action delete = admin
    }

    entity repository {

        relation owner @user @organization#member
        relation parent @organization

        action push = owner
        action read = (owner and (parent.admin and parent.member))
        action delete = (parent.member and (parent.admin or owner))
        action edit = parent.member and not owner
    }

relationships:
    - "organization:1#admin@user:1"
    - "organization:1#member@user:1"
    - "repository:1#owner@user:1"
    - "repository:2#owner@user:2"
    - "repository:2#owner@user:3"
    - "repository:1#parent@organization:1#..."
    - "organization:1#member@user:43"
    - "repository:1#owner@user:43"

assertions:
    - "can user:1 push repository:1": true
    - "can user:1 push repository:2": false
    - "can user:1 push repository:3": false
    - "can user:43 edit repository:1": false
```

You can also test your new authorization model in your local (Permify clone) without using [permify-validate-action] at all. 

For that open up a new file and add a schema yaml file inside. Then build your project with, run `make run` command and run `./permify validate {path of your schema validation file}`. 

If we use the above example schema validation file, after running `./permify validate {path of your schema validation file}` it gives a result on the terminal as:

![schema-validation](https://user-images.githubusercontent.com/34595361/207110538-9837b09d-26b4-409a-a309-a21f66e6677a.png)

## Need any help ?

Our team is happy to help you get started with Permify. If you'd like to learn more about using Permify in your app or have any questions about it, [schedule a call with one of our Permify engineer](https://meetings-eu1.hubspot.com/ege-aytin/call-with-an-expert).




