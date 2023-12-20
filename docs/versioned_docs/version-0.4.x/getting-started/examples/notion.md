# Notion

This is a schema definition of the authorization model for Notion, a popular productivity and organization tool.

### Schema | [Open in playground](https://play.permify.co/?s=BsCvLmd4g81sB20XJZI5p)

```perm
entity user {}

entity workspace {
    // The owner of the workspace
    relation owner @user
    // Members of the workspace
    relation member @user
    // Guests (users with read-only access) of the workspace
    relation guest @user
    // Bots associated with the workspace
    relation bot @user
    // Admin users who have permission to manage the workspace
    relation admin @user

    // Define permissions for workspace actions
    permission create_page = owner or member or admin
    permission invite_member = owner or admin
    permission view_workspace = owner or member or guest or bot
    permission manage_workspace = owner or admin

    // Define permissions that can be inherited by child entities
    permission read = member or guest or bot or admin
    permission write = owner or admin
}

entity page {
    // The workspace associated with the page
    relation workspace @workspace
     // The user who can write to the page
    relation writer @user
     // The user(s) who can read the page (members of the workspace or guests)
    relation reader @user @workspace#member @workspace#guest

    // Define permissions for page actions
    permission read = reader or workspace.read
    permission write = writer or workspace.write
}

entity database {
    // The workspace associated with the database
    relation workspace @workspace
    // The user who can edit the database
    relation editor @user
    // The user(s) who can view the database (members of the workspace or guests)
    relation viewer @user @workspace#member @workspace#guest

    // Define permissions for database actions
    permission read = viewer or workspace.read
    permission write = editor or workspace.write
    permission create = editor or workspace.write
    permission delete = editor or workspace.write
}

entity block {
    // The page associated with the block
    relation page @page
    // The database associated with the block

    relation database @database
    // The user who can edit the block
    relation editor @user
    // The user(s) who can comment on the block (readers of the parent object)
    relation commenter @user @page#reader

    // Define permissions for block actions
    permission read = database.read or commenter
    permission write = editor or database.write
    permission comment = commenter
}

entity comment {
    // The block associated with the comment
    relation block @block

     // The author of the comment
    relation author @user

    // Define permissions for comment actions
    permission read = block.read
    permission write = author
}

entity template {
   // The workspace associated with the template
    relation workspace @workspace
    // The user who creates the template
    relation creator @user

    // The user(s) who can view the page (members of the workspace or guests)
    relation viewer @user @workspace#member @workspace#guest

    // Define permissions for template actions
    permission read = viewer or workspace.read
    permission write = creator or workspace.write
    permission create = creator or workspace.write
    permission delete = creator or workspace.write
}

entity integration {
    // The workspace associated with the integration
    relation workspace @workspace

    // The owner of the integration
    relation owner @user

    // Define permissions for integration actions
    permission read = workspace.read
    permission write = owner or workspace.write
}
```

## Brief Examination of the Model

The model defines several entities, including users, workspaces, pages, databases, blocks, and integrations. It also includes several default roles, such as Admin, Bot, Guest, and Member. Here's a breakdown of the entities:

### Entities & Relations

- **`user`**: Represents a user in the system.

- **`workspace`**: Represents a workspace in which users can collaborate. Each workspace has an owner, members, guests, and bots associated with it. The owner and admin users have permission to manage the workspace. Permissions are defined for creating pages, inviting members, viewing the workspace, and managing the workspace. The read and write permissions can be inherited by child entities.

- **`page`**: Represents a page within a workspace. Each page is associated with a workspace and has a writer and readers. The read and write permissions are defined based on the writer and readers of the page and can be inherited from the workspace.

- **`database`**: Represents a database within a workspace. Each database is associated with a workspace and has an editor and viewers. The read and write permissions are defined based on the editor and viewers of the database and can be inherited from the workspace. Permissions are also defined for creating and deleting databases.

- **`block`**: Represents a block within a page or database. Each block is associated with a page or database and has an editor and commenters. The read and write permissions are defined based on the editor and commenters of the block and can be inherited from the database. Commenters are users who have permission to comment on the block.

