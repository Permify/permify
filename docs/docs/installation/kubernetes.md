---
title: Kubernetes Cluster
---

#  Deploy on Kubernetes Cluster

In this section we’re going to deploy Permify in AWS EKS which is Amazon Elastic Kubernetes Service. EKS is a managed service that you can easily run Kubernetes in AWS.

Here’s what we’re going to do step-by-step;

1. [Configure our AWS IAM credentials](#configure-aws-cli-with-your-iam-account)
3. [Create EKS cluster and configure nodes](#creating-an-aws-eks-cluster)
4. [Deploy Permify to nodes](#deploying--running-permify-in-nodes)

There are a couple of small prerequisites for this tutorial.

### Pre-requisites

- An AWS account.
- The AWS Command Line Interface (CLI) is installed and configured on your local machine. — [Click here](https://us-east-1.console.aws.amazon.com/iamv2/home?region=us-east-1#/home) to go to IAM
- The AWS IAM Authenticator for Kubernetes is installed and configured on your local machine.

## Configure AWS CLI with your IAM account.

The first step is to configure our AWS IAM account into our local terminal so that we can run commands. Most of you probably have a configured AWS account if you ever set up anything into AWS programmatically, so you can skip this. If you don’t follow these steps.

### Create an AWS IAM Programmatic Access Account

First, let’s create IAM credentials for ourselves. Search IAM from the AWS console. You need to write down the account ID if you want to log in AWS console with this account as well. Let’s go over users and start creating our credentials.

![kubernetes-1](https://user-images.githubusercontent.com/34595361/211697636-6e106115-bd68-4909-aea0-5a7b6f8d5e18.png)

At Users screen click to “Add users” — and you’ll end up in your first screen creating user credentials. Here you can define the name of the user. Also there 2 options that you can choose simultaneously.

But you must choose “Access key - Programmatic access” option. It’ll allow us to configure our AWS CLI on our local machine.

You can also choose “Password - AWS Management Console access” if you want to log in to this account through the console. But you’ll need the Account ID that I mentioned in the IAM console screen.

In the next screen, you’ll be asked to create or copy the user-set permissions. For this tutorial, you’ll only need to access EKS resources and features. So lets create group by clicking the “Create group” — and then at pop-up screen search for EKS.

![kubernetes-2](https://user-images.githubusercontent.com/34595361/211697647-f39d73e7-b6e2-40ae-8c3b-ad68032d6b21.png)

I’ll choose all EKS permissions but if you have certain policies internally, just stick with them. You’ll only need following permission to;

- `AmazonEKSClusterPolicy`
- `AmazonEKSServicePolicy`
- `AmazonEKSVPCResourceController`
- `AmazonEKSWorkerNodePolicy`

Then simply you can review and create the user.

![kubernetes-4](https://user-images.githubusercontent.com/34595361/211697655-1b75d4f9-a2ee-4b7e-9e1e-0be0b5aaad7d.png)

Once you created the credentials you’ll prompt the “Access key ID” and “Secret access key”, you should save this down somewhere. We’re going the use these to configure our local machine with AWS CLI.

### **Configure AWS CLI with your IAM account**

Let’s open our local terminal

```jsx
aws configure
```

Next you’ll ask for the following credentials;

- `AWS Access Key ID`
- `AWS Secret Access Key`
- `Default region name`
- `Default output format` (leave it empty)

## Creating an AWS EKS Cluster

For the first step, we need to install [eksctl](https://eksctl.io/) — which is like kubectl but for AWS EKS. It helps us to set up and deploy our cluster and nodes within a fraction of the time.

Let’s download eksctl using brew. 


```jsx
brew tap weaveworks/tap
```

While installing the eksctl, we’ll end up getting kubectl and other dependencies.

```jsx
brew install weaveworks/tap/eksctl
```

Now, we’re ready to create our EKS cluster. You can define certain things while deploying standard the cluster beside the name and version like; the region you want to deploy, the EC2 instance type of each node, and the number of nodes you want to run.

```bash
eksctl create cluster \
--name <your-cluster-name> \
--version 1.24 \
--region <region-of-choice> \
--nodegroup-name permify \
--node-type t2.small \
--nodes 2
```

## Deploying & Running Permify in Nodes

The next stop is applying our manifests which will help us to deploy and configure our container/Permify. 

Let’s create our deployment manifest first.

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
    labels:
        app: permify
    name: permify
spec:
  replicas: 2
  selector:
    matchLabels:
      app: permify
  strategy:
      type: Recreate
  template:
      metadata:
        labels:
          app: permify
      spec:
        containers:
          - image: ghcr.io/permify/permify
            name: permify
            args:
            - "serve"
            - "--database-engine=postgres"
            - "--database-uri=postgres://postgres:nOcodeSTIAnLAba@permify-test.ceuo5kqsxyea.us-east-1.rds.amazonaws.com:5432/demo"
            - "--database-max-open-connections=20"
            ports:
                - containerPort: 3476
                  protocol: TCP
            resources: {}
        restartPolicy: Always
status: {}
```

Now let’s apply our deployment manifest

```jsx
kubectl apply -f deployment.yaml
```

The next step is to create a service manifest, this will allow us to configure our container app.

```jsx
apiVersion: v1
kind: Service
metadata:
  name: permify
spec:
  ports:
      - name: 3476-tcp
        port: 3476
        protocol: TCP
        targetPort: 3476
  selector:
        app: permify
  type: LoadBalancer
status:
  loadBalancer: {}
```

Let’s apply service.yaml to our nodes.

```jsx
kubectl apply -f service.yaml
```

Last but not least, we can check our pods & nodes. And we can start using the container with load balancer