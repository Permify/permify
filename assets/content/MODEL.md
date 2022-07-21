# Modeling Authorization with Permify Schema

## Permify Schema

You can model your authorization with defining your entities, relations between them and access control decisions of each action using Permify Schema.

For creating a Permify Schema file you can use **.perm** file extension. 

Lets create a github example using Permify Schema. To see full implementation you can jump directly to [Github Example](#github-example). 

### Entities

The very first step to building Permify Schema is creating your Entities.

Entities represent your main tables. The table name and the entity name here must be the same. 

```perm
entity user {}

entity organization {}

entity repository {} 
```

→ For our github case, we can create user, organization and repository entities like above. In that case, name of the user entity represents user table in your database.

Entity has 2 different attributes. These are;

- **relations**
- **actions**

### Relations

Relations represent relationships between entities. Attribute ***relation*** need to used in with several options to create a entity relation

**Options:**

- **name:** custom relation name.
- **entity:** the entity it’s related with (e.g. user, organization, repo…)
- **table (optional):** the name of the pivot table. (Only for many-to-many relationships.)
- **rel:(optional):** type of relationship (many-to-many, belongs-to or custom)
- **cols:(optional):** the columns you have created in your database.

→ Let's go back to our github example, organizations and users can have repositories so each repository is related with an organization and user.

```
entity repository {

    relation    owner @user         
    relation    org   @organization   

}
```

#### Actions

Actions describe what relations, or relation’s relation can do. Permify Schema supports ***and*** and ***or*** operators to define actions.

Attribute ***action*** need to used inside of entites with these operators.

```
entity repository {

    ..
    ..

    action push   = owner

}
```

→ For example, only the repository owner can push to
repository.

```
entity repository {

    ..
    ..

    action read   = (owner or org.member) and org.admin

}
```

→ Another one, "user with a admin role and either owner of the repository, or member of the organization which repository belongs to"
can read.

## Github Example 

Here is full implemetation - including non-mandatory options - of Github example with using Permify Schema.

```perm
entity user {} `table:users|identifier:id`

entity organization {

    relation admin @user     `rel:custom`
    relation member @user    `rel:many-to-many|table:org_members|cols:org_id,user_id`

    action create_repository = admin or member
    action delete = admin

} `table:organizations|identifier:id`

entity repository {

    relation    owner @user          `rel:belongs-to|cols:owner_id`
    relation    org   @organization    `rel:belongs-to|cols:organization_id`

    action push   = owner
    action read   = (owner or org.member) and org.admin
    action delete = org.admin or owner

} `table:repositories|identifier:id`
```
