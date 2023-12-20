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

![security-group-1](https://user-images.githubusercontent.com/34595361/208877994-e9461acc-4ffd-4591-b43e-db254366d25d.png)

Search for “Security Groups” in the search bar. And go to the EC2 security groups feature. 

![security-group-2](https://user-images.githubusercontent.com/34595361/208877493-ab11228c-1aa0-4bc5-b41d-4527737028e9.png)

Then start creating a new security group.

![security-group-3](https://user-images.githubusercontent.com/34595361/208877500-2c299883-6107-4b70-aa96-0f28eb00cf3d.png)

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

![create-ecs-cluster-1](https://user-images.githubusercontent.com/34595361/208878666-98c5d3ce-b079-444d-bc66-53f13038a08a.png)

The next step is to create an ECS cluster. From your AWS console search for Elastic Container Service or ECS for short.

![create-ecs-cluster-2](https://user-images.githubusercontent.com/34595361/208878675-2f266cfc-defb-4c7f-9186-b4de39f1743b.png)

Then go over the clusters. As you can see there are 2 types of clusters. One is for ECS and another for EKS. We need to use ECS, EKS stands for Elastic Kubernetes Service. Today we’re not going to cover Kubernetes.

Click **“Create Cluster”**

![create-ecs-cluster-3](https://user-images.githubusercontent.com/34595361/208878685-3edac67b-5b3d-4f0d-b2f7-70a5ec2e4870.png)

Let’s create our first Cluster. Simply you have 3 options; Serverless(Network Only), Linux, and Windows. We’re going to cover EC2 Linux + Networking option.

![create-ecs-cluster-4](https://user-images.githubusercontent.com/34595361/208878681-d98a77db-16b1-42af-a697-3036cc604c85.png)

The next step is to configure our Cluster, starting with your Cluster name. Since we’re deploying Permify, I’ll call it “permify”.

Then choose your instance type. You can take a look at different instances and pricing from [here](https://aws.amazon.com/ec2/pricing/on-demand/). I’m going with the t4 large. For cost purposes, you can choose t2.micro if you’re just trying out. It’s free tier eligible.

Also, if you want to connect this EC2 instance from your local computer. You need to use SSH. Thus choose a key pair. If you have no such intention, leave it “none”.

![create-ecs-cluster-5](https://user-images.githubusercontent.com/34595361/208878989-801839f5-8fce-4410-99e0-0a2dcccb47fa.png)

Now, we need to configure networking. First, choose your VPC, we use the default VPC as we did in the security groups. And choose any subnet on that VPC.

You want to enable auto-assigned IP to make your app reachable from the internet.

Choose the security group we have created previously.

And voila, you can create your cluster. Now, we need to run our container in this cluster. To do that, let’s go over task definitions. And create our container definition.

## 3. Creating and running task definitions

Go over to ECS, and click the task definitions.

![create-run-task-1](https://user-images.githubusercontent.com/34595361/208879726-fe5aac07-16a8-4f8c-9cc9-1c95ca191a42.png)

And create a new task definition.

![create-run-task-2](https://user-images.githubusercontent.com/34595361/208879733-e9aa6fa4-9f66-44e4-8c70-dfa0e33c1b73.png)

Again, you’re going to ask to choose between; FARGATE, EC2, and EXTERNAL (On-premise). We’ll continue with EC2.

Leave everything in default under the “Configure task and container definitions” section.

![create-run-task-3](https://user-images.githubusercontent.com/34595361/208879735-789ec411-5829-47be-9634-c09c7b0c0320.png)

Under the IAM role section you can choose “ecsTaskExecutionRole” if you want to use Cloud Watch later.

You can leave task size in default since it’s optional for EC2.

The critical part over here is to add our container. Click on the “Add Container” button.

![create-run-task-4](https://user-images.githubusercontent.com/34595361/208879740-4515e884-1efd-46fd-8e8c-cfa86634b673.png)

Then we need to add our container details. First, give a name. And then the most important part is our image URI. Permify is registered on the Github Registry so our image is;

```yaml
ghcr.io/permify/permify:latest
```

Then we need to define memory limit for the container, I went with 1024. You can define as much as your instance allows.

Next step is to mapping our ports. As we mentioned in security groups, Permify by default listens;

- `3476 for HTTP port`
- `3478 for RPC port`

![create-run-task-5](https://user-images.githubusercontent.com/34595361/208879746-5991a04c-73d5-4e35-97b0-67aa9ebf61fc.png)

Then we need to define command under the environment section. So, in order to start permify we first need to add “serve” command.

For using properly we need a few other. Here’s the commands we need.

```yaml
serve, --database-engine=postgres, --database-uri=postgres://<user_name>:<password>@<db_endpoint>:<db_port>/<db_name>, --database-pool-max=20
```

- `serve` ⇒ for starting the Permify.
- `--database-engine=postgres` ⇒ for defining the db we use.
- `--database-uri=postgres://<user_name>:password@<db_endpoint>:<db_port>/<db_name>` ⇒ for connecting your database with URI.
- `--database-pool-max=20` ⇒ the depth for running in graph.

We’re nice and clear, add the container and then just create your task definition. We’ll use this definition to run in our cluster.

So, let’s go over and run our task definition.

## 4. Running our task definition

![run-task-definition-1](https://user-images.githubusercontent.com/34595361/208880326-c5ecb48c-e210-47f8-bd92-d1f789be24ff.png)

Let’s go to ECS and enter into our cluster. And go over into the tasks to run our task.

![run-task-definition-2](https://user-images.githubusercontent.com/34595361/208880332-97a5732d-bc7d-401e-bae9-216d4273c5bf.png)

Click to “Run new Task”

![run-task-definition-3](https://user-images.githubusercontent.com/34595361/208880335-b3ce229f-33ff-4f03-90e7-6d6a306928ae.png)

Choose EC2 as a launch type. Then pick the task definition we just created. And leave everything else in the default. You can run your task now.

We have just deployed our container into EC2 instance with ECS. Let’s test it.

Now you can go over into EC2, and click on the running instances. Find the instance named `ECS Instance - EC2ContainerService-<cluster_name>` in the running instances.

![run-task-definition-4](https://user-images.githubusercontent.com/34595361/208880339-a508354c-99ee-4219-8ace-1c7fdbbe90ed.png)

Copy the Public IPv4 DNS from the right corner, and paste it into your browser. But you need to add `:3476` to access our http endpoint. So it should be like this;

`<public_IPv4_DNS>:3476`

and if you add healthz at the end like this;

`<public_IPv4_DNS>:3476/healthz`

you should get Serving status :)

![run-task-definition-5](https://user-images.githubusercontent.com/34595361/208880346-d19a6877-3013-4347-86c9-9f865b8a3e3c.png)

## Need any help ?

Our team is happy to help you to deploy Permify, [schedule a call with an Permify engineer](https://meetings-eu1.hubspot.com/ege-aytin/call-with-an-expert).