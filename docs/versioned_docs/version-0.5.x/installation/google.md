---
title: Deploy on Google Compute Engine
---

This guide outlines the process of deploying Permify, on Google Compute Engine. The steps include setting up Google Cloud SDK and kubectl, managing containers using Google Kubernetes Engine (GKE), deploying Permify, and implementing Permify in a distributed configuration with Serf. By following these steps, you can efficiently deploy Permify on Google's scalable and secure infrastructure.

## Google Cloud SDK Install

1. At the command line, run the following command:

   ```bash
   curl https://sdk.cloud.google.com | bash
   ```
   
2. When prompted, choose a location on your file system (usually your Home directory) to create the `google-cloud-sdk` subdirectory under.
3. If you want to send anonymous usage statistics to help improve gcloud CLI, answer `Y` when prompted.
4. To add gcloud CLI command-line tools to your `PATH` and enable command completion, answer `Y` when prompted
5. Restart your shell:

   ```bash
   exec -l $SHELL
   ```

6. To initialize the Google Cloud CLI environment, run `gcloud init`

## Install kubectl

1. Install the `kubectl` component:

    ```bash
    gcloud components install kubectl
    ```

2. Verify that `kubectl` is installed:

    ```bash
    kubectl version
    ```

3. Install Authn Plug-in

    ```bash
    gcloud components install gke-gcloud-auth-plugin
    ```

   Check the `gke-gcloud-auth-plugin` binary version:

    ```bash
    gke-gcloud-auth-plugin --version
    ```


## Create Containers with GKE

1. Login & Initialize Google Cloud CLI

    ```bash
    gcloud init
    ```

2. Follow configuration instructions
3. Create Container Cluster

    ```bash
    gcloud container clusters create [CLUSTER_NAME]
    ```

4. Authenticate the cluster

    ```bash
    gcloud container clusters get-credentials [CLUSTER_NAME]
    ```


## Deploy Permify

1. Apply deployment config

    ```bash
    kubectl apply -f deployment.yaml
    ```

   - **Deployment.yaml**

       ```yaml
       apiVersion: apps/v1
       kind: Deployment
       metadata:
           labels:
               app: permify
           name: permify
       spec:
         replicas: 3
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
                   - "--database-uri=postgres://user:password@host:5432/db_name"
                   - "--database-max-open-connections=20"
                   ports:
                       - containerPort: 3476
                         protocol: TCP
                   resources: {}
               restartPolicy: Always
       status: {}
       ```

2. Apply service manfiest

    ```bash
    kubectl apply -f service.yaml
    ```

   - **Service Manifest**

       ```yaml
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


## Deploying Permify in a Distributed Configuration

If you aim to deploy Permify in a distributed configuration, you will need to create a Serf deployment. The Serf deployment can be dockerized to our Container Registry under the name permify/serf:v1.0, which is provided by Hashicorp.

Please note: It is crucial to ensure that both Serf and Permify deployments reside within the same namespace for proper operation.

1. Serf Service Create:
   - Serf Deployment&Service yaml

       ```yaml
       apiVersion: apps/v1
       kind: Deployment
       metadata:
         name: serf-deployment
       spec:
         replicas: 1
         selector:
           matchLabels:
             app: serf
         template:
           metadata:
             labels:
               app: serf
           spec:
             containers:
             - name: serf
               image: permify/serf:v1.0
               args: 
                - "-node=main-serf"
               ports:
               - containerPort: 7946
               resources:
                 requests:
                   cpu: 100m
                   memory: 128Mi
                 limits:
                   cpu: 200m
                   memory: 256Mi
       ---
       apiVersion: v1
       kind: Service
       metadata:
         name: serf
       spec:
         selector:
           app: serf
         ports:
         - protocol: TCP
           port: 7946
           targetPort: 7946
           name: serf
         type: ClusterIP
       ```

2. Apply Deployment Manifest
   - Deployment.yaml

       ```yaml
       apiVersion: apps/v1
       kind: Deployment
       metadata:
         name: permify-deployment
       spec:
         replicas: 3
         selector:
           matchLabels:
             app: permify
         template:
           metadata:
             labels:
               app: permify
           spec:
             containers:
               - image: permify/permify:tagname
                 name: permify
                 args:
                   - "serve"
                   - "--database-engine=postgres"
                   - "--database-uri=postgres://user:password@host:5432/db_name"
                   - "--database-max-open-connections=20"
                   - "--distributed-enabled=true"
                   - "--distributed-node=serf:7946"
                   - "--distributed-node-name=main-serf"
                   - "--distributed-protocol=serf"
                 resources:
                    requests:
                      memory: "128Mi"
                      cpu: "200m"
                    limits:
                     memory: "128Mi"
                     cpu: "400m"
                 ports:
                 - containerPort: 3476
                   name: permify-port
                 - containerPort: 7946
                   name: permify-dist
                 - containerPort: 6060
                   name: permify-pprof
       ```

3. Apply Service Manifest
   - Service.yaml

       ```yaml
       apiVersion: v1
       kind: Service
       metadata:
         name: permify
       spec:
         ports:
             - name: permify-port
               port: 3476
               targetPort: 3476
             - name: permify-dist
               port: 7946
               targetPort: 7946
         selector:
               app: permify
         type: LoadBalancer
       ```


## Need any help ?

Our team is happy to help you to deploy Permify, [schedule a call with an Permify engineer](https://meetings-eu1.hubspot.com/ege-aytin/call-with-an-expert).