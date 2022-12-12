
# Ownership 

Granting privileges to the owner of the resource is a common pattern that many applications follow. Generally we want creators of the resource - document, post, comment etc -  have superior power on that resource. Check out the below model see how ownership can be modeled with Permify's DSL, [Permify Schema].

[Permify Schema]: /docs/getting-started/modeling

-------

```perm
entity user {}

entity comment {

	// represents comment's owner
	relation owner @user

	// permissions 
	action edit = owner
    action delete = owner
}

```

## Sample Relational Tuples 

comment:2#owner@user:1

comment:3#owner@user:51

.
.
.

For more details about how relational tuples created and stored your preferred database, see [Relational Tuples].

[Relational Tuples]: ../getting-started/sync-data.md

## Need any help ?

Our team is happy to help you get started with Permify. If you'd like to learn more about using Permify in your app or have any questions about this example, [schedule a call with one of our Permify engineer](https://meetings-eu1.hubspot.com/ege-aytin/call-with-an-expert).

