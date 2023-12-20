# Facebook Groups

This example demonstrate the authorization structure for Facebook groups, which enables users to perform various actions based on their roles and permissions within the group.

### Schema | [Open in playground](https://play.permify.co/?s=XNEAs8dr0AINwCuSMcxHI)

```perm
// Represents a user
entity user {}

// Represents a Facebook group
entity group {

    // Relation to represent the members of the group
    relation member @user
    // Relation to represent the admins of the group
    relation admin @user
    // Relation to represent the moderators of the group
    relation moderator @user

    // Permissions for the group entity
    action create = member
    action join = member
    action leave = member
    action invite_to_group = admin
    action remove_from_group = admin or moderator
    action edit_settings = admin or moderator
    action post_to_group = member
    action comment_on_post = member
    action view_group_insights = admin or moderator
}

// Represents a post in a Facebook group
entity post {

    // Relation to represent the owner of the post
    relation owner @user
    // Relation to represent the group that the post belongs to
    relation group @group

    // Permissions for the post entity
    action view_post = owner or group.member
    action edit_post = owner or group.admin
    action delete_post = owner or group.admin

    permission group_member = group.member
}

// Represents a comment on a post in a Facebook group
entity comment {

    // Relation to represent the owner of the comment
    relation owner @user

    // Relation to represent the post that the comment belongs to
    relation post @post

    // Permissions for the comment entity
    action view_comment = owner or post.group_member
    action edit_comment = owner
    action delete_comment = owner
}

// Represents a comment like on a post in a Facebook group
entity like {

    // Relation to represent the owner of the like
    relation owner @user

    // Relation to represent the post that the like belongs to
    relation post @post

    // Permissions for the like entity
    action like_post = owner or post.group_member
    action unlike_post = owner or post.group_member
}

// Definition of poll entity
entity poll {

     // Relation to represent the owner of the poll
    relation owner @user

    // Relation to represent the group that the poll belongs to
    relation group @group

    // Permissions for the poll entity
    action create_poll = owner or group.admin
    action view_poll = owner or group.member
    action edit_poll = owner or group.admin
    action delete_poll = owner or group.admin
}

// Definition of file entity
entity file {

    // Relation to represent the owner of the file
    relation owner @user

    // Relation to represent the group that the file belongs to
    relation group @group

    // Permissions for the file entity
    action upload_file = owner or group.member
    action view_file = owner or group.member
    action delete_file = owner or group.admin
}

// Definition of event entity
entity event {

    // Relation to represent the owner of the event
    relation owner @user
    // Relation to represent the group that the event belongs to
    relation group @group

    // Permissions for the event entity
    action create_event = owner or group.admin
    action view_event = owner or group.member
    action edit_event = owner or group.admin
    action delete_event = owner or group.admin
    action RSVP_to_event = owner or group.member
}
```

## Brief Examination of the Model

The model defines several entities and relations, as well as actions and permissions that can be taken by users within the group. Let's examine them shortly;

### Entities & Relations

* **`user`** entity represents a user in the Facebook.

* **`group`** entity represents the Facebook group, and it has several relations including member, admin, and moderator to represent the members, admins, and moderators of the group. Additionally, there are relations to represent the posts and comments in the group.

* **`post`** entity represents a post in the Facebook group, and it has relations to represent the owner of the post and the group that the post belongs to.

* **`comment`** entity represents a comment on a post in the Facebook group, and it has relations to represent the owner of the comment, the post that the comment belongs to, and the comment itself.

* **`like`** entity represents a like on a post in the Facebook group, and it has relations to represent the owner of the like and the post that the like belongs to.

* **`poll`** entity represents a poll in the Facebook group, and it has relations to represent the owner of the poll and the group that the poll belongs to.

* **`file`** entity represents a file in the Facebook group, and it has relations to represent the owner of the file and the group that the file belongs to.

* **`event`** entity represents an event in the Facebook group, and it has relations to represent the owner of the event and the group that the event belongs to.

### Permissions 

We have several actions attached with the entities, which are limited by certain permissions. 

For example, the `create_group` action can only be performed by a `member`, as follows:

#### Creating a group permission

```perm
entity group {

    // Relation to represent the members of the group
    relation member @user
    
    ..

    // Create group permission 
    action create_group = member
    
    ..
    ..
}
```

Another example would be given from the `edit_post` action in the post entity, which specifies the permissions required to edit a post in a Facebook group.

#### Editing a post permission

