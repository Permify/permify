# Contributing

Welcome to Permify contribution guidelines, happy to see you here :blush:

Before dive into a contributing flow and steps for creating good issues and pull requests. We must spesificy that all of the contributions must 
follow our [Code of Conduct](https://github.com/Permify/permify/blob/master/CODE_OF_CONDUCT.md). 
Please read it before you make any contributions.

If you need any help or want to talk about about a spesific topic, you can reach out to me. I'm Ege one of the co-founders of Permify and here is my email:
ege@permify.co

You're always more than welcome to our other communication channels.

## Communication Channels

<p align="left">
<a href="https://discord.gg/MJbUjwskdH">
 <img alt="guilyx’s Discord" width="50px" src="https://user-images.githubusercontent.com/34595361/178992169-fba31a7a-fa80-42ba-9d7f-46c9c0b5a9f8.png" />
</a>
<a href="https://twitter.com/GetPermify">
  <img alt="guilyx | Twitter" width="50px" src="https://user-images.githubusercontent.com/43545812/144034996-602b144a-16e1-41cc-99e7-c6040b20dcaf.png"/>
</a>
<a href="https://www.linkedin.com/company/permifyco">
  <img alt="guilyx's LinkdeIN" width="50px" src="https://user-images.githubusercontent.com/43545812/144035037-0f415fc7-9f96-4517-a370-ccc6e78a714b.png" />
</a>
</p>

## Issues

The best way to contribute to Permify is opening a issue. If you found any bug on Permify or mistake in our documents, contents
you can open an issue about it to let us know. Evaluating problems and fixing them is high priority for us. 

### When opening a issue

- If you plan to work on a problem, please check that same problem or topic does not already exist.
- If you plan to work on a new feature, our advice it discuss it with other community members/maintainers who might give you a idea or support.
- If you stuck anywhere, ask for help in our discord community.
- Please relate one bug with one issue, do not use issues as bug lists. 

After issue creation, If you are looking to make your contribution follow the steps below.

## Contribution Steps

- Fork this repository.
- Clone the repository you forked.
- Create a branch with spesified name. Its better to related with your issue title.
- Make necessary changes and commit those changes. Make sure to test your changes. 
- Push changes to your branch.
- Submit your changes for review.

You can create an issue and contribute about anything you want but following above steps
will definitely ease your work and other maintainers too.

## Installation

### Container (Docker)

#### With terminal

1. Open your terminal.
2. Run following line.

```shell
docker run -d -p 3476:3476 --name permify-container -v {YOUR-CONFIG-PATH}:/config permify/permify:0.0.1
```

3. Test your connection.
    - Create an HTTP GET request ~ localhost:3476/v1/status/ping

#### With docker desktop

Setup docker desktop, and run service with the following steps;

1. Open your docker account.
2. Open terminal and run following line

```shell
docker pull permify/permify:0.0.1
```

3. Open images, and find Permify.
4. Run Permify with the following credentials (optional setting)
    - Container Name: authorization-container
      Ports
    - **Local Host:** 3476
      Volumes
    - **Host Path:** choose the config file and folder
    - **Container Path:** /config
5. Test your connection.
    - Create an HTTP GET request ~ localhost:3476/v1/status/ping










