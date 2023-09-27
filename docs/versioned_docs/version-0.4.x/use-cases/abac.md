# Attribute Based Access Control (Beta)

This page explains design approach of Permify ABAC support as well as demonstrates how to create and use attribute based permissions in Permify.

:::info  
You can find Permify's support for ABAC in our [beta release](https://github.com/Permify/permify/pkgs/container/permify-beta) and explore the active [API documentation](https://permify.github.io/permify-swagger/) for the ***beta*** version. 

We are eager to hear your thoughts and looking forward to your feedback!
:::

# What is Attribute Based Access Control (ABAC)?

Attribute-Based Access Control (ABAC) is like a security guard that decides who gets to access what based on specific characteristics or "attributes".

These attributes can be associated with users, resources, or the environment, and their values can influence the outcome of an access request.

Let’s make an analogy, it’s the best of way to understanding complex ideas.

Think about an amusement park, and there are 3 different rides. In order to access each ride, you need to obtain different qualities. For the;

1. first ride you need to be over 6ft tall.
2. second ride you need to be under 200lbs.
3. third ride you need to be between 12 - 18 years old.

Similar to this ABAC check certain qualities that you defined on users, resources, or the environment.

# Why Would Need ABAC?

It’s obvious but simple answer is “use cases”… Sometimes, using ReBAC and RBAC isn't the best fit for the job. It's like using winter tires on a hot desert road, or summer tires in a snowstorm - they're just not the right tools for the conditions.

1. **Geographically Restricted:** Think of ABAC like a bouncer at a club who only lets in people from certain towns. For example, a movie streaming service might only show certain movies in certain countries because of rules about who can watch what and where.
2. **Time-Based:** ABAC can also act like a parent setting rules about when you can use the computer. For example, a system might only let you do certain things during office hours.
3. **Compliance with Privacy Regulations:** ABAC can help follow rules about privacy. For example, a hospital system might need to limit who can see a patient's data based on the patient's permission, why they want to see it, and who the person is.
4. **Limit Range:** ABAC can help you create a rules defining a number limit or range. For instance, a banking system might have limits for wiring or withdrawing money.
5. **Device Information:** ABAC can control access based on attributes of the device, such as the device type, operating system version, or whether the device has the latest security patches.

As you can see ABAC has more contextual approach. You can define access rights regarding context around subject and object in an application.

# Introducing New Key Elements

To support ABAC in Permify, we've added two main components into our DSL: attributes and rules.

## Attribute

Attributes are used to define properties for entities in specific data types. For instance, an attribute could be an IP range associated with an organization, defined as a string array:

```sql
attribute ip_range string[]
```

There are different types of attributes you can use;

### Boolean

For attributes that represent a binary choice or state, such as a yes/no question, the `Boolean` data type is an excellent choice.

```go
entity post {
        attribute is_public boolean
        
        permission view = is_public
}
```

<aside>
⛔ If you don’t create the related attribute data, Permify accounts boolean as `FALSE`

</aside>

### String

String can be used as attribute data type in a variety of scenarios where text-based information is needed to make access control decisions. Here are a few examples:

- **Location:** If you need to control access based on geographical location, you might have a location attribute (e.g., "USA", "EU", "Asia") stored as a string.
- **Device Type**: If access control decisions need to consider the type of device being used, a device type attribute (e.g., "mobile", "desktop", "tablet") could be stored as a string.
- **Time Zone**: If access needs to be controlled based on time zones, a time zone attribute (e.g., "EST", "PST", "GMT") could be stored as a string.
- **Day of the Week:** In a scenario where access to certain resources is determined by the day of the week, the string data type can be used to represent these days (e.g., "Monday", "Tuesday", etc.) as attributes!

```sql
entity user {}

entity organization {
    
    relation admin @user

    attribute location string[]

    permission view = check_location(request.current_location, location) or admin
}

rule check_location(current_location string, location string[]) {
    current_location in location
}
```

<aside>
⛔ If you don’t create the related attribute data, Permify accounts string as `""`

</aside>

### Integer

Integer  can be used as attribute data type in several scenarios where numerical information is needed to make access control decisions. Here are a few examples:

