package main

import (
	"os"

	"github.com/cespare/xxhash/v2"

	"github.com/sercand/kuberesolver/v5"
	"google.golang.org/grpc/balancer"

	consistentbalancer "github.com/Permify/permify/pkg/balancer"
	"github.com/Permify/permify/pkg/cmd"
)

func main() { // Application entry point
	kuberesolver.RegisterInCluster()                               // Register Kubernetes resolver
	balancer.Register(consistentbalancer.NewBuilder(xxhash.Sum64)) // Register consistent hash balancer
	// Setup CLI commands
	root := cmd.NewRootCommand() // Create root command
	// Add serve command
	serve := cmd.NewServeCommand() // Server command
	root.AddCommand(serve)         // Register serve command
	// Add validate command
	validate := cmd.NewValidateCommand() // Schema validation command
	root.AddCommand(validate)            // Register validate command
	// Add coverage command
	coverage := cmd.NewCoverageCommand() // Test coverage command
	root.AddCommand(coverage)            // Register coverage command
	// Add AST generation command
	ast := cmd.NewGenerateASTCommand() // AST generation command
	root.AddCommand(ast)               // Register AST command

	migrate := cmd.NewMigrateCommand() // Database migration command
	root.AddCommand(migrate)           // Register migrate command
	// Version command registration
	version := cmd.NewVersionCommand() // Version info command
	root.AddCommand(version)           // Register version command
	// Config command registration
	config := cmd.NewConfigCommand() // Configuration command
	root.AddCommand(config)          // Register config command
	// Repair command registration
	repair := cmd.NewRepairCommand() // Database repair command
	root.AddCommand(repair)          // Register repair command

	if err := root.Execute(); err != nil { // Run command
		os.Exit(1)
	}
}
