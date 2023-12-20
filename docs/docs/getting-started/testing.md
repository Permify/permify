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

## Adding Validate Action To Your Workflow

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

## Schema Validation File 

Below you can examine an example schema validation yaml file. It consists 3 parts; 
- `schema` which is the authorization model you want to test,
- `relationships` sample data to test your model,
- `scenarios` to test access check queries within created scenarios.

### Defining the Schema:

You can define the `schema` in the YAML file in one of two ways:

1. **Directly in the File:** Define the schema directly within the YAML file.

   ```yaml
   schema: >-
     entity user {}
     entity organization {
       ...
     }
   
2. **Via URL or File Path:** Specify a URL or a file path to an external schema file.
   **Example with URL:**

    ```yaml
    schema: https://example.com/path/to/schema.txt
    ```

    **Example with File Path:**
    ```yaml
    schema: /path/to/your/schema/file.txt
    ```

Here is an example Schema Validation file, 

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
      action edit = parent.member not owner
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
        context:
          - "repository:3#owner@user:1"
        assertions:
          push : true
      - entity: "repository:1"
        subject: "user:43"
        assertions:
          edit : false
    entity_filters:
      - entity_type: "repository"
        subject: "user:1"
        context:
          - "repository:3#owner@user:1"
          - "repository:4#owner@user:1"
          - "repository:5#owner@user:1"
        assertions:
          push : ["1", "3", "4", "5"]
          edit : []
    subject_filters:
      - subject_reference: "user"
        entity: "repository:1"
        context:
          - "organization:1#member@user:58"
        assertions:
          push : ["1", "43"]
          edit : ["58"]
```

Assuming that you're well-familiar with the `schema` and `relationships` sections of the above YAML file. If not, please see the previous sections to learn how to create an authorization model (schema) and generate data (relationships) according to it.

We'll continue by examining how to create scenarios.

## Creating Test Scenarios

You can create multiple access checks at once to test whether your authorization logic behaves as expected or not. 

Besides simple access checks you can also test subject filtering queries and data (entity) filtering with it.

Let's deconstruct the `scenarios`,

### Scenarios

```js
scenarios:
  - name: // name of the scenario
    description: // description of the scenario
    checks: // simple access check case/cases
    entity_filters: // entity (data) filtering query/queries
    subject_filters: // subject filtering query/queries
```

### Access Check

You can create `check` inside `scenarios` to test multiple access check cases,

```js
checks:
   - entity: "repository:3" // resource/entity that you want to check access for
     subject: "user:1" // subject that performs the access check
     context: // additional data provided during an access check to be evaluated
       - "repository:3#owner@user:1" 
     assertions: // expected result/results for specific action/s or an permission/s.
       push : true
```

Semantics for above check is: whether `user:1` can push to `repository:3`, additional to stored tuples take account that user:1 is owner of repository:3 (`repository:3#owner@user:1`). Expected result for that check it **true** - `push : true` 

:::info Contextual Tuples
We use `context` (Contextual Tuples) with simple relational tuples for simplicity in this example. However, it is primarily used for dynamic access checks, such as those involving time, date, or IP address, etc. 

To learn more about how `context` works, see the [Contextual Tuples](../../reference/contextual-tuples) section.
:::

### Entity Filtering

You can create `entity_filters` within `scenarios` to test your data filtering queries.

```js
entity_filters:
      - entity_type: "repository" // entity that you want to filter 
        subject: "user:1" // subject that you want to perform data filtering 
        context: null // additional data provided during an access check to be evaluated
        assertions: 
          push : ["1", "3", "4", "5"] // IDs of the resources that we expected to return
          edit : []
```

The major difference between `check` lies in the assertions part. Since we're performing data filtering with bulk data, instead of a true-false result, we enter the IDs of the resources that we expect to be returned

### Subject Filtering

You can create `subject_filters` within `scenarios` to test your subject filtering queries, a.k.a which users can perform action Y or have permission X on entity:Z?

```js
- subject_reference: "user"
        entity: "repository:1"
        context: null // additional data provided during an access check to be evaluated
        assertions:
          push : ["1", "43"] // IDs of the users that we expected to return
          edit : ["58"]
```

:::info API Endpoints
You can find the related API endpoints for `check`, `entity_filters`, and `subject_filters` in the Permission service in the [Using The API](../../api-overview) section.
:::

## Coverage Analysis

By using the command `permify coverage {path of your schema validation file}`, you can measure the coverage for your schema. 

The coverage is calculated by analyzing the relationships and assertions in your created model, identifying any missing elements. 

The output of the example provided above is as follows.

![schema-coverage](https://user-images.githubusercontent.com/39353278/236303688-15cc2673-05e6-42d3-9ad4-0c538f546fb0.png)

## Testing in Local

You can also test your new authorization model in your local (Permify clone) without using [permify-validate-action] at all. 

For that open up a new file and add a schema yaml file inside. Then build your project with, run `make build` command and run `./permify validate {path of your schema validation file}`. 

If we use the above example schema validation file, after running `./permify validate {path of your schema validation file}` it gives a result on the terminal as:

![schema-validation](https://user-images.githubusercontent.com/39353278/236303542-930de83f-ebdd-4b0a-a09e-5c069744cc5c.png)

[permify-validate-action]: https://github.com/Permify/permify-validate-action

## AST Conversion

By utilizing the command `permify ast {path of your schema validation file}`, you can effortlessly convert your model into an Abstract Syntax Tree (AST) representation.

The conversion to AST provides a structured representation of your model, making it easier to navigate, modify, and analyze. This process ensures that your model is syntactically correct and can be processed by other tools without issues.

The output after running the above example command is illustrated below.


![ast-conversion](https://github.com/Permify/permify/assets/39353278/822902d7-9612-46a6-95e9-1cb09bc0ebb2)

## Unit Tests For Schema Changes

We recommend leveraging Permify's in-memory databases for a simplified and isolated testing environment. These in-memory databases can be easily created and disposed of for each individual unit test, ensuring that your tests do not interfere with each other and each one starts with a clean slate.

For managing permission/relation changes, we suggest storing schema in an abstracted place such as a git repo and centrally checking and approving every change before deploying it via the CI pipeline that utilizes the **Write Schema API**. 

We recommend adding our [schema validator](https://github.com/Permify/permify-validate-action) to the pipeline to ensure that any changes are automatically validated. 

You can find more details about our suggested workflow to handle schema changes in [Write Schema](../../api-overview/schema/write-schema#suggested-workflow-for-schema-changes) section.

## Need any help ?

Our team is happy to help you get started with Permify. If you'd like to learn more about using Permify in your app or have any questions about it, [schedule a call with one of our Permify engineer](https://meetings-eu1.hubspot.com/ege-aytin/call-with-an-expert).