```perm
entity post {

    // Relation to represent the owner of the post
    relation owner @user
    // Relation to represent the group that the post belongs to
    relation group @group

    // Permissions for the post entity
    ..

    action edit_post = owner or group.admin

    ..
    ..
}
```

An **owner** of a post can always edit their own post. In addition, members who are defined as **admin** of the group - which the post belongs to - can also edit the post.

Since most entities are deeply nested together, we also have multiple hierarchical permissions. 

#### Nested Hierarchies

For example, we can define a permission "view_comment" if only user is owner of that comment or user is a member of the group which the comment's post belongs.

```perm
// Represents a post in a Facebook group
entity post {

    ..
    ..

    // Relation to represent the group that the post belongs to
    relation group @group

    // Permissions for the post entity
    
    ..
    ..
    permission group_member = group.member
}

// Represents a comment on a post in a Facebook group
entity comment {

    // Relation to represent the owner of the comment
    relation owner @user

    // Relation to represent the post that the comment belongs to
    relation post @post
    relation comment @comment

    ..
    ..

    // Permissions 
    action view_comment = owner or post.group_member

    ..
    ..
}
```

The `post.group_member` refers to the members of the group to which the post belongs. We defined it as action in **post** entity as,

```perm
permission group_member = group.member
```

Permissions can be inherited as relations in other entities. This allows to form nested hierarchical relationships between entities. 

In this example, a comment belongs to a post which is part of a group. Since there is a **'member'** relation defined for the group entity, we can use the **'group_member'** permission to inherit the **member** relation from the group in the post and then use it in the comment.

## Relationships

Based on our schema, let's create some sample relationships to test both our schema and our authorization logic.

```perm
//group relationships
group:1#member@user:1
group:1#admin@user:2
group:2#moderator@user:3
group:2#member@user:4
group:1#member@user:5

//post relationships
post:1#owner@user:1
post:1#group@group:1
post:2#owner@user:4
post:2#group@group:1

//comment relationships
comment:1#owner@user:2
comment:1#post@post:1
comment:2#owner@user:5
comment:2#post@post:2

//like relationships
like:1#owner@user:3
like:1#post@post:1
like:2#owner@user:4
like:2#post@post:2

//poll relationships
poll:1#owner@user:2
poll:1#group@group:1
poll:2#owner@user:5
poll:2#group@group:1

//like relationships
file:1#owner@user:1
file:1#group@group:1

//event relationships
event:1#owner@user:3
event:1#group@group:1
```

## Test & Validation

Finally, let's check some permissions and test our authorization logic. 

<details><summary>can <strong>user:4 RSVP_to_event event:1</strong> ? </summary>
<p>

```perm
    entity event {

        // Relation to represent the owner of the event
        relation owner @user
        // Relation to represent the group that the event belongs to
        relation group @group

        // Permissions for the event entity

        ..
        ..

        action RSVP_to_event = owner or group.member
    }
```

According to what we have defined for the **'RSVP_to_event'** action, users who are either the owner of `event:1` or a member of the group that belongs to `event:1` can grant access to RSVP to the event.

According to the relation tuples we created, `user:4` is not the **owner** of the event. Furthermore, when we check whether `user:4` is a **member** of the only group (`group:1`) that `event:1` is part of (`event:1#group@group:1`), we see that there is no **member** relation for `user:4` in that group. 

Therefore, the `user:4 RSVP_to_event event:1` check request should yield a **'false'** response.

</p>
</details>

<details><summary>can <strong>user:5 view_comment comment:1</strong> ? </summary>
<p>

```perm
// Represents a post in a Facebook group
entity post {

    ..
    ..

    // Relation to represent the group that the post belongs to
    relation group @group

    // Permissions for the post entity
    
    ..
    ..
    permission group_member = group.member
}

// Represents a comment on a post in a Facebook group
entity comment {

    // Relation to represent the owner of the comment
    relation owner @user

    // Relation to represent the post that the comment belongs to
    relation post @post
    relation comment @comment

    ..
    ..

    // Permissions 
    action view_comment = owner or post.group_member

    ..
    ..
}
```

According to the relation tuples we created, `user:5` is not the **owner** of the comment. But member of the `group:1` and thats grant `user:5` (`group:1#member@user:5`) access to perform view the comment:1. In particularly, `comment:1` is part of the `post:1` (`comment:1#post@post:1`) and `post:1` is part of the group:1 (`post:1#group@group:1`). And from the action definition on above model group:1 members can view the `comment:1`. 

Therefore, the `user:5 view_comment comment:1` check request should yield a **'true'** response.

</p>
</details>

