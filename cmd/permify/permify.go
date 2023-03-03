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

	migrate := cmd.NewMigrateCommand()
	root.AddCommand(migrate)

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
