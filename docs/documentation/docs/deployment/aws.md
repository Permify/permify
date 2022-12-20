---
title: AWS ECS, ECR & EC2
---

#  Deploy on AWS ECS, ECR & EC2

AWS is a piece of cake no one ever said! That’s why today we’re bringing this tutorial to help you deploy Permify in AWS.

There are many ways to deploy and use Permify in AWS. Today we’ll start with Elastic Container Service (ECS). 

ECS is a container management service. You can run your containers as task definitions, and It’s one of the easiest ways to deploy containers.

If you’d like to watch this tutorial rather than reading. Here’s the video version.

<iframe width="100%" height="415" src="https://www.youtube.com/embed/90hCKZLz8jM" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>

There is no prerequisite in this tutorial. You can simply deploy permify by following this step-by-step guide. However, if you want to integrate more advanced AWS security & networking features, we’ll follow up with a new tutorial guideline.

At the end of this tutorial you’ll be able to;

1. [Create a security group](#create-an-ec2-security-group)
2. [Creating and configuring ECS Clusters](#2-creating-an-ecs-cluster)
3. [Creating and defining task definitions](#3-creating-and-running-task-definitions)
4. [Running our task definition](#4-running-our-task-definition)

## 1. Create an EC2 Security Group

So first thing first, let’s go over into security groups and create our security group. We’ll need this security group while creating our cluster.

![security-group-1](https://s3.us-west-2.amazonaws.com/secure.notion-static.com/1768516d-e0df-4041-a4d0-8681bc63672e/Security_Group.png?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Content-Sha256=UNSIGNED-PAYLOAD&X-Amz-Credential=AKIAT73L2G45EIPT3X45%2F20221218%2Fus-west-2%2Fs3%2Faws4_request&X-Amz-Date=20221218T104743Z&X-Amz-Expires=86400&X-Amz-Signature=8fbb96635d2a732099b2bfa0560fa0d0c8e7cb6e2a6eb5e35965d623c7f79ea6&X-Amz-SignedHeaders=host&response-content-disposition=filename%3D%22Security%2520Group.png%22&x-id=GetObject)

Search for “Security Groups” in the search bar. And go to the EC2 security groups feature. 

![security-group-2](https://s3.us-west-2.amazonaws.com/secure.notion-static.com/bdfa8271-b461-4c6c-ae84-2cb365aa9d8d/Security_Group_%28Create%29.png?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Content-Sha256=UNSIGNED-PAYLOAD&X-Amz-Credential=AKIAT73L2G45EIPT3X45%2F20221218%2Fus-west-2%2Fs3%2Faws4_request&X-Amz-Date=20221218T104855Z&X-Amz-Expires=86400&X-Amz-Signature=b77f57a949cd23a0a6c00c07d2e3767dde4ff18de53bbae07705243698bab6ca&X-Amz-SignedHeaders=host&response-content-disposition=filename%3D%22Security%2520Group%2520%28Create%29.png%22&x-id=GetObject)

Then start creating a new security group.

![security-group-3](https://s3.us-west-2.amazonaws.com/secure.notion-static.com/508bed40-c113-4259-bec5-4909dd01e3a8/Security_Group_%28Inbound_Rules%29.png?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Content-Sha256=UNSIGNED-PAYLOAD&X-Amz-Credential=AKIAT73L2G45EIPT3X45%2F20221218%2Fus-west-2%2Fs3%2Faws4_request&X-Amz-Date=20221218T104927Z&X-Amz-Expires=86400&X-Amz-Signature=454ed79eb65c16d88cd2684b3626bce773522b2fa739ebfc55da369445c70387&X-Amz-SignedHeaders=host&response-content-disposition=filename%3D%22Security%2520Group%2520%28Inbound%2520Rules%29.png%22&x-id=GetObject)

You have to name your security group, and give a description. Also, you need to choose the same VPC that you’ll going to use in EC2. So, I choose the default one. And I’m going to use same one while creating the ECS cluster.

The next step is to configure our inbound rules. Here’s the configuration;

```json
//for mapping HTTP request port.
type = "Custom TCP", protocol = "TCP", port_range = "3476",source = "Anywhere", ::/0

type = "Custom TCP", protocol = "TCP", port_range = "3476",source = "Anywhere", 0.0.0.0/0

//for mapping RPC request port.
type = "Custom TCP", protocol = "TCP", port_range = "3478",source = "Anywhere", ::/0

type = "Custom TCP", protocol = "TCP", port_range = "3476",source = "Anywhere", 0.0.0.0/0

//for using SSH for connecting from your local computer.
type = "Custom TCP", protocol = "TCP", port_range = "22",source = "Anywhere", 0.0.0.0/0
```

We have configured the HTTP and RPC ports for Permify. Also, we added port “22” for SSH connection. So, we can connect to EC2 through our local terminal.

Now, we’re good to go. You can create the security group. And it’s ready to use in our ECS.

## 2. Creating an ECS cluster

![create-ecs-cluster-1](https://s3.us-west-2.amazonaws.com/secure.notion-static.com/cbc11e77-b7c0-4dc2-98e5-fef28ea16933/Screenshot_2022-12-16_at_17.50.24.png?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Content-Sha256=UNSIGNED-PAYLOAD&X-Amz-Credential=AKIAT73L2G45EIPT3X45%2F20221218%2Fus-west-2%2Fs3%2Faws4_request&X-Amz-Date=20221218T105054Z&X-Amz-Expires=86400&X-Amz-Signature=f854ad54cf34f1c6fadc78bd9638fe74db8c2e0548ad3060f9d9e256c83cd778&X-Amz-SignedHeaders=host&response-content-disposition=filename%3D%22Screenshot%25202022-12-16%2520at%252017.50.24.png%22&x-id=GetObject)

The next step is to create an ECS cluster. From your AWS console search for Elastic Container Service or ECS for short.

![create-ecs-cluster-2](https://s3.us-west-2.amazonaws.com/secure.notion-static.com/3057fb96-bc44-4600-986c-779afc943001/Screenshot_2022-12-16_at_18.02.14.png?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Content-Sha256=UNSIGNED-PAYLOAD&X-Amz-Credential=AKIAT73L2G45EIPT3X45%2F20221218%2Fus-west-2%2Fs3%2Faws4_request&X-Amz-Date=20221218T105126Z&X-Amz-Expires=86400&X-Amz-Signature=22cb19ae1f141ade93cd826f7c1a3a88c7abf7da60409b233eba6a1f2d0a94be&X-Amz-SignedHeaders=host&response-content-disposition=filename%3D%22Screenshot%25202022-12-16%2520at%252018.02.14.png%22&x-id=GetObject)

Then go over the clusters. As you can see there are 2 types of clusters. One is for ECS and another for EKS. We need to use ECS, EKS stands for Elastic Kubernetes Service. Today we’re not going to cover Kubernetes.

Click **“Create Cluster”**

![create-ecs-cluster-3](https://s3.us-west-2.amazonaws.com/secure.notion-static.com/95b76d3a-a6e6-40fb-baa7-fc0385ba7fe7/Screenshot_2022-12-16_at_23.14.30.png?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Content-Sha256=UNSIGNED-PAYLOAD&X-Amz-Credential=AKIAT73L2G45EIPT3X45%2F20221218%2Fus-west-2%2Fs3%2Faws4_request&X-Amz-Date=20221218T105209Z&X-Amz-Expires=86400&X-Amz-Signature=5b419675be850cbdba6128267b7fc855b76e6207978bd83f00066b1f57d6e84d&X-Amz-SignedHeaders=host&response-content-disposition=filename%3D%22Screenshot%25202022-12-16%2520at%252023.14.30.png%22&x-id=GetObject
)

Let’s create our first Cluster. Simply you have 3 options; Serverless(Network Only), Linux, and Windows. We’re going to cover EC2 Linux + Networking option.

![create-ecs-cluster-4](https://s3.us-west-2.amazonaws.com/secure.notion-static.com/34677e26-58e9-4460-ae6d-89892920c8e6/Screenshot_2022-12-16_at_18.09.16.png?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Content-Sha256=UNSIGNED-PAYLOAD&X-Amz-Credential=AKIAT73L2G45EIPT3X45%2F20221218%2Fus-west-2%2Fs3%2Faws4_request&X-Amz-Date=20221218T105314Z&X-Amz-Expires=86400&X-Amz-Signature=9d5b6a4e61d43c16d3357690155b7b076cc43f0016cc1b51ef8421626c37b6f4&X-Amz-SignedHeaders=host&response-content-disposition=filename%3D%22Screenshot%25202022-12-16%2520at%252018.09.16.png%22&x-id=GetObject)

The next step is to configure our Cluster, starting with your Cluster name. Since we’re deploying Permify, I’ll call it “permify”.

Then choose your instance type. You can take a look at different instances and pricing from [here](https://aws.amazon.com/ec2/pricing/on-demand/). I’m going with the t4 large. For cost purposes, you can choose t2.micro if you’re just trying out. It’s free tier eligible.

Also, if you want to connect this EC2 instance from your local computer. You need to use SSH. Thus choose a key pair. If you have no such intention, leave it “none”.

![create-ecs-cluster-5](https://s3.us-west-2.amazonaws.com/secure.notion-static.com/285a8faf-38bf-4b3f-944d-cb66c041200a/Creating_Cluster_%28Networking%29.png?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Content-Sha256=UNSIGNED-PAYLOAD&X-Amz-Credential=AKIAT73L2G45EIPT3X45%2F20221218%2Fus-west-2%2Fs3%2Faws4_request&X-Amz-Date=20221218T105403Z&X-Amz-Expires=86400&X-Amz-Signature=ae4cc896df889050d9093dca4eb2a348b6bf0d778d1dc38fff1e528a53f5eb66&X-Amz-SignedHeaders=host&response-content-disposition=filename%3D%22Creating%2520Cluster%2520%28Networking%29.png%22&x-id=GetObject)

Now, we need to configure networking. First, choose your VPC, we use the default VPC as we did in the security groups. And choose any subnet on that VPC.

You want to enable auto-assigned IP to make your app reachable from the internet.

Choose the security group we have created previously.

And voila, you can create your cluster. Now, we need to run our container in this cluster. To do that, let’s go over task definitions. And create our container definition.

## 3. Creating and running task definitions

Go over to ECS, and click the task definitions.

![create-run-task-1](https://s3.us-west-2.amazonaws.com/secure.notion-static.com/41d8af12-afa0-4380-880b-ed38cddafba9/Screenshot_2022-12-16_at_23.26.55.png?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Content-Sha256=UNSIGNED-PAYLOAD&X-Amz-Credential=AKIAT73L2G45EIPT3X45%2F20221218%2Fus-west-2%2Fs3%2Faws4_request&X-Amz-Date=20221218T105513Z&X-Amz-Expires=86400&X-Amz-Signature=cfcf80dabf2d2f96028005a14661309aae469abec81815a665235ee15e710967&X-Amz-SignedHeaders=host&response-content-disposition=filename%3D%22Screenshot%25202022-12-16%2520at%252023.26.55.png%22&x-id=GetObject)

And create a new task definition.

![create-run-task-2](https://s3.us-west-2.amazonaws.com/secure.notion-static.com/8967ca83-7a10-459d-81ea-6454141211c8/Screenshot_2022-12-16_at_23.28.10.png?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Content-Sha256=UNSIGNED-PAYLOAD&X-Amz-Credential=AKIAT73L2G45EIPT3X45%2F20221218%2Fus-west-2%2Fs3%2Faws4_request&X-Amz-Date=20221218T105559Z&X-Amz-Expires=86400&X-Amz-Signature=1d79be6afa579db15756a19c0b7d49f5d7705c3c4e65e23d0258a58b211d51fd&X-Amz-SignedHeaders=host&response-content-disposition=filename%3D%22Screenshot%25202022-12-16%2520at%252023.28.10.png%22&x-id=GetObject)

Again, you’re going to ask to choose between; FARGATE, EC2, and EXTERNAL (On-premise). We’ll continue with EC2.

Leave everything in default under the “Configure task and container definitions” section.

![create-run-task-3](https://s3.us-west-2.amazonaws.com/secure.notion-static.com/36a74001-781b-4ee7-9fe3-f682906cd72d/Screenshot_2022-12-16_at_23.30.52.png?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Content-Sha256=UNSIGNED-PAYLOAD&X-Amz-Credential=AKIAT73L2G45EIPT3X45%2F20221218%2Fus-west-2%2Fs3%2Faws4_request&X-Amz-Date=20221218T105704Z&X-Amz-Expires=86400&X-Amz-Signature=c25588f444adb9e3bd9a5d991ee463b49e0d48acd5a173a3ae7fa95628222398&X-Amz-SignedHeaders=host&response-content-disposition=filename%3D%22Screenshot%25202022-12-16%2520at%252023.30.52.png%22&x-id=GetObject)

Under the IAM role section you can choose “ecsTaskExecutionRole” if you want to use Cloud Watch later.

You can leave task size in default since it’s optional for EC2.

The critical part over here is to add our container. Click on the “Add Container” button.

![create-run-task-4](https://s3.us-west-2.amazonaws.com/secure.notion-static.com/8d1935c2-9a9e-4809-a201-5c211062b3ef/Screenshot_2022-12-16_at_23.33.39.png?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Content-Sha256=UNSIGNED-PAYLOAD&X-Amz-Credential=AKIAT73L2G45EIPT3X45%2F20221218%2Fus-west-2%2Fs3%2Faws4_request&X-Amz-Date=20221218T105832Z&X-Amz-Expires=86400&X-Amz-Signature=bf1abbc5cb5d3b4a31bcc3e34dc9be4ecdd1d66314ac5af6dbf534955dc2ace8&X-Amz-SignedHeaders=host&response-content-disposition=filename%3D%22Screenshot%25202022-12-16%2520at%252023.33.39.png%22&x-id=GetObject)

Then we need to add our container details. First, give a name. And then the most important part is our image URI. Permify is registered on the Github Registry so our image is;

```yaml
ghcr.io/permify/permify:latest
```

Then we need to define memory limit for the container, I went with 1024. You can define as much as your instance allows.

Next step is to mapping our ports. As we mentioned in security groups, Permify by default listens;

- `3476 for HTTP port`
- `3478 for RPC port`

![create-run-task-5](https://s3.us-west-2.amazonaws.com/secure.notion-static.com/b5128c68-e546-473f-b6b5-ec8ad5f0acf0/Screenshot_2022-12-16_at_23.44.37.png?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Content-Sha256=UNSIGNED-PAYLOAD&X-Amz-Credential=AKIAT73L2G45EIPT3X45%2F20221218%2Fus-west-2%2Fs3%2Faws4_request&X-Amz-Date=20221218T105957Z&X-Amz-Expires=86400&X-Amz-Signature=cfb0e23d2017fac3b2deeb4f6052b1103f2c1e873e0d9980e4547e4aaa0d37cf&X-Amz-SignedHeaders=host&response-content-disposition=filename%3D%22Screenshot%25202022-12-16%2520at%252023.44.37.png%22&x-id=GetObject)

Then we need to define command under the environment section. So, in order to start permify we first need to add “serve” command.

For using properly we need a few other. Here’s the commands we need.

```yaml
serve, --database-engine=postgres, --database-name=<db_name>, --database-uri=postgres://<user_name>:<password>@<db_endpoint>:<db_port>, --database-pool-max=20
```

- `serve` ⇒ for starting the Permify.
- `--database-engine=postgres` ⇒ for defining the db we use.
- `--database-name=<database_name>` ⇒ name of the database you use.
- `--database-uri=postgres://<user_name>:password@<db_endpoint>:<db_port>` ⇒ for connecting your database with URI.
- `--database-pool-max=20` ⇒ the depth for running in graph.

We’re nice and clear, add the container and then just create your task definition. We’ll use this definition to run in our cluster.

So, let’s go over and run our task definition.

## 4. Running our task definition

![run-task-definition-1](https://s3.us-west-2.amazonaws.com/secure.notion-static.com/0b92b775-fff9-499c-9f14-e93a57db5331/Screenshot_2022-12-17_at_00.36.54.png?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Content-Sha256=UNSIGNED-PAYLOAD&X-Amz-Credential=AKIAT73L2G45EIPT3X45%2F20221218%2Fus-west-2%2Fs3%2Faws4_request&X-Amz-Date=20221218T110225Z&X-Amz-Expires=86400&X-Amz-Signature=561d6e66e1e1c7a9e70f3e0449d55735ff3002854bf4729c56ed7cdbb566db02&X-Amz-SignedHeaders=host&response-content-disposition=filename%3D%22Screenshot%25202022-12-17%2520at%252000.36.54.png%22&x-id=GetObject)

Let’s go to ECS and enter into our cluster. And go over into the tasks to run our task.

![run-task-definition-2](https://s3.us-west-2.amazonaws.com/secure.notion-static.com/e110a801-fdd8-49e7-8f24-b54727c0d254/Screenshot_2022-12-17_at_00.42.57.png?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Content-Sha256=UNSIGNED-PAYLOAD&X-Amz-Credential=AKIAT73L2G45EIPT3X45%2F20221218%2Fus-west-2%2Fs3%2Faws4_request&X-Amz-Date=20221218T110303Z&X-Amz-Expires=86400&X-Amz-Signature=d62b7bc18e00de3e2c8f044d764b116de237f4cf97fe08608f8e752ba2390e82&X-Amz-SignedHeaders=host&response-content-disposition=filename%3D%22Screenshot%25202022-12-17%2520at%252000.42.57.png%22&x-id=GetObject)

Click to “Run new Task”

![run-task-definition-3](https://s3.us-west-2.amazonaws.com/secure.notion-static.com/a7b18288-05df-4a49-b3cc-d79172ffb8c9/Screenshot_2022-12-17_at_00.44.45.png?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Content-Sha256=UNSIGNED-PAYLOAD&X-Amz-Credential=AKIAT73L2G45EIPT3X45%2F20221218%2Fus-west-2%2Fs3%2Faws4_request&X-Amz-Date=20221218T110328Z&X-Amz-Expires=86400&X-Amz-Signature=feb3585d58af8028202c17b3eb109bac66ab40378cb20b20e94b5a405ba98150&X-Amz-SignedHeaders=host&response-content-disposition=filename%3D%22Screenshot%25202022-12-17%2520at%252000.44.45.png%22&x-id=GetObject)

Choose EC2 as a launch type. Then pick the task definition we just created. And leave everything else in the default. You can run your task now.

We have just deployed our container into EC2 instance with ECS. Let’s test it.

Now you can go over into EC2, and click on the running instances. Find the instance named `ECS Instance - EC2ContainerService-<cluster_name>` in the running instances.

![run-task-definition-3](https://s3.us-west-2.amazonaws.com/secure.notion-static.com/f5973284-3439-4ca9-b51c-087205160558/Screenshot_2022-12-17_at_00.51.40.png?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Content-Sha256=UNSIGNED-PAYLOAD&X-Amz-Credential=AKIAT73L2G45EIPT3X45%2F20221218%2Fus-west-2%2Fs3%2Faws4_request&X-Amz-Date=20221218T110410Z&X-Amz-Expires=86400&X-Amz-Signature=553ff2b34d89b60909c3fab2ce8d22279309779b093375921fc505379d697408&X-Amz-SignedHeaders=host&response-content-disposition=filename%3D%22Screenshot%25202022-12-17%2520at%252000.51.40.png%22&x-id=GetObject)

Copy the Public IPv4 DNS from the right corner, and paste it into your browser. But you need to add `:3476` to access our http endpoint. So it should be like this;

`<public_IPv4_DNS>:3476`

and if you add healthz at the end like this;

`<public_IPv4_DNS>:3476/healthz`

you should get Serving status :)

![run-task-definition-4](https://s3.us-west-2.amazonaws.com/secure.notion-static.com/b65f586f-4a0f-4a7c-97b7-9e8fac4caadd/Screenshot_2022-12-17_at_00.55.13.png?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Content-Sha256=UNSIGNED-PAYLOAD&X-Amz-Credential=AKIAT73L2G45EIPT3X45%2F20221218%2Fus-west-2%2Fs3%2Faws4_request&X-Amz-Date=20221218T110444Z&X-Amz-Expires=86400&X-Amz-Signature=6d8d024f3b3c9ee8e644415d995438d9a1d3a4786903701e6033cb25e4f3a61e&X-Amz-SignedHeaders=host&response-content-disposition=filename%3D%22Screenshot%25202022-12-17%2520at%252000.55.13.png%22&x-id=GetObject)

## Need any help ?

Our team is happy to help you to deploy Permify, [schedule a call with an Permify engineer](https://meetings-eu1.hubspot.com/ege-aytin/call-with-an-expert).