- **`comment`**: Represents a comment on a block. Each comment is associated with a block and has an author. The read and write permissions are defined based on the author of the comment and can be inherited from the block.

- **`template`**: Represents a template within a workspace. Each template is associated with a workspace and has a creator and viewers. The read and write permissions are defined based on the creator and viewers of the template and can be inherited from the workspace. Permissions are also defined for creating and deleting templates.

- **`integration`**: Represents an integration within a workspace. Each integration is associated with a workspace and has an owner. Permissions are defined for reading and writing to the integration.

### Permissions

We have several actions attached with the entities, which are limited by certain permissions. Let's examine the **read** permission of the page entity.

#### Page Read Permission

```perm
entity workspace {
    // The owner of the workspace
    relation owner @user
    // Members of the workspace
    relation member @user
    // Guests (users with read-only access) of the workspace
    relation guest @user
    // Bots associated with the workspace
    relation bot @user
    // Admin users who have permission to manage the workspace
    relation admin @user

    // Define permissions for workspace actions

    ..
    ..

    // Define permissions that can be inherited by child entities
    permission read = member or guest or bot or admin
    ..
}

entity page {

    // The workspace associated with the page
    relation workspace @workspace

    ..
    ..

    // The user(s) who can read the page (members of the workspace or guests)
    relation reader @user @workspace#member @workspace#guest

    ..
    ..

    // Define permissions for page actions
    permission read = reader or workspace.read

    ..
    ..
}
```

This permission specifies who can read the contents of the page at Notion.

The `reader` relation specifies the users who are members of the workspace associated with the page (`workspace#member`) or guests of the workspace (`workspace#guest`).

Read permission of the workspace inherited as `workspace.read` in the page entity. THis permission specifies that any user who has been granted read access to the workspace object (i.e., the workspace that the page belongs to) can also read the page.

In summary, any user who is a member or guest of the workspace and has been granted read access to the page through the reader relation, as well as any user who has been granted read access to the workspace itself, can read the contents of the page.

## Relationships

Based on our schema, let's create some sample relationships to test both our schema and our authorization logic.

```perm
// Assign users to different workspaces:
workspace:engineering_team#owner@user:alice
workspace:engineering_team#member@user:bob
workspace:engineering_team#guest@user:charlie
workspace:engineering_team#admin@user:alice
workspace:sales_team#owner@user:david
workspace:sales_team#member@user:eve
workspace:sales_team#guest@user:frank
workspace:sales_team#admin@user:david

// Connect pages, databases, and templates to workspaces:
page:project_plan#workspace@workspace:engineering_team
page:product_spec#workspace@workspace:engineering_team
database:task_list#workspace@workspace:engineering_team
template:weekly_report#workspace@workspace:sales_team
database:customer_list#workspace@workspace:sales_team
template:marketing_campaign#workspace@workspace:sales_team

// Set permissions for pages, databases, and templates:
page:project_plan#writer@user:frank
page:project_plan#reader@user:bob

database:task_list#editor@user:alice
database:task_list#viewer@user:bob

template:weekly_report#creator@user:alice
template:weekly_report#viewer@user:bob

page:product_spec#writer@user:david
page:product_spec#reader@user:eve

database:customer_list#editor@user:david
database:customer_list#viewer@user:eve

template:marketing_campaign#creator@user:david
template:marketing_campaign#viewer@user:eve

// Set relationships for blocks and comments:
block:task_list_1#database@database:task_list
block:task_list_1#editor@user:alice
block:task_list_1#commenter@user:bob
block:task_list_2#database@database:task_list
block:task_list_2#editor@user:alice
block:task_list_2#commenter@user:bob

comment:task_list_1_comment_1#block@block:task_list_1
comment:task_list_1_comment_1#author@user:bob
comment:task_list_1_comment_2#block@block:task_list_1
comment:task_list_1_comment_2#author@user:charlie
comment:task_list_2_comment_1#block@block:task_list_2
comment:task_list_2_comment_1#author@user:bob
comment:task_list_2_comment_2#block@block:task_list_2
comment:task_list_2_comment_2#author@user:charlie
```

## Test & Validation

Since we have our schema and the sample relation tuples, let's check some permissions and test our authorization logic.

