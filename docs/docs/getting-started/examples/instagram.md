# Instagram

This example presents an Instagram Authorization Schema, outlining the intricate relationships between users, accounts, and posts on the platform. It defines user access levels, privacy settings, and interactions, offering insights into how followers, account owners, and post restrictions are managed within the Instagram ecosystem.

## Schema | [Open in playground](https://play.permify.co/?s=instagram&tab=schema)

```perm
entity user {}

entity account {
    // users have accounts
    relation owner @user

    // accounts can follow other users/accounts.
    relation following @user

    // other users/accounts can follow account.
    relation follower @user

    // accounts can be private or public.
    attribute public boolean

    // users can view an account if they're followers, owners, or if the account is not private.
    action view = (owner or follower) or public

}

entity post {
    // posts are linked with accounts.
    relation account @account

    // comments are limited to people followed by the parent account.
    attribute restricted boolean

    // users can view the posts, if they have access to view the linked accounts.
    action view = account.view

    // users can comment and like on unrestricted posts or posts by owners who follow them.
    action comment = account.following not restricted
    action like = account.following not restricted
}
```

## Brief Examination of the Model

The Instagram Authorization Schema models the relationships between users, accounts, and posts in the Instagram platform.

Users can own accounts, follow other accounts, and be followed by other users. Accounts can have public or private settings, and access to view an account is determined by ownership, followers, and privacy settings. Posts are associated with accounts and can have restricted comments and likes based on account privacy.

### Entities & Relations

- **`User`**: Represents a user on the Instagram platform.

- **`Account`**: Represents a user account on Instagram. Accounts have owners, followers, and can follow other accounts.

- **`Post`**: Represents a post on Instagram. Posts are linked to accounts and can have restricted comments and likes.

### Permissions

Users can view an account if they are the owner, a follower, or if the account is public.
Users can comment and like posts if they have access to view the linked account and the post is unrestricted.

### Relationships and Attributes

Based on our schema, let's create some sample relationships to test both our schema and our authorization logic.

```perm
// Relationships
// Users, Accounts and Posts:
  account:1#owner@user:kevin
  account:2#owner@user:george
  account:1#following@user:george
  account:2#follower@user:kevin
  post:1#account@account:1
  post:2#account@account:2

// Attributes
// Accounts and Posts:
  account:1$public|boolean:true
  account:2$public|boolean:false
  post:1$restricted|boolean:false
  post:2$restricted|boolean:true
```

## Test & Validation

To validate our authorization logic, let's run some tests on different scenarios using the Instagram Authorization Schema.

### Test 1: Checking Account Viewing Permissions

<details>
<summary> 
    Can <strong>user:kevin</strong> view <strong>account:1</strong>? 
</summary>
    
<p>

```perm
    entity account {
        relation owner @user
        relation following @user
        relation follower @user
        attribute public boolean
        action view = (owner or follower) or public
    }
```

According to the schema, `user:kevin` is the owner of `account:1`. Hence, `user:kevin` should be able to view `account:1`. The expected result is `'true'`.

</p>
</details>

<details>
<summary> 
    Can <strong>user:kevin</strong> view <strong>account:2</strong>? 
</summary>
    
<p>

```perm
    entity account {
        relation owner @user
        relation following @user
        relation follower @user
        attribute public boolean
        action view = (owner or follower) or public
    }
```

According to the schema, `user:kevin` follows `account:2`. Hence, `user:kevin` should be able to view `account:2` because he is a follower. The expected result is `'true'`.

</p>
</details>

<details>
<summary> 
    Can <strong>user:george</strong> view <strong>account:1</strong>? 
</summary>
    
<p>

```perm
    entity account {
        relation owner @user
        relation following @user
        relation follower @user
        attribute public boolean
        action view = (owner or follower) or public
    }
```

According to the schema, `user:george` can view `account:1`, because the account is public. Hence, `user:george` should be able to view `account:1`. The expected result is `'true'`.

</p>
</details>

<details>
<summary> 
    Can <strong>user:george</strong> view <strong>account:2</strong>? 
</summary>
    
<p>

```perm
    entity account {
        relation owner @user
        relation following @user
        relation follower @user
        attribute public boolean
        action view = (owner or follower) or public
    }
```

According to the schema, `user:george` is the owner of `account:2`. Hence, `user:george` should be able to view `account:2`. The expected result is `'true'`.

</p>
</details>

### Test 2: Checking Post Viewing Permissions

<details><summary>Can <strong>user:george</strong> view <strong>post:1</strong>?</summary>

<p>

```perm
entity post {
    relation account @account
    attribute restricted boolean
    action view = account.view
}
```