Let's test these access checks in our local with using **permify validator**. We'll use the below schema for the schema validation file. 

```yaml
schema: >-
    entity user {}

    entity group {

        // Relation to represent the members of the group
        relation member @user
        // Relation to represent the admins of the group
        relation admin @user
        // Relation to represent the moderators of the group
        relation moderator @user

        // Permissions for the group entity
        action create = member
        action join = member
        action leave = member
        action invite_to_group = admin
        action remove_from_group = admin or moderator
        action edit_settings = admin or moderator
        action post_to_group = member
        action comment_on_post = member
        action view_group_insights = admin or moderator
    }

    entity post {

        // Relation to represent the owner of the post
        relation owner @user
        // Relation to represent the group that the post belongs to
        relation group @group

        // Permissions for the post entity
        action view_post = owner or group.member
        action edit_post = owner or group.admin
        action delete_post = owner or group.admin

        permission group_member = group.member
    }

    entity comment {

        // Relation to represent the owner of the comment
        relation owner @user

        // Relation to represent the post that the comment belongs to
        relation post @post

        // Permissions for the comment entity
        action view_comment = owner or post.group_member
        action edit_comment = owner
        action delete_comment = owner
    }

    entity like {

        // Relation to represent the owner of the like
        relation owner @user

        // Relation to represent the post that the like belongs to
        relation post @post

        // Permissions for the like entity
        action like_post = owner or post.group_member
        action unlike_post = owner or post.group_member
    }

    entity poll {

        // Relation to represent the owner of the poll
        relation owner @user

        // Relation to represent the group that the poll belongs to
        relation group @group

        // Permissions for the poll entity
        action create_poll = owner or group.admin
        action view_poll = owner or group.member
        action edit_poll = owner or group.admin
        action delete_poll = owner or group.admin
    }

    entity file {

        // Relation to represent the owner of the file
        relation owner @user

        // Relation to represent the group that the file belongs to
        relation group @group

        // Permissions for the file entity
        action upload_file = owner or group.member
        action view_file = owner or group.member
        action delete_file = owner or group.admin
    }

    entity event {

        // Relation to represent the owner of the event
        relation owner @user
        // Relation to represent the group that the event belongs to
        relation group @group

        // Permissions for the event entity
        action create_event = owner or group.admin
        action view_event = owner or group.member
        action edit_event = owner or group.admin
        action delete_event = owner or group.admin
        action RSVP_to_event = owner or group.member
    }

relationships:
    - group:1#member@user:1
    - group:1#admin@user:2
    - group:2#moderator@user:3
    - group:2#member@user:4
    - group:1#member@user:5
    - post:1#owner@user:1
    - post:1#group@group:1
    - post:2#owner@user:4
    - post:2#group@group:1
    - comment:1#owner@user:2
    - comment:1#post@post:1
    - comment:2#owner@user:5
    - comment:2#post@post:2
    - like:1#owner@user:3
    - like:1#post@post:1
    - like:2#owner@user:4
    - like:2#post@post:2
    - poll:1#owner@user:2
    - poll:1#group@group:1
    - poll:2#owner@user:5
    - poll:2#group@group:1
    - file:1#owner@user:1
    - file:1#group@group:1
    - event:1#owner@user:3
    - event:1#group@group:1

scenarios:
  - name: "scenario 1"
    description: "test description"
    checks:
      - entity: "event:1"
        subject: "user:4"
        assertions:
          RSVP_to_event : false
      - entity: "comment:1"
        subject: "user:5"
        assertions:
          view_comment : true
```

### Using Schema Validator in Local

After cloning [Permify](https://github.com/Permify/permify), open up a new file and copy the **schema yaml file** content inside. Then, build and run Permify instance using the command `make serve`.

![Running Permify](https://user-images.githubusercontent.com/34595361/233155326-e1d2daf6-2406-4139-b0b3-5f7b54880593.png)

Then run `permify validate {path of your schema validation file}` to start the test process. 

The validation result according to our example schema validation file:

![Screen Shot 2023-04-16 at 15 53 06](https://user-images.githubusercontent.com/34595361/233152003-1fbaf2af-d208-4290-af1f-359870b0de49.png)

## Need any help ?

This is the end of demonstration of the authorization structure for Facebook groups. To install and implement this see the [Set Up Permify](../../installation.md) section.

If you need any kind of help, our team is happy to help you get started with Permify. If you'd like to learn more about using Permify in your app or have any questions about it, [schedule a call with one of our Permify engineer](https://meetings-eu1.hubspot.com/ege-aytin/call-with-an-expert).