<details><summary>can <strong>user:alice write database:task_list</strong> ? </summary>
<p>

```perm
    entity database {
        // The workspace associated with the database
        relation workspace @workspace
        // The user who can edit the database
        relation editor @user

        ..
        ..

        // Define permissions for database actions
        ..
        ..

        permission write = editor or workspace.write

        ..
        ..
    }
```

According to what we have defined for the **'write'** permission, users who are either;

- The editor in task list database (`database:task_list`)
- Have a write permission in the engineering team workspace, which is the only workspace that task list is associated (`database:task_list#workspace@workspace:engineering_team`)

can edit the task list database (`database:task_list`)

Based on the relation tuples we created, `user:alice` doesn't have the **editor** relationship with the `database:task_list`.

Since `user:alice` is the owner and admin in the engineering team workspace (`workspace:engineering_team#admin@user:alice`) it has a write permission defined in the workspace entity, as you can see below:

```perm
entity workspace {
    // The owner of the workspace
    relation owner @user
    ..
    ..
    // Admin users who have permission to manage the workspace
    relation admin @user

    ..
    ..

    // Define permissions that can be inherited by child entities
    ..
    permission write = owner or admin
}
```

And as we mentioned the engineering team workspace is the only workspace that task list is associated (`database:task_list#workspace@workspace:engineering_team`). Therefore, the `user:alice write database:task_list` check request should yield a **'true'** response.

</p>
</details>

<details><summary>can <strong>user:charlie write page:product_spec</strong> ? </summary>
<p>

```perm
entity page {
    // The workspace associated with the page
    relation workspace @workspace
    // The user who can write to the page
    relation writer @user

    ..
    ..

    permission write = writer or workspace.write
}
```

`user:charlie` is guest in the workspace (`workspace:engineering_team#guest@user:charlie`) and the engineering team workspace is the only workspace that `page:product_spec` belongs to.

As we defined, guests doesn't have write permission in a workspace.

```perm
entity workspace {
   // The owner of the workspace
   relation owner @user
   // Admin users who have permission to manage the workspace
   relation admin @user

   ..
   ..

   permission write = owner or admin
}
```

So that, `user:charlie` doesn't have a write relationship in the workspace. And ultimately, the `user:charlie write page:product_spec` check request should yield a **'false'** response.

</p>
</details>

Let's test these access checks in our local with using **permify validator**. We'll use the below schema for the schema validation file.