According to the schema, `post:1` is linked with `account:1`, and it does not have restricted access. Also, `user:george` is following `account:1`. Hence, `user:george` should be able to view `post:1`. The expected result is `'true'`.

</p>
</details>

<details><summary>Can <strong>user:kevin</strong> view <strong>post:2</strong>?</summary>
<p>

```perm
entity post {
    relation account @account
    attribute restricted boolean
    action view = account.view
}
```

According to the schema, `post:2` is linked with `account:2`, and it has restricted access. Also, `user:george` is not following `account:1`. Hence, `user:kevin` should not be able to view `post:2`. The expected result is `'false'`.

</p>
</details>

<details><summary>Can <strong>user:george</strong> view <strong>post:2</strong>?</summary>
<p>

```perm
entity post {
    relation account @account
    attribute restricted boolean
    action view = account.view
}
```

According to the schema, `post:2` is linked with `account:2`, and it is restricted access. Also, `user:george` can view his own `post:2`. The expected result is `'true'`.

</p>
</details>

### Test 3: Checking Post Commenting Permissions

<details><summary>Can <strong>user:george</strong> comment <strong>post:1</strong>?</summary>
<p>

```perm
entity post {
    relation account @account
    attribute restricted boolean
    action comment = account.following not restricted
}
```

According to the schema, `post:1` is linked with `account:1`, and it is not restricted. Also, `user:george` can comment on `post:1`. The expected result is `'true'`.

</p>
</details>

<details><summary>Can <strong>user:kevin</strong> comment <strong>post:2</strong>?</summary>
<p>

```perm
entity post {
    relation account @account
    attribute restricted boolean
    action comment = account.following not restricted
}
```

According to the schema, `post:2` is linked with `account:2`, and it is restricted. `user:kevin` cannot comment on `post:2`. The expected result is `'false'`.

</p>
</details>

Let's test these access checks in our local with using **permify validator**. We'll use the below schema for the schema validation file.

```yaml
schema: |-
  entity user {}

  entity account {
      // users have accounts
      relation owner @user
      
      // accounts can follow other users/accounts.
      relation following @user

      // other users/accounts can follow account.
      relation follower @user

      // accounts can be private or public.
      attribute public boolean

      // users can view an account if they're followers, owners, or if the account is not private.
      action view = (owner or follower) or public
      
  }

  entity post {
      // posts are linked with accounts.
      relation account @account

      // comments are limited to people followed by the parent account.
      attribute restricted boolean

      // users can view the posts, if they have access to view the linked accounts.
      action view = account.view

      // users can comment and like on unrestricted posts or posts by owners who follow them.
      action comment = account.following not restricted
      action like = account.following not restricted
  }
relationships:
  - account:1#owner@user:kevin
  - account:2#owner@user:george
  - account:1#following@user:george
  - account:2#follower@user:kevin
  - post:1#account@account:1
  - post:2#account@account:2
attributes:
  - account:1$public|boolean:true
  - account:2$public|boolean:false
  - post:1$restricted|boolean:false
  - post:2$restricted|boolean:true
scenarios:
  - name: Account Viewing Permissions
    description: Evaluate account viewing permissions for 'kevin' and 'george'.
    checks:
      - entity: account:1
        subject: user:kevin
        assertions:
          view: true
      - entity: account:2
        subject: user:kevin
        assertions:
          view: true
      - entity: account:1
        subject: user:george
        assertions:
          view: true
      - entity: account:2
        subject: user:george
        assertions:
          view: true
  - name: Post Viewing Permissions
    description: Determine post viewing permissions for 'kevin' and 'george'.
    checks:
      - entity: post:1
        subject: user:george
        assertions:
          view: true
      - entity: post:2
        subject: user:kevin
        assertions:
          view: true
      - entity: post:2
        subject: user:george
        assertions:
          view: true
  - name: Post Commenting Permissions
    description: Evaluate post commenting permissions for 'kevin' and 'george'.
    checks:
      - entity: post:1
        subject: user:george
        assertions:
          comment: true
      - entity: post:2
        subject: user:kevin
        assertions:
          comment: false
```

## Using Schema Validator in Local

After cloning [Permify](https://github.com/Permify/permify), open up a new file and copy the **schema yaml file** content inside. Then, build and run Permify instance using the command `make serve`

![Running Permify](https://github.com/Permify/permify/assets/48759364/eb4cde6e-09bf-4e38-88bc-251a811f9c4f)

Then run `permify validate {path of your schema validation file}` to start the test process.

The validation result according to our example schema validation file:

![test-result](https://github.com/Permify/permify/assets/48759364/2fb9a1ab-40d4-48e0-857a-3d59de575134)

## Need any help ?

This is the end of demonstration of the authorization structure for Facebook groups. To install and implement this see the [Set Up Permify](../../installation.md) section.
