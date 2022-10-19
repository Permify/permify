package main

import (
	`log`
	`os`

	`github.com/Permify/permify/internal/config`
	`github.com/Permify/permify/pkg/cmd`
)

func main() {
	// Configuration
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("config error: %s", err)
	}

	root := cmd.NewRootCommand()

	serve := cmd.NewServeCommand(cfg)
	cmd.RegisterServeFlags(serve, cfg)
	root.AddCommand(serve)

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
