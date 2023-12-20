package shapes

import (
	"github.com/Permify/permify/pkg/development/file"
)

// FACEBOOK GROUPS SAMPLE

var InitialFacebookGroupsShape = file.Shape{
	Schema: `
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
    `,
	Relationships: []string{
		// group relationships
		"group:1#member@user:1",
		"group:1#admin@user:2",
		"group:2#moderator@user:3",
		"group:2#member@user:4",
		"group:1#member@user:5",

		// post relationships
		"post:1#owner@user:1",
		"post:1#group@group:1",
		"post:2#owner@user:4",
		"post:2#group@group:1",

		// comment relationships
		"comment:1#owner@user:2",
		"comment:1#post@post:1",
		"comment:2#owner@user:5",
		"comment:2#post@post:2",

		// like relationships
		"like:1#owner@user:3",
		"like:1#post@post:1",
		"like:2#owner@user:4",
		"like:2#post@post:2",

		// poll relationships
		"poll:1#owner@user:2",
		"poll:1#group@group:1",
		"poll:2#owner@user:5",
		"poll:2#group@group:1",

		// like relationships
		"file:1#owner@user:1",
		"file:1#group@group:1",

		// event relationships
		"event:1#owner@user:3",
		"event:1#group@group:1",
	},
	Scenarios: []file.Scenario{
		{
			Name:        "Scenario 1",
			Description: "Scenario Description",
			Checks: []file.Check{
				{
					Entity:  "event:1",
					Subject: "user:4",
					Assertions: map[string]bool{
						"RSVP_to_event": false,
					},
				},
				{
					Entity:  "comment:1",
					Subject: "user:5",
					Assertions: map[string]bool{
						"view_comment": true,
					},
				},
			},
			EntityFilters:  []file.EntityFilter{},
			SubjectFilters: []file.SubjectFilter{},
		},
	},
}