- **Age:** If access to certain resources is age-restricted, an age attribute stored as an integer can be used to control access.
- **Security Clearance Level:** In a system where users have different security clearance levels, these levels can be stored as integer attributes (e.g., 1, 2, 3 with 3 being the highest clearance).
- **Resource Size or Length:** If access to resources is controlled based on their size or length (like a document's length or a file's size), these can be stored as integer attributes.
- **Version Number:** If access control decisions need to consider the version number of a resource (like a software version or a document revision), these can be stored as integer attributes.

```jsx
entity content {
    permission view = check_age(request.age)
}

rule check_age(age integer) {
        age >= 18
}
```

<aside>
⛔ If you don’t create the related attribute data, Permify accounts integer as `0`

</aside>

### Double

Double can be used as attribute data type in several scenarios where precise numerical information is needed to make access control decisions. Here are a few examples:

- **Usage Limit:** If a user has a usage limit (like the amount of storage they can use or the amount of data they can download), and this limit needs to be represented with decimal precision, it can be stored as a double attribute.
- **Transaction Amount:** In a financial system, if access control decisions need to consider the amount of a transaction, and this amount needs to be represented with decimal precision (like $100.50), these amounts can be stored as double attributes.
- **User Rating:** If access control decisions need to consider a user's rating (like a rating out of 5 with decimal points, such as 4.7), these ratings can be stored as double attributes.
- **Geolocation:** If access control decisions need to consider precise geographical coordinates (like latitude and longitude, which are often represented with decimal points), these coordinates can be stored as double attributes.

```sql
entity user {}

entity account {
    relation owner @user
    attribute balance double

    permission withdraw = check_balance(request.amount, balance) and owner
}

rule check_balance(amount double, balance double) {
    (balance >= amount) && (amount <= 5000)
}
```

<aside>
⛔ If you don’t create the related attribute data, Permify accounts double as `0.0`

</aside>

## Rule

Rules are structures that allow you to write specific conditions for the model. They accept parameters and are based on conditions. For example, a rule could be used to check if a given IP address falls within a specified IP range:

```sql
rule check_ip_range(ip string, ip_range string[]) {
    ip in ip_range
}
```

## Evaluation

**Model**

```sql
entity user {}

entity organization {
    
    relation admin @user

    attribute ip_range string[]

    permission view = check_ip_range(request.ip_address, ip_range) or admin
}

rule check_ip_range(ip_address string, ip_range string[]) {
    ip in ip_range
}
```

In this case, the part written as 'context' refers to the context within the request. Any type of data can be added from within the request and can be called within model.

For instance,

```sql
...
"context": {
        "ip_address": "187.182.51.206",
        "day_of_week": "monday"
}
...
```

**Relationships**

- organization:1#admin@user:1

**Attributes**

- organization:1$ip_range|string[]:[‘187.182.51.206’, ‘250.89.38.115’]

**Check request**

```sql
{
    "entity": {
        "type": "organization",
        "id": "1"
    },
    "permission": "view",
    "subject" : {
        "type": "user",
        "id": "1"
    },
    "context": {
        "ip_address": "187.182.51.206"
    }
}
```

**Check Evolution Sub Queries Organization View**
→ organization:1$check_ip_range(context.ip_address,ip_range) → true
→ organization:1#admin@user:1 → true

**Cache Mechanism**
The cache mechanism works by hashing the snapshot of the database, schema version, and sub-queries as keys and adding their results, so it will operate in the same way in calls as in relationships. For example,

**Request keys before hash**

- check_{snapshot}_{schema_version}_{context}_organization:1#admin@user:1 → true
- check_{snapshot}_{schema_version}_{context}_organization:1$check_ip_range(ip_range) → true

## Some Use Cases

### Example of Public/Private Repository

In this example, **`is_public`** is defined as a boolean attribute. If an attribute is boolean, it can be directly written without the need for a rule. This is only applicable for boolean types.

```sql
entity user {}
        
entity post {

  relation owner  @user

    attribute is_public boolean

    permission view   = is_public or owner
  permission edit   = owner
}
```

In this context, if the **`is_public`** attribute of the repository is set to true, everyone can view it. If it's not public (i.e., **`is_public`** is false), only the owner, in this case **`user:1`**, can view it.

The permissions in this model are defined as such:

**`permission view = is_public or owner`**

This means that the 'view' permission is granted if either the repository is public (**`is_public`** is true) or if the current user is the owner of the repository.

