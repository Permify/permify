package main

import (
	"log"
	"os"

	"github.com/cespare/xxhash/v2"
	"github.com/sercand/kuberesolver/v5"
	"google.golang.org/grpc/balancer"

	consistentbalancer "github.com/Permify/permify/pkg/balancer"
	"github.com/Permify/permify/pkg/cmd"
)

func main() {
	// Register Kubernetes resolver
	if err := kuberesolver.RegisterInCluster(); err != nil {
		log.Printf("Failed to register kuberesolver: %v", err)
	}

	// Register gRPC balancer
	if err := balancer.Register(consistentbalancer.NewBuilder(xxhash.Sum64)); err != nil {
		log.Printf("Failed to register balancer: %v", err)
	}

	// Setup CLI commands
	root := cmd.NewRootCommand()

	// Add commands to root
	root.AddCommand(
		cmd.NewServeCommand(),
		cmd.NewValidateCommand(),
		cmd.NewCoverageCommand(),
		cmd.NewGenerateAstCommand(),
		cmd.NewMigrateCommand(),
		cmd.NewVersionCommand(),
		cmd.NewConfigCommand(),
		cmd.NewRepairCommand(),
	)

	// Execute the root command
	if err := root.Execute(); err != nil {
		_, _ = os.Stderr.WriteString("Error: " + err.Error() + "\n")
		os.Exit(1)
	}
}