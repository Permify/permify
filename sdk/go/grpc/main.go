package main

import (
	"context"
	"log"

	"buf.build/gen/go/permifyco/permify/protocolbuffers/go/base/v1"
	pclient "github.com/Permify/permify-go/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	client, err := pclient.NewClient(
		pclient.Config{
			Endpoint: `localhost:3478`,
		},
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("failed to create Permify client: %v", err)
	}

	ct, err := client.Tenancy.Create(context.Background(), &basev1.TenantCreateRequest{
		Id:   "t1",
		Name: "tenant 1",
	})
	if err != nil {
		log.Fatalf("failed to create tenant: %v", err)
	}

	log.Printf("Tenant created: %v", ct)
}