```yaml
schema: >-
  entity user {}

  entity workspace {
      // The owner of the workspace
      relation owner @user
      // Members of the workspace
      relation member @user
      // Guests (users with read-only access) of the workspace
      relation guest @user
      // Bots associated with the workspace
      relation bot @user
      // Admin users who have permission to manage the workspace
      relation admin @user

      // Define permissions for workspace actions
      permission create_page = owner or member or admin
      permission invite_member = owner or admin
      permission view_workspace = owner or member or guest or bot
      permission manage_workspace = owner or admin

      // Define permissions that can be inherited by child entities
      permission read = member or guest or bot or admin
      permission write = owner or admin
  }

  entity page {
      // The workspace associated with the page
      relation workspace @workspace
      // The user who can write to the page
      relation writer @user
      // The user(s) who can read the page (members of the workspace or guests)
      relation reader @user @workspace#member @workspace#guest

      // Define permissions for page actions
      permission read = reader or workspace.read
      permission write = writer or workspace.write
  }

  entity database {
      // The workspace associated with the database
      relation workspace @workspace
      // The user who can edit the database
      relation editor @user
      // The user(s) who can view the database (members of the workspace or guests)
      relation viewer @user @workspace#member @workspace#guest

      // Define permissions for database actions
      permission read = viewer or workspace.read
      permission write = editor or workspace.write
      permission create = editor or workspace.write
      permission delete = editor or workspace.write
  }

  entity block {
      // The page associated with the block
      relation page @page
      // The database associated with the block

      relation database @database
      // The user who can edit the block
      relation editor @user
      // The user(s) who can comment on the block (readers of the parent object)
      relation commenter @user @page#reader

      // Define permissions for block actions
      permission read = database.read or commenter
      permission write = editor or database.write
      permission comment = commenter
  }

  entity comment {
      // The block associated with the comment
      relation block @block

      // The author of the comment
      relation author @user

      // Define permissions for comment actions
      permission read = block.read
      permission write = author
  }

  entity template {
  // The workspace associated with the template
      relation workspace @workspace
      // The user who creates the template
      relation creator @user

      // The user(s) who can view the page (members of the workspace or guests)
      relation viewer @user @workspace#member @workspace#guest

      // Define permissions for template actions
      permission read = viewer or workspace.read
      permission write = creator or workspace.write
      permission create = creator or workspace.write
      permission delete = creator or workspace.write
  }

  entity integration {
      // The workspace associated with the integration
      relation workspace @workspace

      // The owner of the integration
      relation owner @user

      // Define permissions for integration actions
      permission read = workspace.read
      permission write = owner or workspace.write
  }

relationships:
  - workspace:engineering_team#owner@user:alice
  - workspace:engineering_team#member@user:bob
  - workspace:engineering_team#guest@user:charlie
  - workspace:engineering_team#admin@user:alice
  - workspace:sales_team#owner@user:david
  - workspace:sales_team#member@user:eve
  - workspace:sales_team#guest@user:frank
  - workspace:sales_team#admin@user:david
  - page:project_plan#workspace@workspace:engineering_team
  - page:product_spec#workspace@workspace:engineering_team
  - database:task_list#workspace@workspace:engineering_team
  - template:weekly_report#workspace@workspace:sales_team
  - database:customer_list#workspace@workspace:sales_team
  - template:marketing_campaign#workspace@workspace:sales_team
  - page:project_plan#writer@user:frank
  - page:project_plan#reader@user:bob
  - database:task_list#editor@user:alice
  - database:task_list#viewer@user:bob
  - template:weekly_report#creator@user:alice
  - template:weekly_report#viewer@user:bob
  - page:product_spec#writer@user:david
  - page:product_spec#reader@user:eve
  - database:customer_list#editor@user:david
  - database:customer_list#viewer@user:eve
  - template:marketing_campaign#creator@user:david
  - template:marketing_campaign#viewer@user:eve
  - block:task_list_1#database@database:task_list
  - block:task_list_1#editor@user:alice
  - block:task_list_1#commenter@user:bob
  - block:task_list_2#database@database:task_list
  - block:task_list_2#editor@user:alice
  - block:task_list_2#commenter@user:bob
  - comment:task_list_1_comment_1#block@block:task_list_1
  - comment:task_list_1_comment_1#author@user:bob
  - comment:task_list_1_comment_2#block@block:task_list_1
  - comment:task_list_1_comment_2#author@user:charlie
  - comment:task_list_2_comment_1#block@block:task_list_2
  - comment:task_list_2_comment_1#author@user:bob
  - comment:task_list_2_comment_2#block@block:task_list_2
  - comment:task_list_2_comment_2#author@user:charlie

scenarios:
  - name: "scenario 1"
    description: "test description"
    checks:
      - entity: "database:task_list"
        subject: "user:alice"
        assertions:
          write: true
      - entity: "page:product_spec"
        subject: "user:charlie"
        assertions:
          write: false
```

### Using Schema Validator in Local

After cloning [Permify](https://github.com/Permify/permify), open up a new file and copy the **schema yaml file** content inside. Then, build and run Permify instance using the command `make serve`.

![Running Permify](https://user-images.githubusercontent.com/34595361/233155326-e1d2daf6-2406-4139-b0b3-5f7b54880593.png)

Then run `permify validate {path of your schema validation file}` to start the test process.

The validation result according to our example schema validation file:

![Screen Shot 2023-04-16 at 15 53 06](https://user-images.githubusercontent.com/34595361/233154924-c31a76f4-86f5-4ed3-a1ec-750b642927e6.png)

## Need any help ?

This is the end of demonstration of the authorization structure for Facebook groups. To install and implement this see the [Set Up Permify](../../installation.md) section.

If you need any kind of help, our team is happy to help you get started with Permify. If you'd like to learn more about using Permify in your app or have any questions about it, [schedule a call with one of our Permify engineer](https://meetings-eu1.hubspot.com/ege-aytin/call-with-an-expert).
