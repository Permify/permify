package main

import (
	"os"

	"github.com/cespare/xxhash/v2"
	"github.com/sercand/kuberesolver/v5"
	"google.golang.org/grpc/balancer"

	consistentbalancer "github.com/Permify/permify/pkg/balancer"
	"github.com/Permify/permify/pkg/cmd"
)

func main() {
	kuberesolver.RegisterInCluster()
	balancer.Register(consistentbalancer.NewBuilder(xxhash.Sum64))

	root := cmd.NewRootCommand()
	root.AddCommand(cmd.NewServeCommand())
	root.AddCommand(cmd.NewValidateCommand())
	root.AddCommand(cmd.NewCoverageCommand())
	root.AddCommand(cmd.NewGenerateAstCommand())
	root.AddCommand(cmd.NewMigrateCommand())
	root.AddCommand(cmd.NewVersionCommand())
	root.AddCommand(cmd.NewConfigCommand())
	root.AddCommand(cmd.NewRepairCommand())
	
	// Add configure command
	root.AddCommand(cmd.NewConfigureCommand())

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
