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

	serve := cmd.NewServeCommand()
	root.AddCommand(serve)

	validate := cmd.NewValidateCommand()
	root.AddCommand(validate)

	coverage := cmd.NewCoverageCommand()
	root.AddCommand(coverage)

	ast := cmd.NewGenerateASTCommand()
	root.AddCommand(ast)

	migrate := cmd.NewMigrateCommand()
	root.AddCommand(migrate)

	version := cmd.NewVersionCommand()
	root.AddCommand(version)

	config := cmd.NewConfigCommand()
	root.AddCommand(config)

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
