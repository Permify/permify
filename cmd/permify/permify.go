package main

import (
	"os"

	"github.com/sercand/kuberesolver/v5"
	"google.golang.org/grpc/balancer"

	consistentbalancer "github.com/Permify/permify/pkg/balancer"
	"github.com/Permify/permify/pkg/cmd"
)

func main() {
	kuberesolver.RegisterInCluster()
	balancer.Register(consistentbalancer.NewConsistentHashBalancerBuilder())

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

	env := cmd.NewEnvCommand()
	root.AddCommand(env)

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
