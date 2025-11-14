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

	// Setup CLI commands
	root := cmd.NewRootCommand()

	// Add serve command
	serve := cmd.NewServeCommand()
	root.AddCommand(serve)

	// Add validate command
	validate := cmd.NewValidateCommand()
	root.AddCommand(validate)

	// Add coverage command
	coverage := cmd.NewCoverageCommand()
	root.AddCommand(coverage)

	// Add AST generation command
	ast := cmd.NewGenerateAstCommand()
	root.AddCommand(ast)

	// Add migrate command
	migrate := cmd.NewMigrateCommand()
	root.AddCommand(migrate)

	// Add version command
	version := cmd.NewVersionCommand()
	root.AddCommand(version)

	// Add config command
	config := cmd.NewConfigCommand()
	root.AddCommand(config)

	// Add repair command
	repair := cmd.NewRepairCommand()
	root.AddCommand(repair)

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
