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

	ct, err := client.Tenancy.Create(context.Background(), &v1.TenantCreateRequest{
		Id:   "t1",
		Name: "tenant 1",
	})
	if err != nil {
		log.Fatalf("failed to create tenant: %v", err)
	}

	log.Printf("Tenant created: %v", ct)
}
