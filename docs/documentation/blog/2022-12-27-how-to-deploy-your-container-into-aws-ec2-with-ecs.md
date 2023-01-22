---
title: "How to Deploy Your Container into AWS EC2 with ECS"
description: "AWS is a piece of cake no one ever said! That’s why today we’re bringing this tutorial to help you deploy your container into AWS EC2 with Elastic Container Service."
slug: how-to-deploy-your-container-into-aws-ec2-with-ecs
authors:
  - name: Fred Dogan
    image_url: https://user-images.githubusercontent.com/34595361/213848632-4a98f25b-df49-4ee1-ab53-785de24c8388.jpeg
    title: Permify Core Team
    email: firat@permify.co
tags: ["aws, deployment, cloud, docker, kubernetes, ecs, ec2"]
image: https://user-images.githubusercontent.com/34595361/213841676-f2c0505d-c1d4-4c86-bf90-d6365fb5adca.png
hide_table_of_contents: false
---

![How to Deploy your container in AWS EC2 with ECS (Cover)](https://user-images.githubusercontent.com/34595361/213841676-f2c0505d-c1d4-4c86-bf90-d6365fb5adca.png)

AWS is a piece of cake no one ever said! That’s why today we’re bringing this tutorial to help you deploy your container into AWS EC2 with Elastic Container Service.

<!--truncate-->

ECS is a container management service. You can run your containers as task definitions, and It’s one of the easiest ways to deploy containers.

There is no prerequisite in this tutorial. We’ll deploy Permify today, which is an open-source authorization service for building scalable access control and permissions systems in your application.

You can simply deploy any container by following this step-by-step guide. However, if you want to integrate more advanced AWS security & networking features, we’ll follow up with a new tutorial guideline.

At the end of this tutorial you’ll be able to;

1. Create a security group.

2. Creating and configuring ECS Clusters.

3. Creating and defining task definitions.

4. Running our task definition.

## 1. Create an EC2 Security Group

So first thing first, let’s go over into security groups and create our security group. We’ll need this security group while creating our cluster.

![Security Group](https://user-images.githubusercontent.com/34595361/213841794-724cedb4-cacf-4e41-809f-8f35541999d2.png)

Search for “Security Groups” in the search bar. And go to the EC2 security groups feature. 

![Security Group (Create)](https://user-images.githubusercontent.com/34595361/213841792-d9fae458-623b-4504-a45e-8a5972d708c9.png)

Then start creating a new security group.

![Security Group (Inbound Rules)](https://user-images.githubusercontent.com/34595361/213841793-c215d4f5-879a-4d6c-9037-f9f89ca27682.png)

You have to name your security group, and give a description. Also, you need to choose the same VPC that you’ll going to use in EC2. So, I choose the default one. And I’m going to use same one while creating the ECS cluster.

The next step is to configure our inbound rules. Here’s the configuration:

```bash
//for mapping HTTP request port.
type = "Custom TCP", protocol = "TCP", port_range = "3476",source = "Anywhere", ::/0

type = "Custom TCP", protocol = "TCP", port_range = "3476",source = "Anywhere", 0.0.0.0/0

//for mapping RPC request port.
type = "Custom TCP", protocol = "TCP", port_range = "3478",source = "Anywhere", ::/0

type = "Custom TCP", protocol = "TCP", port_range = "3476",source = "Anywhere", 0.0.0.0/0

//for using SSH for connecting from your local computer.
type = "Custom TCP", protocol = "TCP", port_range = "22",source = "Anywhere", 0.0.0.0/0
```

We have configured the HTTP and RPC ports for Permify. Also, we added port “22” for SSH connection. So, we can connect to EC2 through our local terminal. Now, we’re good to go. You can create the security group. And it’s ready to use in our ECS.

## 2. Creating an ECS cluster

![Screenshot 2022-12-16 at 17 50 24](https://user-images.githubusercontent.com/34595361/213841827-db08bbe0-fe60-43cd-bd7e-441e94bd66cc.png)

The next step is to create an ECS cluster. From your AWS console search for Elastic Container Service or ECS for short.

![Screenshot 2022-12-16 at 18 02 14](https://user-images.githubusercontent.com/34595361/213841828-574d1a54-d45a-4478-afe9-3ed1ab198b8f.png)

Then go over the clusters. As you can see there are 2 types of clusters. One is for ECS and another for EKS. We need to use ECS, EKS stands for Elastic Kubernetes Service. Today we’re not going to cover Kubernetes.

Click “Create Cluster”

![Screenshot 2022-12-16 at 23 14 30](https://user-images.githubusercontent.com/34595361/213841830-5177a912-8941-4166-97e1-6ea769adffab.png)

Let’s Create our first Cluster. Simply you have 3 options; Serverless(Network Only), Linux, and Windows. We’re going to cover EC2 Linux + Networking option.

![Screenshot 2022-12-16 at 18 09 16](https://user-images.githubusercontent.com/34595361/213841829-202f498f-2d10-4b44-8ab7-1c4c8563aecf.png)


The next step is to configure our Cluster, starting with your Cluster name. Since we’re deploying Permify, I’ll call it “permify”.

Then choose your instance type. You can take a look at different instances and pricing from here. I’m going with the t4 large. For cost purposes, you can choose t2.micro if you’re just trying out. It’s free tier eligible.


Also, if you want to connect this EC2 instance from your local computer. You need to use SSH. Thus choose a key pair. If you have no such intention, leave it “none”.

![Creating Cluster (Networking)](https://user-images.githubusercontent.com/34595361/213841826-a07c9243-e631-4fee-9556-357a4c8ff61e.png)

Now, we need to configure networking. First, choose your VPC, we use the default VPC as we did in the security groups. And choose any subnet on that VPC.

You want to enable auto-assigned IP if you want to make your app reachable from the internet.

Choose the security group we have created previously.

And voila, you can create your cluster. Now, we need to run our container in this cluster. To do that, let’s go over task definitions. And create our container definition.

## 3. Creating and running task definitions
Go over to ECS, and click the task definitions.

![3-1](https://user-images.githubusercontent.com/34595361/213842064-e4bbf7d3-8286-4141-844c-6a3110ecad03.png)


And create a new task definition.

![3-2](https://user-images.githubusercontent.com/34595361/213842065-e09fd919-2dd2-432d-baf1-ad5ef0197e0b.png)

Again, you’re going to ask to choose between; FARGATE, EC2, and EXTERNAL (On-premise). We’ll continue with EC2.

Leave everything in default under the “Configure task and container definitions” section.

![3-3](https://user-images.githubusercontent.com/34595361/213842066-c621dea3-aa13-4a78-8848-a5e695526c78.png)

Under the IAM role section you can choose “ecsTaskExecutionRole” if you want to use Cloud Watch later.

You can leave task size in default since it’s optional for EC2.

The critical part over here is to add our container. Click on the “Add Container” button.

![3-4](https://user-images.githubusercontent.com/34595361/213842067-1d970960-a931-4ed5-aa83-c264ef690f20.png)

Then we need to add our container details. First, give a name. And then the most important part is our image URI. Permify is registered on the Github Registry so our image URI is;

```bash
ghcr.io/permify/permify:latest
```

Then we need to define memory limit for the container, I went with 1024. You can define as much as your instance allows.

Next step is to mapping our ports. As we mentioned in security groups, Permify by default listens;

- 3476 for HTTP port
- 3478 for RPC port

![3-5](https://user-images.githubusercontent.com/34595361/213842068-e64dcde8-1ee2-4815-9fa7-07fe6c4566d2.png)

Then we need to define command under the environment section. So, in order to start permify we first need to add “serve” command.

For using properly we need a few other. Here’s the commands we need.

```bash
serve, --database-engine=postgres, --database-name=<db_name>, --database-uri=postgres://<user_name>:<password>@<db_endpoint>:<db_port>, --database-pool-max=20

`serve` ⇒ for starting the Permify.
`--database-engine=postgres` ⇒ for defining the db we use.
`--database-name=<database_name>` ⇒ name of the database you use.
`--database-uri=postgres://<user_name>:password@<db_endpoint>:<db_port>` ⇒ for connecting your database with URI.
`--database-pool-max=20` ⇒ the depth for running in the graph.
```

view raw command under the environment section hosted with ❤ by GitHub
We’re nice and clear, add the container and then just create your task definition. We’ll use this definition to run in our cluster.

So, let’s go over and run our task definition.

## 4. Running our task definition

![4-1](https://user-images.githubusercontent.com/34595361/213842157-28b7f33f-4009-4568-ab98-9246398ed3aa.png)

Let’s go to ECS and enter into our cluster. And go over into the tasks to run our task.

![4-2](https://user-images.githubusercontent.com/34595361/213842158-ca729f1e-dbff-47c4-b1f5-f36224f69689.png)

Click to “Run new Task”

![4-3](https://user-images.githubusercontent.com/34595361/213842159-5a07cefb-c8df-47ea-a459-c624c1587266.png)

Choose EC2 as a launch type. Then pick the task definition we just created. And leave everything else in the default. You can run your task now.

We have just deployed our container into an EC2 instance with ECS. Let’s test it.

Now you can go over into EC2, and click on the running instances. Find the instance named ECS Instance:

```bash
EC2ContainerService-<cluster_name>
```

in the running instances.
![4-4](https://user-images.githubusercontent.com/34595361/213842160-cf9d9473-d318-4df6-bab6-8cfeff5734dc.png)

Copy the Public IPv4 DNS from the right corner, and paste it into your browser. But you need to add :3476 to access our http endpoint. So it should be like this;

```bash
<public_IPv4_DNS>:3476
```

and if you add healthz at the end like this;
```bash
<public_IPv4_DNS>:3476/healthz
```

you should get Serving status :)

![4-5](https://user-images.githubusercontent.com/34595361/213842161-8fbb5b02-7d06-45d9-8aef-c0a80753f945.png)

And it’s nice and clear! You can start using Permify with all the endpoints now. For more check our repo and documentation.

Here’s what we had today.

1. We created a security group to reveal 3476 for our HTTP port, 3478 for our RPC port, and 22 for our SSH connection.
2. Then we created an ECS cluster to initialize the EC2 instance and deploy our container into it.
3. We created and defined a task definition to run our container in ECS.
4. And we run our task to deploy our container into an EC2 instance with ECS.

I hope you find this tutorial useful. If you have any questions or suggestions about this tutorial or Permify, feel free to reach me out at ***firat@permify.co***





