# Contributing

Welcome to Permify contribution guidelines, happy to see you here :blush:

Before participating in the community, we must specify that all of the contributions must follow our [Code of Conduct](https://github.com/Permify/permify/blob/master/CODE_OF_CONDUCT.md). Please read it before you make any contributions.

If you need any help or want to talk about about a specific issue, you can always reach out to me from my mail:ege@permify.co.

You're always more than welcome to our other communication channels.

## Communication Channels

<p align="left">
<a href="https://discord.gg/n6KfzYxhPp">
 <img height="70px" width="70px" alt="permify | Discord" src="https://user-images.githubusercontent.com/39353278/187209316-3d01a799-c51b-4eaa-8f52-168047078a14.png" />
</a>
<a href="https://twitter.com/GetPermify">
  <img height="70px" width="70px" alt="permify | Twitter" src="https://user-images.githubusercontent.com/39353278/187209323-23f14261-d406-420d-80eb-1aa707a71043.png"/>
</a>
<a href="https://www.linkedin.com/company/permifyco">
  <img height="70px" width="70px" alt="permify | Linkedin" src="https://user-images.githubusercontent.com/39353278/187209321-03293a24-6f63-4321-b362-b0fc89fdd879.png" />
</a>
</p>

## Ways to contribute

* **Contribute to codebase:** We're collaboratively working with our community to make Permify the best it can be! You can develop new features, fix existing issues or make third-party integrations/packages. 
* **Improve documentation:** Alongside our codebase, documentation is one of the most significant things in our open-source journey. We're trying to give the best DX possible to explain ourselves and Permify. And you can help on that by improving resources or adding new ones.
* **Contribute to playground:** Permify playground allows you to visualize and test your authorization logic. You can contribute to our playground by improving its user interface, fixing glitches, or adding new features.

### Contribution Steps

- Fork this repository.
- Clone the repository you forked.
- Create a branch with specified name. It's better to relate it with your issue title.
- Make necessary changes and commit those changes. Make sure to test your changes. 
- Push changes to your branch.
- Submit your changes for review.

## Commit convention

We use [conventional commits](https://www.conventionalcommits.org/en/v1.0.0/) to keep our commit messages consistent and easy to understand. Here is the applied form of a commit message.

```
<type>(optional scope): <description>
```

**Examples:**

- `feat: added multi tenant authentication support`
- `fix: fixed welcomeServer duplicated syntax`
- `docs: update the deployment options on set up section`

### Types

`fix:`,  `feat:`, `build:`, `chore:`, `ci:`, `docs:`, `style:`, `refactor:`, `perf:`, `test:`

## Running Tests 

In order to contribute and test in our codebase you need to have Go version 1.19 or higher.

```go test -v ./...```

### Adding dependencies
Permify is not using anything other than the standard Go modules toolchain to manage dependencies.

```go get github.com/org/newdependency@version```

### Updating generated Protobuf code
All Protobuf code is managed using buf.

```buf generate```

## Issues

If you found any bug, have feature request or just want to improve our code base, docs or other resources; you can open an issue about it to let us know. If you plan to work on an existing issue, mention us on the issue page before you start working on it so we can assign you to it.

### When opening a issue

- If you plan to work on a problem, please check that same problem or topic does not already exist.
- If you plan to work on a new feature, our advise is to discuss it with other community members/maintainers who might give you an idea or support.
- If you're stuck anywhere, ask for help in our discord community.
- Please relate one bug with one issue, do not use issues as bug lists. 

You can create an issue and contribute to anything you want, but please ensure to follow the steps above. We will definitely ease your work and help on anything when needed.


