# Permify Go SDK

This repository contains a sample usage for the Go gRPC SDK for Permify.

## Getting Started

### Prerequisites

Ensure you have the following installed:
- go

### Cloning the Repository

To get started, clone the repository using the following command:

```sh
git clone https://github.com/ucatbas/permify-sdk-samples.git
cd permify-sdk-samples/go/grpc
```

### Running the Application

After successfully building the project, you can run the application using the following command:

```sh
go run create_tenant.go
```

## For your own Projects

To use the Permify SDK in your project, add the following dependency to your pom.xml file:

```sh
go get github.com/Permify/permify-go/v1
```

Here is a simple permify client:

```go
package main

import (
	"context"
	"log"

	v1 "github.com/Permify/permify-go/generated/base/v1"
	permify "github.com/Permify/permify-go/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	client, err := permify.NewClient(
		permify.Config{
			Endpoint: `localhost:3478`,
		},
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("failed to create Permify client: %v", err)
	}
}

  
```
