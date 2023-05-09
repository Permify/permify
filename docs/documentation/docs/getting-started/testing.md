---
sidebar_position: 4
---

# Testing & Validation

Testing is critical process when building and maintaining an authorization system. This page explains how to ensure the new authorization model and related authorization data works as expected in Permify.

Assuming that you're familiar with creating an authorization model and forming relation tuples in Permify. If not, we're strongly advising you to examine them before testing.

We provide a GitHub action repository called [permify-validate-action] for testing and validation. This repository runs the Permify validate command on the created schema validation yaml file that consists of schema (authorization model) and relationships (sample authorization data) and assertions (sample check queries and results).

:::info 
If you don't know how to create Github action workflow and add a action to it, you can examine [related page](https://docs.github.com/en/actions/quickstart) on Github docs.
:::

## Usage 

### Adding action to your workflow

After adding [permify-validate-action] to your Github Action workflow, you need to define the schema validation yaml file as, 

- **With local file:**
```yaml
steps:
- uses: "permify/permify-validate-action@v1.0.0"
  with:
    validationFile: "test.yaml"
```

- **With external url:**
```yaml
steps:
- uses: "permify/permify-validate-action@v1.0.0"
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

scenarios:
  - name: "scenario 1"
    description: "test description"
    checks:
      - entity: "repository:1"
        subject: "user:1"
        assertions:
          push : true
          owner : true
      - entity: "repository:2"
        subject: "user:1"
        assertions:
          push : false
      - entity: "repository:3"
        subject: "user:1"
        assertions:
          push : false
      - entity: "repository:1"
        subject: "user:43"
        assertions:
          edit : false
    entity_filters:
      - entity_type: "repository"
        subject: "user:1"
        assertions:
          push : ["1"]
          edit : []

```

## Coverage Analysis

By using the command `permify coverage {path of your schema validation file}`, you can measure the coverage for your schema. 
The coverage is calculated by analyzing the relationships and assertions in your created model, identifying any missing elements. 
The output of the example provided above is as follows.

![schema-coverage](https://user-images.githubusercontent.com/39353278/236303688-15cc2673-05e6-42d3-9ad4-0c538f546fb0.png)


## Testing in Local

You can also test your new authorization model in your local (Permify clone) without using [permify-validate-action] at all. 

For that open up a new file and add a schema yaml file inside. Then build your project with, run `make run` command and run `./permify validate {path of your schema validation file}`. 

If we use the above example schema validation file, after running `./permify validate {path of your schema validation file}` it gives a result on the terminal as:

![schema-validation](https://user-images.githubusercontent.com/39353278/236303542-930de83f-ebdd-4b0a-a09e-5c069744cc5c.png)

[permify-validate-action]: https://github.com/Permify/permify-validate-action

## Need any help ?

Our team is happy to help you get started with Permify. If you'd like to learn more about using Permify in your app or have any questions about it, [schedule a call with one of our Permify engineer](https://meetings-eu1.hubspot.com/ege-aytin/call-with-an-expert).




