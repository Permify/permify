# Notion

This is a schema definition of the authorization model for Notion, a popular productivity and organization tool.

The model defines several entities, including users, workspaces, pages, databases, blocks, and integrations. It also includes several default roles, such as Admin, Bot, Guest, and Member.

### Schema | [Open in playground](https://play.permify.co/?s=XNEAs8dr0AINwCuSMcxHI)

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

    // Define permissions for commnet actions
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

   // Define permissions for 'ntegrat'on actions
    permission read = workspace.read
    permission write = owner or workspace.write
}
```

## Brief Examination of the Model