**relationships:**

- post:1#owner@user:1

**attributes:**

- post:1$is_public|boolean:true

**Check Evolution Sub Queries Post View**
→ post:1#is_public → true
→ post:1#admin@user:1 → true

**Request keys before hash**

- check_{snapshot}_{schema_version}_{context}_post:1$is_public → true
- check_{snapshot}_{schema_version}_{context}_post:1#admin@user:1 → true

### Example of Weekday

In this example, to be able to view the repository, it must not be a weekend, and the user must be a member of the organization.

```sql
entity user {}

entity organization {

    relation member @user

    permission view = is_weekday(request.day_of_week) and member
}

entity repository {

    relation organization  @organization

    permission view = organization.view
}

rule is_weekday(day_of_week string) {
      day_of_week != 'saturday' && day_of_week != 'sunday'
}
```

The permissions in this model state that to 'view' the repository, the user must fulfill two conditions: the current day (according to the context data **`day_of_week`**) must not be a weekend (determined by the **`is_weekday`** rule), and the user must be a member of the organization that owns the repository.

**Relationships:**

- organization:1#member@user:1

**Check Evolution Sub Queries Organization View**
→ organization:1$is_weekday(context.day_of_week) → true
→ organization:1#member@user:1 → true

**Request keys before hash**

- check_{snapshot}_{schema_version}_{context}_organization:1$is_weekday(context.day_of_week) → true
- check_{snapshot}_{schema_version}_{context}_post:1#member@user:1 → true

### Example of Banking System

This model represents a banking system with two entities: **`user`** and **`account`**.

1. **`user`**: Represents a customer of the bank.
2. **`account`**: Represents a bank account that has an **`owner`** (which is a **`user`**), and a **`balance`** (amount of money in the account).

```sql
entity user {}

entity account {
    relation owner @user
    attribute balance double

    permission withdraw = check_balance(request.amount, balance) and owner
}

rule check_balance(amount double, balance double) {
    (balance >= amount) && (amount <= 5000)
}
```

**The check_balance rule:** This rule verifies if the withdrawal amount is less than or equal to the account's balance and doesn't exceed 5000 (the maximum amount allowed for a withdrawal). It accepts two parameters, the withdrawal amount (amount) and the account's current balance (balance).
**The owner check:** This condition checks if the person requesting the withdrawal is the owner of the account.

Both of these conditions need to be true for the **`withdraw`** permission to be granted. In other words, a user can withdraw money from an account only if they are the owner of that account, and the amount they want to withdraw is within the account balance and doesn't exceed 5000.

**Relationships**

- account:1#owner@user:1

**Attributes**

- account:1$balance|double:4000

**Check Evolution Sub Queries For Account Withdraw**
→ account:1$check_balance(context.amount,balance) → true
→ account:1#owner@user:1 → true

**Request keys before hash**

- check_{snapshot}_{schema_version}_{context}_account:1$check_balance(context.amount,balance) → true
- check_{snapshot}_{schema_version}_{context}_account:1#owner@user:1 → true

### Hierarchical Usage

In this model:

1. **`employee`**: Represents an individual worker. It has no specific attributes or relations in this case.
2. **`organization`**: Represents an entire organization, which has a **`founding_year`** attribute. The **`view`** permission is granted if the **`check_founding_year`** rule (which checks if the organization was founded after 2000) returns true.
3. **`department`**: Represents a department within the organization. It has a **`budget`** attribute and a relation to its parent **`organization`**. The **`view`** permission is granted if the department's budget is more than 10,000 (checked by the **`check_budget`** rule) and if the **`organization.view`** permission is true.

Note: In this model, permissions can refer to higher-level permissions (like **`organization.view`**). However, you cannot use the attribute of a relation in this way. For example, you cannot directly reference **`organization.founding_year`** in a permission expression. Permissions can depend on permissions in a related entity, but not directly on the related entity's attributes.

```sql
entity employee {}

entity organization {
    attribute founding_year integer

    permission view = check_founding_year(founding_year)
}

entity department {
    relation organization @organization
    attribute budget double

    permission view = check_budget(budget) and organization.view
}

rule check_founding_year(founding_year integer) {
        founding_year > 2000
}

rule check_budget(budget double) {
        budget > 10000
}
```

**Relationships**

