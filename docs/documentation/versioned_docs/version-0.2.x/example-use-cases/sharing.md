
# Sharing & Collaboration

Inviting a team member to a document, project or repository should be hassle free to model. In Permify you can achieve this with simply defining a invite action. Check out the below model block see how sharing can be modeled with Permify's DSL, [Permify Schema].

[Permify Schema]: /docs/getting-started/modeling

-------

```perm
entity user {}

entity organization {

    // organizational roles
    relation admin @user
    relation member @user
    relation manager @user
    
}

entity project {

	// represents project's parent organization
    relation org @organization
    
    // represents owner of this project
    relation owner  @user
    
    // invite permission
    action invite  = org.admin or owner

}

```

## Need any help ?

Our team is happy to help you get started with Permify. If you'd like to learn more about using Permify in your app or have any questions about this example, [schedule a call with one of our Permify engineer](https://meetings-eu1.hubspot.com/ege-aytin/call-with-an-expert).

