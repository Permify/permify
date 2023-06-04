package main

import (
	"os"

	"github.com/Permify/permify/pkg/cmd"
	"github.com/Permify/permify/pkg/cmd/flags"
)

func main() {
	root := cmd.NewRootCommand()

	serve := cmd.NewServeCommand()
	flags.RegisterServeFlags(serve)
	root.AddCommand(serve)

	validate := cmd.NewValidateCommand()
	root.AddCommand(validate)

	coverage := cmd.NewCoverageCommand()
	flags.RegisterCoverageFlags(coverage)
	root.AddCommand(coverage)

	migrate := cmd.NewMigrateCommand()
	root.AddCommand(migrate)

	version := cmd.NewVersionCommand()
	root.AddCommand(version)

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