- department:1#organization@organization:1
- department:1#organization@organization:2

**Attributes**

- department:1$budget|double:20000
- organization:1$organization|integer:2021

**Check Evolution Sub Queries For Department View**
→ department:1$check_budget(budget) → true
→ department:1#organization@user:1 → true
    → organization:2$check_founding_year(founding_year) → false
    → organization:1$check_founding_year(founding_year) → true

**Request keys before hash**

- check_{snapshot}_{schema_version}_{context}_department:1$check_budget(budget) → true
- check_{snapshot}_{schema_version}_{context}_organization:2$check_founding_year(founding_year) → false
- check_{snapshot}_{schema_version}_{context}_organization:1$check_founding_year(founding_year) → true

## How To Use Demo

**Install Permify nightly release**

```yaml
docker pull **ghcr.io/permify/permify-beta:latest**
```

**New Validation Yaml Structure**

```yaml
schema: >-
    {string schem}

relationships:
    - entity_name:entity_id#relation@subject_type:subject_id

attributes:
    - entity_name:entity_id#attribute@attribute_type:attribute_value

scenarios:
  - name: "name"
    description: "description"
    checks:
            - entity: "entity_name:entity_id"
        subject: "subject_name:subject_id"
        context:
          tuples: []
          attributes: []
          data:
            key: {value}
        assertions:
          permission: result
    entity_filters:
            - entity_type: "entity_name"
        subject: "subject_name:subject_id"
        context:
          tuples: []
          attributes: []
          data:
            key: {value}
        assertions:
          permission: result_array
    subject_filters:
            - subject_reference: "subject_name"
        entity: "entity_name:entity_id"
        context:
          tuples: []
          attributes: []
          data:
            key: {value}
        assertions:
          permission: result_array
```

**Note:** The 'data' field within the 'context' can be assigned a desired value as a key-value pair. Later, this value can be retrieved within the model using 'request.key'.

**Example in validation file:** 

```yaml
context:
    tuples: []
    attributes: []
    data:
        day_of_week: "saturday"
```

This YAML snippet specifies a validation context with no tuples or attributes, and a data field indicating the day of the week is Saturday.

**Example in model**

```yaml
permission delete = is_weekday(request.day_of_week)
```

In the model, a **`delete`** permission rule is set. It calls the function **`is_weekday`** with the value of **`day_of_week`** from the context. If **`is_weekday("saturday")`** is true, the delete permission is granted.

**Create Validation File**

```yaml
schema: >-
    entity user {}

    entity organization {

        relation member @user

        attribute credit integer

        permission view = check_credit(credit) and member
    }

    entity repository {

        relation organization  @organization

        attribute is_public boolean

        permission view = is_public
        permission edit = organization.view
        permission delete = is_weekday(request.day_of_week)
    }

    rule check_credit(credit integer) {
        credit > 5000
    }

    rule is_weekday(day_of_week string) {
          day_of_week != 'saturday' && day_of_week != 'sunday'
    }

relationships:
  - organization:1#member@user:1
  - repository:1#organization@organization:1

attributes:
  - organization:1$credit|integer:6000
  - repository:1$is_public|boolean:true

scenarios:
  - name: "scenario 1"
    description: "test description"
    checks:
      - entity: "repository:1"
        subject: "user:1"
        context:
        assertions:
          view: true
      - entity: "repository:1"
        subject: "user:1"
        context:
          tuples: []
          attributes: []
          data:
            day_of_week: "saturday"
        assertions:
          view: true
          delete: false
      - entity: "organization:1"
        subject: "user:1"
        context:
        assertions:
          view: true
    entity_filters:
      - entity_type: "repository"
        subject: "user:1"
        context:
        assertions:
          view : ["1"]
    subject_filters:
      - subject_reference: "user"
        entity: "repository:1"
        context:
        assertions:
          view : ["1"]
          edit : ["1"]
```

**Run validation command**

```yaml
docker run -v {your_config_folder}:/config **ghcr.io/permify/permify-beta:latest validate /config/validation.yaml**
```

## Need any help ?

Our team is happy to help you get started with Permify. If you'd like to learn more about using Permify in your app or have any questions about this example, [schedule a call with one of our Permify engineer](https://meetings-eu1.hubspot.com/ege-aytin/call-with-an-expert).