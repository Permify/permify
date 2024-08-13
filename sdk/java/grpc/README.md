# Permify Java SDK

This repository contains a sample usage for the Java gRPC SDK for Permify.

## Getting Started

### Prerequisites

Ensure you have the following installed:
- Java
- Maven

### Cloning the Repository

To get started, clone the repository using the following command:

```sh
git clone https://github.com/ucatbas/permify-sdk-samples.git
cd permify-sdk-samples/java/grpc
```

### Building the Project

Navigate to the root directory of the project and run the following commands to build and install the SDK:

```sh
mvn clean install
```

### Running the Application

After successfully building the project, you can run the application using the following command:

```sh
java -jar target/your-artifact-name-0.0.1.jar
```

Replace your-artifact-name-0.0.1.jar with the actual name of the jar file generated after the build.

## For your own Projects

To use the Permify SDK in your project, add the following dependency to your pom.xml file:

```xml
<dependency>
    <groupId>build.buf.gen</groupId>
    <artifactId>permifyco_permify_grpc_java</artifactId>
    <version>1.65.1.1.20240628085453.215bbf832f82</version>
</dependency>

<dependency>
    <groupId>build.buf.gen</groupId>
    <artifactId>permifyco_permify_protocolbuffers_java</artifactId>
    <version>27.2.0.1.20240628085453.215bbf832f82</version>
</dependency>
```

Here is a simple permify client:

```java

  // Create the channel
  ManagedChannel channel = ManagedChannelBuilder.forAddress("127.0.0.1", 3478).usePlaintext().build();

  try {
      // Create the blocking stub
      TenancyGrpc.TenancyBlockingStub blockingStub = TenancyGrpc.newBlockingStub(channel);

      String timeStamp = new SimpleDateFormat("yyyy.MM.dd.HH.mm.ss").format(new Date());
      TenantCreateRequest req = TenantCreateRequest.newBuilder()
              .setId("tenant_" + timeStamp)
              .setName("tenant id name")
              .build();

      TenantCreateResponse response = blockingStub.create(req);
      System.out.println(response);

  } finally {
      // Gracefully shutdown the channel
      try {
          channel.shutdown().awaitTermination(5, TimeUnit.SECONDS);
      } catch (InterruptedException e) {
          e.printStackTrace();
          channel.shutdownNow();
      }
  }
  
```
